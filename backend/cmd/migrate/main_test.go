package main

import (
	"testing"

	"github.com/jackc/pgx/v4"
)

func envFrom(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func TestMigrationDSNEscapesPassword(t *testing.T) {
	dsn, err := migrationDSN(envFrom(map[string]string{
		"POSTGRES_DB_HOST":      "db",
		"POSTGRES_DB_NAME":      "giobby",
		"MIGRATION_DB_USER":     "giobby",
		"MIGRATION_DB_PASSWORD": "a@b/c:d",
	}))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse DSN: %v", err)
	}
	if cfg.Password != "a@b/c:d" || cfg.Host != "db" || cfg.Port != 5432 || cfg.Database != "giobby" || cfg.User != "giobby" {
		t.Fatalf("unexpected config: user=%q host=%q port=%d db=%q", cfg.User, cfg.Host, cfg.Port, cfg.Database)
	}
}

func TestMigrationDSNRequiresVariables(t *testing.T) {
	for _, missing := range []string{"POSTGRES_DB_HOST", "POSTGRES_DB_NAME", "MIGRATION_DB_USER", "MIGRATION_DB_PASSWORD"} {
		values := map[string]string{
			"POSTGRES_DB_HOST":      "db",
			"POSTGRES_DB_NAME":      "giobby",
			"MIGRATION_DB_USER":     "giobby",
			"MIGRATION_DB_PASSWORD": "secret",
		}
		delete(values, missing)
		if _, err := migrationDSN(envFrom(values)); err == nil {
			t.Errorf("migrationDSN without %s: want error", missing)
		}
	}
}
