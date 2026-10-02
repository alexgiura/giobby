package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	_ "time/tzdata" // embedded IANA DB so Europe/Bucharest works in minimal images (Alpine) without tzdata

	"github.com/caarlos0/env/v11"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
)

// AppSettings holds configuration for the API server.
type AppSettings struct {
	ServerPort  string `env:"SERVER_PORT" envDefault:"8080"`
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
	DebugMode   bool   `env:"DEBUG_MODE" envDefault:"false"`
	// AppTimezone is an IANA name for logs, DB session SET TIME ZONE, and UI. TZ env (e.g. Docker) overrides when set.
	AppTimezone string `env:"APP_TIMEZONE" envDefault:"Europe/Bucharest"`
}

// DatabaseSettings holds configuration related to the PostgreSQL database.
type DatabaseSettings struct {
	User     string `env:"POSTGRES_DB_USER" envDefault:"postgres"`
	Password string `env:"POSTGRES_DB_PASSWORD" envDefault:"postgres"`
	Host     string `env:"POSTGRES_DB_HOST" envDefault:"localhost"`
	Port     string `env:"POSTGRES_DB_PORT" envDefault:"5432"`
	DbName   string `env:"POSTGRES_DB_NAME" envDefault:"giobby_bff"`
	SSLMode  string `env:"POSTGRES_DB_SSLMODE" envDefault:"disable"`
}

// Config holds configuration for the API and database.
type Config struct {
	AppSettings           AppSettings
	DatabaseSettings      DatabaseSettings
	JWTSecret             string `env:"JWT_SECRET" envDefault:"dev-jwt-secret-change-me"`
	AccessTokenTTLSeconds int    `env:"ACCESS_TOKEN_TTL_SECONDS" envDefault:"3600"`
	RefreshTokenTTLDays   int    `env:"REFRESH_TOKEN_TTL_DAYS" envDefault:"7"`
	CORSAllowedOrigins    string `env:"CORS_ALLOWED_ORIGINS" envDefault:"http://localhost:5173,http://127.0.0.1:5173"`
}

// ConnectPostgreSQL connects to PostgreSQL database and returns a connection pool
func ConnectPostgreSQL(ctx context.Context, cfg *Config) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	dbConnectString := cfg.PostgreSQLConnectionString()

	config, err := pgxpool.ParseConfig(dbConnectString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	// Configure connection pool settings
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	// Retry logic for establishing the connection pool
	const maxRetries = 3
	const retryDelay = 1 * time.Second

	var pool *pgxpool.Pool
	for i := 0; i < maxRetries; i++ {
		pool, err = pgxpool.ConnectConfig(ctx, config)
		if err == nil {
			// Perform a health check
			if pingErr := pool.Ping(ctx); pingErr == nil {
				log.Println("✅ Successfully connected to PostgreSQL")
				return pool, nil
			} else {
				err = pingErr
			}
		}

		log.Printf("Retrying database connection (%d/%d): %v\n", i+1, maxRetries, err)
		time.Sleep(retryDelay)
	}

	return nil, fmt.Errorf("unable to establish database connection after %d retries: %w", maxRetries, err)
}

// Load loads the configuration from environment variables and returns a Config struct.
func Load() (*Config, error) {
	cfg := &Config{}

	// Resolve .env locations: internal/config -> two levels up = backend module root (e.g. /app in Docker);
	// three levels up = monorepo root (skip if that resolves to filesystem root, e.g. /).
	_, currentFilePath, _, ok := runtime.Caller(0)
	if ok {
		configDir := filepath.Dir(currentFilePath)
		backendRoot := filepath.Clean(filepath.Join(configDir, "..", ".."))
		repoRoot := filepath.Clean(filepath.Join(configDir, "..", "..", ".."))

		tryDotEnv := func(path string) {
			if _, statErr := os.Stat(path); statErr != nil {
				return
			}
			if loadErr := godotenv.Load(path); loadErr == nil {
				log.Printf("✅ Loaded .env file: %s\n", path)
			}
		}

		// Monorepo root .env first (e.g. dnsc_microservice/.env), then backend/.env overrides.
		if repoRoot != "/" && repoRoot != backendRoot {
			tryDotEnv(filepath.Join(repoRoot, ".env"))
		}
		tryDotEnv(filepath.Join(backendRoot, ".env"))
	} else {
		_ = godotenv.Load(".env")
	}

	// Parse the configuration from environment variables
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("error loading configuration: %s", err)
	}

	// Basic validation for required app settings
	if cfg.AppSettings.ServerPort == "" {
		return nil, fmt.Errorf("invalid config: SERVER_PORT must not be empty")
	}

	// Single effective timezone: TZ (container/host) wins over APP_TIMEZONE so logs, process, and DB session match.
	effectiveTZ := strings.TrimSpace(os.Getenv("TZ"))
	if effectiveTZ == "" {
		effectiveTZ = strings.TrimSpace(cfg.AppSettings.AppTimezone)
	}
	if effectiveTZ == "" {
		effectiveTZ = "Europe/Bucharest"
	}
	cfg.AppSettings.AppTimezone = effectiveTZ
	_ = os.Setenv("TZ", effectiveTZ)

	if _, err := time.LoadLocation(cfg.AppSettings.AppTimezone); err != nil {
		return nil, fmt.Errorf("invalid timezone %q (TZ / APP_TIMEZONE): %w", cfg.AppSettings.AppTimezone, err)
	}

	log.Printf("✅ Loaded config - ServerPort: %s, Environment: %s, Timezone: %s (TZ + DB session + logs)\n",
		cfg.AppSettings.ServerPort, cfg.AppSettings.Environment, cfg.AppSettings.AppTimezone)

	return cfg, nil
}

// PostgreSQLConnectionString generates the PostgreSQL connection string
func (cfg *Config) PostgreSQLConnectionString() string {
	sslMode := cfg.DatabaseSettings.SSLMode
	if sslMode == "" {
		sslMode = "require" // Default to require SSL for security
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DatabaseSettings.User,
		cfg.DatabaseSettings.Password,
		cfg.DatabaseSettings.Host,
		cfg.DatabaseSettings.Port,
		cfg.DatabaseSettings.DbName,
		sslMode,
	)
}

// AccessTokenTTL returns access token lifetime.
func (cfg *Config) AccessTokenTTL() time.Duration {
	s := cfg.AccessTokenTTLSeconds
	if s <= 0 {
		s = 3600
	}
	return time.Duration(s) * time.Second
}

// RefreshTokenTTL returns refresh token lifetime.
func (cfg *Config) RefreshTokenTTL() time.Duration {
	d := cfg.RefreshTokenTTLDays
	if d <= 0 {
		d = 7
	}
	return time.Duration(d) * 24 * time.Hour
}

// CORSOriginsList splits CORS_ALLOWED_ORIGINS into trimmed origins.
func (cfg *Config) CORSOriginsList() []string {
	s := strings.TrimSpace(cfg.CORSAllowedOrigins)
	if s == "" {
		return []string{"http://localhost:5173"}
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"http://localhost:5173"}
	}
	return out
}
