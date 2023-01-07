package models

import "time"

type PropertyType int8

const (
	Room PropertyType = iota + 1
	Flat
	House
)

type Property struct {
	Id           int64        `json:"id"`
	OwnerId      int64        `json:"ownerId"`
	Latitude     float64      `json:"latitude"`
	Longitude    float64      `json:"longitude"`
	Photos       []string     `json:"photos"`
	Description  string       `json:"description"`
	Price        int32        `json:"price"`
	PropertyType PropertyType `json:"propertyType"`
	ForSell      bool         `json:"forSell"`
	IsDeleted    bool         `json:"isDeleted"`
	WhenCreated  time.Time    `json:"whenCreated"`
	WhenUpdated  time.Time    `json:"whenUpdated"`
}
