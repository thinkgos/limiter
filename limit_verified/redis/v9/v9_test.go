package v9

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/thinkgos/limiter/limit_verified/tests"
)

func Test_RedisV9_Name(t *testing.T) {
	tests.GenericTestName(
		t,
		nil,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})),
	)
}

func Test_RedisV9_SendCode_RedisUnavailable(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	addr := mr.Addr()
	mr.Close()
	tests.GenericTestSendCode_RedisUnavailable(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: addr})),
	)
}

func Test_RedisV9_SendCode_Success(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestSendCode_Success(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9_SendCode_Err_Failure(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestSendCode_Err_Failure(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9_SendCode_MaxSendPerDay(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestSendCode_MaxSendPerDay(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9_SendCode_Concurrency_MaxSendPerDay(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestSendCode_Concurrency_MaxSendPerDay(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9_SendCode_ResendTooFrequently(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestSendCode_ResendTooFrequently(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9_SendCode_Concurrency_ResendTooFrequently(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestSendCode_Concurrency_ResendTooFrequently(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9_VerifyCode_Success(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestVerifyCode_Success(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9_VerifyCode_CodeRequired(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestVerifyCode_CodeRequired(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9_VerifyCode_CodeExpired(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestVerifyCode_CodeExpired(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9_VerifyCode_CodeMaxError(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestVerifyCode_CodeMaxError(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9_VerifyCode_Concurrency_CodeMaxError(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestVerifyCode_Concurrency_CodeMaxError(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9__INCR_MaxSendPerDay(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTest_INCR_MaxSendPerDay(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV9__INCR_DECR(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTest_INCR_DECR(
		t,
		mr,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}
