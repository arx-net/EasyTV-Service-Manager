package main

import (
	"database/sql"
	"log"

	"gitlab.arx.net/easytv/sm/db"
)

func create_table(name, query string, db *sql.DB) {
	log.Printf("Creating table %s...\n", name)
	_, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Completed")
}

func main() {
	pool, err := db.Open()

	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected successfuly")

	init_db(pool)
}
