package api

import (
	"bali-rent/pkg/db/models"
	"net/http"
	"os"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/golang-jwt/jwt/v4"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type app struct {
	db *pg.DB
}

// jwtCustomClaims are custom claims extending default ones.
// See https://github.com/golang-jwt/jwt for more examples
type jwtCustomClaims struct {
	Id    int64  `json:"id"`
	Phone string `json:"phone"`
	Admin bool   `json:"admin"`
	jwt.RegisteredClaims
}

func getJwtSecret() string {
	return os.Getenv("JWT_SECRET")
}

func StartAPI(pgdb *pg.DB) {
	// create new app
	a := app{pgdb}
	e := echo.New()
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	// Register route
	e.POST("/register", a.register)
	// Login route
	e.POST("/login", a.login)
	// Restricted group
	r := e.Group("/user")
	// Configure middleware with the custom claims type
	config := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(jwtCustomClaims)
		},
		SigningKey: []byte(getJwtSecret()),
	}
	r.Use(echojwt.WithConfig(config))
	r.POST("/update", a.updateUser)
	e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))
}

// e.POST("/register", register)
func (a *app) register(c echo.Context) error {
	request := new(UserRequest)
	c.Bind(request)

	user, err := models.AddUser(a.db, &models.User{
		Phone:       request.Phone,
		Email:       request.Email,
		Pwdhash:     request.Password,
		WhenCreated: time.Now(),
	})

	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	return c.JSON(http.StatusCreated, user)
}

// e.POST("/login", login)
func (a *app) login(c echo.Context) error {
	request := new(UserRequest)
	c.Bind(request)

	user, err := models.GetUser(a.db, 0, request.Phone, request.Password)
	// Throws unauthorized error
	if err != nil || user == nil {
		return echo.ErrUnauthorized
	}

	// Set custom claims
	claims := &jwtCustomClaims{
		user.Id,
		user.Phone,
		false,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(getJwtSecret()))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": t,
	})
}

func (a *app) updateUser(c echo.Context) error {
	authData := c.Get("user").(*jwt.Token)
	claims := authData.Claims.(*jwtCustomClaims)

	request := new(UserRequest)
	c.Bind(request)

	if request.Id == nil || *request.Id != claims.Id {
		return echo.ErrUnauthorized
	}

	user, err := models.GetUser(a.db, *request.Id, "", request.Password)
	if err != nil || user == nil {
		return echo.ErrUnauthorized
	}

	user.Phone = request.Phone
	user.Email = request.Email
	user.Pwdhash = request.Password
	user.WhenUpdated = time.Now()

	user, err = models.UpdateUser(a.db, user)
	if err != nil || user == nil {
		return echo.ErrUnauthorized
	}

	return c.JSON(http.StatusOK, user)
}
