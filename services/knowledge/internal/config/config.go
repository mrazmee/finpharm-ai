package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	MigrationsDir string
	SourceDir     string

	GeminiAPIKey             string
	EmbeddingModel           string
	EmbeddingOutputDimension int

	ChunkMaxChars     int
	ChunkOverlapChars int
	BatchSize         int

	DryRun bool
}

func Load() Config {
	return Config{
		AppEnv: getEnv("APP_ENV", "local"),

		DBHost:     getEnv("KNOWLEDGE_DB_HOST", "127.0.0.1"),
		DBPort:     getEnv("KNOWLEDGE_DB_PORT", "55432"),
		DBUser:     getEnv("KNOWLEDGE_DB_USER", "finpharm"),
		DBPassword: getEnv("KNOWLEDGE_DB_PASSWORD", "finpharm"),
		DBName:     getEnv("KNOWLEDGE_DB_NAME", "postgres"),
		DBSSLMode:  getEnv("KNOWLEDGE_DB_SSLMODE", "disable"),

		MigrationsDir: getEnv("KNOWLEDGE_MIGRATIONS_DIR", "./services/knowledge/migrations"),
		SourceDir:     getEnv("KNOWLEDGE_SOURCE_DIR", "./knowledge/sop"),

		GeminiAPIKey:             firstNonEmpty(os.Getenv("GEMINI_API_KEY"), os.Getenv("GOOGLE_API_KEY")),
		EmbeddingModel:           getEnv("KNOWLEDGE_EMBEDDING_MODEL", "models/gemini-embedding-001"),
		EmbeddingOutputDimension: getEnvInt("KNOWLEDGE_EMBEDDING_DIMENSION", 768),

		ChunkMaxChars:     getEnvInt("KNOWLEDGE_CHUNK_MAX_CHARS", 900),
		ChunkOverlapChars: getEnvInt("KNOWLEDGE_CHUNK_OVERLAP_CHARS", 120),
		BatchSize:         getEnvInt("KNOWLEDGE_BATCH_SIZE", 8),

		DryRun: getEnvBool("KNOWLEDGE_DRY_RUN", false),
	}
}

func (c Config) ValidateForMigrate() error {
	if strings.TrimSpace(c.DBHost) == "" {
		return errConfig("KNOWLEDGE_DB_HOST is required")
	}
	if !isPositiveIntegerString(c.DBPort) {
		return errConfig("KNOWLEDGE_DB_PORT must be a positive integer")
	}
	if strings.TrimSpace(c.DBUser) == "" {
		return errConfig("KNOWLEDGE_DB_USER is required")
	}
	if strings.TrimSpace(c.DBName) == "" {
		return errConfig("KNOWLEDGE_DB_NAME is required")
	}
	if strings.TrimSpace(c.DBSSLMode) == "" {
		return errConfig("KNOWLEDGE_DB_SSLMODE is required")
	}
	if strings.TrimSpace(c.MigrationsDir) == "" {
		return errConfig("KNOWLEDGE_MIGRATIONS_DIR is required")
	}
	return nil
}

func (c Config) ValidateForIngest() error {
	if err := c.ValidateForMigrate(); err != nil {
		return err
	}
	if strings.TrimSpace(c.SourceDir) == "" {
		return errConfig("KNOWLEDGE_SOURCE_DIR is required")
	}
	if c.ChunkMaxChars <= 0 {
		return errConfig("KNOWLEDGE_CHUNK_MAX_CHARS must be > 0")
	}
	if c.ChunkOverlapChars < 0 {
		return errConfig("KNOWLEDGE_CHUNK_OVERLAP_CHARS must be >= 0")
	}
	if c.ChunkOverlapChars >= c.ChunkMaxChars {
		return errConfig("KNOWLEDGE_CHUNK_OVERLAP_CHARS must be smaller than KNOWLEDGE_CHUNK_MAX_CHARS")
	}
	if c.BatchSize <= 0 {
		return errConfig("KNOWLEDGE_BATCH_SIZE must be > 0")
	}
	if c.EmbeddingOutputDimension <= 0 {
		return errConfig("KNOWLEDGE_EMBEDDING_DIMENSION must be > 0")
	}
	if strings.TrimSpace(c.EmbeddingModel) == "" {
		return errConfig("KNOWLEDGE_EMBEDDING_MODEL is required")
	}
	if !c.DryRun && strings.TrimSpace(c.GeminiAPIKey) == "" {
		return errConfig("GEMINI_API_KEY or GOOGLE_API_KEY is required when KNOWLEDGE_DRY_RUN=false")
	}
	return nil
}

func (c Config) ValidateForQuery() error {
	if err := c.ValidateForMigrate(); err != nil {
		return err
	}
	if strings.TrimSpace(c.GeminiAPIKey) == "" {
		return errConfig("GEMINI_API_KEY or GOOGLE_API_KEY is required for retrieval query")
	}
	if strings.TrimSpace(c.EmbeddingModel) == "" {
		return errConfig("KNOWLEDGE_EMBEDDING_MODEL is required")
	}
	if c.EmbeddingOutputDimension <= 0 {
		return errConfig("KNOWLEDGE_EMBEDDING_DIMENSION must be > 0")
	}
	return nil
}

func (c Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost,
		c.DBPort,
		c.DBUser,
		c.DBPassword,
		c.DBName,
		c.DBSSLMode,
	)
}

func errConfig(msg string) error {
	return fmt.Errorf("config validation error: %s", msg)
}

func isPositiveIntegerString(v string) bool {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	return err == nil && n > 0
}

func getEnv(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

func getEnvInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func getEnvBool(key string, def bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return def
	}
	switch v {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return def
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}