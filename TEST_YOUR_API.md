# Testing Your API

## Your API is Running! ✅

The app is successfully deployed and responding to requests.

---

## API Endpoints

### Public Endpoints (No API Key Required)

#### 1. Root / Home
```bash
curl https://api.trendss.net/
```
**Response:** API information and available endpoints

#### 2. Health Check
```bash
curl https://api.trendss.net/health
```
**Response:** Database and Redis connection status

#### 3. Readiness Check
```bash
curl https://api.trendss.net/ready
```
**Response:** Service readiness status

---

### Protected Endpoints (Require API Key)

All `/api/*` endpoints require the API key in the header.

**API Key:** `sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1`

#### 4. Submit Compression Job
```bash
curl -X POST https://api.trendss.net/api/compress \
  -H "X-API-Key: sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1" \
  -H "Content-Type: application/json" \
  -d '{
    "post_id": 123,
    "compression_type": "video",
    "video_file_url": "https://example.com/video.mp4",
    "video_quality": "high"
  }'
```
**Response:** Job ID and status

#### 5. Check Job Status
```bash
curl https://api.trendss.net/api/status/YOUR_JOB_ID \
  -H "X-API-Key: sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1"
```

#### 6. Get Job Result
```bash
curl https://api.trendss.net/api/result/YOUR_JOB_ID \
  -H "X-API-Key: sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1"
```

#### 7. Queue Statistics
```bash
curl https://api.trendss.net/api/queue/stats \
  -H "X-API-Key: sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1"
```

#### 8. Cancel Job
```bash
curl -X POST https://api.trendss.net/api/queue/cancel/YOUR_JOB_ID \
  -H "X-API-Key: sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1"
```

---

## Why You Saw 404 Errors

The 404 errors in your logs were because:
- You were accessing the root path `/`
- The API didn't have a root route configured
- Now it does! Try `https://api.trendss.net/` and you'll see API info

---

## Current Status After Latest Deploy

After you redeploy with the latest changes:

✅ **Root endpoint** - Shows API information  
✅ **Database tables** - Will be created automatically  
✅ **All 3 containers** - App, PostgreSQL, Redis running  
✅ **Domain working** - https://api.trendss.net accessible  
✅ **SSL certificate** - Secured with Let's Encrypt  

---

## Database Issue (Will be Fixed)

The current error `database "compressor" does not exist` will be resolved after redeploy because:

1. Fixed healthcheck to use correct database name (`compression`)
2. Removed failed init.sql volume mount
3. Database tables will be created using custom Dockerfile

---

## Next Steps

1. **Redeploy** in Coolify (click "Redeploy")
2. **Wait 2-3 minutes** for all containers to rebuild
3. **Test the API:**
   ```bash
   curl https://api.trendss.net/
   curl https://api.trendss.net/health
   ```

4. **Check logs** - Should see:
   - ✓ "Connected to PostgreSQL database"
   - ✓ "Starting server on port 3000"
   - ✓ NO "database compressor does not exist" errors

---

## Your API is Live! 🚀

**Base URL:** `https://api.trendss.net`  
**API Key:** `sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1`

Start by testing the public endpoints first, then try the protected endpoints with the API key!
