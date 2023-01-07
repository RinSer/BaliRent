package models

import (
	"time"

	"github.com/go-pg/pg/v10"
)

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

func AddProperty(db *pg.DB, req *Property) (*Property, error) {
	property := &Property{}

	_, err := db.Model(property).Insert()

	return property, err
}

func UpdateProperty(db *pg.DB, req *Property) (*Property, error) {
	property := &Property{}

	err := db.Model(property).
		Where("id = ?", req.Id).
		Select()

	return property, err
}

func GetProperty(db *pg.DB, propertyId int64) (*Property, error) {
	property := &Property{}

	err := db.Model(property).
		Where("id = ?", propertyId).
		Select()

	return property, err
}

func GetProperties(db *pg.DB, ownerId int64) ([]*Property, error) {
	properties := []*Property{}

	err := db.Model(properties).
		Where("owner_id = ?", ownerId).
		Select()

	return properties, err
}
