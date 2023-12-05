package initializer

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func RedisDBConnect(config *Config) {
	context := context.TODO()
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.RedisURL,
		Password: "",
		DB:       0,
	})
	response, err := RedisClient.Ping(context).Result()
	if err != nil {
		fmt.Println(response)
	}
	fmt.Println("Redis client is connected successfully")

}
