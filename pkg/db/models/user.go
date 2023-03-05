package models

import (
	"bali-rent/pkg/auth"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id          int64     `json:"id"`
	Phone       string    `json:"phone"`
	Email       string    `json:"email"`
	Pwdhash     string    `json:"-"`
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


func UpdateUser(db *pg.DB, req *User) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Pwdhash), 5)
	if err != nil {
		return nil, err
	}

	req.Pwdhash = string(hash)
	_, err = db.Model(req).WherePK().Update()
	if err != nil {
		return nil, err
	}

	user := &User{}

	err = db.Model(user).
		Where("id = ?", req.Id).
		Select()

	return user, err
}

func GetUser(db *pg.DB, id int64, phone string, password string) (*User, error) {
	user := &User{}

	query := db.Model(user)

	if phone == "" {
		query = query.Where("id = ?", id)
	} else {
		query = query.Where("phone = ?", phone)
	}

	err := query.Select()
	if err != nil {
		return nil, err
	}

	if phone != "" {
		err = bcrypt.CompareHashAndPassword([]byte(user.Pwdhash), []byte(password))
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

func GetUserByEmail(db *pg.DB,email string) (*User, error) {
	user := &User{}

	err := db.Model(user).Where("email = ?", email).Select()
	
	if err != nil {
		return nil, err
	}

	return user, nil
}

func GetUserByPhone(db *pg.DB,phone string) (*User, error) {
	user := &User{}

	err := db.Model(user).Where("phone = ?", phone).Select()
	
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (user *User) GetToken() (*string, error) {
	// Set custom claims
	claims := &auth.JwtCustomClaims{
		Id:    user.Id,
		Phone: user.Phone,
		Admin: false,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(auth.GetJwtSecret()))
	if err != nil {
		return nil,err
	}

	return &t, nil
}