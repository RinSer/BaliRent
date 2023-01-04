package main

import (
	"bali-rent/pkg/api"
	"bali-rent/pkg/db"
	"log"
)

func main() {
	// start the db
	pgdb, err := db.StartDB()
	if err != nil {
		log.Fatalf("error: %v", err)

	}
	// run http server
	api.StartAPI(pgdb)
}
