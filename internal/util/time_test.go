package util_test

import (
	"testing"
	"time"

	"github.com/chrisyxlee/snippets/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTime(t *testing.T) {
	t.Parallel()

	now := time.Now().Truncate(time.Minute)
	formatted := util.FormatLocalTime(now)

	got, err := util.ParseLocalTime(formatted)
	assert.NoError(t, err)
	assert.Equal(t, now, got)

	assert.NotPanics(t, func() {
		got = util.ParseLocalTimeOrDie(formatted)
	})
	assert.Equal(t, now, got)
}

func TestInTimeRange(t *testing.T) {
	t.Parallel()

	now := time.Now()

	assert.True(t, util.InTimeRange(now, now, now.Add(time.Minute)))
	assert.True(t, util.InTimeRange(now, now.Add(-time.Minute), now.Add(time.Second)))
	assert.False(t, util.InTimeRange(now, now.Add(-time.Minute), now))
}

func TestDate(t *testing.T) {
	t.Parallel()

	parsed, err := time.ParseInLocation("2006-01-02", "2222-12-30", time.Local)
	require.NoError(t, err)
	assert.Equal(t, "Dec 30", util.FormatLocalDate(parsed))
}
