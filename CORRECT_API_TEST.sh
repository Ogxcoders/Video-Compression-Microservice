#!/bin/bash

# Video Compression API - Test Script
# This script tests your compression API with the correct format

API_URL="https://api.trendss.net"
API_KEY="sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1"

echo "========================================="
echo "Testing Video Compression API"
echo "========================================="
echo ""

# Test 1: Health Check
echo "Test 1: Health Check"
echo "---------------------"
curl -s $API_URL/health | jq '.'
echo ""
echo ""

# Test 2: Submit Video Compression Job
echo "Test 2: Submit Compression Job"
echo "-------------------------------"
RESPONSE=$(curl -s -X POST $API_URL/api/compress \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "post_id": 1,
    "compression_type": "video",
    "video_data": {
      "file_url": "https://ogtemplate.com/wp-content/uploads/2025/10/7556584672545852733-video.mp4",
      "quality": "medium"
    },
    "priority": 5
  }')

echo "$RESPONSE" | jq '.'
echo ""

# Extract job_id
JOB_ID=$(echo "$RESPONSE" | jq -r '.job_id')

if [ "$JOB_ID" == "null" ] || [ -z "$JOB_ID" ]; then
    echo "❌ Error: Failed to create job!"
    echo "Response: $RESPONSE"
    exit 1
fi

echo "✅ Job created successfully!"
echo "Job ID: $JOB_ID"
echo ""

# Test 3: Check Status
echo "Test 3: Check Job Status"
echo "-------------------------"
curl -s $API_URL/api/status/$JOB_ID \
  -H "X-API-Key: $API_KEY" | jq '.'
echo ""
echo ""

# Test 4: Wait and check status again
echo "Test 4: Waiting 10 seconds, then checking status again..."
echo "----------------------------------------------------------"
sleep 10

curl -s $API_URL/api/status/$JOB_ID \
  -H "X-API-Key: $API_KEY" | jq '.'
echo ""
echo ""

# Test 5: Get Queue Stats
echo "Test 5: Queue Statistics"
echo "------------------------"
curl -s $API_URL/api/queue/stats \
  -H "X-API-Key: $API_KEY" | jq '.'
echo ""
echo ""

echo "========================================="
echo "✅ All tests completed!"
echo "========================================="
echo ""
echo "To check the result later, run:"
echo "curl $API_URL/api/result/$JOB_ID -H \"X-API-Key: $API_KEY\" | jq '.'"
echo ""
echo "Job ID saved for reference: $JOB_ID"
