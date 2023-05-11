package utils

import (
	"github.com/go-redis/redis"
)

// Define Redis Struct
type Redis struct {
	client *redis.Client
}

// Method to set Redis value
func (r *Redis) SetValue(key string, value string) error {
	err := r.client.Set(key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

// Method to get Redis value
func (r *Redis) GetValue(key string) (string, error) {
	val, err := r.client.Get(key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func NewRedis(address, password string) *Redis {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       0,
	})

	return &Redis{
		client: client,
	}
}
