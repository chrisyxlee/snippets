package ratelimit

import (
	"github.com/benbjohnson/clock"
	"github.com/chrisyxlee/snippets/internal"
	"github.com/google/go-github/v45/github"
)

var clk = clock.New()

// Set the clock to something else, say the mock clock, for tests.
func SetClock(c clock.Clock) {
	clk = c
}

// WaitIfRateLimited will return true if the error was a rate limit. If the error
// was a rate limit, it will also block until the rate limit is cleared. If the
// error was not a rate limit, then this will return false.
func WaitIfRateLimited(err error) bool {
	rlErr, ok := err.(*github.RateLimitError)
	if !ok {
		return false
	}

	if dur := clk.Until(rlErr.Rate.Reset.Time); dur > 0 {
		internal.Log().Info().Dur("duration", dur).Time("time", rlErr.Rate.Reset.Time).Msg("waiting for rate limit to continue")
		clk.Sleep(dur)
	}
	return true
}
