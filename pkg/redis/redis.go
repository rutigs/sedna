package redis

import (
	"github.com/go-redis/redis"
)

// RedisSvc - gateway for interacting with the redis instance
type RedisSvc struct {
	// client to redis instance
	cl *redis.Client
}

// NewRedisSvc - Returns a new redis service
func NewRedisSvc(addr string) *RedisSvc {
	gw := &RedisSvc{
		cl: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: "", // no password set
			DB:       0,  // use default DB
		}),
	}

	return gw
}

// Get - attempt to get a value from redis store
func (gw *RedisSvc) Get(key string) (value string, ok bool) {
	val, err := gw.cl.Get(key).Result()
	if err != nil {
		return "", false
	}

	return val, true
}

// Set - set a key value pair in the redis store
func (gw *RedisSvc) Set(key, value string) bool {
	err := gw.cl.Set(key, value, 0).Err()
	if err != nil {
		return false
	}

	return true
}
