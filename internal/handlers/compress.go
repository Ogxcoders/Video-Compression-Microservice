package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/video-compressor/internal/database"
	"github.com/yourusername/video-compressor/internal/models"
	"github.com/yourusername/video-compressor/internal/queue"
	"github.com/yourusername/video-compressor/pkg/config"
)

type CompressHandler struct {
	db     *database.Database
	queue  *queue.RedisQueue
	config *config.Config
}

func NewCompressHandler(db *database.Database, q *queue.RedisQueue, cfg *config.Config) *CompressHandler {
	return &CompressHandler{
		db:     db,
		queue:  q,
		config: cfg,
	}
}

func (h *CompressHandler) Compress(c *gin.Context) {
	var req models.CompressRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	if req.JobID == "" {
		req.JobID = uuid.New().String()
	}

	if err := h.validateRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	job := &models.Job{
		JobID:           req.JobID,
		PostID:          req.PostID,
		UserID:          req.UserID,
		CompressionType: req.CompressionType,
		VideoData:       req.VideoData,
		ImageData:       req.ImageData,
		Priority:        req.Priority,
		Status:          models.JobStatusPending,
		ScheduledTime:   req.ScheduledTime,
		MaxRetries:      h.config.MaxRetries,
	}

	if job.Priority == 0 {
		job.Priority = 5
	}

	if req.CompressionType == models.CompressionTypeVideo || req.CompressionType == models.CompressionTypeBoth {
		status := models.JobStatusPending
		job.VideoStatus = &status
	}

	if req.CompressionType == models.CompressionTypeImage || req.CompressionType == models.CompressionTypeBoth {
		status := models.JobStatusPending
		job.ImageStatus = &status
	}

	if err := h.db.CreateJob(job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create job",
			"details": err.Error(),
		})
		return
	}

	if err := h.queue.Enqueue(job.JobID, job.Priority); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to enqueue job",
			"details": err.Error(),
		})
		return
	}

	queueLength, _ := h.queue.GetQueueLength()

	c.JSON(http.StatusOK, models.CompressResponse{
		Status:          "queued",
		JobID:           job.JobID,
		CompressionType: job.CompressionType,
		QueuePosition:   int(queueLength),
		EstimatedTime:   int(queueLength) * 60,
	})
}

func (h *CompressHandler) validateRequest(req *models.CompressRequest) error {
	switch req.CompressionType {
	case models.CompressionTypeVideo:
		if req.VideoData == nil {
			return ErrVideoDataRequired
		}
	case models.CompressionTypeImage:
		if req.ImageData == nil {
			return ErrImageDataRequired
		}
	case models.CompressionTypeBoth:
		if req.VideoData == nil || req.ImageData == nil {
			return ErrBothDataRequired
		}
	default:
		return ErrInvalidCompressionType
	}

	return nil
}

func (h *CompressHandler) GetStatus(c *gin.Context) {
	jobID := c.Param("job_id")

	cached, _ := h.queue.GetCachedJobStatus(jobID)
	if cached != nil {
		c.JSON(http.StatusOK, cached)
		return
	}

	job, err := h.db.GetJobByID(jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
		})
		return
	}

	response := &models.StatusResponse{
		JobID:           job.JobID,
		CompressionType: job.CompressionType,
		OverallStatus:   job.Status,
		OverallProgress: h.calculateProgress(job),
		EstimatedTime:   h.estimateTime(job),
	}

	if job.VideoStatus != nil {
		response.VideoStatus = job.VideoStatus
		progress := h.calculateVideoProgress(job)
		response.VideoProgress = &progress
	}

	if job.ImageStatus != nil {
		response.ImageStatus = job.ImageStatus
		progress := h.calculateImageProgress(job)
		response.ImageProgress = &progress
	}

	c.JSON(http.StatusOK, response)
}

func (h *CompressHandler) GetResult(c *gin.Context) {
	jobID := c.Param("job_id")

	job, err := h.db.GetJobByID(jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
		})
		return
	}

	response := &models.ResultResponse{
		JobID:           job.JobID,
		CompressionType: job.CompressionType,
		OverallStatus:   job.Status,
		VideoResult:     job.VideoResult,
		ImageResult:     job.ImageResult,
		ErrorMessage:    job.ErrorMessage,
	}

	c.JSON(http.StatusOK, response)
}

func (h *CompressHandler) GetQueueStats(c *gin.Context) {
	stats, err := h.db.GetQueueStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get queue stats",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *CompressHandler) CancelJob(c *gin.Context) {
	jobID := c.Param("job_id")

	job, err := h.db.GetJobByID(jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
		})
		return
	}

	if job.Status == models.JobStatusProcessing {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot cancel job that is currently processing",
		})
		return
	}

	if job.Status == models.JobStatusCompleted || job.Status == models.JobStatusFailed {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Job already finished",
		})
		return
	}

	if err := h.queue.RemoveJob(jobID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to cancel job",
		})
		return
	}

	if err := h.db.UpdateJobStatus(jobID, models.JobStatusCancelled, "Cancelled by user"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update job status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "cancelled",
		"job_id": jobID,
	})
}

func (h *CompressHandler) calculateProgress(job *models.Job) int {
	if job.Status == models.JobStatusCompleted {
		return 100
	}
	if job.Status == models.JobStatusPending {
		return 0
	}

	progress := 0
	count := 0

	if job.VideoStatus != nil {
		progress += h.calculateVideoProgress(job)
		count++
	}

	if job.ImageStatus != nil {
		progress += h.calculateImageProgress(job)
		count++
	}

	if count == 0 {
		return 50
	}

	return progress / count
}

func (h *CompressHandler) calculateVideoProgress(job *models.Job) int {
	if job.VideoStatus == nil {
		return 0
	}

	switch *job.VideoStatus {
	case models.JobStatusCompleted:
		return 100
	case models.JobStatusProcessing:
		return 50
	case models.JobStatusPending:
		return 0
	default:
		return 0
	}
}

func (h *CompressHandler) calculateImageProgress(job *models.Job) int {
	if job.ImageStatus == nil {
		return 0
	}

	switch *job.ImageStatus {
	case models.JobStatusCompleted:
		return 100
	case models.JobStatusProcessing:
		return 50
	case models.JobStatusPending:
		return 0
	default:
		return 0
	}
}

func (h *CompressHandler) estimateTime(job *models.Job) int {
	if job.Status == models.JobStatusCompleted || job.Status == models.JobStatusFailed {
		return 0
	}

	estimatedTime := 0

	if job.VideoStatus != nil && *job.VideoStatus != models.JobStatusCompleted {
		estimatedTime += 300
	}

	if job.ImageStatus != nil && *job.ImageStatus != models.JobStatusCompleted {
		estimatedTime += 30
	}

	return estimatedTime
}

var (
	ErrVideoDataRequired       = &ValidationError{"video_data is required for video compression"}
	ErrImageDataRequired       = &ValidationError{"image_data is required for image compression"}
	ErrBothDataRequired        = &ValidationError{"both video_data and image_data are required"}
	ErrInvalidCompressionType  = &ValidationError{"compression_type must be 'video', 'image', or 'both'"}
)

type ValidationError struct {
	message string
}

func (e *ValidationError) Error() string {
	return e.message
}
