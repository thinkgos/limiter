package v8

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/thinkgos/limiter/verified"
	redisScript "github.com/thinkgos/limiter/verified/redis"
)

// RedisStore verified captcha limit
type RedisStore struct {
	store *redis.Client // store redis client
}

// NewRedisStore new redis store instance.
func NewRedisStore(store *redis.Client) *RedisStore {
	return &RedisStore{store}
}

// Store the arguments.
func (v *RedisStore) Store(ctx context.Context, p *verified.StoreArgs) error {
	if p.DisableOneTime {
		return v.store.Eval(
			ctx,
			redisScript.StorageScript,
			[]string{p.Key},
			[]string{
				p.Answer,
				strconv.Itoa(p.MaxErrQuota),
				strconv.Itoa(int(p.KeyExpires / time.Second)),
			},
		).Err()
	} else {
		return v.store.Set(ctx, p.Key, p.Answer, p.KeyExpires).Err()
	}
}

// Verify the answer.
func (v *RedisStore) Verify(ctx context.Context, p *verified.VerifyArgs) bool {
	if p.DisableOneTime {
		code, err := v.store.Eval(
			ctx,
			redisScript.MatchScript,
			[]string{p.Key},
			[]string{p.Answer},
		).Int64()
		if err != nil {
			return false
		}
		return code == 0
	} else {
		wantAnswer, err := v.store.GetDel(ctx, p.Key).Result()
		return err == nil && wantAnswer == p.Answer
	}
}
