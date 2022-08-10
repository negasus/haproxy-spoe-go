package action

type Type byte

type Scope byte

const (
	nbVarsSetVar   byte = 0x03
	nbVarsUnsetVar byte = 0x02

	TypeSetVar   Type = 0x01
	TypeUnsetVar Type = 0x02

	ScopeProcess     Scope = 0x00
	ScopeSession     Scope = 0x01
	ScopeTransaction Scope = 0x02
	ScopeRequest     Scope = 0x03
	ScopeResponse    Scope = 0x04
)

type Action struct {
	Type  Type
	Scope Scope
	Name  string
	Value interface{}
}

func NewSetVar(scope Scope, name string, value interface{}) Action {
	return Action{
		Type:  TypeSetVar,
		Scope: scope,
		Name:  name,
		Value: value,
	}
}

func NewUnsetVar(scope Scope, name string) Action {
	return Action{
		Type:  TypeUnsetVar,
		Scope: scope,
		Name:  name,
	}
}
