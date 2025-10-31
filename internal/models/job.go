package models

import (
	"encoding/json"
	"time"
)

type CompressionType string

const (
	CompressionTypeVideo CompressionType = "video"
	CompressionTypeImage CompressionType = "image"
	CompressionTypeBoth  CompressionType = "both"
)

type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusScheduled  JobStatus = "scheduled"
	JobStatusCancelled  JobStatus = "cancelled"
)

type VideoQuality string

const (
	VideoQualityLow         VideoQuality = "low"
	VideoQualityMedium      VideoQuality = "medium"
	VideoQualityHigh        VideoQuality = "high"
	VideoQualityUltra       VideoQuality = "ultra"
	VideoQualityHLSAdaptive VideoQuality = "hls-adaptive"
)

type ImageQuality string

const (
	ImageQualityLow    ImageQuality = "low"
	ImageQualityMedium ImageQuality = "medium"
	ImageQualityHigh   ImageQuality = "high"
	ImageQualityUltra  ImageQuality = "ultra"
)

type VideoData struct {
	FileURL     string       `json:"file_url" binding:"required"`
	Quality     VideoQuality `json:"quality" binding:"required"`
	HLSEnabled  bool         `json:"hls_enabled"`
	HLSVariants []string     `json:"hls_variants"`
}

type ImageData struct {
	FileURL  string       `json:"file_url" binding:"required"`
	Quality  ImageQuality `json:"quality" binding:"required"`
	Variants []string     `json:"variants"`
}

type Job struct {
	ID              int              `json:"id"`
	JobID           string           `json:"job_id"`
	PostID          int              `json:"post_id"`
	UserID          *int             `json:"user_id"`
	CompressionType CompressionType  `json:"compression_type"`
	VideoData       *VideoData       `json:"video_data,omitempty"`
	ImageData       *ImageData       `json:"image_data,omitempty"`
	Priority        int              `json:"priority"`
	Status          JobStatus        `json:"status"`
	VideoStatus     *JobStatus       `json:"video_status,omitempty"`
	ImageStatus     *JobStatus       `json:"image_status,omitempty"`
	VideoResult     *VideoResult     `json:"video_result,omitempty"`
	ImageResult     *ImageResult     `json:"image_result,omitempty"`
	ErrorMessage    string           `json:"error_message,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	StartedAt       *time.Time       `json:"started_at,omitempty"`
	CompletedAt     *time.Time       `json:"completed_at,omitempty"`
	ScheduledTime   *time.Time       `json:"scheduled_time,omitempty"`
	RetryCount      int              `json:"retry_count"`
	MaxRetries      int              `json:"max_retries"`
	ProcessingTime  *int             `json:"processing_time,omitempty"`
}

type VideoResult struct {
	Status            string            `json:"status"`
	OriginalSize      int64             `json:"original_size"`
	CompressedSize    int64             `json:"compressed_size"`
	CompressionRatio  float64           `json:"compression_ratio"`
	ProcessingTime    int               `json:"processing_time"`
	CompressedURL     string            `json:"compressed_url,omitempty"`
	HLSPlaylistURL    string            `json:"hls_playlist_url,omitempty"`
	HLSVariants       map[string]string `json:"hls_variants,omitempty"`
}

type ImageResult struct {
	Status           string                    `json:"status"`
	OriginalSize     int64                     `json:"original_size"`
	CompressedSize   int64                     `json:"compressed_size"`
	CompressionRatio float64                   `json:"compression_ratio"`
	ProcessingTime   int                       `json:"processing_time"`
	Variants         map[string]ImageVariant   `json:"variants"`
}

type ImageVariant struct {
	URL        string `json:"url"`
	Size       int64  `json:"size"`
	Dimensions string `json:"dimensions"`
}

type CompressRequest struct {
	JobID           string           `json:"job_id"`
	PostID          int              `json:"post_id" binding:"required"`
	UserID          *int             `json:"user_id"`
	CompressionType CompressionType  `json:"compression_type" binding:"required"`
	VideoData       *VideoData       `json:"video_data,omitempty"`
	ImageData       *ImageData       `json:"image_data,omitempty"`
	Priority        int              `json:"priority"`
	ScheduledTime   *time.Time       `json:"scheduled_time,omitempty"`
}

type CompressResponse struct {
	Status          string          `json:"status"`
	JobID           string          `json:"job_id"`
	CompressionType CompressionType `json:"compression_type"`
	QueuePosition   int             `json:"queue_position"`
	EstimatedTime   int             `json:"estimated_time"`
}

type StatusResponse struct {
	JobID              string          `json:"job_id"`
	CompressionType    CompressionType `json:"compression_type"`
	OverallStatus      JobStatus       `json:"overall_status"`
	OverallProgress    int             `json:"overall_progress"`
	VideoStatus        *JobStatus      `json:"video_status,omitempty"`
	VideoProgress      *int            `json:"video_progress,omitempty"`
	VideoCurrentStep   string          `json:"video_current_step,omitempty"`
	ImageStatus        *JobStatus      `json:"image_status,omitempty"`
	ImageProgress      *int            `json:"image_progress,omitempty"`
	EstimatedTime      int             `json:"estimated_time"`
}

type ResultResponse struct {
	JobID           string          `json:"job_id"`
	CompressionType CompressionType `json:"compression_type"`
	OverallStatus   JobStatus       `json:"overall_status"`
	VideoResult     *VideoResult    `json:"video_result,omitempty"`
	ImageResult     *ImageResult    `json:"image_result,omitempty"`
	ErrorMessage    string          `json:"error_message,omitempty"`
}

type QueueStats struct {
	TotalJobs          int     `json:"total_jobs"`
	PendingJobs        int     `json:"pending_jobs"`
	ProcessingJobs     int     `json:"processing_jobs"`
	CompletedJobs      int     `json:"completed_jobs"`
	FailedJobs         int     `json:"failed_jobs"`
	AvgProcessingTime  float64 `json:"avg_processing_time"`
	QueueDepth         int     `json:"queue_depth"`
	VideoJobs          int     `json:"video_jobs"`
	ImageJobs          int     `json:"image_jobs"`
	CombinedJobs       int     `json:"combined_jobs"`
}

func (v *VideoData) MarshalJSON() ([]byte, error) {
	type Alias VideoData
	return json.Marshal(&struct{ *Alias }{Alias: (*Alias)(v)})
}

func (i *ImageData) MarshalJSON() ([]byte, error) {
	type Alias ImageData
	return json.Marshal(&struct{ *Alias }{Alias: (*Alias)(i)})
}
