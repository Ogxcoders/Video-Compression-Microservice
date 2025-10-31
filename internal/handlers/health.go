package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/video-compressor/internal/database"
	"github.com/yourusername/video-compressor/internal/queue"
)

type HealthHandler struct {
	db    *database.Database
	queue *queue.RedisQueue
}

func NewHealthHandler(db *database.Database, q *queue.RedisQueue) *HealthHandler {
	return &HealthHandler{
		db:    db,
		queue: q,
	}
}

func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "video-compressor-api",
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	queueLength, err := h.queue.GetQueueLength()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  "queue unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "ready",
		"queue_length": queueLength,
	})
}
