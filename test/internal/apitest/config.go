package apitest

import "os"

// Config defines runtime inputs for API tests.
type Config struct {
	BaseURL              string
	AdminUsername        string
	AdminPassword        string
	PlayerUsernamePrefix string
	PlayerEmailSuffix    string
	GameUsername         string
	GamePassword         string
	GameInternalBase     string
}

// LoadConfig reads environment variables with sensible defaults for local docker env.
func LoadConfig() Config {
	return Config{
		BaseURL:              getenv("BASE_URL", "http://localhost:80"),
		AdminUsername:        getenv("ADMIN_USERNAME", "root"),
		AdminPassword:        getenv("ADMIN_PASSWORD", "password"),
		PlayerUsernamePrefix: getenv("PLAYER_USERNAME_PREFIX", "smoke-player"),
		PlayerEmailSuffix:    getenv("PLAYER_EMAIL_SUFFIX", "example.com"),
		GameUsername:         getenv("GAME_USERNAME", ""),
		GamePassword:         getenv("GAME_PASSWORD", ""),
		GameInternalBase:     getenv("GAME_INTERNAL_BASE", "http://localhost:8072"),
	}
}

// getenv fetches an environment variable with fallback to default value.
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
