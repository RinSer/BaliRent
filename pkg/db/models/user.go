package models

import "time"

type User struct {
	Id          int64     `json:"id"`
	Phone       string    `json:"phone"`
	Password    string    `json:"password"`
	WhenCreated time.Time `json:"whenCreated"`
	WhenUpdated time.Time `json:"whenUpdated"`
}
