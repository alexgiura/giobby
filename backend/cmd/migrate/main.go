// Command migrate applies the pending database migrations with the owner role.
// It runs before the API on every deploy; the API itself never changes the schema.
package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"dnsc_microservice/internal/db"

	"github.com/jackc/pgx/v4"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Migration failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Database schema is current.")
}

func run() error {
	dsn, err := migrationDSN(os.Getenv)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	connectCtx, connectCancel := context.WithTimeout(ctx, 30*time.Second)
	defer connectCancel()
	conn, err := pgx.Connect(connectCtx, dsn)
	if err != nil {
		// pgx errors can include the DSN host and user, never the password.
		return fmt.Errorf("connect to database: %w", err)
	}
	defer conn.Close(context.Background())

	applied, err := db.Migrate(ctx, conn, strings.TrimSpace(os.Getenv("MIGRATION_APP_ROLE")))
	if err != nil {
		return err
	}
	fmt.Printf("Applied %d migration(s).\n", applied)
	return nil
}

// migrationDSN builds the owner connection string; url.UserPassword escapes
// passwords containing characters such as '@', '/' or ':'.
func migrationDSN(getenv func(string) string) (string, error) {
	values := map[string]string{}
	for _, key := range []string{"POSTGRES_DB_HOST", "POSTGRES_DB_NAME", "MIGRATION_DB_USER", "MIGRATION_DB_PASSWORD"} {
		values[key] = strings.TrimSpace(getenv(key))
		if values[key] == "" {
			return "", fmt.Errorf("%s is required", key)
		}
	}
	port := strings.TrimSpace(getenv("POSTGRES_DB_PORT"))
	if port == "" {
		port = "5432"
	}
	sslMode := strings.TrimSpace(getenv("POSTGRES_DB_SSLMODE"))
	if sslMode == "" {
		sslMode = "disable"
	}
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(values["MIGRATION_DB_USER"], values["MIGRATION_DB_PASSWORD"]),
		Host:     values["POSTGRES_DB_HOST"] + ":" + port,
		Path:     "/" + values["POSTGRES_DB_NAME"],
		RawQuery: url.Values{"sslmode": {sslMode}}.Encode(),
	}
	return u.String(), nil
}
