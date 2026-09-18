package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/TU-Software/meetings/2026-09-23/db"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: migrate <up|down|version>")
		os.Exit(1)
	}
	cmd := os.Args[1]

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		slog.Error("DATABASE_URL is not set")
		os.Exit(1)
	}

	// The pgx/v5 migrate driver registers under "pgx5://", not "postgres://".
	databaseURL = strings.Replace(databaseURL, "postgres://", "pgx5://", 1)
	databaseURL = strings.Replace(databaseURL, "postgresql://", "pgx5://", 1)

	src, err := iofs.New(db.Migrations, "migrations")
	if err != nil {
		slog.Error("failed to load migration sources", "error", err)
		os.Exit(1)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, databaseURL)
	if err != nil {
		slog.Error("failed to create migrator", "error", err)
		os.Exit(1)
	}
	defer m.Close()

	switch cmd {
	case "up":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			slog.Error("migration up failed", "error", err)
			os.Exit(1)
		}
		slog.Info("migrations applied")

	case "down":
		if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			slog.Error("migration down failed", "error", err)
			os.Exit(1)
		}
		slog.Info("migration rolled back")

	case "version":
		version, dirty, err := m.Version()
		if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
			slog.Error("failed to get version", "error", err)
			os.Exit(1)
		}
		fmt.Printf("version=%d dirty=%v\n", version, dirty)

	default:
		fmt.Fprintf(os.Stderr, "unknown command %q — expected up, down, or version\n", cmd)
		os.Exit(1)
	}
}
