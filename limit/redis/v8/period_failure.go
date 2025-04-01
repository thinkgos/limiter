package v8

import (
	"context"
	"strconv"

	"github.com/go-redis/redis/v8"

	"github.com/thinkgos/limiter/limit"
	redisScript "github.com/thinkgos/limiter/limit/redis"
)

var _ limit.PeriodFailureStorage = (*PeriodFailureStore)(nil)
var _ limit.PeriodFailureLimitDriver = (*limit.PeriodFailureLimit[*PeriodFailureStore])(nil)

// A PeriodFailureStore is used to limit requests when failure during a period of time.
type PeriodFailureStore struct {
	store *redis.Client
}

// NewPeriodFailureStore returns a PeriodFailureLimit with given parameters.
func NewPeriodFailureStore(store *redis.Client) *PeriodFailureStore {
	return &PeriodFailureStore{
		store: store,
	}
}

// Check requests a permit.
func (p *PeriodFailureStore) Check(ctx context.Context, key string, quota, expireSec int, success bool) (int64, error) {
	s := "0"
	if success {
		s = "1"
	}
	return p.store.Eval(ctx,
		redisScript.PeriodFailureLimitFixedScript,
		[]string{key},
		[]string{
			strconv.Itoa(quota),
			strconv.Itoa(expireSec),
			s,
		},
	).Int64()
}

// SetQuotaFull set a permit over quota.
func (p *PeriodFailureStore) SetQuotaFull(ctx context.Context, key string, quota, expireSec int) error {
	err := p.store.Eval(ctx,
		redisScript.PeriodFailureLimitFixedSetQuotaFullScript,
		[]string{key},
		[]string{
			strconv.Itoa(quota),
			strconv.Itoa(expireSec),
		},
	).Err()
	if err == redis.Nil {
		return nil
	}
	return err
}

// Del delete a permit
func (p *PeriodFailureStore) Del(ctx context.Context, key string) error {
	return p.store.Del(ctx, key).Err()
}

// GetRunValue get run value
// Exist: false if key not exist.
// Count: current failure count
// TTL: not set expire time, t = -1
func (p *PeriodFailureStore) GetRunValue(ctx context.Context, key string) ([]int64, error) {
	return p.store.Eval(ctx,
		redisScript.PeriodFailureLimitRunValueScript,
		[]string{key},
	).Int64Slice()
}
