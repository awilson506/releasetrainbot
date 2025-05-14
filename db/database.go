package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/awilson506/releasetrain/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
)

var DB *sql.DB

func Init() {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true",
		config.DBUser,
		config.DBPassword,
		config.DBHost,
		config.DBPort,
		config.DBName,
	)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to open DB:", err)
	}

	if err := DB.Ping(); err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	if err := goose.SetDialect("mysql"); err != nil {
		panic(err)
	}

	// Run migrations
	if err := goose.Up(DB, "migrations"); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}
