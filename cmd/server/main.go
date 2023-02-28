package main

import (
	"bali-rent/pkg/api"
	"bali-rent/pkg/db"
	"bali-rent/pkg/redis"
	"log"
)

func main() {
	// start the db
	pgdb, err := db.StartDB()
	// start the redis
	redis, err := redis.Connect()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	// run http server
	api.StartAPI(pgdb,redis)
}
