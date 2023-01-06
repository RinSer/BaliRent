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
	// Unauthenticated route
	e.GET("/", a.hello)
	// Restricted group
	r := e.Group("/restricted")
	// Configure middleware with the custom claims type
	config := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(jwtCustomClaims)
		},
		SigningKey: []byte(getJwtSecret()),
	}
	r.Use(echojwt.WithConfig(config))
	r.GET("", a.restricted)
	e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))
}

// e.GET("/", hello)
func (a *app) hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

// e.POST("/register", register)
func (a *app) register(c echo.Context) error {
	request := new(UserRequest)
	c.Bind(request)

	user, err := models.AddUser(a.db, &models.User{
		Phone:   request.Phone,
		Pwdhash: request.Password,
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

	user, err := models.GetUser(a.db, request.Phone, request.Password)
	// Throws unauthorized error
	if err != nil || user == nil {
		return echo.ErrUnauthorized
	}

	// Set custom claims
	claims := &jwtCustomClaims{
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

func (a *app) restricted(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*jwtCustomClaims)
	phone := claims.Phone
	return c.String(http.StatusOK, "Your number is "+phone+"!")
}
