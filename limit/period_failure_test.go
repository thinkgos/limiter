package limit_test

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	redisV9 "github.com/thinkgos/limiter/limit/redis/v9"
	"github.com/thinkgos/limiter/limit/tests"
)

func TestPeriodFailureLimit_Check(t *testing.T) {
	mr, err := miniredis.Run()
	assert.Nil(t, err)

	defer mr.Close()
	tests.TestPeriodFailureLimit_Check(t, redisV9.NewPeriodFailureStore(
		redis.NewClient(&redis.Options{Addr: mr.Addr()}),
	))
}

func TestPeriodFailureLimit_CheckWithAlign(t *testing.T) {
	mr, err := miniredis.Run()
	assert.Nil(t, err)

	defer mr.Close()
	tests.TestPeriodFailureLimit_CheckWithAlign(t, redisV9.NewPeriodFailureStore(
		redis.NewClient(&redis.Options{Addr: mr.Addr()}),
	))
}

func TestPeriodFailureLimit_RedisUnavailable(t *testing.T) {
	mr, err := miniredis.Run()
	assert.Nil(t, err)
	addr := mr.Addr()
	mr.Close()

	tests.TestPeriodFailureLimit_RedisUnavailable(
		t,
		redisV9.NewPeriodFailureStore(
			redis.NewClient(&redis.Options{Addr: addr}),
		),
	)
}

func TestPeriodFailureLimit_Check_In_Limit_Failure_Time_Then_Success(t *testing.T) {
	mr, err := miniredis.Run()
	assert.Nil(t, err)

	defer mr.Close()

	tests.TestPeriodFailureLimit_Check_In_Limit_Failure_Time_Then_Success(
		t,
		redisV9.NewPeriodFailureStore(
			redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		),
	)
}

func TestPeriodFailureLimit_Check_Over_Limit_Failure_Time_Then_Success_Always_OverFailureTimeError(t *testing.T) {
	mr, err := miniredis.Run()
	assert.Nil(t, err)

	defer mr.Close()

	tests.TestPeriodFailureLimit_Check_Over_Limit_Failure_Time_Then_Success_Always_OverFailureTimeError(
		t,
		redisV9.NewPeriodFailureStore(
			redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		),
	)
}

func TestPeriodFailureLimit_SetQuotaFull(t *testing.T) {
	mr, err := miniredis.Run()
	assert.Nil(t, err)
	defer mr.Close()

	tests.TestPeriodFailureLimit_SetQuotaFull(
		t,
		redisV9.NewPeriodFailureStore(
			redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		),
	)
}

func TestPeriodFailureLimit_Del(t *testing.T) {
	mr, err := miniredis.Run()
	assert.Nil(t, err)
	defer mr.Close()

	tests.TestPeriodFailureLimit_Del(
		t,
		redisV9.NewPeriodFailureStore(
			redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		),
	)
}
