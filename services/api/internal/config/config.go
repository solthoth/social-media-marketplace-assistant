package config

type Config struct {
	Port         string
	DatabasePath string
}

func Load(getenv func(string) string) Config {
	cfg := Config{
		Port:         "8080",
		DatabasePath: "data/app.db",
	}

	if port := getenv("PORT"); port != "" {
		cfg.Port = port
	}
	if databasePath := getenv("DATABASE_PATH"); databasePath != "" {
		cfg.DatabasePath = databasePath
	}

	return cfg
}
