package v8

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"

	"github.com/thinkgos/limiter/limit/tests"
)

func TestPeriodFailureLimit_Check(t *testing.T) {
	mr, err := miniredis.Run()
	assert.Nil(t, err)

	defer mr.Close()
	tests.TestPeriodFailureLimit_Check(t, NewPeriodFailureStore(
		redis.NewClient(&redis.Options{Addr: mr.Addr()}),
	))
}

func TestPeriodFailureLimit_CheckWithAlign(t *testing.T) {
	mr, err := miniredis.Run()
	assert.Nil(t, err)

	defer mr.Close()
	tests.TestPeriodFailureLimit_CheckWithAlign(t, NewPeriodFailureStore(
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
		NewPeriodFailureStore(
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
		NewPeriodFailureStore(
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
		NewPeriodFailureStore(
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
		NewPeriodFailureStore(
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
		NewPeriodFailureStore(
			redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		),
	)
}
