package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppPort       string
	GinMode       string
	DBURL         string
	DBMaxConns    int
	RedisURL      string
	JWTSecret     string
	JWTAccessTTL  string
	JWTRefreshTTL string
	JWTIssuer     string
	JWTAudience   string
	BcryptCost    int
	CORSOrigins   string
	RunMigrations bool
}

func Load() (*Config, error) {
	dbMaxConns, err := strconv.Atoi(getEnv("DB_MAX_CONNS", "10"))
	if err != nil {
		return nil, err
	}

	bcryptCost, err := strconv.Atoi(getEnv("BCRYPT_COST", "12"))
	if err != nil {
		return nil, err
	}

	runMigrations, err := strconv.ParseBool(getEnv("RUN_MIGRATIONS", "false"))
	if err != nil {
		return nil, err
	}

	return &Config{
		AppPort:       getEnv("APP_PORT", "8080"),
		GinMode:       getEnv("GIN_MODE", "debug"),
		DBURL:         getEnv("DB_URL", ""),
		DBMaxConns:    dbMaxConns,
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		JWTAccessTTL:  getEnv("JWT_ACCESS_TTL", "15m"),
		JWTRefreshTTL: getEnv("JWT_REFRESH_TTL", "168h"),
		JWTIssuer:     getEnv("JWT_ISSUER", "scanner-auth"),
		JWTAudience:   getEnv("JWT_AUDIENCE", "scanner-auth"),
		BcryptCost:    bcryptCost,
		CORSOrigins:   getEnv("CORS_ORIGINS", ""),
		RunMigrations: runMigrations,
	}, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}
