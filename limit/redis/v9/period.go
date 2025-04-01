package v9

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"

	"github.com/thinkgos/limiter/limit"
	redisScript "github.com/thinkgos/limiter/limit/redis"
)

var _ limit.PeriodStorage = (*PeriodStore)(nil)
var _ limit.PeriodLimitDriver = (*limit.PeriodLimit[*PeriodStore])(nil)

// A PeriodStore is used to limit requests during a period of time.
type PeriodStore struct {
	store *redis.Client
}

// NewPeriodStore returns a PeriodLimit with given parameters.
func NewPeriodStore(store *redis.Client) *PeriodStore {
	return &PeriodStore{
		store: store,
	}
}

// Take requests a permit with context, it returns the permit state.
func (p *PeriodStore) Take(ctx context.Context, key string, quota, expireSec int) (int64, error) {
	return p.store.Eval(ctx,
		redisScript.PeriodLimitScript,
		[]string{
			key,
		},
		[]string{
			strconv.Itoa(quota),
			strconv.Itoa(expireSec),
		},
	).Int64()
}

// SetQuotaFull set a permit over quota.
func (p *PeriodStore) SetQuotaFull(ctx context.Context, key string, quota, expireSec int) error {
	return p.store.Eval(ctx,
		redisScript.PeriodLimitSetQuotaFullScript,
		[]string{
			key,
		},
		[]string{
			strconv.Itoa(quota),
			strconv.Itoa(expireSec),
		},
	).Err()
}

// Del delete a permit
func (p *PeriodStore) Del(ctx context.Context, key string) error {
	return p.store.Del(ctx, key).Err()
}

// GetRunValue get run value
// Exist: false if key not exist.
// Count: current count
// TTL: not set expire time, t = -1
func (p *PeriodStore) GetRunValue(ctx context.Context, key string) ([]int64, error) {
	return p.store.Eval(ctx,
		redisScript.PeriodLimitRunValueScript,
		[]string{
			key,
		},
	).Int64Slice()
}
