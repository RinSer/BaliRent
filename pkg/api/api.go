package api

import (
	"bali-rent/pkg/auth"
	"bali-rent/pkg/db/models"
	"bali-rent/pkg/mail"
	"bali-rent/pkg/sms"
	"encoding/base64"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/golang-jwt/jwt/v4"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
)

type app struct {
	db *pg.DB
	redis *redis.Client
	mail *mail.Mail
	sms sms.SmsSender
}

func StartAPI(pgdb *pg.DB, redis *redis.Client,mail *mail.Mail, sms sms.SmsSender) {
	// create new app
	a := app{pgdb,redis,mail,sms}
	e := echo.New()
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	gr := e.Group("/api")
	// Register route
	gr.POST("/register", a.register)
	
	restore := gr.Group("/restore")
	restore.POST("/email/:email", a.restoreByEmail)
	restore.POST("/sms/:phone", a.restoreBySms)

	confirm := gr.Group("/confirm")
	confirm.POST("/email/:email/:code", a.confirmByEmail)
	confirm.POST("/sms/:phone/:code", a.confirmBySms)

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

	confirmation := models.NewConfirmation(user, a.redis, a.mail,a.sms)
	
	if user.Phone != "" {
		err = confirmation.SendSms(c.Request().Context())
	}  
	if user.Email != "" {
		err = confirmation.SendEmail(c.Request().Context())
	} 

	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	} 

	return c.JSON(http.StatusCreated, user)
}

func (a *app) restoreByEmail(c echo.Context) error {
	encodedParam := c.Param("email")

    decodedParam, err := base64.StdEncoding.DecodeString(encodedParam)
    if err != nil {
        return c.JSON(http.StatusBadRequest, err)
    }

	email := string(decodedParam)

	user, err := models.GetUserByEmail(a.db,  email)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	if user == nil {
		return c.JSON(http.StatusNotFound, "User not found")
	}

	confirmation := models.NewConfirmation(user, a.redis, a.mail,a.sms)

	err = confirmation.SendEmail(c.Request().Context())

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	} 

	return c.JSON(http.StatusOK, "Email sent")

}

func (a *app) restoreBySms(c echo.Context) error {
	encodedParam := c.Param("phone")

    decodedParam, err := base64.StdEncoding.DecodeString(encodedParam)
    if err != nil {
        return c.JSON(http.StatusBadRequest, err)
    }

	phone := string(decodedParam)

	user, err := models.GetUserByPhone(a.db,  phone)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	if user == nil {
		return c.JSON(http.StatusNotFound, "User not found")
	}

	confirmation := models.NewConfirmation(user, a.redis, a.mail,a.sms)

	err = confirmation.SendSms(c.Request().Context())

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	} 

	return c.JSON(http.StatusOK, "Sms sent")
}

func (a *app) confirmByEmail(c echo.Context) error {
	encodedParam := c.Param("email")

    decodedParam, err := base64.StdEncoding.DecodeString(encodedParam)
    if err != nil {
        return c.JSON(http.StatusBadRequest, err)
    }

	email := string(decodedParam)

	code := c.Param("code")

	user, err := models.GetUserByEmail(a.db,  email)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	if user == nil {
		return c.JSON(http.StatusNotFound, "User not found")
	}

	confirmation := models.NewConfirmation(user, a.redis, a.mail,a.sms)

	err = confirmation.CheckEmail(c.Request().Context(),code)

	if err != nil {
		return c.JSON(http.StatusForbidden, err)
	}

	t, err := user.GetToken()

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": *t,
	})
}

func (a *app) confirmBySms(c echo.Context) error {

	encodedParam := c.Param("phone")

    decodedParam, err := base64.StdEncoding.DecodeString(encodedParam)
    if err != nil {
        return c.JSON(http.StatusBadRequest, err)
    }

	phone := string(decodedParam)

	code := c.Param("code")

	user, err := models.GetUserByPhone(a.db,  phone)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	if user == nil {
		return c.JSON(http.StatusNotFound, "User not found")
	}

	confirmation := models.NewConfirmation(user, a.redis, a.mail,a.sms)

	err = confirmation.CheckSms(c.Request().Context(),code)

	if err != nil {
		return c.JSON(http.StatusForbidden, err)
	}

	t, err := user.GetToken()

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": *t,
	})
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

	t, err := user.GetToken()

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": *t,
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
