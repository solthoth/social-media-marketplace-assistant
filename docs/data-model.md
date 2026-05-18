# MVP Data Model

The MVP data model keeps inventory as the source of truth. Platform listings are downstream records that reference inventory items.

## Item

An item represents a good that may eventually be published to one or more sales channels.

Initial fields:

- `id`: unique item identifier.
- `title`: required seller-facing item title.
- `description`: item description.
- `category`: flexible text field. This intentionally starts as free text so categories can evolve from real seller usage.
- `size`: flexible text field for clothing, shoes, or other size labels.
- `condition`: flexible text field for the seller's condition notes.
- `original_purchase_price_cents`: optional integer amount in the smallest currency unit. Defaults to `0`.
- `selling_price_cents`: optional integer amount in the smallest currency unit. Defaults to `0`.
- `currency`: three-letter uppercase currency code. Default is `USD`.
- `status`: inventory status.
- `notes`: private seller notes.
- `created_at` and `updated_at`: UTC timestamps.

## Inventory Statuses

Initial statuses:

- `draft`: item is being captured and is not ready to list.
- `ready_to_list`: item details are complete enough for review or publishing.
- `listed`: item has at least one active listing.
- `sold`: item is sold and should be hidden from publish queues by default.
- `archived`: item is retained for history but no longer active.

## Item Photo

An item photo references media stored by the application.

Initial fields:

- `id`
- `item_id`
- `storage_id`
- `filename`
- `mime_type`
- `sort_order`
- `is_primary`
- `created_at`

## Connected Account

A connected account represents a platform destination. Credentials and auth flows are intentionally deferred until the security design is complete.

Initial fields:

- `id`
- `platform`
- `display_name`
- `status`
- `permission_scope`
- `last_validated_at`
- `created_at`
- `updated_at`

## Listing And Attempts

A listing records the app's relationship between an item and a connected account. Listing attempts record publish/update outcomes over time.

Initial listing fields:

- `id`
- `item_id`
- `connected_account_id`
- `external_id`
- `status`
- `published_at`
- `created_at`
- `updated_at`

Initial listing attempt fields:

- `id`
- `listing_id`
- `status`
- `message`
- `created_at`

## Sale

A sale records that an item is no longer available.

Initial fields:

- `id`
- `item_id`
- `sale_price_cents`
- `currency`
- `sold_at`
- `platform`
- `account_id`
- `created_at`
