package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/yourusername/video-compressor/internal/compressor"
	"github.com/yourusername/video-compressor/internal/database"
	"github.com/yourusername/video-compressor/internal/models"
	"github.com/yourusername/video-compressor/internal/queue"
	"github.com/yourusername/video-compressor/internal/storage"
	"github.com/yourusername/video-compressor/pkg/config"
)

type Worker struct {
	config           *Config
	db               *database.Database
	queue            *queue.RedisQueue
	videoCompressor  *compressor.VideoCompressor
	imageCompressor  *compressor.ImageCompressor
	storage          *storage.WordPressStorage
	activeJobs       sync.Map
	maxConcurrentJobs int
	ctx              context.Context
	cancel           context.CancelFunc
}

type Config struct {
	MaxConcurrentJobs int
	JobTimeout        time.Duration
	CheckInterval     time.Duration
	TempDir           string
	MaxRetries        int
	RetryBackoff      []int
}

func NewWorker(
	cfg *config.Config,
	db *database.Database,
	q *queue.RedisQueue,
	videoComp *compressor.VideoCompressor,
	imageComp *compressor.ImageCompressor,
	wpStorage *storage.WordPressStorage,
) *Worker {
	ctx, cancel := context.WithCancel(context.Background())

	return &Worker{
		config: &Config{
			MaxConcurrentJobs: cfg.MaxConcurrentJobs,
			JobTimeout:        time.Duration(cfg.JobTimeout) * time.Second,
			CheckInterval:     time.Duration(cfg.QueueCheckInterval) * time.Second,
			TempDir:           cfg.TempDir,
			MaxRetries:        cfg.MaxRetries,
			RetryBackoff:      cfg.RetryBackoffSeconds,
		},
		db:                db,
		queue:             q,
		videoCompressor:   videoComp,
		imageCompressor:   imageComp,
		storage:           wpStorage,
		maxConcurrentJobs: cfg.MaxConcurrentJobs,
		ctx:               ctx,
		cancel:            cancel,
	}
}

func (w *Worker) Start() {
	log.Println("Worker started, checking queue every", w.config.CheckInterval)

	ticker := time.NewTicker(w.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			log.Println("Worker stopped")
			return
		case <-ticker.C:
			w.processQueue()
		}
	}
}

func (w *Worker) Stop() {
	log.Println("Stopping worker...")
	w.cancel()
}

func (w *Worker) processQueue() {
	activeCount := 0
	w.activeJobs.Range(func(_, _ interface{}) bool {
		activeCount++
		return true
	})

	if activeCount >= w.maxConcurrentJobs {
		return
	}

	availableSlots := w.maxConcurrentJobs - activeCount

	for i := 0; i < availableSlots; i++ {
		jobID, err := w.queue.Dequeue()
		if err != nil {
			log.Printf("Failed to dequeue job: %v", err)
			continue
		}

		if jobID == "" {
			break
		}

		job, err := w.db.GetJobByID(jobID)
		if err != nil {
			log.Printf("Failed to get job %s: %v", jobID, err)
			w.queue.MarkComplete(jobID)
			continue
		}

		w.activeJobs.Store(jobID, true)
		go w.processJob(job)
	}
}

func (w *Worker) processJob(job *models.Job) {
	defer func() {
		w.activeJobs.Delete(job.JobID)
		w.queue.MarkComplete(job.JobID)
	}()

	log.Printf("Processing job %s (type: %s)", job.JobID, job.CompressionType)

	ctx, cancel := context.WithTimeout(w.ctx, w.config.JobTimeout)
	defer cancel()

	if err := w.db.MarkJobStarted(job.JobID); err != nil {
		log.Printf("Failed to mark job %s as started: %v", job.JobID, err)
	}

	startTime := time.Now()

	var wg sync.WaitGroup
	var videoErr, imageErr error

	switch job.CompressionType {
	case models.CompressionTypeVideo:
		videoErr = w.processVideo(ctx, job)

	case models.CompressionTypeImage:
		imageErr = w.processImage(ctx, job)

	case models.CompressionTypeBoth:
		wg.Add(2)
		go func() {
			defer wg.Done()
			videoErr = w.processVideo(ctx, job)
		}()
		go func() {
			defer wg.Done()
			imageErr = w.processImage(ctx, job)
		}()
		wg.Wait()
	}

	processingTime := int(time.Since(startTime).Seconds())

	if videoErr != nil || imageErr != nil {
		errorMsg := ""
		if videoErr != nil {
			errorMsg += fmt.Sprintf("Video: %v. ", videoErr)
		}
		if imageErr != nil {
			errorMsg += fmt.Sprintf("Image: %v", imageErr)
		}

		if job.RetryCount < w.config.MaxRetries {
			log.Printf("Job %s failed (attempt %d/%d): %s", job.JobID, job.RetryCount+1, w.config.MaxRetries, errorMsg)
			w.db.IncrementRetryCount(job.JobID)
			
			backoffIndex := job.RetryCount
			if backoffIndex >= len(w.config.RetryBackoff) {
				backoffIndex = len(w.config.RetryBackoff) - 1
			}
			backoffSeconds := w.config.RetryBackoff[backoffIndex]

			time.AfterFunc(time.Duration(backoffSeconds)*time.Second, func() {
				w.queue.Enqueue(job.JobID, job.Priority)
			})
		} else {
			log.Printf("Job %s failed permanently: %s", job.JobID, errorMsg)
			w.db.MarkJobFailed(job.JobID, errorMsg)
		}
		return
	}

	w.db.MarkJobCompleted(job.JobID, processingTime)
	log.Printf("Job %s completed in %d seconds", job.JobID, processingTime)
}

func (w *Worker) processVideo(ctx context.Context, job *models.Job) error {
	if job.VideoData == nil {
		return nil
	}

	w.db.UpdateVideoStatus(job.JobID, models.JobStatusProcessing)

	jobDir := filepath.Join(w.config.TempDir, job.JobID)
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		return fmt.Errorf("failed to create job directory: %w", err)
	}
	defer os.RemoveAll(jobDir)

	inputPath := filepath.Join(jobDir, "input_video"+filepath.Ext(job.VideoData.FileURL))
	log.Printf("Downloading video from %s", job.VideoData.FileURL)
	if err := w.storage.DownloadFile(job.VideoData.FileURL, inputPath); err != nil {
		return fmt.Errorf("failed to download video: %w", err)
	}

	originalSize, err := w.videoCompressor.GetVideoInfo(inputPath)
	if err != nil {
		return fmt.Errorf("failed to get video info: %w", err)
	}

	startTime := time.Now()
	result := &models.VideoResult{
		Status:       "completed",
		OriginalSize: originalSize,
	}

	if job.VideoData.HLSEnabled && len(job.VideoData.HLSVariants) > 0 {
		log.Printf("Generating HLS variants for job %s", job.JobID)
		masterPlaylist, variantURLs, err := w.videoCompressor.GenerateHLS(inputPath, job.VideoData.HLSVariants)
		if err != nil {
			return fmt.Errorf("failed to generate HLS: %w", err)
		}

		hlsURL, err := w.storage.UploadFile(masterPlaylist)
		if err != nil {
			return fmt.Errorf("failed to upload HLS master playlist: %w", err)
		}

		result.HLSPlaylistURL = hlsURL
		result.HLSVariants = variantURLs
	} else {
		log.Printf("Compressing video with quality %s for job %s", job.VideoData.Quality, job.JobID)
		compressedPath, err := w.videoCompressor.Compress(inputPath, job.VideoData.Quality)
		if err != nil {
			return fmt.Errorf("failed to compress video: %w", err)
		}

		compressedSize, _ := w.videoCompressor.GetVideoInfo(compressedPath)
		result.CompressedSize = compressedSize
		result.CompressionRatio = float64(originalSize-compressedSize) / float64(originalSize)

		compressedURL, err := w.storage.UploadFile(compressedPath)
		if err != nil {
			return fmt.Errorf("failed to upload compressed video: %w", err)
		}

		result.CompressedURL = compressedURL
	}

	result.ProcessingTime = int(time.Since(startTime).Seconds())

	w.db.UpdateVideoResult(job.JobID, result)
	w.db.UpdateVideoStatus(job.JobID, models.JobStatusCompleted)

	log.Printf("Video processing completed for job %s", job.JobID)
	return nil
}

func (w *Worker) processImage(ctx context.Context, job *models.Job) error {
	if job.ImageData == nil {
		return nil
	}

	w.db.UpdateImageStatus(job.JobID, models.JobStatusProcessing)

	jobDir := filepath.Join(w.config.TempDir, job.JobID)
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		return fmt.Errorf("failed to create job directory: %w", err)
	}
	defer os.RemoveAll(jobDir)

	inputPath := filepath.Join(jobDir, "input_image"+filepath.Ext(job.ImageData.FileURL))
	log.Printf("Downloading image from %s", job.ImageData.FileURL)
	if err := w.storage.DownloadFile(job.ImageData.FileURL, inputPath); err != nil {
		return fmt.Errorf("failed to download image: %w", err)
	}

	originalSize, _, err := w.imageCompressor.GetImageInfo(inputPath)
	if err != nil {
		return fmt.Errorf("failed to get image info: %w", err)
	}

	startTime := time.Now()

	variants := job.ImageData.Variants
	if len(variants) == 0 {
		variants = []string{"thumbnail", "medium", "large", "original"}
	}

	log.Printf("Generating image variants for job %s: %v", job.JobID, variants)
	variantPaths, err := w.imageCompressor.CompressWithVariants(inputPath, job.ImageData.Quality, variants)
	if err != nil {
		return fmt.Errorf("failed to compress image: %w", err)
	}

	result := &models.ImageResult{
		Status:       "completed",
		OriginalSize: originalSize,
		Variants:     make(map[string]models.ImageVariant),
	}

	var totalCompressedSize int64
	for variantName, variantPath := range variantPaths {
		size, dimensions, _ := w.imageCompressor.GetImageInfo(variantPath)

		url, err := w.storage.UploadFile(variantPath)
		if err != nil {
			log.Printf("Failed to upload %s variant: %v", variantName, err)
			continue
		}

		result.Variants[variantName] = models.ImageVariant{
			URL:        url,
			Size:       size,
			Dimensions: dimensions,
		}

		totalCompressedSize += size
	}

	result.CompressedSize = totalCompressedSize
	if originalSize > 0 {
		result.CompressionRatio = float64(originalSize-totalCompressedSize) / float64(originalSize)
	}
	result.ProcessingTime = int(time.Since(startTime).Seconds())

	w.db.UpdateImageResult(job.JobID, result)
	w.db.UpdateImageStatus(job.JobID, models.JobStatusCompleted)

	log.Printf("Image processing completed for job %s", job.JobID)
	return nil
}
