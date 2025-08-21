package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Println("DATABASE_URL not set")
		os.Exit(2)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Println("failed to open database:", err)
		os.Exit(2)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, donor_email, amount, created_at FROM donations WHERE amount <= 0 ORDER BY created_at DESC")
	if err != nil {
		fmt.Println("query failed:", err)
		os.Exit(2)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id string
		var email string
		var amount float64
		var createdAt string
		if err := rows.Scan(&id, &email, &amount, &createdAt); err != nil {
			fmt.Println("scan error:", err)
			continue
		}
		fmt.Printf("%s\t%s\t%.2f\t%s\n", id, email, amount, createdAt)
		count++
	}

	fmt.Printf("Found %d zero-or-negative donations\n", count)
}
