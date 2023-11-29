package initializer

import (
	"fmt"

	"github.com/go-redis/redis"
)

func RedisDBConnect(config *Config) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisURL,
		Password: "",
		DB:       0,
	})
	response, err := redisClient.Ping().Result()
	if err != nil {
		fmt.Println(response)
	}
	fmt.Println("Redis client is connected successfully")

}
