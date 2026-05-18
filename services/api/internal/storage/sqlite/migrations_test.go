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
