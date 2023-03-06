package models

import (
	"bali-rent/pkg/mail"
	"bali-rent/pkg/sms"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ConfirmationTtl = time.Second * 60
	ConfirmationDelay = time.Second * 5
)

type Confirmation struct {
	user *User
	redis *redis.Client
	mail *mail.Mail
	sms sms.SmsSender
}

func NewConfirmation(user *User, redis *redis.Client,mail *mail.Mail,sms sms.SmsSender) *Confirmation {
	return &Confirmation{
		user: user,
		redis: redis,
		mail: mail,
		sms: sms,
	}
}


func generateCode(count int) string {
	code := ""
	for i := 0; i < count; i++ {
		code += strconv.Itoa(rand.Intn(10))
	}
	return code
}

func EncodeBase64 (key string) string {
	return base64.StdEncoding.EncodeToString([]byte(key))
}

func (o *Confirmation) SendEmail(ctx context.Context) error {

	if o.user.Email == "" {
		return errors.New("user email is empty")
	}

	code := generateCode(6)
	key := fmt.Sprintf("email:%s",EncodeBase64(o.user.Email))

	if err := o.mail.Send(o.user.Email, "Confirmation code", "Your confirmation code is: "+code); err != nil {
		return err
	}
	
	return o.redis.Set(ctx,key, code,ConfirmationTtl).Err()
}

func (o *Confirmation) SendSms(ctx context.Context) error {

	if o.user.Phone == "" {
		return errors.New("user phone is empty")
	}

	code := generateCode(6)
	key := fmt.Sprintf("sms:%s",EncodeBase64(o.user.Phone))

	if err := o.sms.Send(o.user.Phone, "Your confirmation code is: "+code); err != nil {
		return err
	}

	return o.redis.Set(ctx,key, code,ConfirmationTtl).Err()
}


func (o *Confirmation) CheckEmail(ctx context.Context, code string) error {

	if o.user.Email == "" {
		return errors.New("user email is empty")
	}

	key := fmt.Sprintf("email:%s",EncodeBase64(o.user.Email))

	return o.check(ctx, key, code)
}

func (o *Confirmation) CheckSms(ctx context.Context, code string) error {

	if o.user.Phone == "" {
		return errors.New("user phone is empty")
	}

	key := fmt.Sprintf("sms:%s",EncodeBase64(o.user.Phone))

	return o.check(ctx, key, code)
}

func (o *Confirmation) check(ctx context.Context, key string, code string) error {
	
	val, err := o.redis.Get(ctx, key).Result()

	if err != nil {
		return err
	}

	if val != code {
		return errors.New("invalid code")
	}

	o.redis.Del(ctx, key)

	return nil
}