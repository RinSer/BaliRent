package sms

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type SmsSender interface {
	Send(phone string, message string) error
}

var format = "2006-01-02T15-04-05"
var SmsMockKey = "sms_mock"
type SmsMock struct {
	redis *redis.Client
}

func NewSmsMock(redis *redis.Client) *SmsMock {
	return &SmsMock{
		redis: redis,
	}
}

func (o *SmsMock) Send(phone string, message string) error {
	key := fmt.Sprintf("%s:%s:%s",SmsMockKey,phone,time.Now().Format(format))
	return o.redis.Set(context.Background(), key, message, time.Hour).Err()
}

func (o *SmsMock) GetByPhoneAndTime(phone string, date time.Time) (string, error) {
	key := fmt.Sprintf("%s:%s:%s",SmsMockKey,phone,date.Format(format))
	return o.redis.Get(context.Background(), key).Result()
}

func (o *SmsMock) GetByPhone(phone string) (map[time.Time]string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:%s:*",SmsMockKey,phone)
	keys := o.redis.Keys(ctx, key).Val()
	values,err := o.redis.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	result := make(map[time.Time]string,len(values))
	for i :=0;i<len(keys);i++ {
		t,err := time.Parse(format,strings.Split(keys[i], ":")[2])
		if err != nil {
			continue
		}
		result[time.Time(t)] = fmt.Sprintf("%v",values[i])
	}

	return result, nil
}

func (o *SmsMock) GetByPhoneLast(phone string) (string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:%s:*",SmsMockKey,phone)
	keys := o.redis.Keys(ctx, key).Val()
	maxTime := time.Time{} 
	for i :=0;i<len(keys);i++ {
		t,err := time.Parse(format,strings.Split(keys[i], ":")[2])
		if err != nil {
			continue
		}
		if t.After(maxTime) {
			maxTime = t
		}
	}
	key = fmt.Sprintf("%s:%s:%s",SmsMockKey,phone,maxTime.Format(format))
	return o.redis.Get(ctx, key).Result()
}

func (o *SmsMock) DeleteByPhone(phone string) error {
	ctx := context.Background()
	key := fmt.Sprintf("%s:%s:*",SmsMockKey,phone)
	keys := o.redis.Keys(ctx, key).Val()
	for _,k := range keys {
		o.redis.Del(ctx, k)
	}
	return nil
}