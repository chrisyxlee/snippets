package internal

import (
	"github.com/chrisyxlee/snippets/internal/log"
	"github.com/rs/zerolog"
)

func Log() *zerolog.Logger {
	return log.Log
}
