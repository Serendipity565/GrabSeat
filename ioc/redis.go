package ioc

import (
	"context"
	"fmt"

	"github.com/Serendipity565/GrabSeat/config"
	"github.com/redis/go-redis/v9"
)

func InitRedis(conf *config.RedisConfig) redis.Cmdable {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.DB,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		panic(fmt.Sprintf("Redis 连接失败: %v", err))
	}
	return rdb
}
