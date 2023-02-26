package redis

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type Cache interface {
	redis.UniversalClient
	CheckDuplicate(ctx context.Context, key string, ttl time.Duration) (bool, error)
	TryGetInt(ctx context.Context, key string) int
	TryGetString(ctx context.Context, key string) string
}

type RedisCache struct {
	redis.UniversalClient
}

// NewCacheSingle create a redis client in single mode, useful in unit test
func NewCacheSingle(add string) Cache {
	client := redis.NewClient(&redis.Options{
		Addr: add,
	})

	log.Println("Trying to connect redis...")
	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatalln("Can not connect redis, address=", add, err)
	}

	log.Println("Connect redis successfully")

	return &RedisCache{UniversalClient: client}
}

func (c RedisCache) CheckDuplicate(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	val, err := c.SetNX(ctx, key, true, ttl).Result()
	if err != nil {
		return true, err
	}
	return !val, nil
}

func (c RedisCache) TryGetInt(ctx context.Context, key string) int {
	val, err := c.Get(ctx, key).Result()
	if err != nil {
		log.Println("Get error ", err)
		return 0
	}
	num, err := strconv.Atoi(val)
	if err != nil {
		log.Println("Can not parse number ", err)
		return 0
	}
	return num
}

func (c RedisCache) TryGetString(ctx context.Context, key string) string {
	val, err := c.Get(ctx, key).Result()
	if err != nil {
		log.Println("Get error ", err)
		return ""
	}
	return val
}
