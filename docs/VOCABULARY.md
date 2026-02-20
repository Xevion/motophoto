# Vocabulary

Canonical terms for the MotoPhoto domain. Use these consistently across code, comments, commit messages, UI text, and conversation.

## Core Entities

| Term | Definition | Notes |
|------|-----------|-------|
| **Event** | A sporting occasion where photos are taken — motocross race, BMX competition, rodeo, swim meet, etc. | Top-level organizing entity. Has a date, location, and sport type. |
| **Gallery** | A collection of photos from a single event, shot by one photographer. | An event can have multiple galleries (one per photographer). |
| **Photo** | A single image captured at an event, available for purchase. | Has a storage key, optional watermark, dimensions, price, and metadata tags. |
| **Photographer** | A user who captures and uploads photos to galleries. | A user with `role = 'photographer'`. |
| **Customer** | A user who browses and purchases photos. | A user with `role = 'customer'`. Default role for new signups. |
| **Admin** | A user with platform management privileges. | A user with `role = 'admin'`. |
| **Tag** | A key-value pair attached to a photo for searchability. | Examples: `rider_number: 42`, `color: red`, `position: 1st`. Stored in `photo_tags`. |

## Supporting Concepts

| Term | Definition | Notes |
|------|-----------|-------|
| **Sport** | The type of athletic activity at an event. | Values: `motocross`, `bmx`, `rodeo`, `swimming`, etc. Stored as text, not an enum. |
| **Watermark** | A visual overlay on a photo to prevent unpaid use. | Stored as a separate `watermarked_key` alongside the original `storage_key`. |
| **Storage key** | The identifier for a photo file in object storage. | Opaque string — the storage backend (S3, R2, local) determines the actual URL. |
| **Price** | The cost to purchase a photo, in **cents** (USD). | Always stored as integer cents (`price_cents`) to avoid floating-point issues. |

## Entity Relationships

```
User (photographer) ──creates──► Event
                     ──creates──► Gallery ──belongs to──► Event
                                  Gallery ──contains──► Photo
                                                         Photo ──has many──► Tag

User (customer) ──browses──► Event ──► Gallery ──► Photo
                ──purchases──► Photo
```

## Anti-Patterns

| Don't say | Say instead | Why |
|-----------|-------------|-----|
| image | photo | We sell photos, not generic images |
| album | gallery | Galleries are tied to events and photographers |
| user (when role matters) | photographer, customer, admin | Be specific about which role |
| tournament, game, match | event | Single canonical term for any occasion |
| price (ambiguous) | price in cents, `price_cents` | Always clarify the unit |
| picture | photo | Consistency |
| contest, competition | event | Even competitive events are just "events" |
| folder, collection | gallery | Gallery is the domain term |
| label | tag | Tags are key-value pairs on photos |
