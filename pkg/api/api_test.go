package api

import (
	"bali-rent/pkg/db"
	"bali-rent/pkg/db/models"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/random"
	"github.com/stretchr/testify/assert"
)

var testApp app

func getJwtMiddleware() echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(jwtCustomClaims)
		},
		SigningKey: []byte(getJwtSecret()),
	})
}

func TestMain(m *testing.M) {
	os.Setenv("DATABASE_URL", "localhost")
	os.Setenv("JWT_SECRET", "test-secret")
	// connect the db
	pgdb, err := db.StartDB()
	if err != nil {
		log.Panicf("error: %v", err)
	}
	// create test app
	testApp = app{pgdb}
	os.Exit(m.Run())
}

func TestRegister(t *testing.T) {
	testServer := echo.New()
	testUserRequest := UserRequest{
		Phone:    random.New().String(5, "1234567890"),
		Password: "test_password",
	}
	testRequest, err := json.Marshal(testUserRequest)
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/register",
		strings.NewReader(string(testRequest)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := testServer.NewContext(req, rec)
	// Assertions
	if assert.NoError(t, testApp.register(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		testUser := new(models.User)
		err = json.Unmarshal(rec.Body.Bytes(), testUser)
		assert.NoError(t, err)
		assert.Equal(t, testUserRequest.Phone, testUser.Phone)
		assert.NotNil(t, testUser.Id)
	}
}

func TestLogin(t *testing.T) {
	testServer := echo.New()
	testUserRequest := UserRequest{
		Phone:    random.New().String(5, "1234567890"),
		Password: "test_password",
	}
	testRequest, err := json.Marshal(testUserRequest)
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/login",
		strings.NewReader(string(testRequest)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := testServer.NewContext(req, rec)
	// Assertions for unregistered login
	assert.Error(t, testApp.login(c))
	// Setup register
	req = httptest.NewRequest(http.MethodPost, "/register",
		strings.NewReader(string(testRequest)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = testServer.NewContext(req, rec)
	// Assertions for register
	if assert.NoError(t, testApp.register(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		testUser := new(models.User)
		err = json.Unmarshal(rec.Body.Bytes(), testUser)
		assert.NoError(t, err)
		assert.Equal(t, testUserRequest.Phone, testUser.Phone)
		assert.NotNil(t, testUser.Id)
	}
	// Setup login
	req = httptest.NewRequest(http.MethodPost, "/login",
		strings.NewReader(string(testRequest)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = testServer.NewContext(req, rec)
	// Assertions for register
	if assert.NoError(t, testApp.login(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "token")
	}
}

func TestRestricted(t *testing.T) {
	testServer := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/restricted", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := testServer.NewContext(req, rec)
	// Assertions for unauthorized request
	assert.Error(t, getJwtMiddleware()(testApp.restricted)(c))
	// Setup register
	testUserRequest := UserRequest{
		Phone:    random.New().String(5, "1234567890"),
		Password: "test_password",
	}
	testRequest, err := json.Marshal(testUserRequest)
	assert.NoError(t, err)
	req = httptest.NewRequest(http.MethodPost, "/register",
		strings.NewReader(string(testRequest)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = testServer.NewContext(req, rec)
	// Assertions for register
	if assert.NoError(t, testApp.register(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		testUser := new(models.User)
		err = json.Unmarshal(rec.Body.Bytes(), testUser)
		assert.NoError(t, err)
		assert.Equal(t, testUserRequest.Phone, testUser.Phone)
		assert.NotNil(t, testUser.Id)
	}
	// Setup login
	req = httptest.NewRequest(http.MethodPost, "/login",
		strings.NewReader(string(testRequest)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = testServer.NewContext(req, rec)
	// Assertions for register
	var authResponse struct {
		Token string `json:"token"`
	}
	if assert.NoError(t, testApp.login(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "token")
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &authResponse))
	}
	// Setup restricted
	req = httptest.NewRequest(http.MethodGet, "/restricted", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+authResponse.Token)
	rec = httptest.NewRecorder()
	c = testServer.NewContext(req, rec)
	// Assertions for authorized request
	if assert.NoError(t, getJwtMiddleware()(testApp.restricted)(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), testUserRequest.Phone)
	}
}
