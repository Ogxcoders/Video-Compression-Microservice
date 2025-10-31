package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	APIKey                  string
	AllowedDomains          []string
	Port                    string
	LogLevel                string
	MaxVideoFileSize        int64
	MaxImageFileSize        int64
	TempDir                 string
	RedisURL                string
	DatabaseURL             string
	MaxConcurrentJobs       int
	JobTimeout              int
	QueueCheckInterval      int
	FFmpegPath              string
	ImageMagickPath         string
	WordPressAPIURL         string
	WordPressUsername       string
	WordPressAppPassword    string
	RateLimitPerMinute      int
	RateLimitMaxConcurrent  int
	RateLimitMaxJobsPerDay  int
	MaxRetries              int
	RetryBackoffSeconds     []int
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		APIKey:                  getEnv("API_KEY", ""),
		AllowedDomains:          getEnvAsSlice("ALLOWED_DOMAINS", []string{}, ","),
		Port:                    getEnv("PORT", "3000"),
		LogLevel:                getEnv("LOG_LEVEL", "info"),
		MaxVideoFileSize:        getEnvAsInt64("MAX_VIDEO_FILE_SIZE", 5000000000),
		MaxImageFileSize:        getEnvAsInt64("MAX_IMAGE_FILE_SIZE", 500000000),
		TempDir:                 getEnv("TEMP_DIR", "/tmp/compression"),
		RedisURL:                getEnv("REDIS_URL", "redis://localhost:6379"),
		DatabaseURL:             getEnv("DATABASE_URL", ""),
		MaxConcurrentJobs:       getEnvAsInt("MAX_CONCURRENT_JOBS", 5),
		JobTimeout:              getEnvAsInt("JOB_TIMEOUT", 3600),
		QueueCheckInterval:      getEnvAsInt("QUEUE_CHECK_INTERVAL", 5),
		FFmpegPath:              getEnv("FFMPEG_PATH", "/usr/bin/ffmpeg"),
		ImageMagickPath:         getEnv("IMAGEMAGICK_PATH", "/usr/bin/convert"),
		WordPressAPIURL:         getEnv("WORDPRESS_API_URL", ""),
		WordPressUsername:       getEnv("WORDPRESS_USERNAME", ""),
		WordPressAppPassword:    getEnv("WORDPRESS_APP_PASSWORD", ""),
		RateLimitPerMinute:      getEnvAsInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 10),
		RateLimitMaxConcurrent:  getEnvAsInt("RATE_LIMIT_MAX_CONCURRENT", 100),
		RateLimitMaxJobsPerDay:  getEnvAsInt("RATE_LIMIT_MAX_JOBS_PER_DAY", 1000),
		MaxRetries:              getEnvAsInt("MAX_RETRIES", 3),
		RetryBackoffSeconds:     getEnvAsIntSlice("RETRY_BACKOFF_SECONDS", []int{60, 300, 900}, ","),
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsInt64(key string, defaultVal int64) int64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsSlice(key string, defaultVal []string, sep string) []string {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultVal
	}
	return strings.Split(valueStr, sep)
}

func getEnvAsIntSlice(key string, defaultVal []int, sep string) []int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultVal
	}
	
	parts := strings.Split(valueStr, sep)
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		if val, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
			result = append(result, val)
		}
	}
	
	if len(result) == 0 {
		return defaultVal
	}
	return result
}

func (c *Config) Validate() error {
	if c.APIKey == "" {
		log.Println("WARNING: API_KEY is not set")
	}
	if len(c.AllowedDomains) == 0 {
		log.Println("WARNING: ALLOWED_DOMAINS is not set")
	}
	if c.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	return nil
}
