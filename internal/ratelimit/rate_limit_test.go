package ratelimit_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/chrisyxlee/snippets/internal/ratelimit"
	"github.com/google/go-github/v45/github"
	"github.com/stretchr/testify/assert"
)

func TestRateLimit(t *testing.T) {
	t.Parallel()

	mc := clock.NewMock()
	ratelimit.SetClock(mc)

	t.Run("not rate limited", func(t *testing.T) {
		t.Parallel()

		assert.False(t, ratelimit.WaitIfRateLimited(fmt.Errorf("blah")))
	})

	t.Run("rate limit time already passed", func(t *testing.T) {
		t.Parallel()
		assert.True(t, ratelimit.WaitIfRateLimited(&github.RateLimitError{
			Rate: github.Rate{
				Limit:     100,
				Remaining: 0,
				Reset: github.Timestamp{
					Time: mc.Now().Add(-5 * time.Minute),
				},
			},
		}))
	})

	t.Run("rate limited", func(t *testing.T) {
		t.Parallel()

		mc := clock.NewMock()
		ratelimit.SetClock(mc)

		now := mc.Now()

		go func() {
			mc.Set(now.Add(10 * time.Minute))
		}()

		assert.True(t, ratelimit.WaitIfRateLimited(&github.RateLimitError{
			Rate: github.Rate{
				Limit:     100,
				Remaining: 0,
				Reset: github.Timestamp{
					Time: now.Add(5 * time.Minute),
				},
			},
		}))
	})
}
