package api

import (
	"bali-rent/pkg/auth"
	"bali-rent/pkg/db"
	"bali-rent/pkg/db/models"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/random"
	"github.com/stretchr/testify/assert"
)

var testApp app

func authorizeRequest(next echo.HandlerFunc) echo.HandlerFunc {
	return echojwt.WithConfig(echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(auth.JwtCustomClaims)
		},
		SigningKey: []byte(auth.GetJwtSecret()),
	})(next)
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
		Phone:    random.New().String(25, "1234567890"),
		Email:    random.New().String(25, "1234567890"),
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
		Phone:    random.New().String(25, "1234567890"),
		Email:    random.New().String(25, "1234567890"),
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

func TestUserUpdate(t *testing.T) {
	testUserRequest := UserRequest{
		Phone:    random.New().String(25, "1234567890"),
		Email:    random.New().String(25, "1234567890"),
		Password: "test_password",
	}
	testRequest, err := json.Marshal(testUserRequest)
	assert.NoError(t, err)
	testServer := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/user/update",
		strings.NewReader(string(testRequest)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := testServer.NewContext(req, rec)
	// Assertions for unauthorized request
	assert.Error(t, authorizeRequest(testApp.updateUser)(c))
	// Setup register
	req = httptest.NewRequest(http.MethodPost, "/register",
		strings.NewReader(string(testRequest)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = testServer.NewContext(req, rec)
	// Assertions for register
	var testUser *models.User
	if assert.NoError(t, testApp.register(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		testUser = new(models.User)
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
	// Assertions for login
	var authResponse struct {
		Token string `json:"token"`
	}
	if assert.NoError(t, testApp.login(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "token")
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &authResponse))
	}
	// Setup restricted
	testUserRequest = UserRequest{
		Id:       &testUser.Id,
		Phone:    random.New().String(25, "1234567890"),
		Email:    random.New().String(25, "1234567890"),
		Password: "new_test_password",
	}
	testRequest, err = json.Marshal(testUserRequest)
	assert.NoError(t, err)
	req = httptest.NewRequest(http.MethodGet, "/user/update",
		strings.NewReader(string(testRequest)))
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+authResponse.Token)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = testServer.NewContext(req, rec)
	// Assertions for update user request
	if assert.NoError(t, authorizeRequest(testApp.updateUser)(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		updatedUser := new(models.User)
		err = json.Unmarshal(rec.Body.Bytes(), updatedUser)
		assert.NoError(t, err)
		assert.Equal(t, *testUserRequest.Id, updatedUser.Id)
		assert.Equal(t, testUserRequest.Phone, updatedUser.Phone)
		assert.Equal(t, testUserRequest.Email, updatedUser.Email)
		assert.NotNil(t, testUser.WhenUpdated)
	}
}

func TestProperty(t *testing.T) {
	// Setup user
	testServer := echo.New()
	testUserRequest := UserRequest{
		Phone:    random.New().String(25, "1234567890"),
		Email:    random.New().String(25, "1234567890"),
		Password: "test_password",
	}
	testRequest, err := json.Marshal(testUserRequest)
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/register",
		strings.NewReader(string(testRequest)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := testServer.NewContext(req, rec)
	// Assertions for user
	var testUser *models.User
	if assert.NoError(t, testApp.register(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		testUser = new(models.User)
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
	// Assertions for login
	var authResponse struct {
		Token string `json:"token"`
	}
	if assert.NoError(t, testApp.login(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "token")
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &authResponse))
	}
	// Setup property
	propertyData := &models.Property{
		OwnerId:      testUser.Id,
		Latitude:     rand.Float64(),
		Longitude:    rand.Float64(),
		Photos:       nil,
		Description:  random.New().String(100, "1234567890"),
		Price:        100,
		PropertyType: models.Room,
		ForSell:      false,
		IsDeleted:    false,
	}
	testPropertyRequest, err := json.Marshal(propertyData)
	assert.NoError(t, err)
	// Assertions for property create
	req = httptest.NewRequest(http.MethodPost, "/property",
		strings.NewReader(string(testPropertyRequest)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+authResponse.Token)
	rec = httptest.NewRecorder()
	c = testServer.NewContext(req, rec)
	var testProperty *models.Property
	if assert.NoError(t, authorizeRequest(testApp.createProperty)(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		testProperty = new(models.Property)
		err = json.Unmarshal(rec.Body.Bytes(), testProperty)
		assert.NoError(t, err)
		assert.Equal(t, propertyData.Price, testProperty.Price)
		assert.NotNil(t, testProperty.Id)
	}
	// Assertions for property get
	req = httptest.NewRequest(http.MethodGet, "/property?id="+strconv.FormatInt(testProperty.Id, 10), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = testServer.NewContext(req, rec)
	if assert.NoError(t, testApp.getProperties(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		getTestProperty := new(models.Property)
		err = json.Unmarshal(rec.Body.Bytes(), getTestProperty)
		assert.NoError(t, err)
		assert.Equal(t, testProperty.Id, getTestProperty.Id)
	}
	// Assertions for property update
	propertyData.Id = testProperty.Id
	propertyData.Price *= 100
	testPropertyRequest, err = json.Marshal(propertyData)
	assert.NoError(t, err)
	req = httptest.NewRequest(http.MethodPost, "/property",
		strings.NewReader(string(testPropertyRequest)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+authResponse.Token)
	rec = httptest.NewRecorder()
	c = testServer.NewContext(req, rec)
	if assert.NoError(t, authorizeRequest(testApp.updateProperty)(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		testProperty = new(models.Property)
		err = json.Unmarshal(rec.Body.Bytes(), testProperty)
		assert.NoError(t, err)
		assert.Equal(t, propertyData.Price, testProperty.Price)
	}
	// Assertions for properties get
	req = httptest.NewRequest(http.MethodGet, "/property?ownerId="+strconv.FormatInt(testUser.Id, 10), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+authResponse.Token)
	rec = httptest.NewRecorder()
	c = testServer.NewContext(req, rec)
	if assert.NoError(t, authorizeRequest(testApp.getProperties)(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		testProperties := make([]models.Property, 0)
		err = json.Unmarshal(rec.Body.Bytes(), &testProperties)
		assert.NoError(t, err)
		assert.NotEmpty(t, testProperties)
		assert.Equal(t, testProperties[0].Price, testProperty.Price)
	}
}
