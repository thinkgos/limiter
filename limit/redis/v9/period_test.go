package v9

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"github.com/thinkgos/limiter/limit/tests"
)

func TestPeriodLimit_Take(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	tests.TestPeriodLimit_Take(
		t,
		NewPeriodStore(
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
		NewPeriodStore(
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
		NewPeriodStore(
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
		NewPeriodStore(
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
		NewPeriodStore(
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
		NewPeriodStore(
			redis.NewClient(&redis.Options{Addr: mr.Addr()}),
		),
	)
}
