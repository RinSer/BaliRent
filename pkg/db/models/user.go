package models

import (
	"time"

	"github.com/go-pg/pg/v10"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id          int64     `json:"id"`
	Phone       string    `json:"phone"`
	Pwdhash     string    `json:"pwdhash"`
	WhenCreated time.Time `json:"whenCreated"`
	WhenUpdated time.Time `json:"whenUpdated"`
}

func AddUser(db *pg.DB, req *User) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Pwdhash), 5)
	if err != nil {
		return nil, err
	}

	req.Pwdhash = string(hash)
	_, err = db.Model(req).Insert()
	if err != nil {
		return nil, err
	}

	user := &User{}

	err = db.Model(user).
		Where("id = ?", req.Id).
		Select()

	return user, err
}

func GetUser(db *pg.DB, phone string, password string) (*User, error) {
	user := &User{}

	err := db.Model(user).
		Where("phone = ?", phone).
		Select()
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Pwdhash), []byte(password))
	if err != nil {
		return nil, err
	}

	return user, nil
}
