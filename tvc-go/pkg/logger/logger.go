package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	zerolog.Logger
}

func New(level string, format string) *Logger {
	var writer io.Writer

	if format == "console" {
		writer = zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		}
	} else {
		writer = os.Stderr
	}

	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	l := zerolog.New(writer).
		Level(lvl).
		With().
		Timestamp().
		Caller().
		Logger()

	return &Logger{Logger: l}
}

func Default() *Logger {
	return New("info", "console")
}
