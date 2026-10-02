package db

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
)

// Integration tests: they need a PostgreSQL superuser DSN in GIOBBY_TEST_DATABASE_URL
// and are skipped without it. Each test recreates a dedicated database owned by a
// non-superuser role, as in production.
const (
	testDatabase  = "giobby_migrate_test"
	testOwnerRole = "giobby_test_owner"
	testAppRole   = "giobby_test_app"
	testPassword  = "test-only-password"
)

func adminConfig(t *testing.T) *pgx.ConnConfig {
	t.Helper()
	dsn := os.Getenv("GIOBBY_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("GIOBBY_TEST_DATABASE_URL not set")
	}
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse GIOBBY_TEST_DATABASE_URL: %v", err)
	}
	return cfg
}

// freshDatabase recreates the test database and roles and returns the config to
// connect to it as the given role.
func freshDatabase(t *testing.T) func(role string) *pgx.ConnConfig {
	t.Helper()
	admin := adminConfig(t)
	ctx := context.Background()
	conn, err := pgx.ConnectConfig(ctx, admin)
	if err != nil {
		t.Fatalf("connect as admin: %v", err)
	}
	defer conn.Close(ctx)
	for _, stmt := range []string{
		"DROP DATABASE IF EXISTS " + testDatabase + " WITH (FORCE)",
		"DROP ROLE IF EXISTS " + testAppRole,
		"DROP ROLE IF EXISTS " + testOwnerRole,
		"CREATE ROLE " + testOwnerRole + " LOGIN PASSWORD '" + testPassword + "'",
		"CREATE ROLE " + testAppRole + " LOGIN PASSWORD '" + testPassword + "'",
		"CREATE DATABASE " + testDatabase + " OWNER " + testOwnerRole,
	} {
		if _, err := conn.Exec(ctx, stmt); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}
	return func(role string) *pgx.ConnConfig {
		cfg := admin.Copy()
		cfg.User = role
		cfg.Password = testPassword
		cfg.Database = testDatabase
		return cfg
	}
}

func connectAs(t *testing.T, cfg *pgx.ConnConfig) *pgx.Conn {
	t.Helper()
	conn, err := pgx.ConnectConfig(context.Background(), cfg)
	if err != nil {
		t.Fatalf("connect as %s: %v", cfg.User, err)
	}
	t.Cleanup(func() { conn.Close(context.Background()) })
	return conn
}

func migrationFileCount(t *testing.T) int {
	t.Helper()
	files, err := fs.Glob(migrationFiles, "migrations/*.sql")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no embedded migrations")
	}
	return len(files)
}

func countRows(t *testing.T, conn *pgx.Conn) int {
	t.Helper()
	var n int
	if err := conn.QueryRow(context.Background(), "SELECT count(*) FROM schema_migrations").Scan(&n); err != nil {
		t.Fatalf("count schema_migrations: %v", err)
	}
	return n
}

func TestMigrateAppliesAllThenNothing(t *testing.T) {
	as := freshDatabase(t)
	conn := connectAs(t, as(testOwnerRole))
	ctx := context.Background()
	want := migrationFileCount(t)

	applied, err := Migrate(ctx, conn, "")
	if err != nil {
		t.Fatalf("first Migrate: %v", err)
	}
	if applied != want || countRows(t, conn) != want {
		t.Fatalf("first Migrate applied %d, recorded %d, want %d", applied, countRows(t, conn), want)
	}

	applied, err = Migrate(ctx, conn, "")
	if err != nil || applied != 0 {
		t.Fatalf("second Migrate = %d, %v; want 0, nil", applied, err)
	}
}

func TestMigrateRejectsChangedChecksum(t *testing.T) {
	as := freshDatabase(t)
	conn := connectAs(t, as(testOwnerRole))
	ctx := context.Background()
	if _, err := Migrate(ctx, conn, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(ctx, "UPDATE schema_migrations SET checksum='x' WHERE version=10"); err != nil {
		t.Fatal(err)
	}
	before := countRows(t, conn)

	_, err := Migrate(ctx, conn, "")
	if !errors.Is(err, ErrChecksumMismatch) {
		t.Fatalf("Migrate error = %v, want ErrChecksumMismatch", err)
	}
	if countRows(t, conn) != before {
		t.Fatal("schema_migrations changed after a rejected run")
	}
}

func TestMigrateRejectsUnknownVersion(t *testing.T) {
	as := freshDatabase(t)
	conn := connectAs(t, as(testOwnerRole))
	ctx := context.Background()
	if _, err := Migrate(ctx, conn, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(ctx, "INSERT INTO schema_migrations(version, name, checksum) VALUES (99999, 'gone.sql', 'x')"); err != nil {
		t.Fatal(err)
	}

	if _, err := Migrate(ctx, conn, ""); !errors.Is(err, ErrUnknownVersion) {
		t.Fatalf("Migrate error = %v, want ErrUnknownVersion", err)
	}
}

func TestMigrateConcurrent(t *testing.T) {
	as := freshDatabase(t)
	conns := []*pgx.Conn{connectAs(t, as(testOwnerRole)), connectAs(t, as(testOwnerRole))}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var wg sync.WaitGroup
	applied := make([]int, len(conns))
	errs := make([]error, len(conns))
	for i, c := range conns {
		wg.Add(1)
		go func(i int, c *pgx.Conn) {
			defer wg.Done()
			applied[i], errs[i] = Migrate(ctx, c, "")
		}(i, c)
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			t.Fatalf("concurrent Migrate: %v", err)
		}
	}
	if total := applied[0] + applied[1]; total != migrationFileCount(t) {
		t.Fatalf("concurrent runs applied %d in total, want %d", total, migrationFileCount(t))
	}
}

func TestAppRoleCanWriteDataNotSchema(t *testing.T) {
	as := freshDatabase(t)
	ctx := context.Background()
	if _, err := Migrate(ctx, connectAs(t, as(testOwnerRole)), testAppRole); err != nil {
		t.Fatal(err)
	}
	app := connectAs(t, as(testAppRole))

	if _, err := app.Exec(ctx, "INSERT INTO users(username, password_hash) VALUES ('app-check', 'x')"); err != nil {
		t.Fatalf("app role INSERT into users: %v", err)
	}
	if _, err := app.Exec(ctx, "CREATE TABLE app_ddl_check(x int)"); err == nil {
		t.Fatal("app role could CREATE TABLE")
	}
	if _, err := app.Exec(ctx, "INSERT INTO schema_migrations(version, name, checksum) VALUES (1, 'x', 'x')"); err == nil {
		t.Fatal("app role could write schema_migrations")
	}
	var n int
	if err := app.QueryRow(ctx, "SELECT count(*) FROM schema_migrations").Scan(&n); err != nil {
		t.Fatalf("app role cannot read schema_migrations: %v", err)
	}
}
