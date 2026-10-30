package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// A place where two messages of one type differ: its JSON path, such as priorityList[3].action.castSpell, and the
// value each holds there, written as JSON, or "unset".
type difference struct {
	path     string
	old, new string
}

// Where a and b differ, in field order: a field set in only one of them, a value that differs, or a list element
// one of them lacks. Messages present in both are compared field by field, so each difference is the deepest one.
func messageDiff(a, b protoreflect.Message) []difference {
	var out []difference
	diffMessages("", a, b, &out)
	return out
}

func diffMessages(path string, a, b protoreflect.Message, out *[]difference) {
	fields := a.Descriptor().Fields()
	for i := range fields.Len() {
		fd := fields.Get(i)
		at := fd.JSONName()
		if path != "" {
			at = path + "." + at
		}
		switch {
		case !a.Has(fd) && !b.Has(fd):
		case fd.IsList():
			diffLists(at, fd, a.Get(fd).List(), b.Get(fd).List(), out)
		case fd.IsMap():
			diffMaps(at, fd, a.Get(fd).Map(), b.Get(fd).Map(), out)
		default:
			diffValues(at, fd, value(a, fd), value(b, fd), out)
		}
	}
}

// The value of a field, invalid when it is unset.
func value(m protoreflect.Message, fd protoreflect.FieldDescriptor) protoreflect.Value {
	if !m.Has(fd) {
		return protoreflect.Value{}
	}
	return m.Get(fd)
}

func diffLists(path string, fd protoreflect.FieldDescriptor, a, b protoreflect.List, out *[]difference) {
	for i := range max(a.Len(), b.Len()) {
		var x, y protoreflect.Value
		if i < a.Len() {
			x = a.Get(i)
		}
		if i < b.Len() {
			y = b.Get(i)
		}
		diffValues(fmt.Sprintf("%s[%d]", path, i), fd, x, y, out)
	}
}

func diffMaps(path string, fd protoreflect.FieldDescriptor, a, b protoreflect.Map, out *[]difference) {
	var keys []protoreflect.MapKey
	collect := func(k protoreflect.MapKey, _ protoreflect.Value) bool {
		if !slices.ContainsFunc(keys, func(seen protoreflect.MapKey) bool { return seen.Interface() == k.Interface() }) {
			keys = append(keys, k)
		}
		return true
	}
	a.Range(collect)
	b.Range(collect)
	slices.SortFunc(keys, func(x, y protoreflect.MapKey) int { return strings.Compare(x.String(), y.String()) })
	for _, k := range keys {
		diffValues(fmt.Sprintf("%s[%s]", path, strconv.Quote(k.String())), fd.MapValue(), a.Get(k), b.Get(k), out)
	}
}

// Compares one value of field fd, an element of it when it is a list; an invalid value is unset.
func diffValues(path string, fd protoreflect.FieldDescriptor, x, y protoreflect.Value, out *[]difference) {
	switch {
	case x.IsValid() && y.IsValid() && fd.Message() != nil:
		diffMessages(path, x.Message(), y.Message(), out)
	case x.IsValid() && y.IsValid() && x.Equal(y):
	default:
		*out = append(*out, difference{path: path, old: valueText(fd, x), new: valueText(fd, y)})
	}
}

// A value as JSON, as protojson writes it; "unset" for an invalid value.
func valueText(fd protoreflect.FieldDescriptor, v protoreflect.Value) string {
	switch {
	case !v.IsValid():
		return "unset"
	case fd.Message() != nil:
		data, err := protojson.Marshal(v.Message().Interface())
		var compact bytes.Buffer
		if err != nil || json.Compact(&compact, data) != nil {
			return fmt.Sprint(v.Interface())
		}
		return compact.String()
	case fd.Enum() != nil:
		if ev := fd.Enum().Values().ByNumber(v.Enum()); ev != nil {
			return strconv.Quote(string(ev.Name()))
		}
		return strconv.Itoa(int(v.Enum()))
	case fd.Kind() == protoreflect.StringKind:
		return strconv.Quote(v.String())
	}
	return fmt.Sprint(v.Interface())
}
