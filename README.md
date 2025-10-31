# Video Compression Microservice

A production-ready Go microservice for video and image compression with job queue management, HLS streaming support, and WordPress integration.

## Features

- **Video Compression**: FFmpeg-based compression with multiple quality presets (low/medium/high/ultra)
- **Image Compression**: ImageMagick-based compression with responsive variants (thumbnail/medium/large/original)
- **Combined Processing**: Process video and image in parallel
- **Job Queue System**: Redis-backed queue with PostgreSQL persistence
- **HLS Streaming**: Generate adaptive bitrate HLS streams with multiple quality variants
- **WordPress Integration**: Seamless file download/upload via WordPress REST API
- **Retry Logic**: Automatic retry with exponential backoff (3 attempts)
- **Security**: API key authentication, domain whitelist, rate limiting
- **Docker Ready**: Complete Docker Compose setup for easy deployment

## Architecture

```
├── cmd/api/              # Application entry point
├── internal/
│   ├── compressor/       # Video & image compression logic
│   ├── database/         # PostgreSQL database layer
│   ├── handlers/         # API endpoint handlers
│   ├── middleware/       # Authentication & rate limiting
│   ├── models/           # Data structures
│   ├── queue/            # Redis queue management
│   ├── storage/          # WordPress file operations
│   └── worker/           # Job processing worker
├── pkg/config/           # Configuration management
├── scripts/              # Database initialization
├── nginx/                # Nginx reverse proxy config
└── docker-compose.yml    # Docker orchestration
```

## Prerequisites

- Docker & Docker Compose
- Go 1.21+ (for local development)
- FFmpeg
- ImageMagick

## Quick Start

### 1. Clone and Configure

```bash
git clone <your-repo>
cd video-compressor

# Copy environment template
cp .env.example .env

# Edit .env with your configuration
nano .env
```

### 2. Configure Environment Variables

```env
API_KEY=your-secure-api-key-here
ALLOWED_DOMAINS=https://wp.yourdomain.com
WORDPRESS_API_URL=https://wp.yourdomain.com/wp-json/wp/v2
WORDPRESS_USERNAME=admin
WORDPRESS_APP_PASSWORD=your-app-password
```

### 3. Deploy with Docker Compose

```bash
# Build and start all services
docker-compose up -d

# Check logs
docker-compose logs -f app

# Check status
docker-compose ps
```

### 4. Deploy to Coolify

1. **Create a new project** in Coolify
2. **Add Docker Compose** deployment
3. **Upload** your project files
4. **Set environment variables** in Coolify UI
5. **Deploy** and access via your domain

## API Endpoints

### Compress Video/Image

```bash
POST /api/compress
X-API-Key: your-api-key

{
  "post_id": 12345,
  "compression_type": "both",
  "video_data": {
    "file_url": "https://wp.yourdomain.com/uploads/video.mp4",
    "quality": "medium",
    "hls_enabled": false
  },
  "image_data": {
    "file_url": "https://wp.yourdomain.com/uploads/image.jpg",
    "quality": "high",
    "variants": ["thumbnail", "medium", "large", "original"]
  },
  "priority": 5
}
```

**Response:**
```json
{
  "status": "queued",
  "job_id": "uuid-v4",
  "compression_type": "both",
  "queue_position": 3,
  "estimated_time": 180
}
```

### Get Job Status

```bash
GET /api/status/:job_id
X-API-Key: your-api-key
```

**Response:**
```json
{
  "job_id": "uuid-v4",
  "compression_type": "both",
  "overall_status": "processing",
  "overall_progress": 55,
  "video_status": "processing",
  "video_progress": 45,
  "image_status": "completed",
  "image_progress": 100,
  "estimated_time": 300
}
```

### Get Job Result

```bash
GET /api/result/:job_id
X-API-Key: your-api-key
```

**Response:**
```json
{
  "job_id": "uuid-v4",
  "compression_type": "both",
  "overall_status": "completed",
  "video_result": {
    "status": "completed",
    "original_size": 1000000000,
    "compressed_size": 250000000,
    "compression_ratio": 0.75,
    "processing_time": 300,
    "compressed_url": "https://wp.yourdomain.com/uploads/video-compressed.mp4"
  },
  "image_result": {
    "status": "completed",
    "original_size": 5000000,
    "compressed_size": 1500000,
    "compression_ratio": 0.70,
    "processing_time": 15,
    "variants": {
      "thumbnail": {
        "url": "https://wp.yourdomain.com/uploads/image-thumbnail.jpg",
        "size": 12000,
        "dimensions": "150x150"
      }
    }
  }
}
```

### Get Queue Statistics

```bash
GET /api/queue/stats
X-API-Key: your-api-key
```

### Cancel Job

```bash
POST /api/queue/cancel/:job_id
X-API-Key: your-api-key
```

## Compression Types

### Video Compression

- **compression_type**: `"video"`
- **Quality Presets**:
  - `low`: 480p @ 1000kbps
  - `medium`: 720p @ 2500kbps
  - `high`: 1080p @ 5000kbps
  - `ultra`: Original resolution @ 8000kbps

### Image Compression

- **compression_type**: `"image"`
- **Quality Presets**:
  - `low`: 60% quality
  - `medium`: 75% quality
  - `high`: 85% quality
  - `ultra`: 95% quality
- **Variants**:
  - `thumbnail`: 150x150px
  - `medium`: 400x300px
  - `large`: 800x600px
  - `original`: Original size

### Combined Compression

- **compression_type**: `"both"`
- Process video and image in parallel

## HLS Streaming (Future Phase)

Enable adaptive bitrate streaming:

```json
{
  "video_data": {
    "hls_enabled": true,
    "hls_variants": ["480p", "720p", "1080p"]
  }
}
```

## WordPress Integration

### Setup WordPress

1. **Install Application Passwords** plugin (or use WordPress 5.6+)
2. **Generate App Password** for API user
3. **Enable REST API**
4. **Set permissions** for media uploads

### Configuration

```env
WORDPRESS_API_URL=https://wp.yourdomain.com/wp-json/wp/v2
WORDPRESS_USERNAME=admin
WORDPRESS_APP_PASSWORD=xxxx xxxx xxxx xxxx xxxx xxxx
```

## Security

- **API Key Authentication**: Required via `X-API-Key` header
- **Domain Whitelist**: Only allowed domains can access API
- **Rate Limiting**: 10 requests/minute per IP
- **Input Validation**: File size and format validation
- **CORS**: Configured for allowed domains only

## Monitoring

### Health Check

```bash
curl http://localhost:3000/health
```

### Queue Metrics

```bash
curl http://localhost:3000/api/queue/stats \
  -H "X-API-Key: your-api-key"
```

### Docker Logs

```bash
# All services
docker-compose logs -f

# Just the app
docker-compose logs -f app

# Redis
docker-compose logs -f redis

# PostgreSQL
docker-compose logs -f db
```

## Development

### Local Development

```bash
# Install dependencies
go mod download

# Run locally
go run cmd/api/main.go

# Build binary
make build

# Run tests
make test
```

### Database Migrations

The database schema is automatically initialized on first startup via `scripts/init.sql`.

## Troubleshooting

### Jobs stuck in queue

```bash
# Check worker logs
docker-compose logs -f app

# Restart worker
docker-compose restart app
```

### FFmpeg errors

```bash
# Check FFmpeg is installed
docker exec compressor-api ffmpeg -version

# Check temp directory permissions
docker exec compressor-api ls -la /tmp/compression
```

### Database connection issues

```bash
# Check PostgreSQL status
docker-compose ps db

# Connect to database
docker exec -it compressor-db psql -U compressor -d compression
```

## Performance Tuning

### Concurrent Jobs

Increase parallel processing:

```env
MAX_CONCURRENT_JOBS=10
```

### Queue Check Interval

Faster queue processing:

```env
QUEUE_CHECK_INTERVAL=3
```

### Job Timeout

For large files:

```env
JOB_TIMEOUT=7200
```

## Production Deployment

### SSL Configuration

1. Place SSL certificates in `nginx/ssl/`
2. Update `nginx/nginx.conf` with your domain
3. Restart Nginx: `docker-compose restart nginx`

### Backup

```bash
# Backup PostgreSQL
docker exec compressor-db pg_dump -U compressor compression > backup.sql

# Backup Redis
docker exec compressor-redis redis-cli SAVE
```

## License

MIT License

## Support

For issues and questions, please create an issue in the repository.
