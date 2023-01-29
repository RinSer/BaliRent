package files

import (
	"bali-rent/pkg/auth"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var filesDir string

func RunFileServer() {

	filesDir = os.Getenv("FILES_DIR")

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	gr := e.Group("/files")
	// serve all files
	gr.Static("/images", filesDir)
	// upload files
	r := gr.Group("/upload")
	// Configure middleware with the custom claims type
	config := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(auth.JwtCustomClaims)
		},
		SigningKey: []byte(auth.GetJwtSecret()),
	}
	r.Use(echojwt.WithConfig(config))
	r.POST("", uploadFile)

	// run server
	e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))

}

func uploadFile(c echo.Context) error {
	// Source
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Destination
	ext := filepath.Ext(file.Filename)
	filename := uuid.New().String() + ext
	dst, err := os.Create(filesDir + "/" + filename)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	return c.String(http.StatusCreated, filename)
}
