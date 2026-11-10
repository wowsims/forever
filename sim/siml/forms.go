package siml

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// The hand-kept part of the language (SPEC.md §8.1). Everything else the schema holds is derived from the
// proto descriptors.
//
//go:embed forms.json
var formsJSON []byte

type forms struct {
	Keywords []string   `json:"keywords"`
	Heads    []formHead `json:"heads"`
	Units    unitForms  `json:"units"`
	Short    struct {
		Action []string `json:"action"`
		Value  []string `json:"value"`
	} `json:"short"` // The kinds the keyword actions and operators write.
	Sections   []formSection     `json:"sections"`
	Refs       map[string]string `json:"refs"`        // Field pattern → item, enchant, suffix, stat or number.
	Indexed    map[string]string `json:"indexed"`     // Field pattern → the enum that indexes the array.
	FlagValues map[string]string `json:"flag_values"` // Enum → the value a bare field name sets.
	Skip       []string          `json:"skip"`        // Field patterns, or type:<message> for every field of that type.
}

type formHead struct {
	Name    string `json:"name"`
	Subject string `json:"subject"` // The printed type of the field the head's argument fills.
}

type unitForms struct {
	Words   map[string]json.RawMessage `json:"words"`
	Indexed map[string]string          `json:"indexed"`
}

type formSection struct {
	Name     string       `json:"name"`
	Category string       `json:"category"`
	Body     []string     `json:"body"`
	Sources  []formSource `json:"sources"`
	Presence string       `json:"presence"`
}

// Paths are dotted proto field names from IndividualSimSettings; "*" stands for the field of Player's spec oneof
// that the file's spec line names.
type formSource struct {
	Path    string   `json:"path"`
	Fields  []string `json:"fields"`
	Flatten string   `json:"flatten"`
}

var (
	categories   = []string{"g", "t", "r", "c", "m", "x", "e", "u"}
	bodyKinds    = []string{"entries", "slot_rows", "rows", "var"}
	flattenKinds = []string{"", "spec_options", "spec_rotation_json"}
	refKinds     = []string{"item", "enchant", "suffix", "stat", "number"}
)

var formsValue lazy[*forms]

func loadForms() (*forms, error) { return formsValue.get(readForms) }

func readForms() (*forms, error) {
	f := &forms{}
	d := json.NewDecoder(bytes.NewReader(formsJSON))
	d.DisallowUnknownFields()
	if err := d.Decode(f); err != nil {
		return nil, fmt.Errorf("forms.json: %w", err)
	}
	if err := f.check(); err != nil {
		return nil, fmt.Errorf("forms.json: %w", err)
	}
	return f, nil
}

// The checks that need no descriptors; the schema builder checks every name against the protos.
func (f *forms) check() error {
	for _, s := range f.Sections {
		if !slices.Contains(categories, s.Category) {
			return fmt.Errorf("section %s: unknown category %q", s.Name, s.Category)
		}
		for _, b := range s.Body {
			if !slices.Contains(bodyKinds, b) {
				return fmt.Errorf("section %s: unknown body kind %q", s.Name, b)
			}
		}
		for _, src := range s.Sources {
			if !slices.Contains(flattenKinds, src.Flatten) {
				return fmt.Errorf("section %s: unknown flatten %q", s.Name, src.Flatten)
			}
		}
	}
	for pattern, kind := range f.Refs {
		if !slices.Contains(refKinds, kind) {
			return fmt.Errorf("refs %s: unknown kind %q", pattern, kind)
		}
	}
	for _, patterns := range [][]string{slices.Collect(maps.Keys(f.Refs)), slices.Collect(maps.Keys(f.Indexed)), f.Skip} {
		for _, p := range patterns {
			if _, err := path.Match(strings.TrimPrefix(p, typePrefix), ""); err != nil {
				return fmt.Errorf("pattern %q: %w", p, err)
			}
		}
	}
	return nil
}

const typePrefix = "type:"

// A forms.json field pattern: a glob over the field's full name (proto.ConsumesSpec.*_id), or type:<full name>
// for every field of that message or enum type.
func fieldMatches(pattern string, fd protoreflect.FieldDescriptor) bool {
	if typeName, ok := strings.CutPrefix(pattern, typePrefix); ok {
		return fieldTypeName(fd) == typeName
	}
	ok, _ := path.Match(pattern, string(fd.FullName()))
	return ok
}

// The value from the one pattern in table that matches fd, or "" when none does.
func matchField(table map[string]string, fd protoreflect.FieldDescriptor) (string, error) {
	found, value := "", ""
	for pattern, v := range table {
		if !fieldMatches(pattern, fd) {
			continue
		}
		if found != "" && v != value {
			return "", fmt.Errorf("%s matches both %q and %q in forms.json", fd.FullName(), found, pattern)
		}
		found, value = pattern, v
	}
	return value, nil
}

func (f *forms) keywordSet() map[string]bool {
	set := make(map[string]bool, len(f.Keywords))
	for _, k := range f.Keywords {
		set[k] = true
	}
	return set
}

func (f *forms) skips(fd protoreflect.FieldDescriptor) bool {
	return slices.ContainsFunc(f.Skip, func(p string) bool { return fieldMatches(p, fd) })
}

// The full name of a message or enum field's type, "" for a scalar.
func fieldTypeName(fd protoreflect.FieldDescriptor) string {
	switch {
	case fd.Message() != nil:
		return string(fd.Message().FullName())
	case fd.Enum() != nil:
		return string(fd.Enum().FullName())
	}
	return ""
}
