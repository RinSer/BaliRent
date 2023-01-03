package api

import (
	"net/http"
	"os"

	"github.com/go-pg/pg/v10"
	"github.com/labstack/echo/v4"
)

type app struct {
	db *pg.DB
}

func StartAPI(pgdb *pg.DB) {
	// create new app
	a := app{pgdb}
	e := echo.New()
	e.GET("/", a.hello)
	e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))
}

// e.GET("/", hello)
func (a *app) hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
