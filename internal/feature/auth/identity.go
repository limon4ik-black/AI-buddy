package auth

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type UserStore interface {
	FindOrCreate(provider, providerUserID, email, name string) int64
}

type RedisUserStore struct {
	rdb *redis.Client
}

func NewRedisUserStore(rdb *redis.Client) *RedisUserStore {
	return &RedisUserStore{rdb: rdb}
}

// Связка (provider, providerUserID) -> внутренний user_id, без TTL:
// аккаунт не должен протухать вместе с сессией.
func (s *RedisUserStore) FindOrCreate(provider, providerUserID, _, _ string) int64 {
	ctx := context.Background()
	key := "auth:user:" + provider + ":" + providerUserID

	if id, err := s.rdb.Get(ctx, key).Int64(); err == nil {
		return id
	}

	id, err := s.rdb.Incr(ctx, "auth:user:next_id").Result()
	if err != nil {
		return 0
	}
	// SetNX: при гонке двух первых входов побеждает записавший первым.
	ok, err := s.rdb.SetNX(ctx, key, id, 0).Result()
	if err != nil {
		return 0
	}
	if !ok {
		if existing, err := s.rdb.Get(ctx, key).Int64(); err == nil {
			return existing
		}
	}
	return id
}
