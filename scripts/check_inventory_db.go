package main

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	connString := "postgres://finpharm:finpharm@127.0.0.1:55432/inventory_db?sslmode=disable"
	fmt.Println("trying:", connString)

	db, err := sqlx.Open("postgres", connString)
	if err != nil {
		log.Fatalf("open error: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping error: %v", err)
	}

	fmt.Println("db ping ok")
}