package database

import (
	"fmt"
	"math"
)

// Decodes the CBOR (RFC 8949) the baked data holds: integers as int64, byte and text strings as string, floats as
// float64, false, true, null, lists as []any and maps as map[any]any with int64 or string keys. Tags and indefinite
// lengths fail.
func decodeCBOR(data []byte) (any, error) {
	decoder := &cborDecoder{data: data}
	value, err := decoder.value()
	if err != nil {
		return nil, err
	}
	if decoder.pos != len(data) {
		return nil, fmt.Errorf("CBOR: %d bytes after the value", len(data)-decoder.pos)
	}
	return value, nil
}

type cborDecoder struct {
	data []byte
	pos  int
}

func (d *cborDecoder) next(n uint64) ([]byte, error) {
	if n > uint64(len(d.data)-d.pos) {
		return nil, fmt.Errorf("CBOR: %d bytes past the end at offset %d", n, d.pos)
	}
	chunk := d.data[d.pos : d.pos+int(n)]
	d.pos += int(n)
	return chunk, nil
}

func (d *cborDecoder) value() (any, error) {
	head, err := d.next(1)
	if err != nil {
		return nil, err
	}
	major, info := head[0]>>5, head[0]&31
	var arg uint64
	switch {
	case info < 24:
		arg = uint64(info)
	case info <= 27:
		chunk, err := d.next(1 << (info - 24))
		if err != nil {
			return nil, err
		}
		for _, b := range chunk {
			arg = arg<<8 | uint64(b)
		}
	default:
		return nil, fmt.Errorf("CBOR: additional information %d at offset %d", info, d.pos-1)
	}

	switch major {
	case 0, 1:
		if arg > math.MaxInt64 {
			return nil, fmt.Errorf("CBOR: integer out of range at offset %d", d.pos)
		}
		if major == 1 {
			return -1 - int64(arg), nil
		}
		return int64(arg), nil
	case 2, 3:
		chunk, err := d.next(arg)
		return string(chunk), err
	case 4:
		// Every value takes a byte at least, which bounds the length before allocating.
		if arg > uint64(len(d.data)-d.pos) {
			return nil, fmt.Errorf("CBOR: list of %d past the end at offset %d", arg, d.pos)
		}
		list := make([]any, arg)
		for i := range list {
			if list[i], err = d.value(); err != nil {
				return nil, err
			}
		}
		return list, nil
	case 5:
		if arg > uint64(len(d.data)-d.pos)/2 {
			return nil, fmt.Errorf("CBOR: map of %d past the end at offset %d", arg, d.pos)
		}
		table := make(map[any]any, arg)
		for range arg {
			key, err := d.value()
			if err != nil {
				return nil, err
			}
			switch key.(type) {
			case int64, string:
			default:
				return nil, fmt.Errorf("CBOR: map key %v at offset %d", key, d.pos)
			}
			if table[key], err = d.value(); err != nil {
				return nil, err
			}
		}
		return table, nil
	case 7:
		switch info {
		case 20:
			return false, nil
		case 21:
			return true, nil
		case 22:
			return nil, nil
		case 25:
			return halfFloat(uint16(arg)), nil
		case 26:
			return float64(math.Float32frombits(uint32(arg))), nil
		case 27:
			return math.Float64frombits(arg), nil
		}
	}
	return nil, fmt.Errorf("CBOR: major type %d, additional information %d at offset %d", major, info, d.pos)
}

// An IEEE 754 binary16 value.
func halfFloat(bits uint16) float64 {
	sign := 1.0
	if bits&0x8000 != 0 {
		sign = -1
	}
	exponent, fraction := int(bits>>10)&0x1f, float64(bits&0x3ff)
	switch exponent {
	case 0:
		return sign * math.Ldexp(fraction, -24)
	case 0x1f:
		if fraction != 0 {
			return math.NaN()
		}
		return math.Inf(int(sign))
	}
	return sign * math.Ldexp(fraction+1024, exponent-25)
}
