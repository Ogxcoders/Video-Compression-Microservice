# How Video Compression API Works - Complete Guide

## 🎯 Overview

This is a microservice architecture for compressing videos and images, uploading them to WordPress, and returning optimized URLs.

---

## 🔄 Complete Flow Diagram

```
WordPress/Client → API → Redis Queue → Worker → FFmpeg/ImageMagick → WordPress Storage → Database
```

### Step-by-Step Flow:

1. **Client Submits Job** → POST `/api/compress`
2. **API Creates Job** → Saves to PostgreSQL
3. **Job Added to Queue** → Redis priority queue
4. **Worker Picks Up Job** → Background processing (every 5 seconds)
5. **Download Original File** → From source URL
6. **Compress File** → FFmpeg (video) or ImageMagick (image)
7. **Upload to WordPress** → Via WordPress REST API
8. **Save Results** → PostgreSQL database
9. **Client Checks Status** → GET `/api/status/:job_id`
10. **Client Gets Result** → GET `/api/result/:job_id`

---

## 📋 System Components

### 1. **API Layer** (Gin Framework - Go)
- **Endpoints**: `/api/compress`, `/api/status/:job_id`, `/api/result/:job_id`
- **Authentication**: API Key (X-API-Key header)
- **Rate Limiting**: 10 requests/minute per IP
- **Domain Whitelisting**: Only allowed domains can access

### 2. **Redis Queue**
- **Priority Queue**: Jobs with higher priority (1-10) processed first
- **Job Tracking**: Stores job IDs and queue positions
- **Status Caching**: Caches job status for quick lookups

### 3. **Worker** (Background Processor)
- **Concurrent Jobs**: Processes 5 jobs simultaneously (configurable)
- **Queue Polling**: Checks Redis every 5 seconds
- **Retry Logic**: Retries failed jobs 3 times with exponential backoff (60s, 300s, 900s)
- **Timeout**: Jobs timeout after 3600 seconds (1 hour)

### 4. **Compressors**

#### Video Compressor (FFmpeg)
- **Qualities**:
  - `low`: 480p, 500 kbps
  - `medium`: 720p, 1500 kbps
  - `high`: 1080p, 3000 kbps
  - `ultra`: 4K, 6000 kbps
  - `hls-adaptive`: HLS streaming with multiple variants

- **HLS Variants** (Adaptive Streaming):
  - 240p (400 kbps)
  - 360p (800 kbps)
  - 480p (1200 kbps)
  - 720p (2500 kbps)
  - 1080p (5000 kbps)

#### Image Compressor (ImageMagick)
- **Qualities**:
  - `low`: 60% quality
  - `medium`: 75% quality
  - `high`: 85% quality
  - `ultra`: 95% quality

- **Variants**: Generates multiple sizes (thumbnail, medium, large, full)

### 5. **WordPress Storage**
- **Upload Method**: WordPress REST API (`/wp-json/wp/v2/media`)
- **Authentication**: App Password (not regular password!)
- **Returns**: Media IDs and URLs

### 6. **PostgreSQL Database**
- **Jobs Table**: Stores all compression jobs
- **Queue Stats**: Daily statistics
- **Indexes**: Optimized for job_id, status, post_id, scheduled_time

---

## 🔧 How to Use the API

### Authentication

All `/api/*` endpoints require an API key:

```bash
X-API-Key: sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1
```

---

### 1. Submit a Compression Job

#### Video Compression

```bash
curl -X POST https://api.trendss.net/api/compress \
  -H "X-API-Key: sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1" \
  -H "Content-Type: application/json" \
  -d '{
    "post_id": 123,
    "compression_type": "video",
    "video_data": {
      "file_url": "https://example.com/video.mp4",
      "quality": "medium",
      "hls_enabled": false
    },
    "priority": 5
  }'
```

**Response:**
```json
{
  "status": "queued",
  "job_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "compression_type": "video",
  "queue_position": 3,
  "estimated_time": 180
}
```

#### Image Compression

```bash
curl -X POST https://api.trendss.net/api/compress \
  -H "X-API-Key: sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1" \
  -H "Content-Type: application/json" \
  -d '{
    "post_id": 456,
    "compression_type": "image",
    "image_data": {
      "file_url": "https://example.com/image.jpg",
      "quality": "high",
      "variants": ["thumbnail", "medium", "large"]
    },
    "priority": 5
  }'
```

#### Both Video + Image

```bash
curl -X POST https://api.trendss.net/api/compress \
  -H "X-API-Key: sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1" \
  -H "Content-Type: application/json" \
  -d '{
    "post_id": 789,
    "compression_type": "both",
    "video_data": {
      "file_url": "https://example.com/video.mp4",
      "quality": "high",
      "hls_enabled": true,
      "hls_variants": ["240p", "360p", "480p", "720p", "1080p"]
    },
    "image_data": {
      "file_url": "https://example.com/thumbnail.jpg",
      "quality": "medium"
    },
    "priority": 8
  }'
```

---

### 2. Check Job Status

```bash
curl https://api.trendss.net/api/status/a1b2c3d4-e5f6-7890-abcd-ef1234567890 \
  -H "X-API-Key: sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1"
```

**Response:**
```json
{
  "job_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "compression_type": "video",
  "overall_status": "processing",
  "overall_progress": 50,
  "video_status": "processing",
  "video_progress": 50,
  "estimated_time": 120
}
```

**Status Values:**
- `pending`: Waiting in queue
- `processing`: Currently being compressed
- `completed`: Successfully finished
- `failed`: Error occurred (check error_message)
- `cancelled`: Cancelled by user

---

### 3. Get Compression Results

```bash
curl https://api.trendss.net/api/result/a1b2c3d4-e5f6-7890-abcd-ef1234567890 \
  -H "X-API-Key: sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1"
```

**Response:**
```json
{
  "job_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "compression_type": "video",
  "overall_status": "completed",
  "video_result": {
    "status": "completed",
    "original_size": 50000000,
    "compressed_size": 15000000,
    "compression_ratio": 0.3,
    "processing_time": 180,
    "compressed_url": "https://capcut.ogtemplate.com/wp-content/uploads/2025/11/compressed-video.mp4",
    "hls_playlist_url": "https://capcut.ogtemplate.com/wp-content/uploads/2025/11/playlist.m3u8",
    "hls_variants": {
      "240p": "https://capcut.ogtemplate.com/uploads/variants/240p.m3u8",
      "720p": "https://capcut.ogtemplate.com/uploads/variants/720p.m3u8"
    }
  }
}
```

---

## 🔌 WordPress Plugin Integration

### Installation

1. Create file: `wp-content/plugins/video-compressor/video-compressor.php`
2. Activate plugin in WordPress admin

### Plugin Code

```php
<?php
/**
 * Plugin Name: Video Compressor API
 * Description: Automatically compress videos uploaded to WordPress
 * Version: 1.0.0
 * Author: Your Name
 */

defined('ABSPATH') || exit;

class VideoCompressorAPI {
    private $api_url = 'https://api.trendss.net/api';
    private $api_key = 'sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1';
    
    public function __construct() {
        // Auto-compress when video is uploaded
        add_action('add_attachment', [$this, 'auto_compress_video']);
        
        // Admin menu
        add_action('admin_menu', [$this, 'add_admin_menu']);
        
        // AJAX handlers
        add_action('wp_ajax_check_compression_status', [$this, 'ajax_check_status']);
    }
    
    /**
     * Automatically compress video when uploaded
     */
    public function auto_compress_video($attachment_id) {
        $mime_type = get_post_mime_type($attachment_id);
        
        // Only process videos
        if (strpos($mime_type, 'video/') !== 0) {
            return;
        }
        
        $file_url = wp_get_attachment_url($attachment_id);
        
        $response = wp_remote_post($this->api_url . '/compress', [
            'headers' => [
                'X-API-Key' => $this->api_key,
                'Content-Type' => 'application/json',
            ],
            'body' => json_encode([
                'post_id' => $attachment_id,
                'compression_type' => 'video',
                'video_data' => [
                    'file_url' => $file_url,
                    'quality' => 'medium',
                    'hls_enabled' => false
                ],
                'priority' => 5
            ]),
            'timeout' => 30
        ]);
        
        if (is_wp_error($response)) {
            error_log('Compression API error: ' . $response->get_error_message());
            return;
        }
        
        $body = json_decode(wp_remote_retrieve_body($response), true);
        
        if (isset($body['job_id'])) {
            // Save job ID in post meta
            update_post_meta($attachment_id, '_compression_job_id', $body['job_id']);
            update_post_meta($attachment_id, '_compression_status', 'queued');
        }
    }
    
    /**
     * Check compression status
     */
    public function get_status($job_id) {
        $response = wp_remote_get($this->api_url . '/status/' . $job_id, [
            'headers' => ['X-API-Key' => $this->api_key]
        ]);
        
        if (is_wp_error($response)) {
            return false;
        }
        
        return json_decode(wp_remote_retrieve_body($response), true);
    }
    
    /**
     * Get compression result
     */
    public function get_result($job_id) {
        $response = wp_remote_get($this->api_url . '/result/' . $job_id, [
            'headers' => ['X-API-Key' => $this->api_key]
        ]);
        
        if (is_wp_error($response)) {
            return false;
        }
        
        $result = json_decode(wp_remote_retrieve_body($response), true);
        
        // If completed, save compressed URL
        if ($result['overall_status'] === 'completed' && isset($result['video_result']['compressed_url'])) {
            return $result['video_result']['compressed_url'];
        }
        
        return false;
    }
    
    /**
     * Add admin menu
     */
    public function add_admin_menu() {
        add_menu_page(
            'Video Compressor',
            'Compressor',
            'manage_options',
            'video-compressor',
            [$this, 'admin_page'],
            'dashicons-video-alt3'
        );
    }
    
    /**
     * Admin page
     */
    public function admin_page() {
        ?>
        <div class="wrap">
            <h1>Video Compression Status</h1>
            <table class="wp-list-table widefat fixed striped">
                <thead>
                    <tr>
                        <th>Video</th>
                        <th>Status</th>
                        <th>Progress</th>
                        <th>Compressed URL</th>
                    </tr>
                </thead>
                <tbody>
                    <?php
                    $args = [
                        'post_type' => 'attachment',
                        'post_mime_type' => 'video',
                        'posts_per_page' => 50,
                        'meta_query' => [
                            [
                                'key' => '_compression_job_id',
                                'compare' => 'EXISTS'
                            ]
                        ]
                    ];
                    
                    $videos = get_posts($args);
                    
                    foreach ($videos as $video) {
                        $job_id = get_post_meta($video->ID, '_compression_job_id', true);
                        $status = $this->get_status($job_id);
                        
                        echo '<tr>';
                        echo '<td>' . esc_html($video->post_title) . '</td>';
                        echo '<td>' . ($status ? esc_html($status['overall_status']) : 'Unknown') . '</td>';
                        echo '<td>' . ($status ? intval($status['overall_progress']) . '%' : '-') . '</td>';
                        
                        if ($status && $status['overall_status'] === 'completed') {
                            $result = $this->get_result($job_id);
                            echo '<td><a href="' . esc_url($result) . '" target="_blank">View</a></td>';
                        } else {
                            echo '<td>-</td>';
                        }
                        echo '</tr>';
                    }
                    ?>
                </tbody>
            </table>
        </div>
        <?php
    }
    
    /**
     * AJAX: Check status
     */
    public function ajax_check_status() {
        $job_id = $_POST['job_id'] ?? '';
        
        if (!$job_id) {
            wp_send_json_error('No job ID provided');
        }
        
        $status = $this->get_status($job_id);
        wp_send_json_success($status);
    }
}

// Initialize plugin
new VideoCompressorAPI();
```

---

## 🧪 Testing if Compression Works

### Test 1: Health Check
```bash
curl https://api.trendss.net/health
```
Should return: `{"database":"connected","redis":"connected","status":"healthy"}`

### Test 2: Submit Test Job
```bash
curl -X POST https://api.trendss.net/api/compress \
  -H "X-API-Key: sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1" \
  -H "Content-Type: application/json" \
  -d '{
    "post_id": 1,
    "compression_type": "video",
    "video_data": {
      "file_url": "https://ogtemplate.com/wp-content/uploads/2025/10/7556584672545852733-video.mp4",
      "quality": "medium"
    }
  }'
```

Save the `job_id` from the response!

### Test 3: Check Status
```bash
# Replace JOB_ID with actual job_id from step 2
curl https://api.trendss.net/api/status/JOB_ID \
  -H "X-API-Key: sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1"
```

### Test 4: Get Result (when completed)
```bash
curl https://api.trendss.net/api/result/JOB_ID \
  -H "X-API-Key: sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1"
```

---

## 🎬 What Happens During Compression

### Video Compression Process:

1. **Download** original video from URL
2. **Analyze** video (resolution, bitrate, codec, duration)
3. **Compress** using FFmpeg with H.264 codec
4. **Generate HLS** variants (if enabled)
5. **Upload** compressed files to WordPress
6. **Save** URLs and metadata to database
7. **Clean up** temporary files

### Time Estimates:

- **Small video** (< 50MB): 1-2 minutes
- **Medium video** (50-500MB): 3-10 minutes
- **Large video** (> 500MB): 10-30 minutes
- **Image**: 5-30 seconds

---

## 📊 Priority System

Jobs with **higher priority** (1-10) are processed first:

- **Priority 10**: Urgent/premium users
- **Priority 5**: Normal users (default)
- **Priority 1**: Low priority/batch jobs

---

## 🔒 Security Features

1. **API Key Authentication**: Required for all API endpoints
2. **Domain Whitelisting**: Only allowed domains can access
3. **Rate Limiting**: 10 requests/minute per IP
4. **CORS Protection**: Configured allowed origins
5. **Input Validation**: All requests validated
6. **SQL Injection Protection**: Parameterized queries
7. **File Type Validation**: Only allowed file types processed

---

## 🚀 Production Checklist

- [ ] Set strong API key (not the test key!)
- [ ] Configure allowed domains
- [ ] Set WordPress credentials
- [ ] Configure SSL/HTTPS
- [ ] Set up monitoring
- [ ] Configure backups
- [ ] Test compression with real files
- [ ] Monitor queue depth
- [ ] Set up alerts for failed jobs

---

Your compression API is now ready to use! 🎉
