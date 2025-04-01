package tests

import (
	"context"
	"math/bits"
	"math/rand"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/thinkgos/limiter/verified"
)

const defaultKind = "default"
const unsupportedKind = "unsupported"

const question = "1+1"
const answer = "2"
const badAnswer = "3"

var defaultAlphabet = []byte("QWERTYUIOPLKJHGFDSAZXCVBNMabcdefghijklmnopqrstuvwxyz")

func randString(length int) string {
	b := make([]byte, length)
	bn := bits.Len(uint(len(defaultAlphabet)))
	mask := int64(1)<<bn - 1
	max := 63 / bn
	r := rand.New(rand.NewSource(time.Now().UnixNano() + rand.Int63() + rand.Int63()))

	// A rand.Int63() generates 63 random bits, enough for alphabets letters!
	for i, cache, remain := 0, r.Int63(), max; i < length; {
		if remain == 0 {
			cache, remain = r.Int63(), max
		}
		if idx := int(cache & mask); idx < len(defaultAlphabet) {
			b[i] = defaultAlphabet[idx]
			i++
		}
		cache >>= bn
		remain--
	}
	return string(b)
}

var _ verified.CaptchaProvider = (*TestVerifiedCaptchaProvider)(nil)

type TestVerifiedCaptchaProvider struct{}

func (t TestVerifiedCaptchaProvider) AcquireDriver(kind string) verified.CaptchaDriver {
	if kind == unsupportedKind {
		return new(verified.UnsupportedVerifiedCaptchaDriver)
	}
	return new(TestVerifiedCaptchaDriver)
}

type TestVerifiedCaptchaDriver struct{}

func (t TestVerifiedCaptchaDriver) Name() string { return "test_provider" }

func (t TestVerifiedCaptchaDriver) GenerateQuestionAnswer() (*verified.QuestionAnswer, error) {
	return &verified.QuestionAnswer{
		Id:       randString(6),
		Question: question,
		Answer:   answer,
	}, nil
}

func GenericTestCaptcha_Improve_Cover[S verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	l := verified.NewVerifiedCaptcha(
		new(TestVerifiedCaptchaProvider),
		store,
	)
	l.Name(defaultKind)
}

func GenericTestCaptcha_Unsupported_Driver[S verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	l := verified.NewVerifiedCaptcha(
		new(TestVerifiedCaptchaProvider),
		store,
	)

	_, _, err := l.Generate(context.Background(), unsupportedKind)
	assert.Error(t, err)
}

func GenericTestCaptcha_RedisUnavailable[S verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	l := verified.NewVerifiedCaptcha(
		new(TestVerifiedCaptchaProvider),
		store,
	)
	_, _, err := l.Generate(context.Background(), defaultKind)
	assert.Error(t, err)
}

func GenericTestCaptcha_OneTime[S verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	l := verified.NewVerifiedCaptcha(
		new(TestVerifiedCaptchaProvider),
		store,
		verified.WithKeyPrefix("verified:captcha:"),
		verified.WithKeyExpires(time.Minute*3),
	)

	id, _, err := l.Generate(context.Background(), defaultKind, verified.WithGenerateKeyExpires(time.Minute*5))
	assert.NoError(t, err)

	b := l.Verify(context.Background(), defaultKind, id, answer)
	require.True(t, b)

	b = l.Verify(context.Background(), defaultKind, id, answer)
	require.False(t, b)
}

func GenericTestCaptcha_In_Quota[S verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	l := verified.NewVerifiedCaptcha(
		new(TestVerifiedCaptchaProvider),
		store,
		verified.WithKeyPrefix("verified:captcha:"),
		verified.WithKeyExpires(time.Minute*3),
		verified.WithMaxErrQuota(3),
	)

	id, _, err := l.Generate(context.Background(), defaultKind,
		verified.WithGenerateKeyExpires(time.Minute*5),
	)
	assert.NoError(t, err)

	b := l.Verify(context.Background(), defaultKind, id, badAnswer)
	require.False(t, b)
	b = l.Verify(context.Background(), defaultKind, id, badAnswer)
	require.False(t, b)
	b = l.Verify(context.Background(), defaultKind, id, answer)
	require.True(t, b)
}

func GenericTestCaptcha_Over_Quota[S verified.Storage](t *testing.T, _ *miniredis.Miniredis, store S) {
	l := verified.NewVerifiedCaptcha(
		new(TestVerifiedCaptchaProvider),
		store,
		verified.WithKeyPrefix("verified:captcha:"),
		verified.WithKeyExpires(time.Minute*3),
		verified.WithMaxErrQuota(3),
	)

	id, _, err := l.Generate(context.Background(), defaultKind,
		verified.WithGenerateKeyExpires(time.Minute*5),
		verified.WithGenerateMaxErrQuota(6),
	)
	assert.NoError(t, err)

	for i := 0; i < 6; i++ {
		b := l.Verify(context.Background(), defaultKind, id, badAnswer)
		require.False(t, b)
	}
	b := l.Verify(context.Background(), defaultKind, id, answer)
	require.False(t, b)
}

func GenericTestCaptcha_Onetime_Timeout[S verified.Storage](t *testing.T, mr *miniredis.Miniredis, store S) {
	l := verified.NewVerifiedCaptcha(
		new(TestVerifiedCaptchaProvider),
		store,
	)
	id, _, err := l.Generate(context.Background(), defaultKind, verified.WithGenerateKeyExpires(time.Second*1))
	assert.NoError(t, err)

	mr.FastForward(time.Second * 2) // time.Sleep(time.Second * 2)

	b := l.Verify(context.Background(), defaultKind, id, "2")
	require.False(t, b)
}
