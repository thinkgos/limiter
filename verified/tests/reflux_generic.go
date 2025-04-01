package tests

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/thinkgos/limiter/verified"
)

var _ verified.RefluxProvider = (*TestVerifiedRefluxProvider)(nil)

type TestVerifiedRefluxProvider struct{}

func (t TestVerifiedRefluxProvider) Name() string { return "test_provider" }

func (t TestVerifiedRefluxProvider) GenerateUniqueId() string {
	return randString(6)
}

func GenericTestReflux_Improve_Cover[S verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	l := verified.NewVerifiedReflux(
		new(TestVerifiedRefluxProvider),
		store,
	)
	l.Name()
}

func GenericTestReflux_RedisUnavailable[S verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	l := verified.NewVerifiedReflux(
		new(TestVerifiedRefluxProvider),
		store,
	)

	randKey := randString(6)
	_, err := l.Generate(context.Background(), defaultKind, randKey)
	assert.Error(t, err)
}

func GenericTestReflux_One_Time[S verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	l := verified.NewVerifiedReflux(
		new(TestVerifiedRefluxProvider),
		store,
		verified.WithKeyPrefix("verified:reflux:"),
		verified.WithKeyExpires(time.Minute*3),
	)

	randKey := randString(6)

	value, err := l.Generate(context.Background(), defaultKind, randKey, verified.WithGenerateKeyExpires(time.Minute*5))
	assert.NoError(t, err)

	b := l.Verify(context.Background(), defaultKind, randKey, value)
	require.True(t, b)

	b = l.Verify(context.Background(), defaultKind, randKey, value)
	require.False(t, b)
}

func GenericTestReflux_In_Quota[S verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	l := verified.NewVerifiedReflux(
		new(TestVerifiedRefluxProvider),
		store,
		verified.WithKeyPrefix("verified:reflux:"),
		verified.WithKeyExpires(time.Minute*3),
		verified.WithMaxErrQuota(3),
	)

	randKey := randString(6)
	value, err := l.Generate(context.Background(), defaultKind, randKey, verified.WithGenerateKeyExpires(time.Minute*5))
	assert.NoError(t, err)

	badValue := value + "xxx"

	b := l.Verify(context.Background(), defaultKind, randKey, badValue)
	require.False(t, b)
	b = l.Verify(context.Background(), defaultKind, randKey, badValue)
	require.False(t, b)
	b = l.Verify(context.Background(), defaultKind, randKey, value)
	require.True(t, b)
}

func GenericTestReflux_Over_Quota[S verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	l := verified.NewVerifiedReflux(
		new(TestVerifiedRefluxProvider),
		store,
		verified.WithKeyPrefix("verified:reflux:"),
		verified.WithKeyExpires(time.Minute*3),
		verified.WithMaxErrQuota(3),
	)

	randKey := randString(6)
	value, err := l.Generate(context.Background(), defaultKind, randKey,
		verified.WithGenerateKeyExpires(time.Minute*5),
		verified.WithGenerateMaxErrQuota(6),
	)
	assert.NoError(t, err)

	badValue := value + "xxx"

	for i := 0; i < 6; i++ {
		b := l.Verify(context.Background(), defaultKind, randKey, badValue)
		require.False(t, b)
	}
	b := l.Verify(context.Background(), defaultKind, randKey, value)
	require.False(t, b)
}

func GenericTestReflux_OneTime_Timeout[S verified.Storage](t *testing.T, mr *miniredis.Miniredis, store S) {
	l := verified.NewVerifiedReflux(
		new(TestVerifiedRefluxProvider),
		store,
	)
	randKey := randString(6)
	value, err := l.Generate(context.Background(), defaultKind, randKey, verified.WithGenerateKeyExpires(time.Second*1))
	assert.NoError(t, err)

	mr.FastForward(time.Second) // time.Sleep(time.Second * 2)

	b := l.Verify(context.Background(), defaultKind, randKey, value)
	require.False(t, b)
}
