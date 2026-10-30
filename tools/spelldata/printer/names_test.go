package printer

import "testing"

func TestStatMaskName(t *testing.T) {
	cases := []struct {
		mask int32
		want string
	}{
		{8, "intellect"},
		{5, "strength and stamina"},
		{31, "to all stats"},
		{32, "stat 5"},
		{33, "strength and stat 5"},
	}
	for _, c := range cases {
		if got := statMaskName(c.mask); got != c.want {
			t.Errorf("statMaskName(%d) = %q, want %q", c.mask, got, c.want)
		}
	}
}
