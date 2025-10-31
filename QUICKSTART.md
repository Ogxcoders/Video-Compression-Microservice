# Quick Start Guide

## 🚀 Deploy in 5 Minutes

### Prerequisites
- VPS with Docker installed
- Domain name pointed to your VPS
- WordPress site with REST API enabled

### Step 1: Clone & Configure (2 minutes)

```bash
# SSH to your VPS
ssh user@your-vps-ip

# Clone repository
git clone <your-repo-url> /opt/video-compressor
cd /opt/video-compressor

# Copy environment template
cp .env.example .env

# Edit environment variables
nano .env
```

**Minimal Configuration:**
```env
API_KEY=your-secure-random-key-here
ALLOWED_DOMAINS=https://yourdomain.com
WORDPRESS_API_URL=https://yourdomain.com/wp-json/wp/v2
WORDPRESS_USERNAME=admin
WORDPRESS_APP_PASSWORD=xxxx xxxx xxxx xxxx
```

### Step 2: Generate SSL Certificates (1 minute)

```bash
# Option A: Let's Encrypt (Recommended)
sudo certbot certonly --standalone -d compress.yourdomain.com
sudo cp /etc/letsencrypt/live/compress.yourdomain.com/fullchain.pem nginx/ssl/cert.pem
sudo cp /etc/letsencrypt/live/compress.yourdomain.com/privkey.pem nginx/ssl/key.pem

# Option B: Self-Signed (Development)
mkdir -p nginx/ssl
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout nginx/ssl/key.pem -out nginx/ssl/cert.pem
```

### Step 3: Update Nginx Config (30 seconds)

```bash
# Replace compress.yourdomain.com with your actual domain
sed -i 's/compress.yourdomain.com/your-actual-domain.com/g' nginx/nginx.conf
```

### Step 4: Deploy (1 minute)

```bash
# Build and start all services
docker-compose up -d --build

# Check status
docker-compose ps

# View logs
docker-compose logs -f app
```

### Step 5: Test (30 seconds)

```bash
# Test health endpoint
curl https://compress.yourdomain.com/health

# Expected response:
# {"status":"healthy","service":"video-compressor-api"}
```

## ✅ Verify Installation

### Test Compression API

```bash
curl -X POST https://compress.yourdomain.com/api/compress \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "post_id": 1,
    "compression_type": "video",
    "video_data": {
      "file_url": "https://yourdomain.com/test-video.mp4",
      "quality": "medium"
    }
  }'
```

**Expected Response:**
```json
{
  "status": "queued",
  "job_id": "uuid-here",
  "compression_type": "video",
  "queue_position": 1,
  "estimated_time": 60
}
```

### Check Queue Stats

```bash
curl https://compress.yourdomain.com/api/queue/stats \
  -H "X-API-Key: your-api-key"
```

## 🔧 WordPress Integration

### Generate WordPress App Password

1. Login to WordPress admin
2. Go to **Users → Profile**
3. Scroll to **Application Passwords**
4. Enter name: "Video Compressor"
5. Click **Add New Application Password**
6. Copy the generated password

### Test WordPress Connection

```bash
# From your VPS
docker exec compressor-app curl -u "admin:your-app-password" \
  https://yourdomain.com/wp-json/wp/v2/media
```

## 📊 Monitoring

### View Logs

```bash
# All services
docker-compose logs -f

# Just the app
docker-compose logs -f app

# Last 100 lines
docker-compose logs --tail=100 app
```

### Check Service Status

```bash
docker-compose ps
```

### Access Database

```bash
docker exec -it compressor-db psql -U compressor -d compression
```

## 🛠 Common Tasks

### Restart Services

```bash
docker-compose restart app
```

### Update Application

```bash
cd /opt/video-compressor
git pull
docker-compose up -d --build
```

### View Active Jobs

```bash
docker exec -it compressor-db psql -U compressor -d compression \
  -c "SELECT job_id, status, compression_type, created_at FROM jobs ORDER BY created_at DESC LIMIT 10;"
```

### Clear Queue

```bash
docker exec compressor-redis redis-cli DEL compression:queue
```

## 🐛 Troubleshooting

### Service Won't Start

```bash
# Check logs for errors
docker-compose logs app

# Rebuild
docker-compose down
docker-compose up -d --build
```

### Database Connection Error

```bash
# Test database
docker exec compressor-db psql -U compressor -d compression -c "SELECT 1"

# Reset if needed
docker-compose down -v
docker-compose up -d
```

### Jobs Not Processing

```bash
# Check worker logs
docker-compose logs -f app | grep "Processing job"

# Check queue
docker exec compressor-redis redis-cli ZCARD compression:queue

# Restart worker
docker-compose restart app
```

### Permission Errors

```bash
# Fix temp directory permissions
sudo chown -R 1000:1000 /opt/video-compressor/tmp
chmod 755 /opt/video-compressor/tmp
```

## 📚 Next Steps

1. **Read Full Documentation:**
   - [README.md](README.md) - Complete feature overview
   - [API_DOCUMENTATION.md](API_DOCUMENTATION.md) - API reference
   - [DEPLOYMENT.md](DEPLOYMENT.md) - Advanced deployment

2. **Secure Your Installation:**
   - Change default PostgreSQL password
   - Set strong API_KEY
   - Configure firewall (allow only 80, 443, 22)
   - Set up automatic backups

3. **Optimize Performance:**
   - Increase MAX_CONCURRENT_JOBS for high traffic
   - Configure queue check interval
   - Set up monitoring alerts

4. **WordPress Plugin:**
   - Install companion plugin (see DEPLOYMENT.md)
   - Configure auto-compression on upload
   - Set up cron jobs for bulk processing

## 🎉 You're Ready!

Your video compression microservice is now running and ready to process jobs!

**API Endpoint:** `https://compress.yourdomain.com/api`  
**Health Check:** `https://compress.yourdomain.com/health`

For support and advanced features, check the full documentation.
