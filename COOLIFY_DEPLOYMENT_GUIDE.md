# Coolify Deployment Guide

This guide will help you deploy the Video Compression API to Coolify successfully.

## Important: Coolify handles reverse proxy automatically

**Do NOT use the nginx service when deploying to Coolify.** Coolify provides its own reverse proxy and domain management. Using the nginx container will cause conflicts.

## Step 1: Add Environment Variables in Coolify UI

In your Coolify project, go to **Environment Variables** and add the following (copy-paste each one):

```
API_KEY=sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1
ALLOWED_DOMAINS=https://capcut.ogtemplate.com/,https://ogtemplate.com/
PORT=3000
LOG_LEVEL=info
MAX_VIDEO_FILE_SIZE=5000000000
MAX_IMAGE_FILE_SIZE=500000000
TEMP_DIR=/tmp/compression
REDIS_URL=redis://redis:6379
DATABASE_URL=postgres://compressor:compressor_secure_pw_9x8c7v6b5n4m3@db:5432/compression?sslmode=disable
MAX_CONCURRENT_JOBS=5
JOB_TIMEOUT=3600
QUEUE_CHECK_INTERVAL=5
FFMPEG_PATH=/usr/bin/ffmpeg
IMAGEMAGICK_PATH=/usr/bin/convert
WORDPRESS_API_URL=https://capcut.ogtemplate.com/wp-json/wp/v2
WORDPRESS_USERNAME=vps
WORDPRESS_APP_PASSWORD=bisf lAxw AsTk Jm2t ytUb 3ENg
RATE_LIMIT_REQUESTS_PER_MINUTE=10
RATE_LIMIT_MAX_CONCURRENT=100
RATE_LIMIT_MAX_JOBS_PER_DAY=1000
MAX_RETRIES=3
RETRY_BACKOFF_SECONDS=60,300,900
POSTGRES_DB=compression
POSTGRES_USER=compressor
POSTGRES_PASSWORD=compressor_secure_pw_9x8c7v6b5n4m3
```

## Step 2: Configure Domain in Coolify

1. In Coolify, go to your application settings
2. Under **Domains**, add your domain: `https://api.trendss.net`
3. Make sure SSL/TLS is enabled (Coolify handles this automatically)
4. Set the port to **3000** (this is where the Go API listens)

## Step 3: Deploy

Use the `docker-compose.coolify.yml` file for deployment:

1. In Coolify, set the **Docker Compose File Path** to: `docker-compose.coolify.yml`
2. Click **Deploy**
3. Wait for the deployment to complete

## Step 4: Verify

After deployment, check the logs in Coolify. You should see:
- No "WARNING: API_KEY is not set" messages
- No "WARNING: ALLOWED_DOMAINS is not set" messages
- No "DATABASE_URL is required" errors
- "Connected to PostgreSQL database" message

## Troubleshooting

### If you still see environment variable warnings:

1. Make sure all environment variables are added in Coolify UI (not just in .env file)
2. Verify the variables are referenced in `docker-compose.coolify.yml` using `${VARIABLE_NAME}` syntax
3. Force a fresh deployment (disable cache)

### If the domain doesn't work:

1. Check that you're using `docker-compose.coolify.yml` (NOT the regular `docker-compose.yml` with nginx)
2. Verify the domain is correctly configured in Coolify
3. Make sure port 3000 is exposed in the app service
4. Check Coolify's reverse proxy logs

### If the database connection fails:

1. Wait 30-60 seconds after deployment for PostgreSQL to initialize
2. Check that `POSTGRES_PASSWORD` matches in both `DATABASE_URL` and `POSTGRES_PASSWORD` variables
3. View container logs for the `db` service

## API Endpoint

Once deployed, your API will be available at: `https://api.trendss.net/`

Test the health endpoint: `https://api.trendss.net/health`
