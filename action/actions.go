package action

type Actions []*Action

func (actions *Actions) SetVar(scope Scope, name string, value interface{}) {
	a := AcquireAction()
	a.SetVar(scope, name, value)
	*actions = append(*actions, a)
}

func (actions *Actions) UnsetVar(scope Scope, name string) {
	a := AcquireAction()
	a.UnsetVar(scope, name)
	*actions = append(*actions, a)
}

func (actions *Actions) Reset() {
	for _, action := range *actions {
		ReleaseAction(action)
	}

	*actions = (*actions)[:0]
}

func NewActions() *Actions {
	return &Actions{}
}
