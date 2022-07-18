package logger

import "log"

var _ Logger = &Log{}

// defaultLog is singleton representing default logger of Go log package.
var defaultLog = NewLog(log.Default())

// Log is Logger using standard logger provided by Go standard log package.
type Log struct {
	l *log.Logger
}

// NewLog creates new Log using l for logging.
func NewLog(l *log.Logger) *Log { return &Log{l: l} }

// NewDefaultLog returns instance of Log using default logger of Go standard log
// package.
func NewDefaultLog() *Log { return defaultLog }

func (l *Log) Errorf(format string, args ...interface{}) {
	l.l.Printf("error: "+format, args...)
}

func (l *Log) Warnf(format string, args ...interface{}) {
	l.l.Printf("warning: "+format, args...)
}
