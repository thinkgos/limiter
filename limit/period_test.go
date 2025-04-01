package limit_test

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	redisV9 "github.com/thinkgos/limiter/limit/redis/v9"
	"github.com/thinkgos/limiter/limit/tests"
)

func TestPeriodLimit_Take(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	tests.TestPeriodLimit_Take(
		t,
		redisV9.NewPeriodStore(
			redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		),
	)
}

func TestPeriodLimit_TakeWithAlign(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	tests.TestPeriodLimit_TakeWithAlign(
		t,
		redisV9.NewPeriodStore(
			redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		),
	)
}

func TestPeriodLimit_RedisUnavailable(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	addr := mr.Addr()

	mr.Close()
	tests.TestPeriodLimit_RedisUnavailable(
		t,
		redisV9.NewPeriodStore(
			redis.NewClient(&redis.Options{Addr: addr}),
		),
	)
}

func TestPeriodLimit_QuotaFull(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	tests.TestPeriodLimit_QuotaFull(
		t,
		redisV9.NewPeriodStore(
			redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		),
	)
}

func TestPeriodLimit_SetQuotaFull(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	tests.TestPeriodLimit_SetQuotaFull(
		t,
		redisV9.NewPeriodStore(
			redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		),
	)
}

func TestPeriodLimit_Del(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	tests.TestPeriodLimit_Del(
		t,
		redisV9.NewPeriodStore(
			redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		),
	)
}
