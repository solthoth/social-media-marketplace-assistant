package sqlite

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestApplyMigrationsCreatesInitialTables(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	if err := ApplyMigrations(context.Background(), db); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	expectedTables := []string{
		"schema_migrations",
		"items",
		"item_photos",
		"connected_accounts",
		"listings",
		"listing_attempts",
		"sales",
	}

	for _, table := range expectedTables {
		var name string
		err := db.QueryRowContext(
			context.Background(),
			"select name from sqlite_master where type = 'table' and name = ?",
			table,
		).Scan(&name)
		if err != nil {
			t.Fatalf("expected table %q to exist: %v", table, err)
		}
	}
}
