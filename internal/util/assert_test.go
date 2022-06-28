package util_test

import (
	"testing"

	"github.com/chrisyxlee/snippets/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestXxx(t *testing.T) {
	assert.Panics(t, func() {
		util.Assert(false, "some output")
	})

	assert.NotPanics(t, func() {
		util.Assert(true, "yay")
	})
}
