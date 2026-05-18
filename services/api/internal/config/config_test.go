package config

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ConfigSuite struct {
	suite.Suite
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}

func (s *ConfigSuite) TestLoadUsesDefaults() {
	cfg := Load(func(string) string {
		return ""
	})

	s.Equal("8080", cfg.Port)
	s.Equal("data/app.db", cfg.DatabasePath)
}

func (s *ConfigSuite) TestLoadUsesEnvironmentOverrides() {
	values := map[string]string{
		"PORT":          "9090",
		"DATABASE_PATH": "/tmp/marketplace.db",
	}

	cfg := Load(func(key string) string {
		return values[key]
	})

	s.Equal("9090", cfg.Port)
	s.Equal("/tmp/marketplace.db", cfg.DatabasePath)
}
