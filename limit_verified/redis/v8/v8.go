package v8

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/thinkgos/limiter/limit_verified"
	redisScript "github.com/thinkgos/limiter/limit_verified/redis"
)

// RedisStore verified captcha limit
type RedisStore struct {
	store *redis.Client // store client
}

// NewRedisStore
func NewRedisStore(store *redis.Client) *RedisStore {
	return &RedisStore{store}
}

func (v *RedisStore) Store(ctx context.Context, p *limit_verified.StoreArgs) error {
	sts, err := v.store.Eval(
		ctx,
		redisScript.LimitVerifiedSendCodeScript,
		[]string{
			p.KeyPrefix,
			p.Kind,
			p.Target,
		},
		[]string{
			p.Code,
			strconv.Itoa(p.MaxSendPerDay),
			strconv.Itoa(p.CodeMaxSendPerDay),
			strconv.Itoa(p.CodeMaxErrorQuota),
			strconv.Itoa(p.CodeAvailWindowSecond),
			strconv.Itoa(p.CodeResendIntervalSecond),
			p.NowSecond,
			strconv.FormatInt(int64(p.KeyExpires/time.Second), 10),
		},
	).Int64()
	if err != nil {
		return err
	}
	switch sts {
	case redisScript.InnerLimitVerifiedSuccess:
		return nil
	case redisScript.InnerLimitVerifiedOfSendCodeReachMaxSendPerDay:
		err = limit_verified.ErrMaxSendPerDay
	case redisScript.InnerLimitVerifiedOfSendCodeResendTooFrequently:
		err = limit_verified.ErrResendTooFrequently
	default:
		err = limit_verified.ErrUnknownCode
	}
	return err
}

func (v *RedisStore) Rollback(ctx context.Context, p *limit_verified.RollbackArgs) error {
	return v.store.Eval(
		ctx,
		redisScript.LimitVerifiedRollbackSendCntAndCodeCntScript,
		[]string{
			p.KeyPrefix,
			p.Kind,
			p.Target,
		},
		[]string{
			p.Code,
			p.NowSecond,
		},
	).Err()
}

// VerifyCode verify code from redis cache.
func (v *RedisStore) Verify(ctx context.Context, p *limit_verified.VerifyArgs) error {
	sts, err := v.store.Eval(
		ctx,
		redisScript.LimitVerifiedVerifyCodeScript,
		[]string{
			p.KeyPrefix,
			p.Kind,
			p.Target,
		},
		[]string{
			p.Code,
			p.NowSecond,
		},
	).Int64()
	if err != nil {
		return err
	}
	switch sts {
	case redisScript.InnerLimitVerifiedSuccess:
		err = nil
	case redisScript.InnerLimitVerifiedOfVerifyCodeRequiredOrExpired:
		err = limit_verified.ErrCodeRequiredOrExpired
	case redisScript.InnerLimitVerifiedOfVerifyCodeReachMaxError:
		err = limit_verified.ErrCodeMaxErrorQuota
	case redisScript.InnerLimitVerifiedOfVerifyCodeVerificationFailure:
		err = limit_verified.ErrCodeVerification
	default:
		err = limit_verified.ErrUnknownCode
	}
	return err
}

func (v *RedisStore) Incr(ctx context.Context, p *limit_verified.IncrArgs) error {
	sts, err := v.store.Eval(
		ctx,
		redisScript.LimitVerifiedIncrSendCntScript,
		[]string{
			p.KeyPrefix,
			p.Target,
		},
		[]string{
			strconv.Itoa(p.MaxSendPerDay),
			strconv.FormatInt(int64(p.KeyExpires/time.Second), 10),
		},
	).Int64()
	if err != nil {
		return err
	}
	switch sts {
	case redisScript.InnerLimitVerifiedSuccess:
		err = nil
	case redisScript.InnerLimitVerifiedOfSendCodeReachMaxSendPerDay:
		err = limit_verified.ErrMaxSendPerDay
	default:
		err = limit_verified.ErrUnknownCode
	}
	return err
}
func (v *RedisStore) Decr(ctx context.Context, p *limit_verified.DecrArgs) error {
	return v.store.Eval(ctx, redisScript.LimitVerifiedDecrSendCntScript,
		[]string{
			p.KeyPrefix,
			p.Target,
		},
	).Err()
}
