package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/thinkgos/limiter/limit"
)

const (
	seconds = time.Second
	quota   = 5
	total   = 100
)

func TestPeriodLimit_Take[S limit.PeriodStorage](t *testing.T, store S) {
	testPeriodLimit(
		t,
		store,
		limit.WithKeyPrefix("limit:period"),
		limit.WithPeriod(seconds),
		limit.WithQuota(quota),
	)
}

func TestPeriodLimit_TakeWithAlign[S limit.PeriodStorage](t *testing.T, store S) {
	testPeriodLimit(
		t,
		store,
		limit.WithKeyPrefix("limit:period"),
		limit.WithAlign(),
		limit.WithPeriod(seconds),
		limit.WithQuota(quota),
	)
}

func TestPeriodLimit_RedisUnavailable[S limit.PeriodStorage](t *testing.T, store S) {
	l := limit.NewPeriodLimit(store)
	val, err := l.Take(context.Background(), "first")
	assert.Error(t, err)
	assert.Equal(t, limit.PeriodLimitStsUnknown, val)
}

func testPeriodLimit[S limit.PeriodStorage](t *testing.T, store S, opts ...limit.PeriodLimitOption) {
	l := limit.NewPeriodLimit(store, opts...)
	var allowed, hitQuota, overQuota int
	for i := 0; i < total; i++ {
		val, err := l.Take(context.Background(), "first")
		assert.NoError(t, err)
		switch val {
		case limit.PeriodLimitStsAllowed:
			allowed++
		case limit.PeriodLimitStsHitQuota:
			hitQuota++
		case limit.PeriodLimitStsOverQuota:
			overQuota++
		case limit.PeriodLimitStsUnknown:
			fallthrough
		default:
			t.Error("unknown status")
		}
	}

	assert.Equal(t, quota-1, allowed)
	assert.Equal(t, 1, hitQuota)
	assert.Equal(t, total-quota, overQuota)
}

func TestPeriodLimit_QuotaFull[S limit.PeriodStorage](t *testing.T, store S) {
	l := limit.NewPeriodLimit(
		store,
		limit.WithPeriod(1),
		limit.WithQuota(1),
	)
	val, err := l.Take(context.Background(), "first")
	assert.NoError(t, err)
	assert.True(t, val.IsHitQuota())
}

func TestPeriodLimit_SetQuotaFull[S limit.PeriodStorage](t *testing.T, store S) {
	l := limit.NewPeriodLimit(store)

	err := l.SetQuotaFull(context.Background(), "first")
	assert.NoError(t, err)

	val, err := l.Take(context.Background(), "first")
	assert.NoError(t, err)
	assert.Equal(t, limit.PeriodLimitStsOverQuota, val)
}

func TestPeriodLimit_Del[S limit.PeriodStorage](t *testing.T, store S) {
	l := limit.NewPeriodLimit(
		store,
		limit.WithPeriod(seconds),
		limit.WithQuota(quota),
	)

	// 第一次, key不存在
	rv, err := l.GetRunValue(context.Background(), "first")
	assert.NoError(t, err)
	assert.False(t, rv.Exist)
	assert.Zero(t, rv.Count)
	assert.Zero(t, int(rv.TTL))

	runValue, err := l.GetRunValue(context.Background(), "first")
	assert.Nil(t, err)
	assert.Equal(t, runValue.Exist, false)
	assert.Equal(t, runValue.Count, int64(0))
	assert.Equal(t, runValue.TTL, time.Duration(0))

	err = l.SetQuotaFull(context.Background(), "first")
	assert.NoError(t, err)

	// 第二次, key 存在
	rv, err = l.GetRunValue(context.Background(), "first")
	assert.NoError(t, err)
	assert.True(t, rv.Exist)
	assert.Equal(t, int64(quota), rv.Count)
	assert.LessOrEqual(t, seconds, rv.TTL)

	runValue, err = l.GetRunValue(context.Background(), "first")
	assert.Nil(t, err)
	assert.Equal(t, runValue.Exist, true)
	assert.Equal(t, runValue.Count, int64(quota))
	assert.Equal(t, runValue.TTL, seconds)

	val, err := l.Take(context.Background(), "first")
	assert.NoError(t, err)
	assert.True(t, val.IsOverQuota())

	err = l.Del(context.Background(), "first")
	assert.NoError(t, err)

	val, err = l.Take(context.Background(), "first")
	assert.NoError(t, err)
	assert.True(t, val.IsAllowed())
}
