package logger

var _ Logger = &Nop{}

var nop = &Nop{}

// Nop is Logger implementation which never logs.
type Nop struct{}

// NewNop returns a Nop logger.
func NewNop() *Nop { return nop }

func (*Nop) Errorf(_ string, _ ...interface{}) {}
