package log

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.DebugLevel)

	Log = &l
}

var Log *zerolog.Logger

func IsSilent() bool {
	return Log.GetLevel() == zerolog.Disabled
}

func SetLevel(level string) error {
	lv, err := zerolog.ParseLevel(strings.ToLower(level))
	if err != nil {
		return err
	}

	l := Log.Level(lv)
	Log = &l
	Log.Debug().Str("lvl", level).Send()
	return nil
}
