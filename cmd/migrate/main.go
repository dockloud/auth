package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"

	"github.com/dockloud/auth/internal/infrastructure/database/postgres/migrations"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: migrate <up|down|status>")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}

	goose.SetBaseFS(migrations.FS)

	switch os.Args[1] {
	case "up":
		err = goose.Up(db, ".")
	case "down":
		err = goose.Down(db, ".")
	case "status":
		err = goose.Status(db, ".")
	default:
		log.Fatalf("unknown command: %s", os.Args[1])
	}

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Migration command completed")
}
