package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./database/receipts.db")
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	createTables()
}

func createTables() {
	createReceiptFileTable := `
	CREATE TABLE IF NOT EXISTS receipt_file (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_name TEXT,
		file_path TEXT,
		is_valid BOOLEAN DEFAULT FALSE,
		invalid_reason TEXT,
		is_processed BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	createReceiptTable := `
	CREATE TABLE IF NOT EXISTS receipt (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		purchased_at TEXT,
		merchant_name TEXT,
		total_amount REAL,
		file_path TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := DB.Exec(createReceiptFileTable); err != nil {
		log.Fatalf("❌ Failed to create receipt_file table: %v", err)
	}

	if _, err := DB.Exec(createReceiptTable); err != nil {
		log.Fatalf("❌ Failed to create receipt table: %v", err)
	}

	log.Println("✅ Database and tables initialized.")
}
