package config

import "os"

type Config struct {
	DummyJSONBaseURL   string
	OpenMeteoBaseURL   string
	FrankfurterBaseURL string
}

func Load() Config {
	return Config{
		DummyJSONBaseURL:   getenv("DUMMYJSON_BASE_URL", "https://dummyjson.com"),
		OpenMeteoBaseURL:   getenv("OPEN_METEO_BASE_URL", "https://api.open-meteo.com"),
		FrankfurterBaseURL: getenv("FRANKFURTER_BASE_URL", "https://api.frankfurter.dev/v2"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
