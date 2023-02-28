package redis

import (
	"context"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

func Connect() (*redis.Client,error) {

	u, err := url.Parse(os.Getenv("REDIS_URL"))
    if err != nil {
        return nil, err
    }
    
    // Create a Redis Options object
    options := &redis.Options{
        Addr: u.Host,
    }
    
    // Set the password if provided
    if u.User != nil {
        password, ok := u.User.Password()
        if ok {
            options.Password = password
        }
    }
    
    // Set the Redis database if provided
    if len(u.Path) > 1 {
        options.DB, err = strconv.Atoi(strings.TrimLeft(u.Path, "/"))
        if err != nil {
            return nil, err
        }
    }
    
    // Create a new Redis client with the options
    client := redis.NewClient(options)
    
    // Ping the Redis server to check if the connection is successful
    err = client.Ping(context.Background()).Err()
    if err != nil {
        return nil, err
    }
    
    return client, nil
}