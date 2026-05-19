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
