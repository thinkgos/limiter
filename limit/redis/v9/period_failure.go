package v9

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/things-go/limiter/limit"
	redisScript "github.com/things-go/limiter/limit/redis"
)

var _ limit.PeriodFailureLimitDriver = (*PeriodFailureLimit)(nil)

// A PeriodFailureLimit is used to limit requests when failure during a period of time.
type PeriodFailureLimit struct {
	// a period seconds of time
	period int
	// limit quota requests during a period seconds of time.
	quota int
	// keyPrefix in redis
	keyPrefix string
	store     *redis.Client
	isAlign   bool
}

// NewPeriodFailureLimit returns a PeriodFailureLimit with given parameters.
func NewPeriodFailureLimit(store *redis.Client, opts ...PeriodLimitOption) *PeriodFailureLimit {
	limiter := &PeriodFailureLimit{
		period:    int(24 * time.Hour / time.Second),
		quota:     6,
		keyPrefix: "limit:period:failure:", // limit:period:failure:
		store:     store,
	}
	for _, opt := range opts {
		opt(limiter)
	}
	return limiter
}

func (p *PeriodFailureLimit) align()                { p.isAlign = true }
func (p *PeriodFailureLimit) setKeyPrefix(k string) { p.keyPrefix = k }
func (p *PeriodFailureLimit) setPeriod(v time.Duration) {
	if vv := int(v / time.Second); vv > 0 {
		p.period = int(v / time.Second)
	}
}
func (p *PeriodFailureLimit) setQuota(v int) { p.quota = v }

// CheckErr requests a permit state.
// same as Check
func (p *PeriodFailureLimit) CheckErr(ctx context.Context, key string, err error) (limit.PeriodFailureLimitState, error) {
	return p.Check(ctx, key, err == nil)
}

// Check requests a permit.
func (p *PeriodFailureLimit) Check(ctx context.Context, key string, success bool) (limit.PeriodFailureLimitState, error) {
	s := "0"
	if success {
		s = "1"
	}
	code, err := p.store.Eval(ctx,
		redisScript.PeriodFailureLimitFixedScript,
		[]string{p.formatKey(key)},
		[]string{
			strconv.Itoa(p.quota),
			strconv.Itoa(p.calcExpireSeconds()),
			s,
		},
	).Int64()
	if err != nil {
		return limit.PeriodFailureLimitStsUnknown, err
	}
	switch code {
	case redisScript.InnerPeriodFailureLimitCodeSuccess:
		return limit.PeriodFailureLimitStsSuccess, nil
	case redisScript.InnerPeriodFailureLimitCodeInQuota:
		return limit.PeriodFailureLimitStsInQuota, nil
	case redisScript.InnerPeriodFailureLimitCodeOverQuota:
		return limit.PeriodFailureLimitStsOverQuota, nil
	default:
		return limit.PeriodFailureLimitStsUnknown, limit.ErrUnknownCode
	}
}

// SetQuotaFull set a permit over quota.
func (p *PeriodFailureLimit) SetQuotaFull(ctx context.Context, key string) error {
	err := p.store.Eval(ctx,
		redisScript.PeriodFailureLimitFixedSetQuotaFullScript,
		[]string{p.formatKey(key)},
		[]string{
			strconv.Itoa(p.quota),
			strconv.Itoa(p.calcExpireSeconds()),
		},
	).Err()
	if err == redis.Nil {
		return nil
	}
	return err
}

// Del delete a permit
func (p *PeriodFailureLimit) Del(ctx context.Context, key string) error {
	return p.store.Del(ctx, p.formatKey(key)).Err()
}

// TTL get key ttl
// if key not exist, time = -2.
// if key exist, but not set expire time, t = -1.
func (p *PeriodFailureLimit) TTL(ctx context.Context, key string) (time.Duration, error) {
	return p.store.TTL(ctx, p.formatKey(key)).Result()
}

// GetInt get current failure count
func (p *PeriodFailureLimit) GetInt(ctx context.Context, key string) (int, bool, error) {
	v, err := p.store.Get(ctx, p.formatKey(key)).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, false, nil
		}
		return 0, false, err
	}
	return v, true, nil
}

// GetRunValue get run value
// Exist: false if key not exist.
// Count: current failure count
// TTL: not set expire time, t = -1
func (p *PeriodFailureLimit) GetRunValue(ctx context.Context, key string) (*limit.RunValue, error) {
	tb, err := p.store.Eval(ctx,
		redisScript.PeriodFailureLimitRunValueScript,
		[]string{
			p.formatKey(key),
		},
	).Int64Slice()
	if err != nil {
		return nil, err
	}
	switch {
	case len(tb) == 1 && tb[0] == 0:
		return &limit.RunValue{
			Exist: false,
			Count: 0,
			TTL:   0,
		}, nil
	case len(tb) == 3:
		var t time.Duration

		switch n := tb[2]; n {
		// -2 if the key does not exist
		// -1 if the key exists but has no associated expire
		case -2, -1:
			t = time.Duration(n)
		default:
			t = time.Duration(n) * time.Second
		}
		return &limit.RunValue{
			Exist: tb[0] == 1 && t != -2,
			Count: tb[1],
			TTL:   t,
		}, nil
	default:
		return nil, limit.ErrUnknownCode
	}
}

func (p *PeriodFailureLimit) formatKey(key string) string {
	return p.keyPrefix + key
}

func (p *PeriodFailureLimit) calcExpireSeconds() int {
	if p.isAlign {
		now := time.Now()
		_, offset := now.Zone()
		unix := now.Unix() + int64(offset)
		return p.period - int(unix%int64(p.period))
	}
	return p.period
}
