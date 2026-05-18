package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"finpharm-ai/services/knowledge/internal/config"

	_ "github.com/lib/pq"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})).
		With("service", "knowledge-migrate")
	slog.SetDefault(logger)

	cfg := config.Load()
	if err := cfg.ValidateForMigrate(); err != nil {
		slog.Error("config_invalid", "error", err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		slog.Error("db_open_error", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		slog.Error("db_ping_error", "error", err)
		os.Exit(1)
	}

	files, err := collectUpSQLFiles(cfg.MigrationsDir)
	if err != nil {
		slog.Error("collect_migrations_error", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	for _, file := range files {
		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			slog.Error("read_migration_error", "file", file, "error", err)
			os.Exit(1)
		}
		if _, err := db.ExecContext(ctx, string(sqlBytes)); err != nil {
			slog.Error("apply_migration_error", "file", file, "error", err)
			os.Exit(1)
		}
		slog.Info("migration_applied", "file", filepath.Base(file))
	}

	slog.Info("knowledge_migrations_complete", "count", len(files))
}

func collectUpSQLFiles(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".up.sql") {
			files = append(files, filepath.Join(root, name))
		}
	}

	sort.Strings(files)
	return files, nil
}