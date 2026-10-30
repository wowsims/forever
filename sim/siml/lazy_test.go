package siml

import "testing"

func TestLazyKeepsAPanicAsItsError(t *testing.T) {
	var l lazy[int]
	for range 2 {
		v, err := l.get(func() (int, error) { panic("no table") })
		if v != 0 || err == nil || err.Error() != "no table" {
			t.Fatalf("get gives %d, %v; want 0 and the error no table", v, err)
		}
	}
}
