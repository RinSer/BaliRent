package api 

import (
	"bali-rent/pkg/auth"
	"bali-rent/pkg/db"
	"bali-rent/pkg/db/models"
	"bali-rent/pkg/mail"
	"bali-rent/pkg/redis"
	"bali-rent/pkg/sms"
	"context"
	"encoding/json"
	"fmt"
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
	rds "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

var testApp app
var r *rds.Client

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
	os.Setenv("MAIL_FROM", "admin@bali-rent.com")
	os.Setenv("MAIL_PASSWORD", "")
	os.Setenv("MAIL_SERVER", "localhost:1025")
	os.Setenv("REDIS_URL", "localhost:6379")

	// connect the db
	pgdb, err := db.StartDB()

	if err != nil {
		log.Panicf("error: %v", err)
	}

	r, err = redis.Connect()
	mail := mail.NewMail()
	sms := sms.NewSmsMock(r)
	if err != nil {
		log.Panicf("error: %v", err)
	}

	// create test app
	testApp = app{pgdb, r, mail, sms}
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

		v, err := r.Get(context.Background(), fmt.Sprintf("email:%s", models.EncodeBase64(testUserRequest.Email))).Result()
		assert.NoError(t, err)
		assert.NotEmpty(t, v)

		v, err = r.Get(context.Background(), fmt.Sprintf("sms:%s", models.EncodeBase64(testUserRequest.Phone))).Result()
		assert.NoError(t, err)
		assert.NotEmpty(t, v)
	}
}

func TestConfirmEmail(t *testing.T) {
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

	if !assert.NoError(t, testApp.register(c)) {
		return
	}

	encoded := models.EncodeBase64(testUserRequest.Email)
	req = httptest.NewRequest(http.MethodPost, "/api/restore", nil)

	rec = httptest.NewRecorder()
	c = testServer.NewContext(req, rec)
	c.SetPath("/email/:email")
	c.SetParamNames("email")
	c.SetParamValues(encoded)

	if !assert.NoError(t, testApp.restoreByEmail(c)) {
		return
	}

	code, err := r.Get(context.Background(), fmt.Sprintf("email:%s", models.EncodeBase64(testUserRequest.Email))).Result()
	assert.NoError(t, err)

	req = httptest.NewRequest(http.MethodPost, "/api/confirm", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = testServer.NewContext(req, rec)
	c.SetPath("/email/:email/:code")
	c.SetParamNames("email", "code")
	c.SetParamValues(encoded, code)

	if assert.NoError(t, testApp.confirmByEmail(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "token")
	}
}

func TestConfirmPhone(t *testing.T) {

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

	if !assert.NoError(t, testApp.register(c)) {
		return
	}

	encoded := models.EncodeBase64(testUserRequest.Phone)
	req = httptest.NewRequest(http.MethodPost, "/api/restore", nil)

	rec = httptest.NewRecorder()
	c = testServer.NewContext(req, rec)
	c.SetPath("/sms/:phone")
	c.SetParamNames("phone")
	c.SetParamValues(encoded)

	if !assert.NoError(t, testApp.restoreBySms(c)) {
		return
	}

	code, err := r.Get(context.Background(), fmt.Sprintf("sms:%s", models.EncodeBase64(testUserRequest.Phone))).Result()
	assert.NoError(t, err)

	req = httptest.NewRequest(http.MethodPost, "/api/confirm", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = testServer.NewContext(req, rec)
	c.SetPath("/sms/:phone/:code")
	c.SetParamNames("phone", "code")
	c.SetParamValues(encoded, code)

	if assert.NoError(t, testApp.confirmBySms(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "token")
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
	// Assertions for properties get by owner
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
	// Assertions for properties get paged
	count := 10
	req = httptest.NewRequest(http.MethodGet, "/property?from=0&count="+strconv.Itoa(count), nil)
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
		assert.GreaterOrEqual(t, count, len(testProperties))
	}
	// Assertions for filtered by location position
	latStart := testProperty.Latitude - 0.0000001
	latEnd := testProperty.Latitude + 0.0000001
	longStart := testProperty.Longitude - 0.0000001
	longEnd := testProperty.Longitude + 0.0000001
	req = httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/property?latStart=%.8f&latEnd=%.8f&longStart=%.8f&longEnd=%.8f",
			latStart, latEnd, longStart, longEnd), nil)
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
		assert.Contains(t, testProperties, *testProperty)
	}
}
