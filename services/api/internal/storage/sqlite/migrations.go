package sqlite

import (
	"context"
	"database/sql"
)

type migration struct {
	Version int
	SQL     string
}

var migrations = []migration{
	{
		Version: 1,
		SQL: `
create table if not exists schema_migrations (
  version integer primary key,
  applied_at text not null default current_timestamp
);

create table if not exists items (
  id text primary key,
  title text not null,
  description text not null default '',
  category text not null default '',
  size text not null default '',
  condition text not null default '',
  price_cents integer not null default 0,
  currency text not null default 'USD',
  status text not null,
  notes text not null default '',
  created_at text not null,
  updated_at text not null
);

create table if not exists item_photos (
  id text primary key,
  item_id text not null references items(id) on delete cascade,
  storage_id text not null,
  filename text not null,
  mime_type text not null,
  sort_order integer not null default 0,
  is_primary integer not null default 0,
  created_at text not null
);

create table if not exists connected_accounts (
  id text primary key,
  platform text not null,
  display_name text not null,
  status text not null,
  permission_scope text not null default '',
  last_validated_at text,
  created_at text not null,
  updated_at text not null
);

create table if not exists listings (
  id text primary key,
  item_id text not null references items(id) on delete cascade,
  connected_account_id text not null references connected_accounts(id),
  external_id text not null default '',
  status text not null,
  published_at text,
  created_at text not null,
  updated_at text not null
);

create table if not exists listing_attempts (
  id text primary key,
  listing_id text not null references listings(id) on delete cascade,
  status text not null,
  message text not null default '',
  created_at text not null
);

create table if not exists sales (
  id text primary key,
  item_id text not null references items(id),
  sale_price_cents integer not null,
  currency text not null default 'USD',
  sold_at text not null,
  platform text not null default '',
  account_id text not null default '',
  created_at text not null
);
`,
	},
	{
		Version: 2,
		SQL: `
alter table items add column original_purchase_price_cents integer not null default 0;
alter table items add column selling_price_cents integer not null default 0;
update items set selling_price_cents = price_cents where selling_price_cents = 0;
`,
	},
}

func ApplyMigrations(ctx context.Context, db *sql.DB) error {
	for _, migration := range migrations {
		applied, err := migrationApplied(ctx, db, migration.Version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		if _, err := db.ExecContext(ctx, migration.SQL); err != nil {
			return err
		}
		if _, err := db.ExecContext(
			ctx,
			"insert or ignore into schema_migrations (version) values (?)",
			migration.Version,
		); err != nil {
			return err
		}
	}
	return nil
}

func migrationApplied(ctx context.Context, db *sql.DB, version int) (bool, error) {
	var count int
	err := db.QueryRowContext(
		ctx,
		"select count(*) from schema_migrations where version = ?",
		version,
	).Scan(&count)
	if err != nil {
		return false, nil
	}
	return count > 0, nil
}
