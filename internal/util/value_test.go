package util_test

import (
	"testing"

	"github.com/chrisyxlee/snippets/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestNewValue(t *testing.T) {
	t.Parallel()

	assert.Equal(t, int64(0), *util.NewValue(int64(0)))
	assert.Equal(t, "bleh", *util.NewValue("bleh"))
	assert.InDelta(t, float64(0.123), *util.NewValue(float64(0.123)), 0.00001)
}
