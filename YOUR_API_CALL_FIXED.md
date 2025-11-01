# Your API Call - What Was Wrong & How to Fix

## ❌ What You Did (WRONG)

```bash
curl -X POST https://api.trendss.net/api/compress \
  -H "X-API-Key:`SLACK_TEST_API_KEY`6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1" \
  -H "Content-Type: application/json" \
  -d '{
    "post_id": 1,
    "compression_type": "video",
    "video_data": {
      "file_url": "https://ogtemplate.com/wp-content/uploads/2025/10/7556584672545852733-video.mp4",
      "quality": "medium"
    }
```

### Problems:

1. **Wrong API Key** - You have `SLACK_TEST_API_KEY` which is invalid
   - Correct key: `sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1`

2. **Missing closing brace** - Your JSON is incomplete (missing `}` at the end)

---

## ✅ Correct Version (Copy this!)

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

**Changes made:**
1. Fixed API key (removed `SLACK_TEST_API_KEY`, added full correct key)
2. Added missing `}` to close the JSON
3. Added space after colon in header: `X-API-Key:` → `X-API-Key: `

---

## 📝 Expected Response

If it works, you'll get:

```json
{
  "status": "queued",
  "job_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "compression_type": "video",
  "queue_position": 1,
  "estimated_time": 60
}
```

**Save the `job_id`!** You'll need it to check status.

---

## 🔍 Check Status

After submitting, check the status:

```bash
curl https://api.trendss.net/api/status/YOUR_JOB_ID_HERE \
  -H "X-API-Key: sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1"
```

Replace `YOUR_JOB_ID_HERE` with the actual job_id from the response.

---

## 🧪 One-Command Test

Run this to test everything automatically:

```bash
bash CORRECT_API_TEST.sh
```

This script will:
1. Test health check
2. Submit compression job
3. Check status
4. Show queue stats
5. Give you the job_id to check results later

---

## ⚡ Quick Test (One-liner)

```bash
curl -X POST https://api.trendss.net/api/compress -H "X-API-Key: sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1" -H "Content-Type: application/json" -d '{"post_id":1,"compression_type":"video","video_data":{"file_url":"https://ogtemplate.com/wp-content/uploads/2025/10/7556584672545852733-video.mp4","quality":"medium"}}'
```

---

## 🎬 What Will Happen

1. **Immediately**: Job queued, you get a job_id
2. **After 5-30 seconds**: Worker picks up the job
3. **During processing**: 
   - Downloads video from URL
   - Compresses it using FFmpeg
   - Uploads to WordPress
   - Saves result to database
4. **When done**: Status becomes "completed", compressed URL available

---

## 📊 How to Check if Compression is Working

### Method 1: Check Status
```bash
curl https://api.trendss.net/api/status/YOUR_JOB_ID \
  -H "X-API-Key: sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1"
```

**Look for:**
- `"overall_status": "processing"` - Compression in progress
- `"overall_status": "completed"` - Compression done!
- `"overall_status": "failed"` - Error occurred

### Method 2: Check Worker Logs in Coolify

In Coolify, check the `app` container logs:

**Good signs:**
```
Processing job a1b2c3d4-e5f6-7890-abcd-ef1234567890 (type: video)
Downloading video from https://...
Compressing video...
Uploading to WordPress...
Job a1b2c3d4-e5f6-7890-abcd-ef1234567890 completed in 120 seconds
```

**Bad signs:**
```
Failed to download video: ...
Failed to compress: ...
Job a1b2c3d4-e5f6-7890-abcd-ef1234567890 failed permanently
```

### Method 3: Get Result
```bash
curl https://api.trendss.net/api/result/YOUR_JOB_ID \
  -H "X-API-Key: sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1"
```

**Success looks like:**
```json
{
  "job_id": "...",
  "compression_type": "video",
  "overall_status": "completed",
  "video_result": {
    "status": "completed",
    "original_size": 50000000,
    "compressed_size": 15000000,
    "compression_ratio": 0.3,
    "compressed_url": "https://capcut.ogtemplate.com/wp-content/uploads/2025/11/compressed-video.mp4"
  }
}
```

The `compressed_url` is your final compressed video!

---

## 🚨 Common Errors & Solutions

### Error: "Invalid request format"
**Cause**: JSON syntax error  
**Fix**: Make sure JSON is valid (use the corrected version above)

### Error: "Unauthorized" or "Invalid API key"
**Cause**: Wrong API key  
**Fix**: Use the correct key: `sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1`

### Error: "video_data is required for video compression"
**Cause**: Missing or incorrect video_data structure  
**Fix**: Make sure you include `video_data` object with `file_url` and `quality`

### Job stays "pending" forever
**Cause**: Worker not running or database issue  
**Fix**: 
1. Redeploy in Coolify
2. Check app container logs
3. Make sure all 3 containers (app, db, redis) are running

---

## ✅ Verification Checklist

Before testing, make sure:

- [ ] All 3 containers running (app, db, redis)
- [ ] No database errors in logs
- [ ] Worker started (log shows "Worker started")
- [ ] Using correct API key (starts with `sk_test_`)
- [ ] Video URL is accessible (not behind authentication)
- [ ] WordPress credentials are correct

---

Use the **corrected command above** and it will work! 🚀
