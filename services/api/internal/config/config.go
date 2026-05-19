package config

type Config struct {
	Port             string
	DatabasePath     string
	PhotoStoragePath string
}

func Load(getenv func(string) string) Config {
	cfg := Config{
		Port:             "8080",
		DatabasePath:     "data/app.db",
		PhotoStoragePath: "data/photos",
	}

	if port := getenv("PORT"); port != "" {
		cfg.Port = port
	}
	if databasePath := getenv("DATABASE_PATH"); databasePath != "" {
		cfg.DatabasePath = databasePath
	}
	if photoStoragePath := getenv("PHOTO_STORAGE_PATH"); photoStoragePath != "" {
		cfg.PhotoStoragePath = photoStoragePath
	}

	return cfg
}
