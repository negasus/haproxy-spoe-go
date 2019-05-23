package action

import (
	"sync"
)

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

var pool = sync.Pool{
	New: func() interface{} {
		return newAction()
	},
}

type Action struct {
	Type  Type
	Scope Scope
	Name  string
	Value interface{}
}

func (action *Action) SetVar(scope Scope, name string, value interface{}) {
	action.Type = TypeSetVar
	action.Scope = scope
	action.Name = name
	action.Value = value
}

func (action *Action) UnsetVar(scope Scope, name string) {
	action.Type = TypeUnsetVar
	action.Scope = scope
	action.Name = name
}

func newAction() *Action {
	m := &Action{}

	return m
}

func AcquireAction() *Action {
	m := pool.Get()
	if m == nil {
		return newAction()
	}

	return m.(*Action)
}

func ReleaseAction(m *Action) {
	m.Reset()
	pool.Put(m)
}

func (action *Action) Reset() {
	action.Type = 0
	action.Scope = 0
	action.Name = ""
	action.Value = nil
}
