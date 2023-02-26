package redis

import (
	"context"
	"log"
	"strings"

	"github.com/go-redis/redis/extra/redisotel/v8"
	redisLib "github.com/go-redis/redis/v8"
	"github.com/nmtri1912/go-common/pkg/redis"
	"github.com/nmtri1912/go-common/pkg/redisprom"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

func NewCache(lifecycle fx.Lifecycle) redis.Cache {
	addresses := viper.GetString("redis.addresses")

	if len(addresses) == 0 {
		log.Fatal("Invalid redis address")
	}

	client := redisLib.NewUniversalClient(&redisLib.UniversalOptions{
		Addrs: strings.Split(addresses, ","),
	})

	//redis.monitor-hook = true => using monitor hook
	monitorHook := viper.GetBool("redis.monitor-hook")
	if monitorHook {
		hook := redisprom.NewHook()
		client.AddHook(hook)
	}
	client.AddHook(redisotel.NewTracingHook())

	log.Println("Trying to connect redis...")
	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatalln("Can not connect redis, address=", addresses, err)
	}

	log.Println("Connect redis successfully")

	lifecycle.Append(fx.Hook{OnStop: func(ctx context.Context) error {
		log.Println("Closing redis connection")
		return client.Close()
	}})

	return &redis.RedisCache{UniversalClient: client}
}
