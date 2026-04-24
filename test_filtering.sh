#!/bin/bash
# Testing the photo filtering feature with seeded test data

# Configuration
BACKEND_URL="http://localhost:3001"
ENVIRONMENT="development"

echo "=== Photo Filtering Test Guide ==="
echo ""
echo "Step 1: Seed test data"
echo "Running: POST $BACKEND_URL/api/v1/test/seed"
echo ""

# Export environment variable for development
export ENVIRONMENT=development

# Call the seed endpoint
SEED_RESPONSE=$(curl -s -X POST "$BACKEND_URL/api/v1/test/seed" -H "Content-Type: application/json")

echo "Seed Response:"
echo "$SEED_RESPONSE" | jq '.'
echo ""

# Extract IDs from response
EVENT_ID=$(echo "$SEED_RESPONSE" | jq -r '.event_id')
GALLERY_ID=$(echo "$SEED_RESPONSE" | jq -r '.gallery_id')
TEST_URL=$(echo "$SEED_RESPONSE" | jq -r '.test_url')

echo "Created:"
echo "  Event ID: $EVENT_ID"
echo "  Gallery ID: $GALLERY_ID"
echo ""
echo "=== Now testing photo filtering endpoints ==="
echo ""

# Test 1: Get all photos
echo "Test 1: List all photos (no filters)"
echo "GET $BACKEND_URL$TEST_URL"
curl -s "$BACKEND_URL$TEST_URL" | jq '.data | length as $count | "Total photos: \($count)"'
echo ""

# Test 2: Filter by taken_after (morning only)
MORNING_START=$(date -u -d "7 days ago 00:00:00" +"%Y-%m-%dT%H:%M:%SZ")
MORNING_END=$(date -u -d "7 days ago 012:00:00" +"%Y-%m-%dT%H:%M:%SZ")

echo "Test 2: Filter morning photos (taken_after to taken_before)"
echo "GET $BACKEND_URL$TEST_URL?taken_after=$MORNING_START&taken_before=$MORNING_END"
curl -s "$BACKEND_URL$TEST_URL?taken_after=$MORNING_START&taken_before=$MORNING_END" | jq '.data | map(.filename)'
echo ""

# Test 3: Filter afternoon only
AFTERNOON_START=$(date -u -d "7 days ago 12:00:00" +"%Y-%m-%dT%H:%M:%SZ")
AFTERNOON_END=$(date -u -d "7 days ago 24:00:00" +"%Y-%m-%dT%H:%M:%SZ")

echo "Test 3: Filter afternoon photos"
echo "GET $BACKEND_URL$TEST_URL?taken_after=$AFTERNOON_START&taken_before=$AFTERNOON_END"
curl -s "$BACKEND_URL$TEST_URL?taken_after=$AFTERNOON_START&taken_before=$AFTERNOON_END" | jq '.data | map(.filename)'
echo ""

# Test 4: Get gallery with time range
echo "Test 4: Get gallery with earliest and latest photo times"
echo "GET $BACKEND_URL/api/v1/events/$EVENT_ID/galleries/$GALLERY_ID"
curl -s "$BACKEND_URL/api/v1/events/$EVENT_ID/galleries/$GALLERY_ID" | jq '.data | {name, earliest_photo_time, latest_photo_time}'
echo ""

echo "=== All tests complete! ==="
