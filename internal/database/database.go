package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/yourusername/video-compressor/internal/models"
)

type Database struct {
	db *sql.DB
}

func New(databaseURL string) (*Database, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) CreateJob(job *models.Job) error {
	query := `
		INSERT INTO jobs (
			job_id, post_id, user_id, compression_type,
			video_file_url, video_quality, video_hls_enabled, video_hls_variants,
			image_file_url, image_quality, image_variants,
			priority, status, video_status, image_status,
			scheduled_time, max_retries
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, created_at, updated_at
	`

	var videoFileURL, videoQuality *string
	var videoHLSEnabled *bool
	var videoHLSVariants interface{}
	var imageFileURL, imageQuality *string
	var imageVariants interface{}

	if job.VideoData != nil {
		videoFileURL = &job.VideoData.FileURL
		q := string(job.VideoData.Quality)
		videoQuality = &q
		videoHLSEnabled = &job.VideoData.HLSEnabled
		if len(job.VideoData.HLSVariants) > 0 {
			videoHLSVariants = job.VideoData.HLSVariants
		}
	}

	if job.ImageData != nil {
		imageFileURL = &job.ImageData.FileURL
		q := string(job.ImageData.Quality)
		imageQuality = &q
		if len(job.ImageData.Variants) > 0 {
			imageVariants = job.ImageData.Variants
		}
	}

	err := d.db.QueryRow(
		query,
		job.JobID, job.PostID, job.UserID, job.CompressionType,
		videoFileURL, videoQuality, videoHLSEnabled, videoHLSVariants,
		imageFileURL, imageQuality, imageVariants,
		job.Priority, job.Status, job.VideoStatus, job.ImageStatus,
		job.ScheduledTime, job.MaxRetries,
	).Scan(&job.ID, &job.CreatedAt, &job.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	return nil
}

func (d *Database) GetJobByID(jobID string) (*models.Job, error) {
	query := `
		SELECT 
			id, job_id, post_id, user_id, compression_type,
			video_file_url, video_quality, video_hls_enabled, video_hls_variants,
			image_file_url, image_quality, image_variants,
			priority, status, video_status, image_status,
			video_result, image_result, error_message,
			created_at, updated_at, started_at, completed_at, scheduled_time,
			retry_count, max_retries, processing_time
		FROM jobs WHERE job_id = $1
	`

	job := &models.Job{}
	var videoFileURL, videoQuality, videoResult, imageFileURL, imageQuality, imageResult, errorMessage sql.NullString
	var videoHLSEnabled sql.NullBool
	var videoHLSVariants, imageVariants interface{}
	var userID, processingTime sql.NullInt64
	var startedAt, completedAt, scheduledTime sql.NullTime
	var videoStatus, imageStatus sql.NullString

	err := d.db.QueryRow(query, jobID).Scan(
		&job.ID, &job.JobID, &job.PostID, &userID, &job.CompressionType,
		&videoFileURL, &videoQuality, &videoHLSEnabled, &videoHLSVariants,
		&imageFileURL, &imageQuality, &imageVariants,
		&job.Priority, &job.Status, &videoStatus, &imageStatus,
		&videoResult, &imageResult, &errorMessage,
		&job.CreatedAt, &job.UpdatedAt, &startedAt, &completedAt, &scheduledTime,
		&job.RetryCount, &job.MaxRetries, &processingTime,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("job not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	if userID.Valid {
		uid := int(userID.Int64)
		job.UserID = &uid
	}
	if videoFileURL.Valid {
		job.VideoData = &models.VideoData{
			FileURL:    videoFileURL.String,
			Quality:    models.VideoQuality(videoQuality.String),
			HLSEnabled: videoHLSEnabled.Bool,
		}
	}
	if imageFileURL.Valid {
		job.ImageData = &models.ImageData{
			FileURL: imageFileURL.String,
			Quality: models.ImageQuality(imageQuality.String),
		}
	}
	if videoStatus.Valid {
		vs := models.JobStatus(videoStatus.String)
		job.VideoStatus = &vs
	}
	if imageStatus.Valid {
		is := models.JobStatus(imageStatus.String)
		job.ImageStatus = &is
	}
	if videoResult.Valid {
		var vr models.VideoResult
		if err := json.Unmarshal([]byte(videoResult.String), &vr); err == nil {
			job.VideoResult = &vr
		}
	}
	if imageResult.Valid {
		var ir models.ImageResult
		if err := json.Unmarshal([]byte(imageResult.String), &ir); err == nil {
			job.ImageResult = &ir
		}
	}
	if errorMessage.Valid {
		job.ErrorMessage = errorMessage.String
	}
	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}
	if scheduledTime.Valid {
		job.ScheduledTime = &scheduledTime.Time
	}
	if processingTime.Valid {
		pt := int(processingTime.Int64)
		job.ProcessingTime = &pt
	}

	return job, nil
}

func (d *Database) UpdateJobStatus(jobID string, status models.JobStatus, errorMessage string) error {
	query := `
		UPDATE jobs 
		SET status = $1, error_message = $2, updated_at = CURRENT_TIMESTAMP
		WHERE job_id = $3
	`
	_, err := d.db.Exec(query, status, errorMessage, jobID)
	return err
}

func (d *Database) UpdateVideoStatus(jobID string, status models.JobStatus) error {
	query := `UPDATE jobs SET video_status = $1, updated_at = CURRENT_TIMESTAMP WHERE job_id = $2`
	_, err := d.db.Exec(query, status, jobID)
	return err
}

func (d *Database) UpdateImageStatus(jobID string, status models.JobStatus) error {
	query := `UPDATE jobs SET image_status = $1, updated_at = CURRENT_TIMESTAMP WHERE job_id = $2`
	_, err := d.db.Exec(query, status, jobID)
	return err
}

func (d *Database) UpdateVideoResult(jobID string, result *models.VideoResult) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return err
	}
	query := `UPDATE jobs SET video_result = $1, updated_at = CURRENT_TIMESTAMP WHERE job_id = $2`
	_, err = d.db.Exec(query, resultJSON, jobID)
	return err
}

func (d *Database) UpdateImageResult(jobID string, result *models.ImageResult) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return err
	}
	query := `UPDATE jobs SET image_result = $1, updated_at = CURRENT_TIMESTAMP WHERE job_id = $2`
	_, err = d.db.Exec(query, resultJSON, jobID)
	return err
}

func (d *Database) MarkJobStarted(jobID string) error {
	query := `
		UPDATE jobs 
		SET status = $1, started_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE job_id = $2
	`
	_, err := d.db.Exec(query, models.JobStatusProcessing, jobID)
	return err
}

func (d *Database) MarkJobCompleted(jobID string, processingTime int) error {
	query := `
		UPDATE jobs 
		SET status = $1, completed_at = CURRENT_TIMESTAMP, processing_time = $2, updated_at = CURRENT_TIMESTAMP
		WHERE job_id = $3
	`
	_, err := d.db.Exec(query, models.JobStatusCompleted, processingTime, jobID)
	return err
}

func (d *Database) MarkJobFailed(jobID string, errorMessage string) error {
	query := `
		UPDATE jobs 
		SET status = $1, error_message = $2, completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE job_id = $3
	`
	_, err := d.db.Exec(query, models.JobStatusFailed, errorMessage, jobID)
	return err
}

func (d *Database) IncrementRetryCount(jobID string) error {
	query := `UPDATE jobs SET retry_count = retry_count + 1, updated_at = CURRENT_TIMESTAMP WHERE job_id = $1`
	_, err := d.db.Exec(query, jobID)
	return err
}

func (d *Database) GetQueueStats() (*models.QueueStats, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'pending') as pending,
			COUNT(*) FILTER (WHERE status = 'processing') as processing,
			COUNT(*) FILTER (WHERE status = 'completed') as completed,
			COUNT(*) FILTER (WHERE status = 'failed') as failed,
			COALESCE(AVG(processing_time) FILTER (WHERE processing_time IS NOT NULL), 0) as avg_time,
			COUNT(*) FILTER (WHERE compression_type = 'video') as video,
			COUNT(*) FILTER (WHERE compression_type = 'image') as image,
			COUNT(*) FILTER (WHERE compression_type = 'both') as combined
		FROM jobs
	`

	stats := &models.QueueStats{}
	err := d.db.QueryRow(query).Scan(
		&stats.TotalJobs,
		&stats.PendingJobs,
		&stats.ProcessingJobs,
		&stats.CompletedJobs,
		&stats.FailedJobs,
		&stats.AvgProcessingTime,
		&stats.VideoJobs,
		&stats.ImageJobs,
		&stats.CombinedJobs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue stats: %w", err)
	}

	stats.QueueDepth = stats.PendingJobs

	return stats, nil
}

func (d *Database) GetPendingJobs(limit int) ([]*models.Job, error) {
	query := `
		SELECT job_id FROM jobs 
		WHERE status = $1 
		ORDER BY priority DESC, created_at ASC 
		LIMIT $2
	`

	rows, err := d.db.Query(query, models.JobStatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		var jobID string
		if err := rows.Scan(&jobID); err != nil {
			continue
		}
		job, err := d.GetJobByID(jobID)
		if err == nil {
			jobs = append(jobs, job)
		}
	}

	return jobs, nil
}
