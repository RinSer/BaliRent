package models

import "time"

type Property struct {
	Id          int64     `json:"id"`
	OwnerId     int64     `json:"ownerId"`
	Address     string    `json:"address"`
	Photos      []string  `json:"photos"`
	Description string    `json:"description"`
	Price       int32     `json:"price"`
	Type        string    `json:"type"`
	RentOrSell  string    `json:"rentOrSell"`
	WhenCreated time.Time `json:"whenCreated"`
	WhenUpdated time.Time `json:"whenUpdated"`
}
