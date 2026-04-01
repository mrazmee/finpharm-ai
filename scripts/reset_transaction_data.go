package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	yes := flag.Bool("yes", false, "confirm reset transaction data")
	timeoutMs := flag.Int("timeout-ms", 10000, "database operation timeout in milliseconds")
	flag.Parse()

	if !*yes {
		fmt.Println("Reset dibatalkan.")
		fmt.Println("Jalankan dengan konfirmasi:")
		fmt.Println("  go run ./scripts/reset_transaction_data.go --yes")
		os.Exit(1)
	}

	cfg := loadConfig()
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeoutMs)*time.Millisecond)
	defer cancel()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal membuka koneksi database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "gagal konek ke database: %v\n", err)
		os.Exit(1)
	}

	beforeTxCount, err := countRows(ctx, db, "transactions")
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal count transactions sebelum reset: %v\n", err)
		os.Exit(1)
	}

	beforeItemCount, err := countRows(ctx, db, "transaction_items")
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal count transaction_items sebelum reset: %v\n", err)
		os.Exit(1)
	}

	sqlTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal begin transaction reset: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = sqlTx.Rollback()
	}()

	_, err = sqlTx.ExecContext(
		ctx,
		`TRUNCATE TABLE transaction_items, transactions RESTART IDENTITY CASCADE`,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal truncate tabel transaksi: %v\n", err)
		os.Exit(1)
	}

	if err := sqlTx.Commit(); err != nil {
		fmt.Fprintf(os.Stderr, "gagal commit reset: %v\n", err)
		os.Exit(1)
	}

	afterTxCount, err := countRows(ctx, db, "transactions")
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal count transactions setelah reset: %v\n", err)
		os.Exit(1)
	}

	afterItemCount, err := countRows(ctx, db, "transaction_items")
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal count transaction_items setelah reset: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== RESET TRANSACTION DATA SUCCESS ===")
	fmt.Printf("DB       : %s:%s / %s\n", cfg.DBHost, cfg.DBPort, cfg.DBName)
	fmt.Printf("Before   : transactions=%d, transaction_items=%d\n", beforeTxCount, beforeItemCount)
	fmt.Printf("After    : transactions=%d, transaction_items=%d\n", afterTxCount, afterItemCount)
	fmt.Println("Catatan  : reset ini hanya membersihkan data transaksi.")
	fmt.Println("           Stock medicines TIDAK di-reset oleh script ini.")
}

type config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

func loadConfig() config {
	return config{
		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBPort:     getEnv("DB_PORT", "55432"),
		DBUser:     getEnv("DB_USER", "finpharm"),
		DBPassword: getEnv("DB_PASSWORD", "finpharm"),
		DBName:     getEnv("DB_NAME", "transaction_db"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

func getEnv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func countRows(ctx context.Context, db *sql.DB, table string) (int, error) {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	err := db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}