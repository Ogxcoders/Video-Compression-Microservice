# API Documentation

## Base URL

```
https://compress.yourdomain.com/api
```

## Authentication

All API requests require an API key sent via the `X-API-Key` header.

```bash
X-API-Key: your-api-key-here
```

## Endpoints

### 1. Compress Media

Enqueue a compression job for video, image, or both.

**Endpoint:** `POST /api/compress`

**Headers:**
```
X-API-Key: your-api-key
Content-Type: application/json
```

**Request Body:**

```json
{
  "job_id": "optional-custom-uuid",
  "post_id": 12345,
  "user_id": 1,
  "compression_type": "both",
  "video_data": {
    "file_url": "https://wp.yourdomain.com/uploads/video.mp4",
    "quality": "medium",
    "hls_enabled": false,
    "hls_variants": ["480p", "720p", "1080p"]
  },
  "image_data": {
    "file_url": "https://wp.yourdomain.com/uploads/poster.jpg",
    "quality": "high",
    "variants": ["thumbnail", "medium", "large", "original"]
  },
  "priority": 5,
  "scheduled_time": "2025-01-15T14:00:00Z"
}
```

**Parameters:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `job_id` | string | No | Custom job ID (auto-generated if not provided) |
| `post_id` | integer | Yes | WordPress post ID |
| `user_id` | integer | No | WordPress user ID |
| `compression_type` | string | Yes | `"video"`, `"image"`, or `"both"` |
| `video_data` | object | Conditional | Required if compression_type is "video" or "both" |
| `image_data` | object | Conditional | Required if compression_type is "image" or "both" |
| `priority` | integer | No | Priority (1-10, default: 5) |
| `scheduled_time` | string | No | ISO 8601 timestamp for scheduled compression |

**Video Data:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `file_url` | string | Yes | Full URL to video file |
| `quality` | string | Yes | `"low"`, `"medium"`, `"high"`, `"ultra"` |
| `hls_enabled` | boolean | No | Enable HLS streaming (default: false) |
| `hls_variants` | array | No | HLS quality variants: `["480p", "720p", "1080p"]` |

**Image Data:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `file_url` | string | Yes | Full URL to image file |
| `quality` | string | Yes | `"low"`, `"medium"`, `"high"`, `"ultra"` |
| `variants` | array | No | Image sizes: `["thumbnail", "medium", "large", "original"]` |

**Response:**

```json
{
  "status": "queued",
  "job_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "compression_type": "both",
  "queue_position": 3,
  "estimated_time": 180
}
```

**Status Codes:**
- `200 OK` - Job created successfully
- `400 Bad Request` - Invalid request
- `401 Unauthorized` - Invalid API key
- `403 Forbidden` - Domain not allowed
- `413 Payload Too Large` - File too large
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

---

### 2. Get Job Status

Check the current status of a compression job.

**Endpoint:** `GET /api/status/:job_id`

**Headers:**
```
X-API-Key: your-api-key
```

**Response:**

```json
{
  "job_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "compression_type": "both",
  "overall_status": "processing",
  "overall_progress": 55,
  "video_status": "processing",
  "video_progress": 45,
  "video_current_step": "encoding_720p",
  "image_status": "completed",
  "image_progress": 100,
  "estimated_time": 300
}
```

**Status Values:**
- `pending` - Waiting in queue
- `processing` - Currently being compressed
- `completed` - Successfully compressed
- `failed` - Compression failed
- `cancelled` - Job was cancelled

**Status Codes:**
- `200 OK` - Status retrieved
- `404 Not Found` - Job not found

---

### 3. Get Job Result

Retrieve the result of a completed compression job.

**Endpoint:** `GET /api/result/:job_id`

**Headers:**
```
X-API-Key: your-api-key
```

**Response:**

```json
{
  "job_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "compression_type": "both",
  "overall_status": "completed",
  "video_result": {
    "status": "completed",
    "original_size": 1000000000,
    "compressed_size": 250000000,
    "compression_ratio": 0.75,
    "processing_time": 300,
    "compressed_url": "https://wp.yourdomain.com/uploads/video-compressed.mp4",
    "hls_playlist_url": null,
    "hls_variants": null
  },
  "image_result": {
    "status": "completed",
    "original_size": 5000000,
    "compressed_size": 1500000,
    "compression_ratio": 0.70,
    "processing_time": 15,
    "variants": {
      "thumbnail": {
        "url": "https://wp.yourdomain.com/uploads/poster-thumbnail.jpg",
        "size": 12000,
        "dimensions": "150x150"
      },
      "medium": {
        "url": "https://wp.yourdomain.com/uploads/poster-medium.jpg",
        "size": 45000,
        "dimensions": "400x300"
      },
      "large": {
        "url": "https://wp.yourdomain.com/uploads/poster-large.jpg",
        "size": 120000,
        "dimensions": "800x600"
      },
      "original": {
        "url": "https://wp.yourdomain.com/uploads/poster-original.jpg",
        "size": 4500000,
        "dimensions": "original"
      }
    }
  },
  "error_message": null
}
```

---

### 4. Get Queue Statistics

Retrieve overall queue statistics.

**Endpoint:** `GET /api/queue/stats`

**Headers:**
```
X-API-Key: your-api-key
```

**Response:**

```json
{
  "total_jobs": 1523,
  "pending_jobs": 12,
  "processing_jobs": 3,
  "completed_jobs": 1480,
  "failed_jobs": 28,
  "avg_processing_time": 245.7,
  "queue_depth": 12,
  "video_jobs": 890,
  "image_jobs": 320,
  "combined_jobs": 313
}
```

---

### 5. Cancel Job

Cancel a pending job (cannot cancel processing jobs).

**Endpoint:** `POST /api/queue/cancel/:job_id`

**Headers:**
```
X-API-Key: your-api-key
```

**Response:**

```json
{
  "status": "cancelled",
  "job_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

**Status Codes:**
- `200 OK` - Job cancelled
- `400 Bad Request` - Job cannot be cancelled
- `404 Not Found` - Job not found

---

### 6. Health Check

Check API health status.

**Endpoint:** `GET /health`

**Response:**

```json
{
  "status": "healthy",
  "service": "video-compressor-api"
}
```

---

### 7. Readiness Check

Check if API is ready to accept requests.

**Endpoint:** `GET /ready`

**Response:**

```json
{
  "status": "ready",
  "queue_length": 5
}
```

---

## Quality Presets

### Video Quality

| Preset | Resolution | Bitrate | Use Case |
|--------|-----------|---------|----------|
| `low` | 480p (854x480) | 1000 kbps | Mobile, low bandwidth |
| `medium` | 720p (1280x720) | 2500 kbps | Standard web playback |
| `high` | 1080p (1920x1080) | 5000 kbps | HD playback |
| `ultra` | Original | 8000 kbps | Archive quality |

### Image Quality

| Preset | Quality | Compression | Use Case |
|--------|---------|-------------|----------|
| `low` | 60% | 70-80% reduction | Thumbnails |
| `medium` | 75% | 50-65% reduction | Web display |
| `high` | 85% | 30-45% reduction | High quality |
| `ultra` | 95% | 10-20% reduction | Archive |

### Image Variants

| Variant | Dimensions | Description |
|---------|-----------|-------------|
| `thumbnail` | 150x150px | Gallery thumbnails |
| `medium` | 400x300px | Blog posts |
| `large` | 800x600px | Full-width display |
| `original` | Original size | Archive/download |

---

## Error Responses

```json
{
  "error": "Error message here",
  "details": "Additional error details"
}
```

**Common Error Codes:**

| Code | Description |
|------|-------------|
| 400 | Invalid request format or parameters |
| 401 | Missing or invalid API key |
| 403 | Domain not in whitelist |
| 404 | Resource not found |
| 413 | File size exceeds limit |
| 429 | Rate limit exceeded |
| 500 | Internal server error |
| 503 | Service unavailable (queue full) |
| 504 | Request timeout |
| 507 | Insufficient storage |

---

## Rate Limits

- **Requests per minute**: 10 (configurable)
- **Concurrent compressions**: 100 (configurable)
- **Jobs per day**: 1000 (configurable)

When rate limit is exceeded:

```json
{
  "error": "Rate limit exceeded",
  "retry_after": 45.2
}
```

---

## Examples

### cURL Examples

**Compress Video Only:**
```bash
curl -X POST https://compress.yourdomain.com/api/compress \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "post_id": 123,
    "compression_type": "video",
    "video_data": {
      "file_url": "https://wp.example.com/video.mp4",
      "quality": "medium"
    }
  }'
```

**Compress Image Only:**
```bash
curl -X POST https://compress.yourdomain.com/api/compress \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "post_id": 123,
    "compression_type": "image",
    "image_data": {
      "file_url": "https://wp.example.com/image.jpg",
      "quality": "high",
      "variants": ["thumbnail", "medium", "large"]
    }
  }'
```

**Check Status:**
```bash
curl https://compress.yourdomain.com/api/status/job-id-here \
  -H "X-API-Key: your-api-key"
```

---

## Webhooks (Future Phase)

Coming soon: Receive notifications when jobs complete.

```json
{
  "event": "job.completed",
  "job_id": "uuid",
  "status": "completed",
  "result": { }
}
```
