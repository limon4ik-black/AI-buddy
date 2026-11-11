package auth

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type RefreshStore interface {
	Save(ctx context.Context, token string, userID int64) error
	Get(ctx context.Context, token string) (userID int64, ok bool, err error)
	Delete(ctx context.Context, token string) error
}

// Redis снимает ключи по TTL: истекший токен == ключ не найден.
type RedisRefreshStore struct {
	rdb *redis.Client
}

func NewRedisRefreshStore(rdb *redis.Client) *RedisRefreshStore {
	return &RedisRefreshStore{rdb: rdb}
}

func refreshKey(token string) string {
	return "auth:refresh:" + token
}

func (s *RedisRefreshStore) Save(ctx context.Context, token string, userID int64) error {
	return s.rdb.Set(ctx, refreshKey(token), userID, RefreshTokenTTL).Err()
}

func (s *RedisRefreshStore) Get(ctx context.Context, token string) (int64, bool, error) {
	val, err := s.rdb.Get(ctx, refreshKey(token)).Result()
	if err == redis.Nil {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	userID, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, false, err
	}
	return userID, true, nil
}

func (s *RedisRefreshStore) Delete(ctx context.Context, token string) error {
	return s.rdb.Del(ctx, refreshKey(token)).Err()
}
