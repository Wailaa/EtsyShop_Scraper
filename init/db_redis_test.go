package initializer_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	initializer "EtsyScraper/init"
)

func TestValidRedisDBConnect(t *testing.T) {

	config := &initializer.Config{
		RedisURL: "localhost:6379",
	}

	initializer.RedisDBConnect(config)

	if initializer.RedisClient == nil {
		t.Errorf("RedisClient is nil")
	}

	context := context.TODO()
	response, err := initializer.RedisClient.Ping(context).Result()
	if err != nil {
		t.Errorf("Failed to connect to Redis database: %s", err.Error())
	}
	if response != "PONG" {
		t.Errorf("Invalid response from Redis database: %s", response)
	}
}

func TestValidRedisDBConnectWrongPort(t *testing.T) {

	config := &initializer.Config{
		RedisURL: "localhost:1111",
	}

	initializer.RedisDBConnect(config)

	if initializer.RedisClient == nil {
		t.Errorf("RedisClient is nil ")
	}

	context := context.TODO()
	_, err := initializer.RedisClient.Ping(context).Result()

	assert.Error(t, err)

}

func TestValidRedisDBConnectIntegretionTest(t *testing.T) {

	config := initializer.LoadProjConfig(".")

	initializer.RedisDBConnect(&config)

	if initializer.RedisClient == nil {
		t.Errorf("RedisClient is nil")
	}

	context := context.TODO()
	response, err := initializer.RedisClient.Ping(context).Result()
	if err != nil {
		t.Errorf("Failed to connect to Redis database: %s", err.Error())
	}
	if response != "PONG" {
		t.Errorf("Invalid response from Redis database: %s", response)
	}
}
