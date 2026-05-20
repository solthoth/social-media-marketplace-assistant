package config

type Config struct {
	Port                string
	DatabasePath        string
	PhotoStoragePath    string
	AIEnrichmentEnabled bool
	AIProvider          string
	AIModel             string
	OpenAIAPIKey        string
}

func Load(getenv func(string) string) Config {
	cfg := Config{
		Port:             "8080",
		DatabasePath:     "data/app.db",
		PhotoStoragePath: "data/photos",
		AIProvider:       "fake",
		AIModel:          "fake-vision",
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
	if aiEnrichmentEnabled := getenv("AI_ENRICHMENT_ENABLED"); aiEnrichmentEnabled == "true" {
		cfg.AIEnrichmentEnabled = true
	}
	if aiProvider := getenv("AI_PROVIDER"); aiProvider != "" {
		cfg.AIProvider = aiProvider
	}
	if aiModel := getenv("AI_MODEL"); aiModel != "" {
		cfg.AIModel = aiModel
	}
	if openAIAPIKey := getenv("OPENAI_API_KEY"); openAIAPIKey != "" {
		cfg.OpenAIAPIKey = openAIAPIKey
	}

	return cfg
}
