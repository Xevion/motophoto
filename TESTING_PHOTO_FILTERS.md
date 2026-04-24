# Photo Filtering Feature - Testing Guide

## Quick Start Testing

### 1. **Seed Test Data**

First, set the environment and call the seed endpoint:

```bash
export ENVIRONMENT=development
curl -X POST http://localhost:3001/api/v1/test/seed
```

**Response (example):**
```json
{
  "data": {
    "message": "Successfully seeded 10 test photos",
    "user_id": "xyz123",
    "event_id": "abc456",
    "gallery_id": "def789",
    "photo_count": 10,
    "test_url": "/api/v1/events/abc456/galleries/def789/photos"
  }
}
```

### 2. **Test the Photo Filtering API**

The seed endpoint creates 10 photos with timestamps spanning 8 hours:
- 3 morning photos (0-1 hour)
- 1 lunch break photo (3 hours)
- 3 afternoon photos (5-6 hours)
- 3 finals photos (7-8 hours)

Save your IDs, then test these endpoints:

#### **Get all photos (no filters)**
```bash
curl http://localhost:3001/api/v1/events/{eventId}/galleries/{galleryId}/photos
```

#### **Filter: Morning photos only** (first hour)
```bash
# Get photos from 7 days ago, 00:00 to 01:00
curl "http://localhost:3001/api/v1/events/{eventId}/galleries/{galleryId}/photos?taken_after=2026-03-27T00:00:00Z&taken_before=2026-03-27T01:00:00Z"
```

#### **Filter: Afternoon photos** (5:00-7:00)
```bash
curl "http://localhost:3001/api/v1/events/{eventId}/galleries/{galleryId}/photos?taken_after=2026-03-27T05:00:00Z&taken_before=2026-03-27T07:00:00Z"
```

#### **Filter: From a specific time onward** (after lunch)
```bash
curl "http://localhost:3001/api/v1/events/{eventId}/galleries/{galleryId}/photos?taken_after=2026-03-27T03:00:00Z"
```

#### **Get gallery with time range**
```bash
curl http://localhost:3001/api/v1/events/{eventId}/galleries/{galleryId}
```

This shows `earliest_photo_time` and `latest_photo_time` extracted from the photos.

### 3. **Frontend Testing**

Once the backend is working, test the UI:

1. Start frontend: `cd web && npm run dev`
2. Navigate to: `http://localhost:5173/events/{eventId}`
3. Click on a gallery card → should show photos with time range picker
4. Adjust the time filters → photos should update without page reload

## What Gets Seeded

The seed endpoint creates:

| Timestamp | Photo Name | Hours |
|-----------|-----------|-------|
| 00:00 | morning-1 | 0 |
| 00:30 | morning-2 | 0.5 |
| 01:00 | morning-3 | 1 |
| 03:00 | lunch-break | 3 |
| 05:00 | afternoon-1 | 5 |
| 05:30 | afternoon-2 | 5.5 |
| 06:00 | afternoon-3 | 6 |
| 07:00 | finals-1 | 7 |
| 07:30 | finals-2 | 7.5 |
| 08:00 | finals-3 | 8 |

All timestamps are relative to 7 days ago from "now".

## Response Examples

### ✅ List photos response:
```json
{
  "data": [
    {
      "id": "photo-xyz",
      "filename": "morning-1.jpg",
      "preview_url": "https://cdn.example.com/...",
      "taken_at": "2026-03-27T00:00:00Z",
      "width": 1920,
      "height": 1080,
      "content_type": "image/jpeg",
      "size_bytes": 102400
    }
  ]
}
```

### ✅ Gallery response:
```json
{
  "data": {
    "id": "gallery-123",
    "name": "Morning Heats",
    "photo_count": 10,
    "earliest_photo_time": "2026-03-27T00:00:00Z",
    "latest_photo_time": "2026-03-27T08:00:00Z"
  }
}
```

## Troubleshooting

**"Test endpoint only available in development mode"**
- Make sure to set: `export ENVIRONMENT=development`

**"Photos returned but no taken_at"**
- This shouldn't happen with the seed endpoint, but in production photos without EXIF won't have `taken_at`

**"No photos returned with filters"**
- Double-check your timestamp format (must be ISO-8601: `YYYY-MM-DDTHH:MM:SSZ`)
- Verify the times are within the photo range (0-8 hours relative to 7 days ago)

## Testing Checklist

- [ ] Seed endpoint returns 10 photos
- [ ] List all photos works
- [ ] Filter by `taken_after` works
- [ ] Filter by `taken_before` works
- [ ] Combining both filters works
- [ ] Gallery response includes `earliest_photo_time` and `latest_photo_time`
- [ ] Frontend displays gallery with time picker
- [ ] Frontend time picker filters photos
- [ ] Photos update without page reload

