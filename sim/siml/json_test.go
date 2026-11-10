package siml

import (
	"math"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

// Numbers as ECMAScript writes them, which spec_rotation_json and raw JSON hold (JSON.stringify), and as SIML writes
// them, the same digits with any exponent written out (SPEC.md §5.4, §12.3).
func TestNumbers(t *testing.T) {
	for _, c := range []struct {
		value      float64
		bits       int
		json, siml string
	}{
		{0, 64, "0", "0"},
		{math.Copysign(0, -1), 64, "0", "0"},
		{3, 64, "3", "3"},
		{-250, 64, "-250", "-250"},
		{0.1, 64, "0.1", "0.1"},
		{0.30000000000000004, 64, "0.30000000000000004", "0.30000000000000004"},
		{123.456, 64, "123.456", "123.456"},
		{1e20, 64, "100000000000000000000", "100000000000000000000"},
		{1e21, 64, "1e+21", "1000000000000000000000"},
		{1.5e21, 64, "1.5e+21", "1500000000000000000000"},
		{1e-6, 64, "0.000001", "0.000001"},
		{1e-7, 64, "1e-7", "0.0000001"},
		{-1.5e-7, 64, "-1.5e-7", "-0.00000015"},
		{9007199254740993, 64, "9007199254740992", "9007199254740992"},
		{math.MaxFloat64, 64, "1.7976931348623157e+308", "17976931348623157" + strings.Repeat("0", 292)},
		{5e-324, 64, "5e-324", "0." + strings.Repeat("0", 323) + "5"},
		{float64(float32(0.1)), 32, "0.1", "0.1"},
		{float64(float32(2.5)), 32, "2.5", "2.5"},
	} {
		if got := jsNumber(c.value, c.bits); got != c.json {
			t.Errorf("jsNumber(%v) = %s, want %s", c.value, got, c.json)
		}
		if got := numberText(c.value, c.bits); got != c.siml {
			t.Errorf("numberText(%v) = %s, want %s", c.value, got, c.siml)
		}
	}
}

// A float field holds the float32 nearest the number written, prints as the shortest decimal of that float32, and
// refuses a number past its range (SPEC.md §5.2). No proto holds a float field, so no corpus case can.
func TestFloatFields(t *testing.T) {
	f, ok := parseFloat("0.123456789")
	if !ok || numberText(float64(f), 32) != "0.12345679" {
		t.Errorf("0.123456789 in a float field is %v (%v), want it printed as 0.12345679", f, ok)
	}
	if f, ok := parseFloat("340282356779733661637539395458142568448"); ok {
		t.Errorf("a number past a float32's range reads as %v", f)
	}
}

// A double reaches spec_rotation_json, raw JSON and an entry in those forms, and reads back to itself.
func TestNumbersInFiles(t *testing.T) {
	ix, err := loadIndex()
	if err != nil {
		t.Fatal(err)
	}
	encounter := &proto.Encounter{Duration: 1e21, ExecuteProportion_20: 1e-7}
	if got, want := ix.jsonText(encounter.ProtoReflect()), `{"duration":1e+21,"executeProportion20":1e-7}`; got != want {
		t.Errorf("the encounter's JSON is %s, want %s", got, want)
	}
	settings := &proto.IndividualSimSettings{Encounter: encounter}
	text := string(Print(Kind{Section: "encounter"}, proto.Spec_SpecUnknown, settings, nil, &Names{}))
	if want := "siml 1\n\nduration 1000000000000000000000\nexecute_proportion_20 0.0000001\n"; text != want {
		t.Errorf("the encounter prints as\n%s\nwant\n%s", text, want)
	}
	f, err := Parse("x.encounter.siml", []byte(text), &Names{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := f.Settings.GetEncounter(); got.GetDuration() != 1e21 || got.GetExecuteProportion_20() != 1e-7 {
		t.Errorf("the encounter reads back as %v", got)
	}
}

// Strings as JSON.stringify writes them: only ", \ and the control characters escaped.
func TestJSONString(t *testing.T) {
	if got, want := jsonString("a\"b\\c\b\f\n\r\t\x01\x1f\x7f\u2028é😀"), `"a\"b\\c\b\f\n\r\t\u0001\u001f`+"\x7f\u2028é😀\""; got != want {
		t.Errorf("jsonString = %s, want %s", got, want)
	}
	if got, want := rawText(`{"a":"x`+"`"+`y"}`), "`{\"a\":\"x\\u0060y\"}`"; got != want {
		t.Errorf("rawText = %s, want %s", got, want)
	}
}
