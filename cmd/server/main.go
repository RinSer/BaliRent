package main

import (
	"bali-rent/pkg/api"
	"bali-rent/pkg/db"
	"bali-rent/pkg/mail"
	"bali-rent/pkg/redis"
	"bali-rent/pkg/sms"
	"log"
)

func main() {
	// start the db
	pgdb, err := db.StartDB()
	// start the redis
	redis, err := redis.Connect()
	// start the mail
	mail := mail.NewMail()
	// start the mail
	sms := sms.NewSmsMock(redis)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// run http server
	api.StartAPI(pgdb,redis,mail,sms)
}
