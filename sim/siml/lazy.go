package siml

import (
	"fmt"
	"sync"
)

// A value made on first use. A package-level sync.OnceValues keeps the function it wraps, and all that function
// calls, in every binary that imports the package, since the package's init refers to it; a lazy read only inside a
// function is kept only where that function is reachable. lib.wasm reads the schema index through ParseRequest and
// PrintRequest, so it carries that code; sim/core's preset loaders, which it imports and never calls, add none.
type lazy[T any] struct {
	once  sync.Once
	value T
	err   error
}

// The value make gives on the first call; a panic in make is its error, on that call and every later one.
func (l *lazy[T]) get(make func() (T, error)) (T, error) {
	l.once.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				l.err = fmt.Errorf("%v", r)
			}
		}()
		l.value, l.err = make()
	})
	return l.value, l.err
}
