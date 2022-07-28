package action_test

import (
	"runtime"
	"testing"

	"github.com/negasus/haproxy-spoe-go/action"
)

func BenchmarkActionsPool(b *testing.B) {
	const str = "foo"
	as := action.NewActions()

	for i := 0; i < b.N; i++ {
		for j := 0; j < 200; j++ {
			as.SetVar(action.ScopeSession, str, nil)
		}
		as.Reset()

		if i%150 == 0 {
			runtime.GC()
		}
	}
}
