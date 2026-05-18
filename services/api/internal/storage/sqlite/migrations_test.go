package sqlite

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/suite"
	_ "modernc.org/sqlite"
)

type MigrationSuite struct {
	suite.Suite
	db *sql.DB
}

func TestMigrationSuite(t *testing.T) {
	suite.Run(t, new(MigrationSuite))
}

func (s *MigrationSuite) SetupTest() {
	db, err := sql.Open("sqlite", ":memory:")
	s.Require().NoError(err)
	s.db = db
}

func (s *MigrationSuite) TearDownTest() {
	s.Require().NoError(s.db.Close())
}

func (s *MigrationSuite) TestApplyMigrationsCreatesInitialTables() {
	s.Require().NoError(ApplyMigrations(context.Background(), s.db))

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
		err := s.db.QueryRowContext(
			context.Background(),
			"select name from sqlite_master where type = 'table' and name = ?",
			table,
		).Scan(&name)
		s.Require().NoError(err, "expected table %q to exist", table)
	}
}

func (s *MigrationSuite) TestApplyMigrationsAddsPurchaseAndSellingPrices() {
	s.Require().NoError(ApplyMigrations(context.Background(), s.db))

	columns := map[string]bool{}
	rows, err := s.db.QueryContext(context.Background(), "pragma table_info(items)")
	s.Require().NoError(err)
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var columnType string
		var notNull int
		var defaultValue any
		var pk int
		s.Require().NoError(rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk))
		columns[name] = true
	}
	s.Require().NoError(rows.Err())

	s.True(columns["original_purchase_price_cents"])
	s.True(columns["selling_price_cents"])
}
