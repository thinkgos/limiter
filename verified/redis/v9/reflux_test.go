package v9

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thinkgos/limiter/verified/tests"
)

func TestReflux_Improve_Cover(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()
	tests.GenericTestReflux_Improve_Cover(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func TestReflux_RedisUnavailable(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	addr := mr.Addr()
	mr.Close()
	tests.GenericTestReflux_RedisUnavailable(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: addr})),
	)
}

func TestReflux_One_Time(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	tests.GenericTestReflux_One_Time(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func TestReflux_In_Quota(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	tests.GenericTestReflux_In_Quota(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func TestReflux_Over_Quota(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	tests.GenericTestReflux_Over_Quota(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func TestReflux_OneTime_Timeout(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	tests.GenericTestReflux_OneTime_Timeout(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}
