package sms 

import (
	rc "bali-rent/pkg/redis"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

var r *redis.Client

func TestMain(m *testing.M) {
	err := os.Setenv("REDIS_URL", "localhost:6379")
	if err != nil {
		panic(err)
	}

	r, err = rc.Connect()
	if err != nil {
		panic(err)
	}

	m.Run()

	defer func() { _ = r.Close() }()
}

func TestSend(t *testing.T) {

	phone := uuid.New().String()
	sms := NewSmsMock(r)
	err := sms.Send(phone, "test")

	assert.NoError(t, err)

	_ = sms.DeleteByPhone(phone)

}

func TestGetByPhone(t *testing.T) {

	phone := uuid.New().String()
	sms := NewSmsMock(r)
	_ = sms.Send(phone, "test1")
	time.Sleep(2 * time.Second)
	_ = sms.Send(phone, "test2")

	act, err := sms.GetByPhone(phone)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(act))

	_ = sms.DeleteByPhone(phone)

}

func TestGetByPhoneAndTime(t *testing.T) {

	phone := uuid.New().String()
	code := "test"
	sms := NewSmsMock(r)
	n := time.Now()
	_ = sms.Send(phone, code)

	act, err := sms.GetByPhoneAndTime(phone, n)

	assert.NoError(t, err)
	assert.Equal(t, code, act)

	_ = sms.DeleteByPhone(phone)

}

func TestGetByPhoneLast(t *testing.T) {

	phone := uuid.New().String()
	sms := NewSmsMock(r)
	_ = sms.Send(phone, "test1")
	time.Sleep(time.Second)
	_ = sms.Send(phone, "test2")

	act, err := sms.GetByPhoneLast(phone)

	assert.NoError(t, err)
	assert.Equal(t, "test2", act)

	_ = sms.DeleteByPhone(phone)

}
