# Video Compression Microservice

## Project Overview

A production-ready Go microservice for video and image compression with job queue management, built for deployment on VPS via Coolify/Docker.

**Created:** October 31, 2025  
**Language:** Go 1.21+  
**Deployment:** Docker Compose (Coolify-ready)  
**Database:** PostgreSQL + Redis

## Project Purpose

This microservice provides:
- Video compression with FFmpeg (4 quality presets: low/medium/high/ultra)
- Image compression with ImageMagick (4 variants: thumbnail/medium/large/original)
- Combined video+image processing with parallel execution
- Job queue system with Redis and PostgreSQL persistence
- WordPress REST API integration for file operations
- Retry logic with exponential backoff
- API security (API key + domain whitelist + rate limiting)

## Architecture

### Core Components

1. **API Layer** (`cmd/api/main.go`)
   - Gin web framework
   - RESTful endpoints
   - Middleware for auth, CORS, rate limiting

2. **Database Layer** (`internal/database/`)
   - PostgreSQL for job persistence
   - CRUD operations for jobs
   - Queue statistics tracking

3. **Queue System** (`internal/queue/`)
   - Redis-backed job queue
   - FIFO with priority support
   - Job status caching

4. **Worker System** (`internal/worker/`)
   - Background job processor
   - Concurrent job execution (configurable MAX_CONCURRENT_JOBS)
   - Retry logic with exponential backoff
   - Parallel video+image processing

5. **Compression Engines**
   - **Video** (`internal/compressor/video.go`): FFmpeg-based compression & HLS generation
   - **Image** (`internal/compressor/image.go`): ImageMagick-based compression & variants

6. **Storage Integration** (`internal/storage/`)
   - WordPress REST API file download/upload
   - File size validation

## Key Features

### Implemented (MVP)
- [x] Complete REST API with 7 endpoints
- [x] Video compression (low/medium/high/ultra presets)
- [x] Image compression with 4 responsive variants
- [x] Combined video+image processing
- [x] Job queue with Redis + PostgreSQL
- [x] Worker pool with configurable concurrency
- [x] Retry mechanism (3 attempts with backoff)
- [x] WordPress integration (download/upload)
- [x] API key authentication
- [x] Domain whitelist security
- [x] Rate limiting (10 req/min configurable)
- [x] Docker Compose setup
- [x] Nginx reverse proxy with SSL
- [x] Health & readiness endpoints
- [x] Queue statistics API

### Next Phase
- [ ] HLS streaming (adaptive bitrate with .m3u8 playlists)
- [ ] Scheduled compression (cron-like scheduler)
- [ ] Real-time WebSocket status updates
- [ ] Webhook callbacks on completion
- [ ] Multiple worker instances (horizontal scaling)
- [ ] S3/object storage integration
- [ ] Advanced queue monitoring dashboard

## Project Structure

```
.
├── cmd/api/                    # Application entry point
│   └── main.go                # Main server & worker initialization
├── internal/
│   ├── compressor/            # Compression logic
│   │   ├── video.go          # FFmpeg video compression
│   │   └── image.go          # ImageMagick image compression
│   ├── database/             # Database layer
│   │   └── database.go       # PostgreSQL operations
│   ├── handlers/             # API handlers
│   │   ├── compress.go       # Compression endpoints
│   │   └── health.go         # Health checks
│   ├── middleware/           # HTTP middleware
│   │   ├── auth.go          # API key + domain whitelist
│   │   └── ratelimit.go     # Rate limiting
│   ├── models/              # Data structures
│   │   └── job.go           # Job, request, response models
│   ├── queue/               # Queue management
│   │   └── redis.go         # Redis queue operations
│   ├── storage/             # File operations
│   │   └── wordpress.go     # WordPress REST API integration
│   └── worker/              # Background worker
│       └── worker.go        # Job processor with retry logic
├── pkg/config/              # Configuration
│   └── config.go           # Environment config loader
├── scripts/                # Database scripts
│   └── init.sql           # PostgreSQL schema
├── nginx/                  # Nginx configuration
│   └── nginx.conf         # Reverse proxy config
├── docker-compose.yml      # Docker orchestration
├── Dockerfile             # Go app container
├── .env.example          # Environment template
├── README.md            # User documentation
├── DEPLOYMENT.md        # Deployment guide
└── API_DOCUMENTATION.md # API reference
```

## API Endpoints

1. **POST /api/compress** - Enqueue compression job
2. **GET /api/status/:job_id** - Get job status
3. **GET /api/result/:job_id** - Get compression results
4. **GET /api/queue/stats** - Queue statistics
5. **POST /api/queue/cancel/:job_id** - Cancel pending job
6. **GET /health** - Health check
7. **GET /ready** - Readiness check

## Environment Configuration

### Required Variables
- `API_KEY` - Secure API key for authentication
- `ALLOWED_DOMAINS` - Comma-separated allowed domains
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string
- `WORDPRESS_API_URL` - WordPress REST API endpoint
- `WORDPRESS_USERNAME` - WordPress admin username
- `WORDPRESS_APP_PASSWORD` - WordPress application password

### Optional Variables
- `MAX_CONCURRENT_JOBS` - Max parallel jobs (default: 5)
- `JOB_TIMEOUT` - Job timeout in seconds (default: 3600)
- `QUEUE_CHECK_INTERVAL` - Worker check interval (default: 5s)
- `MAX_RETRIES` - Max retry attempts (default: 3)

## Deployment

### Local Docker Development
```bash
cp .env.example .env
# Edit .env with your settings
docker-compose up -d
```

### Production (Coolify/VPS)
See `DEPLOYMENT.md` for complete guide.

## Recent Changes

**October 31, 2025**
- Initial project creation
- Complete MVP implementation
- Docker Compose setup for Coolify deployment
- Comprehensive documentation (README, API docs, deployment guide)
- PostgreSQL database schema with job tracking
- Redis queue with priority support
- Worker system with retry logic
- Video compression (FFmpeg, 4 quality presets)
- Image compression (ImageMagick, 4 variants)
- WordPress REST API integration
- Security middleware (API key, domain whitelist, rate limiting)

## Technical Notes

### Why No Docker in Replit?
This project is designed for deployment on external VPS via Coolify, not for running in Replit. Docker is not supported in Replit's environment.

### Database Schema
- `jobs` table tracks all compression jobs
- `queue_stats` table stores daily statistics
- Auto-updating timestamps via PostgreSQL triggers

### Queue Processing
- Worker checks queue every 5 seconds (configurable)
- Respects MAX_CONCURRENT_JOBS limit
- Failed jobs retry with exponential backoff: 60s → 300s → 900s

### File Handling
- Downloads to `/tmp/compression/{job_id}/`
- Cleans up after processing
- Validates file sizes against limits

## Development Notes

This is a production-ready microservice meant for VPS deployment. For local development or testing without Docker:

1. Install Go 1.21+
2. Install PostgreSQL and Redis
3. Install FFmpeg and ImageMagick
4. Run: `go run cmd/api/main.go`

The project uses Go modules for dependency management. All dependencies are listed in `go.mod`.
