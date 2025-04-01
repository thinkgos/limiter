package tests

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/thinkgos/limiter/limit_verified"
)

const (
	target  = "112233"
	code    = "123456"
	badCode = "654321"
)

var _ limit_verified.LimitVerifiedProvider = (*TestProvider)(nil)

type TestProvider struct{}

func (t TestProvider) Name() string { return "test_provider" }

func (t TestProvider) SendCode(c limit_verified.CodeParam) error { return nil }

type TestErrProvider struct{}

func (t TestErrProvider) Name() string { return "test_provider" }

func (t TestErrProvider) SendCode(c limit_verified.CodeParam) error {
	return errors.New("发送失败")
}

func GenericTestName[S limit_verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	l := limit_verified.NewLimitVerified(
		new(TestProvider),
		store,
	)
	require.Equal(t, "test_provider", l.Name())
}

func GenericTestSendCode_RedisUnavailable[S limit_verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	l := limit_verified.NewLimitVerified(
		new(TestProvider),
		store,
	)

	err := l.SendCode(context.Background(), limit_verified.CodeParam{Target: target, Code: code})
	assert.NotNil(t, err)
}

func GenericTestSendCode_Success[S limit_verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	l := limit_verified.NewLimitVerified(
		new(TestProvider),
		store,
		limit_verified.WithKeyPrefix("verification"),
		limit_verified.WithKeyExpires(time.Hour),
	)
	err := l.SendCode(
		context.Background(),
		limit_verified.CodeParam{Target: target, Code: code},
		limit_verified.WithAvailWindowSecond(3),
	)
	require.NoError(t, err)
}

func GenericTestSendCode_Err_Failure[S limit_verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	l := limit_verified.NewLimitVerified(
		new(TestErrProvider),
		store,
		limit_verified.WithKeyPrefix("verification:err"),
		limit_verified.WithKeyExpires(time.Hour),
	)
	err := l.SendCode(context.Background(), limit_verified.CodeParam{Target: target, Code: code})
	require.Error(t, err)
}

func GenericTestSendCode_MaxSendPerDay[S limit_verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	l := limit_verified.NewLimitVerified(
		new(TestProvider),
		store,
		limit_verified.WithMaxSendPerDay(1),
		limit_verified.WithCodeMaxSendPerDay(1),
		limit_verified.WithCodeResendIntervalSecond(1),
	)

	err := l.SendCode(context.Background(), limit_verified.CodeParam{Target: target, Code: code})
	require.NoError(t, err)

	time.Sleep(time.Second + time.Millisecond*10)
	err = l.SendCode(context.Background(), limit_verified.CodeParam{Target: target, Code: code})
	require.ErrorIs(t, err, limit_verified.ErrMaxSendPerDay)
}

func GenericTestSendCode_Concurrency_MaxSendPerDay[S limit_verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	var success uint32
	var failed uint32

	l := limit_verified.NewLimitVerified(new(TestProvider),
		store,
		limit_verified.WithMaxSendPerDay(1),
	)

	wg := &sync.WaitGroup{}
	wg.Add(15)
	for i := 0; i < 15; i++ {
		go func() {
			defer wg.Done()

			err := l.SendCode(context.Background(), limit_verified.CodeParam{Target: target, Code: code})
			if err != nil {
				require.ErrorIs(t, err, limit_verified.ErrMaxSendPerDay)
				atomic.AddUint32(&failed, 1)
			} else {
				atomic.AddUint32(&success, 1)
			}
		}()
	}

	wg.Wait()
	require.Equal(t, uint32(1), success)
	require.Equal(t, uint32(14), failed)
}

func GenericTestSendCode_ResendTooFrequently[S limit_verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	l := limit_verified.NewLimitVerified(
		new(TestProvider),
		store,
	)

	err := l.SendCode(context.Background(), limit_verified.CodeParam{Target: target, Code: code}, limit_verified.WithResendIntervalSecond(1))
	require.NoError(t, err)
	err = l.SendCode(context.Background(), limit_verified.CodeParam{Target: target, Code: code}, limit_verified.WithResendIntervalSecond(1))
	require.ErrorIs(t, err, limit_verified.ErrResendTooFrequently)
}

func GenericTestSendCode_Concurrency_ResendTooFrequently[S limit_verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	var success uint32
	var failed uint32

	l := limit_verified.NewLimitVerified(
		new(TestProvider),
		store,
		limit_verified.WithCodeResendIntervalSecond(3),
	)

	wg := &sync.WaitGroup{}
	wg.Add(15)
	for i := 0; i < 15; i++ {
		go func() {
			defer wg.Done()

			err := l.SendCode(context.Background(), limit_verified.CodeParam{Target: target, Code: code})
			if err != nil {
				require.ErrorIs(t, err, limit_verified.ErrResendTooFrequently)
				atomic.AddUint32(&failed, 1)
			} else {
				atomic.AddUint32(&success, 1)
			}
		}()
	}

	wg.Wait()
	require.Equal(t, uint32(1), success)
	require.Equal(t, uint32(14), failed)
}

func GenericTestVerifyCode_Success[S limit_verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	l := limit_verified.NewLimitVerified(
		new(TestProvider),
		store,
	)

	err := l.SendCode(context.Background(), limit_verified.CodeParam{Target: target, Code: code})
	require.Nil(t, err)

	err = l.VerifyCode(context.Background(), limit_verified.CodeParam{Target: target, Code: code})
	assert.NoError(t, err)
}

func GenericTestVerifyCode_CodeRequired[S limit_verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	l := limit_verified.NewLimitVerified(
		new(TestProvider),
		store,
	)

	err := l.VerifyCode(context.Background(), limit_verified.CodeParam{Target: target, Code: code})
	assert.ErrorIs(t, err, limit_verified.ErrCodeRequiredOrExpired)
}

func GenericTestVerifyCode_CodeExpired[S limit_verified.Storage](t *testing.T, mr *miniredis.Miniredis, store S) {
	l := limit_verified.NewLimitVerified(
		new(TestProvider),
		store,
	)
	err := l.SendCode(context.Background(), limit_verified.CodeParam{Target: target, Code: code}, limit_verified.WithAvailWindowSecond(1))
	require.Nil(t, err)

	mr.FastForward(time.Second * 3) // time.Sleep(time.Second * 3)
	err = l.VerifyCode(context.Background(), limit_verified.CodeParam{Target: target, Code: code})
	assert.ErrorIs(t, err, limit_verified.ErrCodeRequiredOrExpired)
}

func GenericTestVerifyCode_CodeMaxError[S limit_verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	l := limit_verified.NewLimitVerified(
		new(TestProvider),
		store,
	)
	err := l.SendCode(context.Background(), limit_verified.CodeParam{Target: target, Code: code}, limit_verified.WithMaxErrorQuota(6))
	require.Nil(t, err)

	for i := 0; i < 6; i++ {
		err = l.VerifyCode(context.Background(), limit_verified.CodeParam{Target: target, Code: badCode})
		assert.ErrorIs(t, err, limit_verified.ErrCodeVerification)
	}
	err = l.VerifyCode(context.Background(), limit_verified.CodeParam{Target: target, Code: badCode})
	assert.ErrorIs(t, err, limit_verified.ErrCodeMaxErrorQuota)
}

func GenericTestVerifyCode_Concurrency_CodeMaxError[S limit_verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	var failedMaxError uint32
	var failedVerify uint32

	l := limit_verified.NewLimitVerified(
		new(TestProvider),
		store,
		limit_verified.WithCodeMaxErrorQuota(3),
		limit_verified.WithCodeAvailWindowSecond(180),
	)

	err := l.SendCode(context.Background(), limit_verified.CodeParam{Target: target, Code: code})
	require.Nil(t, err)

	wg := &sync.WaitGroup{}
	wg.Add(15)
	for i := 0; i < 15; i++ {
		go func() {
			defer wg.Done()

			err := l.VerifyCode(context.Background(), limit_verified.CodeParam{Target: target, Code: badCode})
			if err != nil {
				if errors.Is(err, limit_verified.ErrCodeMaxErrorQuota) {
					atomic.AddUint32(&failedMaxError, 1)
				} else {
					atomic.AddUint32(&failedVerify, 1)
				}
			}
		}()
	}

	wg.Wait()
	require.Equal(t, uint32(3), failedVerify)
	require.Equal(t, uint32(12), failedMaxError)
}

func GenericTest_INCR_MaxSendPerDay[S limit_verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	l := limit_verified.NewLimitVerified(
		new(TestProvider),
		store,
		limit_verified.WithMaxSendPerDay(10),
	)
	for i := 0; i < 10; i++ {
		err := l.Incr(context.Background(), target)
		require.Nil(t, err)
	}
	err := l.Incr(context.Background(), target)
	require.ErrorIs(t, err, limit_verified.ErrMaxSendPerDay)
}

func GenericTest_INCR_DECR[S limit_verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	l := limit_verified.NewLimitVerified(
		new(TestProvider),
		store,
	)
	err := l.Incr(context.Background(), target)
	require.Nil(t, err)

	err = l.Decr(context.Background(), target)
	require.Nil(t, err)
}
