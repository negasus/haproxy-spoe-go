package logger

var _ Logger = Channel{}

// LogMessage represents a single log message to be logger.
//
// This structure groups parameters of a single log line to allow sending them
// over a channel.
type LogMessage struct {
	format string
	args   []interface{}
}

// Channel is Logger sensing messages to be logged over a channel.
type Channel struct {
	ch chan<- LogMessage
}

// NewChannel creates a new Channel Logger logging to ch.
//
// WARNING: If the channel gets full, the write of log message will block.
// Consequently, it's absolutely necessary to provide a channel with sufficient
// capacity and to guarantee that messages from it are consumed.
func NewChannel(ch chan<- LogMessage) Channel {
	return Channel{ch: ch}
}

func (c Channel) Errorf(format string, args ...interface{}) {
	c.ch <- LogMessage{
		format: format,
		args:   args,
	}
}
