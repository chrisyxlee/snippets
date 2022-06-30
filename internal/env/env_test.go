package env_test

import (
	"os"
	"testing"

	"github.com/chrisyxlee/snippets/internal/env"
	"github.com/stretchr/testify/assert"
)

func TestGetenvOrDefault(t *testing.T) {
	key := "SOME_KEY"

	assert.Equal(t, env.GetenvOrDefault(key, "default"), "default")

	os.Setenv(key, "some value")
	defer os.Unsetenv(key)

	assert.Equal(t, env.GetenvOrDefault(key, "default"), "some value")
}

func TestGetenvOrDie(t *testing.T) {
	key := "SOME_KEY"

	assert.Panics(t, func() { env.GetenvOrDie(key, "message to print") })

	os.Setenv(key, "some value")
	defer os.Unsetenv(key)

	assert.Equal(t, env.GetenvOrDie(key, "message to print"), "some value")
}
