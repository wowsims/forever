package database

import (
	"math"
	"testing"
)

func TestDecodeCBOR(t *testing.T) {
	for _, test := range []struct {
		data []byte
		want any
	}{
		{[]byte{0x19, 0x01, 0xf4}, int64(500)},
		{[]byte{0x39, 0x01, 0xf3}, int64(-500)},
		{[]byte{0x63, 'a', 'b', 'c'}, "abc"},
		{[]byte{0xf9, 0x3e, 0x00}, 1.5},
		{[]byte{0xf9, 0xc4, 0x00}, -4.0},
		{[]byte{0xf9, 0x00, 0x01}, math.Ldexp(1, -24)},
		{[]byte{0xfa, 0x3f, 0xc0, 0x00, 0x00}, 1.5},
		{[]byte{0xfb, 0x40, 0x39, 0x0f, 0x5c, 0x28, 0xf5, 0xc2, 0x8f}, 25.06},
		{[]byte{0xf5}, true},
	} {
		got, err := decodeCBOR(test.data)
		if err != nil || got != test.want {
			t.Errorf("% x = %v (%v), want %v", test.data, got, err, test.want)
		}
	}

	for name, data := range map[string][]byte{
		"indefinite length": {0x9f, 0x01, 0xff},
		"tag":               {0xc1, 0x00},
		"truncated":         {0x18},
		"trailing bytes":    {0x01, 0x02},
		"list past the end": {0x99, 0xff, 0xff},
		"list as map key":   {0xa1, 0x80, 0x01},
	} {
		if _, err := decodeCBOR(data); err == nil {
			t.Errorf("%s: decoded without an error", name)
		}
	}
}
