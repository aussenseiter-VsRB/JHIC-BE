package database

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	type migration struct {
		version int
		name    string
	}

	var migrations []migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		parts := strings.SplitN(e.Name(), "_", 2)
		if len(parts) < 2 {
			continue
		}
		v, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		migrations = append(migrations, migration{version: v, name: e.Name()})
	}
	slices.SortFunc(migrations, func(a, b migration) int {
		return cmp.Compare(a.version, b.version)
	})

	for _, m := range migrations {
		var count int
		if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE version = $1", m.version).Scan(&count); err != nil {
			return fmt.Errorf("check migration %d: %w", m.version, err)
		}
		if count > 0 {
			continue
		}

		sql, err := os.ReadFile(filepath.Join(dir, m.name))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", m.name, err)
		}

		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("apply migration %s: %w", m.name, err)
		}

		if _, err := pool.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", m.version); err != nil {
			return fmt.Errorf("record migration %d: %w", m.version, err)
		}
	}
	return nil
}
