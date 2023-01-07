package models

import "time"

type Booking struct {
	Id          int64     `json:"id"`
	UserId      int64     `json:"userId"`
	PropertyId  int64     `json:"propertyId"`
	StartDate   time.Time `json:"startDate"`
	EndDate     time.Time `json:"endDate"`
	IsDeleted   bool      `json:"isDeleted"`
	WhenCreated time.Time `json:"whenCreated"`
	WhenUpdated time.Time `json:"whenUpdated"`
}
