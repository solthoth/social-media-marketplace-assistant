# API

The API is a JSON HTTP service served by the Go backend.

## Swagger

Swagger/OpenAPI documentation is served by the backend:

- `GET /swagger/index.html`: Swagger UI.
- `GET /swagger/doc.json`: OpenAPI JSON document.

## Health

### `GET /healthz`

Returns service health.

## Items

Items are the central inventory records. Deleting an item archives it instead of physically removing it.

Inventory status changes are validated as a workflow:

- `draft` can move to `ready_to_list` or `archived`.
- `ready_to_list` can move to `draft`, `listed`, or `archived`.
- `listed` can move to `ready_to_list`, `sold`, or `archived`.
- `sold` can move back to `listed` for correction or to `archived`.
- `archived` can restore to `draft`.

### Create Item

`POST /items`

Request body:

```json
{
  "title": "Denim jacket",
  "description": "Medium wash",
  "category": "Clothing",
  "size": "M",
  "condition": "Good",
  "original_purchase_price_cents": 1800,
  "selling_price_cents": 3200,
  "currency": "USD",
  "notes": "Steam before photos"
}
```

Response: `201 Created` with the item.

Validation:

- `title` is required.
- `original_purchase_price_cents` defaults to `0` and must be zero or greater.
- `selling_price_cents` defaults to `0` and must be zero or greater.
- `currency` defaults to `USD` when omitted. Supported values: `USD`.

### List Items

`GET /items`

Optional query parameters:

- `status`: one of `draft`, `ready_to_list`, `listed`, `sold`, or `archived`.

Response body:

```json
{
  "items": []
}
```

### Get Item

`GET /items/{id}`

Returns `404 Not Found` when the item does not exist.

### Update Item

`PATCH /items/{id}`

Accepts partial item fields:

```json
{
  "title": "Listed denim jacket",
  "status": "ready_to_list"
}
```

Invalid status changes return `400 Bad Request` with error code `invalid_status_transition`.

### Archive Item

`DELETE /items/{id}`

Archives the item by setting status to `archived`.

Response: `204 No Content`.

## Item Photos

Photo bytes are uploaded with multipart form data and served back through API routes. The backend stores metadata in SQLite and stores bytes through the configured photo storage adapter.

Supported upload content types:

- `image/jpeg`
- `image/png`
- `image/webp`

### Upload Photo

`POST /items/{id}/photos`

Request body: `multipart/form-data` with a `photo` file field.

Response: `201 Created` with the photo metadata.

```json
{
  "id": "photo-id",
  "item_id": "item-id",
  "filename": "front.png",
  "mime_type": "image/png",
  "sort_order": 0,
  "is_primary": true,
  "content_urls": {
    "original": "/items/item-id/photos/photo-id/content?variant=original",
    "medium": "/items/item-id/photos/photo-id/content?variant=medium",
    "thumbnail": "/items/item-id/photos/photo-id/content?variant=thumbnail"
  },
  "created_at": "2026-05-19T00:00:00Z"
}
```

The first uploaded photo for an item is marked primary. Later image normalization and generated variants are planned; until variants are generated, variant content routes may return the original stored bytes.

### List Photos

`GET /items/{id}/photos`

Response body:

```json
{
  "photos": []
}
```

### Get Photo Content

`GET /items/{id}/photos/{photoId}/content`

Optional query parameters:

- `variant`: one of `original`, `medium`, or `thumbnail`. Defaults to `original`.

Returns the image bytes with the stored image content type.

### Reorder Photos

`PATCH /items/{id}/photos/order`

Request body:

```json
{
  "photo_ids": ["photo-2", "photo-1"]
}
```

The request must include the full set of photo IDs for the item.

Response: `200 OK` with reordered photo metadata.

### Set Primary Photo

`PATCH /items/{id}/photos/{photoId}/primary`

Marks the selected photo as the item primary photo and clears the primary flag
from the other photos for the same item.

Response: `200 OK` with the current photo metadata list.

### Delete Photo

`DELETE /items/{id}/photos/{photoId}`

Deletes the photo metadata and stored photo bytes.

Response: `204 No Content`.

## Item Enrichment

AI enrichment jobs generate missing item details from an item title and photos. The first implementation stores job history and applies completed suggestions only to empty item fields.

### Create Enrichment Job

`POST /items/{id}/enrichment-jobs`

Creates a queued enrichment job for an item. The item must have a title and at least one photo.

Response: `201 Created` with the enrichment job metadata.

```json
{
  "id": "job-id",
  "item_id": "item-id",
  "status": "queued",
  "provider": "fake",
  "model": "fake-vision",
  "requested_at": "2026-05-19T00:00:00Z",
  "started_at": null,
  "completed_at": null,
  "applied_at": null,
  "error_message": "",
  "input_snapshot": {},
  "suggestion": {
    "description": "",
    "category": "",
    "size": "",
    "condition": "",
    "notes": ""
  }
}
```

### List Enrichment Jobs

`GET /items/{id}/enrichment-jobs`

Response body:

```json
{
  "jobs": []
}
```

### Get Enrichment Job

`GET /items/{id}/enrichment-jobs/{jobId}`

Response: `200 OK` with the enrichment job metadata.

### Apply Enrichment Job

`POST /items/{id}/enrichment-jobs/{jobId}/apply`

Applies completed suggestions to empty item fields only. Existing item values are not overwritten.

Response body:

```json
{
  "item": {},
  "job": {},
  "applied_fields": ["description", "category"]
}
```

## Error Shape

Errors use a consistent response shape:

```json
{
  "error": {
    "code": "invalid_item",
    "message": "item request is invalid"
  }
}
```
