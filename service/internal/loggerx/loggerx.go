package loggerx

import (
	"fmt"
	"io"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Ensure we always implement io.WriteCloser
var _ io.WriteCloser = (*Logger)(nil)

type Logger struct {
	*lumberjack.Logger
	mu      sync.Mutex
	enabled bool
}

func New(logger *lumberjack.Logger) *Logger {
	l := new(Logger)
	l.Logger = logger
	return l
}

// Write implements io.Writer
func (l *Logger) Write(p []byte) (int, error) {
	if l.enabled {
		if n, err := l.Logger.Write(p); err != nil {
			return n, err
		}
	}
	return fmt.Print(string(p))
}

// Close implements io.Closer
func (l *Logger) Close() error {
	if l.enabled {
		return l.Logger.Close()
	} else {
		return nil
	}
}

func (l *Logger) SetEnabled(value bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = value
}
