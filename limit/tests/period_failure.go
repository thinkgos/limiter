package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/thinkgos/limiter/limit"
)

var errInternal = errors.New("internal error")

func TestPeriodFailureLimit_Check[S limit.PeriodFailureStorage](t *testing.T, store S) {
	TestPeriodFailureLimit(
		t,
		store,
		limit.WithKeyPrefix("limit:period:failure:"),
		limit.WithPeriod(seconds),
		limit.WithQuota(quota),
	)
}

func TestPeriodFailureLimit_CheckWithAlign[S limit.PeriodFailureStorage](t *testing.T, store S) {
	TestPeriodFailureLimit(
		t,
		store,
		limit.WithAlign(),
		limit.WithKeyPrefix("limit:period:failure:"),
		limit.WithAlign(),
		limit.WithPeriod(seconds),
		limit.WithQuota(quota),
	)
}

func TestPeriodFailureLimit_RedisUnavailable[S limit.PeriodFailureStorage](t *testing.T, store S) {
	l := limit.NewPeriodFailureLimit(store)
	sts, err := l.CheckErr(context.Background(), "first", nil)
	assert.Error(t, err)
	assert.Equal(t, limit.PeriodFailureLimitStsUnknown, sts)
}

func TestPeriodFailureLimit[S limit.PeriodFailureStorage](t *testing.T, store S, opts ...limit.PeriodLimitOption) {
	l := limit.NewPeriodFailureLimit(
		store,
		opts...,
	)
	var inLimitCnt, overFailureTimeCnt int
	for i := 0; i < total; i++ {
		sts, err := l.CheckErr(context.Background(), "first", errInternal)
		assert.NoError(t, err)
		switch sts { // nolint: exhaustive
		case limit.PeriodFailureLimitStsInQuota:
			inLimitCnt++
		case limit.PeriodFailureLimitStsOverQuota:
			overFailureTimeCnt++
		default:
			t.Errorf("unknown status, must be on of [%d, %d]", limit.PeriodFailureLimitStsInQuota, limit.PeriodFailureLimitStsOverQuota)
		}
	}
	assert.Equal(t, quota, inLimitCnt)
	assert.Equal(t, total-quota, overFailureTimeCnt)

	sts, err := l.CheckErr(context.Background(), "first", nil)
	assert.NoError(t, err)
	assert.Equal(t, limit.PeriodFailureLimitStsOverQuota, sts)
}

func TestPeriodFailureLimit_Check_In_Limit_Failure_Time_Then_Success[S limit.PeriodFailureStorage](t *testing.T, store S) {
	l := limit.NewPeriodFailureLimit(
		store,
	)
	var inLimitCnt, overFailureTimeCnt int
	for i := 0; i < quota-1; i++ {
		sts, err := l.CheckErr(context.Background(), "first", errInternal)
		assert.NoError(t, err)
		switch sts { // nolint: exhaustive
		case limit.PeriodFailureLimitStsInQuota:
			inLimitCnt++
		case limit.PeriodFailureLimitStsOverQuota:
			overFailureTimeCnt++
		default:
			t.Errorf("unknown status, must be on of [%d, %d]", limit.PeriodFailureLimitStsInQuota, limit.PeriodFailureLimitStsOverQuota)
		}
	}
	assert.Equal(t, quota-1, inLimitCnt)
	assert.Equal(t, 0, overFailureTimeCnt)

	sts, err := l.CheckErr(context.Background(), "first", nil)
	assert.NoError(t, err)
	assert.Equal(t, limit.PeriodFailureLimitStsSuccess, sts)

	rv, err := l.GetRunValue(context.Background(), "first")
	assert.NoError(t, err)
	assert.False(t, rv.Exist)
	assert.Zero(t, rv.Count)
}

func TestPeriodFailureLimit_Check_Over_Limit_Failure_Time_Then_Success_Always_OverFailureTimeError[S limit.PeriodFailureStorage](t *testing.T, store S) {
	l := limit.NewPeriodFailureLimit(
		store,
		limit.WithQuota(quota),
	)
	var inLimitCnt, overFailureTimeCnt int
	for i := 0; i < quota+1; i++ {
		sts, err := l.CheckErr(context.Background(), "first", errInternal)
		assert.NoError(t, err)
		switch sts { // nolint: exhaustive
		case limit.PeriodFailureLimitStsInQuota:
			inLimitCnt++
		case limit.PeriodFailureLimitStsOverQuota:
			overFailureTimeCnt++
		default:
			t.Errorf("unknown status, must be on of [%d, %d]", limit.PeriodFailureLimitStsInQuota, limit.PeriodFailureLimitStsOverQuota)
		}
	}
	assert.Equal(t, quota, inLimitCnt)
	assert.Equal(t, 1, overFailureTimeCnt)

	sts, err := l.CheckErr(context.Background(), "first", nil)
	assert.NoError(t, err)
	assert.Equal(t, limit.PeriodFailureLimitStsOverQuota, sts)

	rv, err := l.GetRunValue(context.Background(), "first")
	assert.NoError(t, err)
	assert.True(t, rv.Exist)
	assert.Equal(t, int64(quota+1), rv.Count)
}

func TestPeriodFailureLimit_SetQuotaFull[S limit.PeriodFailureStorage](t *testing.T, store S) {
	l := limit.NewPeriodFailureLimit(store)

	err := l.SetQuotaFull(context.Background(), "first")
	assert.Nil(t, err)

	sts, err := l.CheckErr(context.Background(), "first", nil)
	assert.NoError(t, err)
	assert.Equal(t, limit.PeriodFailureLimitStsOverQuota, sts)
}

func TestPeriodFailureLimit_Del[S limit.PeriodFailureStorage](t *testing.T, store S) {
	l := limit.NewPeriodFailureLimit(
		store,
		limit.WithPeriod(seconds),
		limit.WithQuota(quota),
	)

	// 第一次, key不存在
	rv, err := l.GetRunValue(context.Background(), "first")
	assert.Nil(t, err)
	assert.False(t, rv.Exist)
	assert.Zero(t, rv.Count)
	assert.Zero(t, int(rv.TTL))

	runValue, err := l.GetRunValue(context.Background(), "first")
	assert.Nil(t, err)
	assert.Equal(t, runValue.Exist, false)
	assert.Equal(t, runValue.Count, int64(0))
	assert.Equal(t, runValue.TTL, time.Duration(0))

	err = l.SetQuotaFull(context.Background(), "first")
	assert.Nil(t, err)

	// 第二次, key 存在
	rv, err = l.GetRunValue(context.Background(), "first")
	assert.Nil(t, err)
	assert.Equal(t, int64(quota), rv.Count)
	assert.LessOrEqual(t, seconds, rv.TTL)

	runValue, err = l.GetRunValue(context.Background(), "first")
	assert.Nil(t, err)
	assert.Equal(t, runValue.Exist, true)
	assert.Equal(t, runValue.Count, int64(quota))
	assert.Equal(t, runValue.TTL, seconds)

	sts, err := l.CheckErr(context.Background(), "first", errInternal)
	assert.NoError(t, err)
	assert.Equal(t, limit.PeriodFailureLimitStsOverQuota, sts)
	assert.True(t, sts.IsOverQuota())

	err = l.Del(context.Background(), "first")
	assert.Nil(t, err)

	sts, err = l.CheckErr(context.Background(), "first", errInternal)
	assert.NoError(t, err)
	assert.Equal(t, limit.PeriodFailureLimitStsInQuota, sts)
	assert.True(t, sts.IsWithinQuota())

	sts, err = l.CheckErr(context.Background(), "first", nil)
	assert.NoError(t, err)
	assert.Equal(t, limit.PeriodFailureLimitStsSuccess, sts)
	assert.True(t, sts.IsSuccess())
}
