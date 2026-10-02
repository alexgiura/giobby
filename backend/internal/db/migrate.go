package db

import (
	"context"
	"crypto/sha256"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// migrationLockID serializes concurrent runs (two deploys at once) on the same database.
const migrationLockID = 7303101

var (
	// ErrChecksumMismatch means an already applied migration file was edited.
	ErrChecksumMismatch = errors.New("migration checksum mismatch")
	// ErrUnknownVersion means the database has a migration that no longer exists in the code.
	ErrUnknownVersion = errors.New("applied migration missing from code")
)

type migration struct {
	version  int
	name     string
	sql      string
	checksum string
}

func loadMigrations() ([]migration, error) {
	files, err := fs.Glob(migrationFiles, "migrations/*.sql")
	if err != nil {
		return nil, err
	}
	migrations := make([]migration, 0, len(files))
	seen := map[int]string{}
	for _, file := range files {
		name := path.Base(file)
		prefix, _, ok := strings.Cut(name, "_")
		version, err := strconv.Atoi(prefix)
		if !ok || err != nil {
			return nil, fmt.Errorf("migration %s: name must start with a numeric version and '_'", name)
		}
		if other, dup := seen[version]; dup {
			return nil, fmt.Errorf("migrations %s and %s share version %d", other, name, version)
		}
		seen[version] = name
		body, err := migrationFiles.ReadFile(file)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, migration{
			version:  version,
			name:     name,
			sql:      string(body),
			checksum: fmt.Sprintf("%x", sha256.Sum256(body)),
		})
	}
	sort.Slice(migrations, func(i, j int) bool { return migrations[i].version < migrations[j].version })
	return migrations, nil
}

// Migrate applies the pending embedded migrations in one transaction and returns how
// many it applied. If appRole is not empty, it (re)grants that role read/write access
// to the data, without DDL and without write access to schema_migrations.
func Migrate(ctx context.Context, conn *pgx.Conn, appRole string) (int, error) {
	migrations, err := loadMigrations()
	if err != nil {
		return 0, err
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin migration transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", migrationLockID); err != nil {
		return 0, fmt.Errorf("acquire migration lock: %w", err)
	}
	if _, err := tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS public.schema_migrations (
		version    integer PRIMARY KEY,
		name       text NOT NULL,
		checksum   text NOT NULL,
		applied_at timestamptz NOT NULL DEFAULT now()
	)`); err != nil {
		return 0, fmt.Errorf("create schema_migrations: %w", err)
	}

	applied, err := appliedChecksums(ctx, tx)
	if err != nil {
		return 0, err
	}
	known := map[int]bool{}
	for _, m := range migrations {
		known[m.version] = true
		if sum, ok := applied[m.version]; ok && sum != m.checksum {
			return 0, fmt.Errorf("%w: %s", ErrChecksumMismatch, m.name)
		}
	}
	for version := range applied {
		if !known[version] {
			return 0, fmt.Errorf("%w: version %d", ErrUnknownVersion, version)
		}
	}

	count := 0
	for _, m := range migrations {
		if _, ok := applied[m.version]; ok {
			continue
		}
		if _, err := tx.Exec(ctx, m.sql); err != nil {
			return 0, fmt.Errorf("apply %s: %w", m.name, err)
		}
		if _, err := tx.Exec(ctx, "INSERT INTO public.schema_migrations(version, name, checksum) VALUES ($1, $2, $3)",
			m.version, m.name, m.checksum); err != nil {
			return 0, fmt.Errorf("record %s: %w", m.name, err)
		}
		count++
	}

	if appRole != "" {
		if err := grantAppRole(ctx, tx, appRole); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit migrations: %w", err)
	}
	return count, nil
}

func appliedChecksums(ctx context.Context, tx pgx.Tx) (map[int]string, error) {
	rows, err := tx.Query(ctx, "SELECT version, checksum FROM public.schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("read schema_migrations: %w", err)
	}
	defer rows.Close()
	applied := map[int]string{}
	for rows.Next() {
		var version int
		var sum string
		if err := rows.Scan(&version, &sum); err != nil {
			return nil, fmt.Errorf("read schema_migrations: %w", err)
		}
		applied[version] = sum
	}
	return applied, rows.Err()
}

func grantAppRole(ctx context.Context, tx pgx.Tx, appRole string) error {
	role := pgx.Identifier{appRole}.Sanitize()
	for _, stmt := range []string{
		"GRANT USAGE ON SCHEMA public TO " + role,
		"GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO " + role,
		"GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO " + role,
		"REVOKE INSERT, UPDATE, DELETE, TRUNCATE ON public.schema_migrations FROM " + role,
	} {
		if _, err := tx.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("grant %s: %w", appRole, err)
		}
	}
	return nil
}
