package config

import (
	"context"
	"github.com/trae2api/pkg/logger"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

var RDB redis.Cmdable
var RedisConnString = getEnv("REDIS_CONN_STRING", "")

// InitRedisClient This function is called after init()
func InitRedisClient() (err error) {
	if RedisConnString == "" {
		logger.Log.Debug("REDIS_CONN_STRING not set, Redis is not enabled")
		return nil
	}

	redisConnString := os.Getenv("REDIS_CONN_STRING")

	logger.Log.Println("Redis is enabled")
	opt, err := redis.ParseURL(redisConnString)
	if err != nil {
		logger.Log.Fatalln("failed to parse Redis connection string: " + err.Error())
	}
	RDB = redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = RDB.Ping(ctx).Result()
	if err != nil {
		logger.Log.Fatalln("Redis ping test failed: " + err.Error())
	}
	return err
}

func ParseRedisOption() *redis.Options {
	opt, err := redis.ParseURL(os.Getenv("REDIS_CONN_STRING"))
	if err != nil {
		logger.Log.Fatalln("failed to parse Redis connection string: " + err.Error())
	}
	return opt
}

func RedisSet(key string, value string, expiration time.Duration) error {
	ctx := context.Background()
	return RDB.Set(ctx, key, value, expiration).Err()
}

func RedisGet(key string) (string, error) {
	ctx := context.Background()
	return RDB.Get(ctx, key).Result()
}

func RedisDel(key string) error {
	ctx := context.Background()
	return RDB.Del(ctx, key).Err()
}

func RedisDecrease(key string, value int64) error {
	ctx := context.Background()
	return RDB.DecrBy(ctx, key, value).Err()
}
