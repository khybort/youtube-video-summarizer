package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
	YouTube  YouTubeConfig
	LLM      LLMConfig
	Whisper  WhisperConfig
}

type ServerConfig struct {
	Port int
	Mode string
}

type DatabaseConfig struct {
	URL string
}

type RedisConfig struct {
	URL string
}

type KafkaConfig struct {
	Brokers       []string
	EnableKafka   bool
	ConsumerGroup string
}

type YouTubeConfig struct {
	APIKey string
}

type LLMConfig struct {
	Provider    string // "gemini" | "ollama"
	GeminiKey   string
	OllamaURL   string
	OllamaModel string
}

type WhisperConfig struct {
	Provider        string // "groq" | "local" | "huggingface"
	GroqKey         string
	HuggingFaceKey  string
	LocalWhisperURL string
	LocalModel      string
}

func Load() (*Config, error) {
	// Determine environment (development, production, or custom)
	env := getEnv("APP_ENV", getEnv("ENV", "development"))
	
	// Load environment-specific .env file
	envFile := fmt.Sprintf(".env.%s", env)
	if err := godotenv.Load(envFile); err != nil {
		// Fallback to .env if environment-specific file doesn't exist
		_ = godotenv.Load(".env")
	}
	
	// Also load base .env if it exists (for overrides)
	_ = godotenv.Load(".env")

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnvAsInt("PORT", 8080),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/youtube_analyzer?sslmode=disable"),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", "redis://localhost:6379"),
		},
		Kafka: KafkaConfig{
			Brokers:       getEnvAsStringSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
			EnableKafka:   getEnvAsBool("KAFKA_ENABLED", true),
			ConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "youtube-analyzer"),
		},
		YouTube: YouTubeConfig{
			APIKey: getEnv("YOUTUBE_API_KEY", ""),
		},
		LLM: LLMConfig{
			Provider:    getEnv("DEFAULT_LLM_PROVIDER", "gemini"),
			GeminiKey:   getEnv("GEMINI_API_KEY", ""),
			OllamaURL:   getEnv("OLLAMA_URL", "http://localhost:11434"),
			OllamaModel: getEnv("OLLAMA_MODEL", "llama3.2"),
		},
		Whisper: WhisperConfig{
			Provider:        getEnv("DEFAULT_WHISPER_PROVIDER", "groq"),
			GroqKey:         getEnv("GROQ_API_KEY", ""),
			HuggingFaceKey:  getEnv("HUGGINGFACE_API_KEY", ""),
			LocalWhisperURL: getEnv("LOCAL_WHISPER_URL", "http://localhost:8001"),
			LocalModel:      getEnv("WHISPER_MODEL", "base"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsStringSlice(key string, defaultValue []string) []string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	// Split by comma
	var result []string
	for _, v := range splitString(valueStr, ",") {
		if trimmed := trimString(v); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return defaultValue
	}
	return result
}

func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimString(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

