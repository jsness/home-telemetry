package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.LUTC)

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		logger.Fatal("DATABASE_URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		logger.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	if err := ensureMigrationsTable(ctx, pool); err != nil {
		logger.Fatalf("ensure migrations: %v", err)
	}

	migrations, err := readMigrations("migrations")
	if err != nil {
		logger.Fatalf("read migrations: %v", err)
	}

	applied, err := appliedMigrations(ctx, pool)
	if err != nil {
		logger.Fatalf("applied migrations: %v", err)
	}

	for _, m := range migrations {
		if applied[m.Version] {
			continue
		}
		logger.Printf("applying %s", m.Name)
		if err := applyMigration(ctx, pool, m); err != nil {
			logger.Fatalf("apply %s: %v", m.Name, err)
		}
	}

	logger.Println("migrations complete")
}

type migration struct {
	Name    string
	Version string
	SQL     string
}

func readMigrations(dir string) ([]migration, error) {
	entries := []migration{}
	walk := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".sql") {
			return nil
		}
		bytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		name := d.Name()
		version := strings.SplitN(name, "_", 2)[0]
		entries = append(entries, migration{
			Name:    name,
			Version: version,
			SQL:     string(bytes),
		})
		return nil
	}

	if err := filepath.WalkDir(dir, walk); err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	return entries, nil
}

func ensureMigrationsTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
  version TEXT PRIMARY KEY,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`)
	return err
}

func appliedMigrations(ctx context.Context, pool *pgxpool.Pool) (map[string]bool, error) {
	rows, err := pool.Query(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]bool{}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out[v] = true
	}
	return out, rows.Err()
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, m migration) error {
	if strings.TrimSpace(m.SQL) == "" {
		return fmt.Errorf("empty migration: %s", m.Name)
	}

	_, err := pool.Exec(ctx, m.SQL)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", m.Version)
	return err
}
