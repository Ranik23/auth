package redis

import "github.com/go-redis/redis"



type Redis interface {
	Put(key, value string) error
	Get(key string) (string, error)
}

type RedisImpl struct {
	redisClient *redis.Client
}

func NewRedisImpl(dsn string) *RedisImpl {
	return &RedisImpl{
		redisClient: redis.NewClient(&redis.Options{
			Addr:     dsn,
			Password: "", 
			DB:       0,
		}),
	}
}

func (r *RedisImpl) Put(key, value string) error {
	return nil
}

func (r *RedisImpl) Get(key string) (string, error) {
	return "", nil
}

