# Deployment Guide

## Deployment to Coolify (VPS)

### Step 1: Prepare Your VPS

1. **Install Docker & Docker Compose** on your VPS:
```bash
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
```

2. **Install Coolify** (if not already installed):
```bash
curl -fsSL https://cdn.coollabs.io/coolify/install.sh | bash
```

### Step 2: Clone Repository to VPS

```bash
ssh user@your-vps-ip
git clone <your-repo-url> /opt/video-compressor
cd /opt/video-compressor
```

### Step 3: Configure Environment

```bash
cp .env.example .env
nano .env
```

**Required Variables:**
```env
API_KEY=generate-strong-key-here
ALLOWED_DOMAINS=https://wp.yourdomain.com,https://wordpress.yourdomain.com
WORDPRESS_API_URL=https://wp.yourdomain.com/wp-json/wp/v2
WORDPRESS_USERNAME=admin
WORDPRESS_APP_PASSWORD=your-wordpress-app-password
DATABASE_URL=postgres://compressor:CHANGE_THIS_PASSWORD@db:5432/compression?sslmode=disable
```

**Update docker-compose.yml PostgreSQL Password:**
```yaml
db:
  environment:
    - POSTGRES_PASSWORD=CHANGE_THIS_PASSWORD
```

### Step 4: SSL Certificates

Create SSL certificates directory:
```bash
mkdir -p nginx/ssl
```

**Option A: Use Let's Encrypt (Recommended)**
```bash
sudo apt install certbot
sudo certbot certonly --standalone -d compress.yourdomain.com
sudo cp /etc/letsencrypt/live/compress.yourdomain.com/fullchain.pem nginx/ssl/cert.pem
sudo cp /etc/letsencrypt/live/compress.yourdomain.com/privkey.pem nginx/ssl/key.pem
```

**Option B: Self-Signed Certificate (Development)**
```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout nginx/ssl/key.pem \
  -out nginx/ssl/cert.pem
```

### Step 5: Update Nginx Configuration

Edit `nginx/nginx.conf` and replace `compress.yourdomain.com` with your actual domain.

### Step 6: Deploy

```bash
# Build and start
docker-compose up -d --build

# Check status
docker-compose ps

# View logs
docker-compose logs -f app
```

### Step 7: Verify Deployment

```bash
# Test health endpoint
curl http://localhost:3000/health

# Test with domain
curl https://compress.yourdomain.com/health
```

## Coolify UI Deployment

### Method 1: Docker Compose in Coolify

1. **Login to Coolify Dashboard**
2. **Create New Project**
3. **Select "Docker Compose" deployment**
4. **Upload Files**:
   - `docker-compose.yml`
   - `Dockerfile`
   - All source code

5. **Set Environment Variables** in Coolify UI:
   - Add all variables from `.env.example`
   - Set secure passwords

6. **Configure Domain**:
   - Add your domain (e.g., `compress.yourdomain.com`)
   - Enable SSL/TLS
   - Point DNS to your VPS IP

7. **Deploy**:
   - Click "Deploy" button
   - Monitor deployment logs
   - Wait for "Running" status

### Method 2: Git Integration

1. **Push to Git Repository** (GitHub/GitLab/Bitbucket)
2. **In Coolify**:
   - Select "Git Repository"
   - Connect your repository
   - Set branch (e.g., `main`)
   - Configure build settings
3. **Auto-Deploy**:
   - Enable automatic deployments
   - Push to trigger rebuild

## WordPress Plugin Setup

### Install Companion WordPress Plugin

Create a WordPress plugin to integrate with the compression API:

**File: `wp-content/plugins/video-compressor/video-compressor.php`**

```php
<?php
/**
 * Plugin Name: Video Compressor
 * Description: Integrate with video compression API
 * Version: 1.0.0
 */

defined('ABSPATH') || exit;

class VideoCompressor {
    private $api_url = 'https://compress.yourdomain.com/api';
    private $api_key = 'your-api-key';

    public function compress_media($post_id, $file_url, $type = 'both') {
        $response = wp_remote_post($this->api_url . '/compress', [
            'headers' => [
                'X-API-Key' => $this->api_key,
                'Content-Type' => 'application/json',
            ],
            'body' => json_encode([
                'post_id' => $post_id,
                'compression_type' => $type,
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
            return false;
        }

        $body = json_decode(wp_remote_retrieve_body($response), true);
        return $body['job_id'] ?? false;
    }

    public function get_status($job_id) {
        $response = wp_remote_get($this->api_url . '/status/' . $job_id, [
            'headers' => ['X-API-Key' => $this->api_key]
        ]);

        if (is_wp_error($response)) {
            return false;
        }

        return json_decode(wp_remote_retrieve_body($response), true);
    }

    public function get_result($job_id) {
        $response = wp_remote_get($this->api_url . '/result/' . $job_id, [
            'headers' => ['X-API-Key' => $this->api_key]
        ]);

        if (is_wp_error($response)) {
            return false;
        }

        return json_decode(wp_remote_retrieve_body($response), true);
    }
}

// Usage
add_action('add_attachment', function($attachment_id) {
    $compressor = new VideoCompressor();
    $file_url = wp_get_attachment_url($attachment_id);
    $job_id = $compressor->compress_media($attachment_id, $file_url, 'video');
    
    if ($job_id) {
        update_post_meta($attachment_id, '_compression_job_id', $job_id);
    }
});
```

## Monitoring & Maintenance

### Health Checks

Add to your monitoring system:

```bash
# Endpoint
https://compress.yourdomain.com/health

# Expected Response
{"status":"healthy","service":"video-compressor-api"}
```

### Logs

```bash
# Application logs
docker-compose logs -f app

# All services
docker-compose logs -f

# Last 100 lines
docker-compose logs --tail=100
```

### Backup Strategy

**Daily Database Backup:**
```bash
#!/bin/bash
# /opt/scripts/backup-compressor-db.sh

DATE=$(date +%Y%m%d_%H%M%S)
docker exec compressor-db pg_dump -U compressor compression > /backups/compression_$DATE.sql
find /backups -name "compression_*.sql" -mtime +7 -delete
```

**Add to crontab:**
```bash
0 2 * * * /opt/scripts/backup-compressor-db.sh
```

### Scaling

**Increase Workers:**
```yaml
# docker-compose.yml
services:
  app:
    environment:
      - MAX_CONCURRENT_JOBS=10
```

**Multiple Instances:**
```bash
docker-compose up -d --scale app=3
```

## Troubleshooting

### Service Won't Start

```bash
# Check logs
docker-compose logs app

# Check environment
docker-compose config

# Rebuild
docker-compose down
docker-compose up -d --build
```

### Database Connection Error

```bash
# Check database is running
docker-compose ps db

# Test connection
docker exec compressor-db psql -U compressor -d compression -c "SELECT 1"

# Reset database
docker-compose down -v
docker-compose up -d
```

### Queue Not Processing

```bash
# Check Redis
docker exec compressor-redis redis-cli PING

# Check queue length
docker exec compressor-redis redis-cli ZCARD compression:queue

# Restart worker
docker-compose restart app
```

## Security Checklist

- [ ] Change default PostgreSQL password
- [ ] Generate strong API_KEY
- [ ] Configure domain whitelist
- [ ] Enable SSL/TLS
- [ ] Set up firewall (allow only 80, 443, 22)
- [ ] Regular security updates
- [ ] Monitor access logs
- [ ] Implement backup strategy

## Performance Optimization

### For Large Files

```env
MAX_VIDEO_FILE_SIZE=10000000000
JOB_TIMEOUT=7200
MAX_CONCURRENT_JOBS=5
```

### For High Traffic

```env
MAX_CONCURRENT_JOBS=15
QUEUE_CHECK_INTERVAL=3
```

### Database Connection Pool

```yaml
db:
  command: postgres -c max_connections=100
```

## Updates & Maintenance

### Update Application

```bash
cd /opt/video-compressor
git pull
docker-compose up -d --build
```

### Update Dependencies

```bash
go get -u ./...
go mod tidy
```

### Database Migrations

Add new migrations to `scripts/migrations/` and run:

```bash
docker exec compressor-db psql -U compressor -d compression -f /migrations/001_add_new_field.sql
```
