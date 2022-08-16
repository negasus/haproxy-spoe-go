package logger

// Logger is any object with ability to log internal errors and warnings.
type Logger interface {
	Errorf(format string, args ...interface{})
}
