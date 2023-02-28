package api

import (
	"bali-rent/pkg/auth"
	"bali-rent/pkg/db/models"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/redis/go-redis/v9"
	"github.com/golang-jwt/jwt/v4"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type app struct {
	db *pg.DB
	redis *redis.Client
}

func StartAPI(pgdb *pg.DB, redis *redis.Client) {
	// create new app
	a := app{pgdb,redis}
	e := echo.New()
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	gr := e.Group("/api")
	// Register route
	gr.POST("/register", a.register)
	// Login route
	gr.POST("/login", a.login)
	// Configure middleware with the custom claims type
	config := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(auth.JwtCustomClaims)
		},
		SigningKey: []byte(auth.GetJwtSecret()),
	}
	// Restricted group
	r := gr.Group("/user", echojwt.WithConfig(config))
	r.POST("/update", a.updateUser)
	// Property REST
	p := r.Group("/property")
	p.GET("", a.getProperties)
	p.POST("", a.createProperty)
	p.PUT("", a.updateProperty)
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
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": t,
	})
}

func (a *app) updateUser(c echo.Context) error {
	authData := c.Get("user").(*jwt.Token)
	claims := authData.Claims.(*auth.JwtCustomClaims)

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

func (a *app) getProperties(c echo.Context) error {
	propertyId := c.QueryParam("id")
	if propertyId != "" {
		id, err := strconv.ParseInt(propertyId, 10, 64)
		if err != nil {
			return echo.ErrBadRequest
		}
		property, err := models.GetProperty(a.db, id)
		if err != nil {
			return echo.ErrBadRequest
		}
		return c.JSON(http.StatusOK, property)
	} else {
		ownerId := c.QueryParam("ownerId")
		if ownerId != "" {
			id, err := strconv.ParseInt(ownerId, 10, 64)
			if err != nil {
				return echo.ErrBadRequest
			}
			properties, err := models.GetOwnersProperties(a.db, id)
			if err != nil {
				return echo.ErrBadRequest
			}
			return c.JSON(http.StatusOK, properties)
		} else {
			fromStr := c.QueryParam("from")
			countStr := c.QueryParam("count")
			if fromStr != "" && countStr != "" {
				from, err := strconv.Atoi(fromStr)
				if err != nil {
					return echo.ErrBadRequest
				}
				count, err := strconv.Atoi(countStr)
				if err != nil {
					return echo.ErrBadRequest
				}
				properties, err := models.GetPagedProperties(a.db, from, count)
				if err != nil {
					return echo.ErrBadRequest
				}
				return c.JSON(http.StatusOK, properties)
			} else {
				latStartStr := c.QueryParam("latStart")
				latEndStr := c.QueryParam("latEnd")
				longStartStr := c.QueryParam("longStart")
				longEndStr := c.QueryParam("longEnd")
				if latStartStr != "" && latEndStr != "" &&
					longStartStr != "" && longEndStr != "" {
					latStart, err := strconv.ParseFloat(latStartStr, 64)
					if err != nil {
						return echo.ErrBadRequest
					}
					latEnd, err := strconv.ParseFloat(latEndStr, 64)
					if err != nil {
						return echo.ErrBadRequest
					}
					longStart, err := strconv.ParseFloat(longStartStr, 64)
					if err != nil {
						return echo.ErrBadRequest
					}
					longEnd, err := strconv.ParseFloat(longEndStr, 64)
					if err != nil {
						return echo.ErrBadRequest
					}
					properties, err := models.GetMapProperties(
						a.db, latStart, latEnd, longStart, longEnd)
					if err != nil {
						return echo.ErrBadRequest
					}
					return c.JSON(http.StatusOK, properties)
				} else {
					return echo.ErrBadRequest
				}
			}
		}
	}
}

func (a *app) createProperty(c echo.Context) error {
	authData := c.Get("user").(*jwt.Token)
	claims := authData.Claims.(*auth.JwtCustomClaims)

	request := new(models.Property)
	c.Bind(request)

	if request.OwnerId != claims.Id {
		return echo.ErrUnauthorized
	}

	property, err := models.AddProperty(a.db, request)
	if err != nil || property == nil {
		return echo.ErrBadRequest
	}

	return c.JSON(http.StatusCreated, property)
}

func (a *app) updateProperty(c echo.Context) error {
	authData := c.Get("user").(*jwt.Token)
	claims := authData.Claims.(*auth.JwtCustomClaims)

	request := new(models.Property)
	c.Bind(request)

	if request.OwnerId != claims.Id {
		return echo.ErrUnauthorized
	}

	property, err := models.UpdateProperty(a.db, request)
	if err != nil || property == nil {
		return echo.ErrBadRequest
	}

	return c.JSON(http.StatusOK, property)
}
