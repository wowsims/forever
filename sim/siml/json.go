package siml

import (
	"encoding/base64"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// Protojson as the site writes it, JSON.stringify of the protojson of a message (SPEC.md §5.4, §12.3): JSON names
// in declaration order, default values left out, the skipped fields left out, enum values by name, 64-bit integers
// as strings, numbers in ECMAScript's form, and strings escaped as JSON.stringify escapes them. Raw JSON,
// spec_rotation_json and the printer's read-back check all write it.

// The JSON of m.
func (ix *schemaIndex) jsonText(m protoreflect.Message) string {
	var b strings.Builder
	ix.writeMessage(&b, m)
	return b.String()
}

// The JSON of a field's whole value, or of one element of a list or map when item.
func (ix *schemaIndex) valueJSON(fd protoreflect.FieldDescriptor, v protoreflect.Value, item bool) string {
	var b strings.Builder
	switch {
	case !v.IsValid() || !item && isDefault(fd, v):
		b.WriteString("null")
	case item:
		ix.writeValue(&b, fd, v)
	default:
		ix.writeField(&b, fd, v)
	}
	return b.String()
}

func (ix *schemaIndex) writeMessage(b *strings.Builder, m protoreflect.Message) {
	if !m.IsValid() {
		b.WriteString("null")
		return
	}
	b.WriteByte('{')
	first := true
	for _, f := range ix.messages[string(m.Descriptor().FullName())].fields {
		v, ok := fieldValue(m, f.desc)
		if !ok {
			continue
		}
		if !first {
			b.WriteByte(',')
		}
		first = false
		writeString(b, f.JSON)
		b.WriteByte(':')
		ix.writeField(b, f.desc, v)
	}
	b.WriteByte('}')
}

func (ix *schemaIndex) writeField(b *strings.Builder, fd protoreflect.FieldDescriptor, v protoreflect.Value) {
	switch {
	case fd.IsList():
		b.WriteByte('[')
		for i := range v.List().Len() {
			if i > 0 {
				b.WriteByte(',')
			}
			ix.writeValue(b, fd, v.List().Get(i))
		}
		b.WriteByte(']')
	case fd.IsMap():
		entries := map[string]protoreflect.Value{}
		var keys []string
		v.Map().Range(func(k protoreflect.MapKey, x protoreflect.Value) bool {
			entries[k.String()] = x
			keys = append(keys, k.String())
			return true
		})
		slices.SortFunc(keys, compareObjectKeys)
		b.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				b.WriteByte(',')
			}
			writeString(b, k)
			b.WriteByte(':')
			ix.writeValue(b, fd.MapValue(), entries[k])
		}
		b.WriteByte('}')
	default:
		ix.writeValue(b, fd, v)
	}
}

// The order ECMAScript keeps the keys of a JSON object in: those that read as array indices first, by number, then
// the others, which a map holds in no order, by text.
func compareObjectKeys(a, b string) int {
	index := func(k string) (uint64, bool) {
		n, err := strconv.ParseUint(k, 10, 32)
		return n, err == nil && n < math.MaxUint32 && strconv.FormatUint(n, 10) == k
	}
	na, ia := index(a)
	nb, ib := index(b)
	switch {
	case ia && ib:
		return int(na) - int(nb)
	case ia:
		return -1
	case ib:
		return 1
	}
	return strings.Compare(a, b)
}

func (ix *schemaIndex) writeValue(b *strings.Builder, fd protoreflect.FieldDescriptor, v protoreflect.Value) {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		b.WriteString(strconv.FormatBool(v.Bool()))
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		b.WriteString(strconv.FormatInt(v.Int(), 10))
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		b.WriteString(strconv.FormatUint(v.Uint(), 10))
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		writeString(b, strconv.FormatInt(v.Int(), 10))
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		writeString(b, strconv.FormatUint(v.Uint(), 10))
	case protoreflect.FloatKind:
		writeFloat(b, v.Float(), 32)
	case protoreflect.DoubleKind:
		writeFloat(b, v.Float(), 64)
	case protoreflect.StringKind:
		writeString(b, v.String())
	case protoreflect.BytesKind:
		writeString(b, base64.StdEncoding.EncodeToString(v.Bytes()))
	case protoreflect.EnumKind:
		if e := ix.enums[string(fd.Enum().FullName())]; e != nil && e.byNumber[int32(v.Enum())] != nil {
			writeString(b, e.byNumber[int32(v.Enum())].Name)
		} else {
			b.WriteString(strconv.Itoa(int(v.Enum())))
		}
	default:
		ix.writeMessage(b, v.Message())
	}
}

// A float as protojson writes one: NaN and the infinities as strings.
func writeFloat(b *strings.Builder, f float64, bits int) {
	switch {
	case math.IsNaN(f):
		b.WriteString(`"NaN"`)
	case math.IsInf(f, 1):
		b.WriteString(`"Infinity"`)
	case math.IsInf(f, -1):
		b.WriteString(`"-Infinity"`)
	default:
		b.WriteString(jsNumber(f, bits))
	}
}

// The value of fd in m as protojson writes it; false when fd is unset, or holds the default value of a field with no
// presence. A -0 is the default too, as it is in a JSON number.
func fieldValue(m protoreflect.Message, fd protoreflect.FieldDescriptor) (protoreflect.Value, bool) {
	if !m.IsValid() || !m.Has(fd) {
		return protoreflect.Value{}, false
	}
	v := m.Get(fd)
	if isDefault(fd, v) {
		return protoreflect.Value{}, false
	}
	return v, true
}

// Whether v is the value protojson leaves out of fd: an empty list or map, or the zero of a scalar with no presence.
func isDefault(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
	switch {
	case fd.IsList():
		return v.List().Len() == 0
	case fd.IsMap():
		return v.Map().Len() == 0
	case fd.HasPresence():
		return false
	}
	switch fd.Kind() {
	case protoreflect.BoolKind:
		return !v.Bool()
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return v.Float() == 0
	case protoreflect.StringKind:
		return v.String() == ""
	case protoreflect.BytesKind:
		return len(v.Bytes()) == 0
	case protoreflect.EnumKind:
		return v.Enum() == 0
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind, protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return v.Uint() == 0
	}
	return v.Int() == 0
}

// A JSON string as JSON.stringify writes it: " and \ escaped, \b \f \n \r \t for those control characters,
// \u00xx for the others, and every other character as itself.
func jsonString(s string) string {
	var b strings.Builder
	writeString(&b, s)
	return b.String()
}

func writeString(b *strings.Builder, s string) {
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 {
				fmt.Fprintf(b, `\u%04x`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
}

// JSON in backticks, a raw token (SPEC.md §2): a backtick, which can only sit in a JSON string, is escaped.
func rawText(json string) string {
	return "`" + strings.ReplaceAll(json, "`", "\\u0060") + "`"
}

// The shortest decimal digits that read back as f, a float of bits bits, and where its point goes: |f| is
// 0.digits × 10^point.
func shortestDigits(f float64, bits int) (digits string, point int) {
	mantissa, exponent, _ := strings.Cut(strconv.FormatFloat(math.Abs(f), 'e', -1, bits), "e")
	e, _ := strconv.Atoi(exponent)
	return strings.Replace(mantissa, ".", "", 1), e + 1
}

// A finite number as ECMAScript's Number#toString writes it, which JSON.stringify uses: 1e+21 and 1e-7 in
// exponent form, and -0 as 0.
func jsNumber(f float64, bits int) string {
	if f == 0 {
		return "0"
	}
	digits, n := shortestDigits(f, bits)
	sign := ""
	if f < 0 {
		sign = "-"
	}
	k := len(digits)
	switch {
	case k <= n && n <= 21:
		return sign + digits + strings.Repeat("0", n-k)
	case 0 < n && n <= 21:
		return sign + digits[:n] + "." + digits[n:]
	case -6 < n && n <= 0:
		return sign + "0." + strings.Repeat("0", -n) + digits
	}
	exponent := fmt.Sprintf("e%+d", n-1)
	if k == 1 {
		return sign + digits + exponent
	}
	return sign + digits[:1] + "." + digits[1:] + exponent
}

// A finite number as SIML writes it (SPEC.md §12.3): the shortest decimal that reads back to it, never with an
// exponent.
func numberText(f float64, bits int) string {
	if f == 0 {
		return "0"
	}
	digits, n := shortestDigits(f, bits)
	sign := ""
	if f < 0 {
		sign = "-"
	}
	switch {
	case n <= 0:
		return sign + "0." + strings.Repeat("0", -n) + digits
	case n >= len(digits):
		return sign + digits + strings.Repeat("0", n-len(digits))
	}
	return sign + digits[:n] + "." + digits[n:]
}
