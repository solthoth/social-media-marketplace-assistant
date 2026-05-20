package sqlite

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type DatabaseSuite struct {
	suite.Suite
}

func TestDatabaseSuite(t *testing.T) {
	suite.Run(t, new(DatabaseSuite))
}

func (s *DatabaseSuite) TestOpenSerializesSQLiteConnections() {
	db, err := Open(context.Background(), filepath.Join(s.T().TempDir(), "test.db"))
	s.Require().NoError(err)
	defer db.Close()

	s.Equal(1, db.Stats().MaxOpenConnections)
}
