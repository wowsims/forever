package siml

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// SPEC.md §15 item 6, the schema property test: every action and value kind and every section field in the schema,
// with each field set to a typical value (and every kind with none set too), prints in the form §8.3 or §12.2 gives
// it and reads back to itself. A kind or field added to the protos is covered the day it is added.

// Whether a printed row or value has a form.
type form func(string) bool

func matches(pattern string) form { return regexp.MustCompile(pattern).MatchString }

func startsWith(prefix, unless string) form {
	return func(s string) bool {
		return strings.HasPrefix(s, prefix) && (unless == "" || !strings.HasPrefix(s, unless))
	}
}

var (
	sampleSpell = &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: 4242}} // testdata/names.json names 4242 sample in every namespace.
	named       = regexp.MustCompile(`^\w+=`)
	// The keyword or operator each short form starts with, or holds (SPEC.md §6.4, §6.5).
	shortActionForms = map[string]form{
		"cast_spell":      startsWith("cast ", ""),
		"channel_spell":   startsWith("channel ", ""),
		"wait":            startsWith("wait ", "wait until"),
		"wait_until":      startsWith("wait until ", ""),
		"move":            startsWith("move ", "move for"),
		"move_duration":   startsWith("move for ", ""),
		"group_reference": startsWith(`group "`, ""),
		"schedule":        startsWith(`schedule "`, ""),
		"sequence":        startsWith(`sequence "`, ""),
	}
	shortValueForms = map[string]form{
		"const":                matches(`^1\.5s$`),
		"and":                  matches(` and `),
		"or":                   matches(` or `),
		"not":                  startsWith("not ", ""),
		"cmp":                  matches(` > `),
		"math":                 matches(` \+ `),
		"variable_ref":         startsWith("$", ""),
		"variable_placeholder": startsWith("%", ""),
	}
	// The fields each short form writes (SPEC.md §6.4, §6.5). A kind that sets another prints in its generic form.
	shortFields = map[string][]protoreflect.Name{
		"cast_spell":           {"spell_id", "target"},
		"channel_spell":        {"spell_id", "target", "interrupt_if", "allow_recast"},
		"wait":                 {"duration"},
		"wait_until":           {"condition"},
		"move":                 {"range_from_target", "recheck_while_moving"},
		"move_duration":        {"duration"},
		"group_reference":      {"group_name", "variables"},
		"schedule":             {"schedule", "inner_action"},
		"sequence":             {"name", "actions"},
		"const":                {"val"},
		"and":                  {"vals"},
		"or":                   {"vals"},
		"not":                  {"val"},
		"cmp":                  {"op", "lhs", "rhs"},
		"math":                 {"op", "lhs", "rhs"},
		"variable_ref":         {"name"},
		"variable_placeholder": {"name"},
	}
)

// Sets f in m to a typical value: its last enum value, a set ActionID, the current target, a duration, an item id,
// some entries of an indexed array, and a message with each of its fields set; two of them in a list.
func setSample(t *testing.T, ix *schemaIndex, m protoreflect.Message, f *field, depth int) {
	one := func() protoreflect.Value {
		switch f.Printed {
		case "bool":
			return protoreflect.ValueOfBool(true)
		case "number", "map":
			return sampleNumber(f.desc)
		case "string":
			return protoreflect.ValueOfString("Sample")
		case "enum", "flag_enum":
			values := ix.enums[f.Type].values
			return protoreflect.ValueOfEnum(protoreflect.EnumNumber(values[len(values)-1].Number))
		case "stat":
			return intValue(f.desc, 14)
		case "ref":
			return intValue(f.desc, 4242)
		case "action_id":
			return protoreflect.ValueOfMessage(googleProto.Clone(sampleSpell).ProtoReflect())
		case "unit":
			return protoreflect.ValueOfMessage((&proto.UnitReference{Type: proto.UnitReference_CurrentTarget}).ProtoReflect())
		case "value":
			return protoreflect.ValueOfMessage(constValue("1.5s").ProtoReflect())
		case "action":
			cast := &proto.APLAction{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: sampleSpell}}}
			return protoreflect.ValueOfMessage(cast.ProtoReflect())
		case "indexed":
			l := newList(f)
			for i := range ix.enums[f.Index].values {
				if i%5 == 1 {
					l.Append(protoreflect.ValueOfFloat64(float64(i) / 2))
				} else {
					l.Append(protoreflect.ValueOfFloat64(0))
				}
			}
			return protoreflect.ValueOfList(l)
		case "message":
			body := newMessage(f.desc.Message())
			if depth <= 2 {
				for _, x := range ix.messages[f.Type].fields {
					setSample(t, ix, body, x, depth+1)
				}
			}
			return protoreflect.ValueOfMessage(body)
		}
		t.Fatalf("%s is printed as %s, which has no sample", f.desc.FullName(), f.Printed)
		return protoreflect.Value{}
	}
	switch {
	case f.desc.IsMap():
		m.Mutable(f.desc).Map().Set(protoreflect.ValueOfInt32(6).MapKey(), one())
	case f.isList():
		l := m.Mutable(f.desc).List()
		l.Append(one())
		l.Append(one())
	default:
		m.Set(f.desc, one())
	}
}

// 2.5 for a float, 3 for an integer, and 50 for a map of integers.
func sampleNumber(fd protoreflect.FieldDescriptor) protoreflect.Value {
	if fd.IsMap() {
		fd = fd.MapValue()
		if fd.Kind() != protoreflect.DoubleKind && fd.Kind() != protoreflect.FloatKind {
			return intValue(fd, 50)
		}
	}
	switch fd.Kind() {
	case protoreflect.DoubleKind:
		return protoreflect.ValueOfFloat64(2.5)
	case protoreflect.FloatKind:
		return protoreflect.ValueOfFloat32(2.5)
	}
	return intValue(fd, 3)
}

// A kind's message with each field set, the short forms' operators at values that print as the patterns above say.
func kindSample(t *testing.T, ix *schemaIndex, k *aplKind) protoreflect.Message {
	body := newMessage(k.message.desc)
	for _, f := range k.message.fields {
		setSample(t, ix, body, f, 0)
	}
	switch m := body.Interface().(type) {
	case *proto.APLValueConst:
		m.Val = "1.5s"
	case *proto.APLValueCompare:
		m.Op = proto.APLValueCompare_OpGt
	case *proto.APLValueMath:
		m.Op = proto.APLValueMath_OpAdd
	}
	return body
}

// Prints a hunter rotation with row as its one priority row, checks that it reads back, and gives the row's text.
func printRow(t *testing.T, row *proto.APLListItem) string {
	t.Helper()
	r := &proto.APLRotation{Type: proto.APLRotation_TypeAPL, PriorityList: []*proto.APLListItem{row}}
	names := corpusNames(t)
	text := printRotation(proto.Spec_SpecHunter, r, names)
	f, err := Parse("x.apl.siml", text, names, nil)
	if err != nil {
		t.Fatalf("%v\n%s", err, text)
	}
	if !googleProto.Equal(f.Rotation(), r) {
		t.Fatalf("the print reads back differently\n%s\ngot:  %s\nwant: %s", text, protojson.Format(f.Rotation()), protojson.Format(r))
	}
	lines := strings.Split(string(text), "\n")
	start := slices.Index(lines, "priority:") + 1
	end := start + slices.Index(lines[start:], "")
	for i := range lines[start:end] {
		lines[start+i] = strings.TrimSpace(lines[start+i])
	}
	return strings.Join(lines[start:end], " ")
}

// The form SPEC.md §8.3 derives for a kind with every field set: a short kind's own, unless it has a field its
// short form does not write.
func derivedForm(k *aplKind, short map[string]form) form {
	args := func(s string) (string, bool) { return strings.CutPrefix(s, k.Kind+"(") }
	switch k.Form {
	case "short":
		if !slices.ContainsFunc(k.message.fields, func(f *field) bool { return !slices.Contains(shortFields[k.Kind], f.desc.Name()) }) {
			return short[k.Kind]
		}
	case "accessor":
		return matches(`^(target\.)?` + k.Accessor.Head + `\(.+\)\.` + k.Accessor.Prop + `(\(.+\))?$`)
	case "bare":
		return func(s string) bool { return s == k.Kind }
	case "unit_prefix":
		return func(s string) bool { return s == "target."+k.Kind }
	case "list_args":
		return func(s string) bool {
			rest, ok := args(s)
			return ok && strings.HasSuffix(rest, ")") && strings.Contains(rest, ", ") && !named.MatchString(rest)
		}
	}
	return func(s string) bool {
		rest, ok := args(s)
		return ok && named.MatchString(rest) == (k.Positional == "")
	}
}

func TestEveryActionKind(t *testing.T) {
	ix, err := loadIndex()
	if err != nil {
		t.Fatal(err)
	}
	for i := range actionMessage.Oneofs().ByName("action").Fields().Len() {
		k := ix.actionKinds[string(actionMessage.Oneofs().ByName("action").Fields().Get(i).Name())]
		t.Run(k.Kind, func(t *testing.T) {
			full := printRow(t, &proto.APLListItem{Action: kindAction(k, kindSample(t, ix, k))})
			if !derivedForm(k, shortActionForms)(full) {
				t.Errorf("%s prints as %q, which is not of the %s form", k.Kind, full, k.Form)
			}
			printRow(t, &proto.APLListItem{Action: kindAction(k, newMessage(k.message.desc))})
		})
	}
}

func TestEveryValueKind(t *testing.T) {
	ix, err := loadIndex()
	if err != nil {
		t.Fatal(err)
	}
	cast := func(condition *proto.APLValue) *proto.APLListItem {
		return &proto.APLListItem{Action: &proto.APLAction{
			Condition: condition,
			Action:    &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: sampleSpell}},
		}}
	}
	for _, k := range ix.valueKindList {
		t.Run(k.Kind, func(t *testing.T) {
			full := printRow(t, cast(kindValue(k, kindSample(t, ix, k))))
			condition := full[strings.Index(full, " if ")+len(" if "):]
			if !derivedForm(k, shortValueForms)(condition) {
				t.Errorf("%s prints as %q, which is not of the %s form", k.Kind, condition, k.Form)
			}
			printRow(t, cast(kindValue(k, newMessage(k.message.desc))))
		})
	}
}

// The entry form SPEC.md §12.2 gives a field set alone in its section, as the line or lines it prints.
func entryForm(f *field) *regexp.Regexp {
	n := regexp.QuoteMeta(f.Siml)
	switch {
	case f.Type == itemSpecName:
		return regexp.MustCompile(`^head +\w+ \w+ \+ \w+\nneck +\w+ \w+ \+ \w+$`)
	case topWords[f.Siml]:
		return regexp.MustCompile(`^` + n + `=`)
	case f.Printed == "bool":
		return regexp.MustCompile(`^` + n + `$`)
	case f.Printed == "flag_enum":
		return regexp.MustCompile(`^` + n + `=\w+$`)
	case f.Printed == "message" && f.Repeated:
		return regexp.MustCompile(`^` + n + `\(.+\)\n` + n + `\(.+\)$`)
	case f.Printed == "message":
		return regexp.MustCompile(`^` + n + `\(.+\)$`)
	case f.Printed == "indexed":
		return regexp.MustCompile(`^` + n + `\(\w+=[0-9.]+(, \w+=[0-9.]+)*\)$`)
	case f.Printed == "map":
		return regexp.MustCompile(`^` + n + " `")
	case f.Repeated:
		return regexp.MustCompile(`^` + n + ` \S+, \S+$`)
	}
	return regexp.MustCompile(`^` + n + ` \S+$`)
}

// The body lines of a section in a printed file, a header line's body included and continuation lines joined.
func sectionBody(text, section string) string {
	lines := strings.Split(text, "\n")
	start := slices.IndexFunc(lines, func(l string) bool { return l == section+":" || strings.HasPrefix(l, section+": ") })
	var body []string
	if len(lines[start]) > len(section)+1 {
		body = append(body, lines[start][len(section)+2:])
	}
	for _, line := range lines[start+1:] {
		switch {
		case !strings.HasPrefix(line, "  "):
			return strings.Join(body, "\n")
		case strings.HasPrefix(line, "      "):
			body[len(body)-1] += " " + strings.TrimSpace(line)
		default:
			body = append(body, strings.TrimSpace(line))
		}
	}
	return strings.Join(body, "\n")
}

// Settings with p set alone in the section, and what reading the section sets beside it: the spec's message and
// class, the section's message, the spec's class options, and the rotation type a rotation section infers.
func sectionSample(t *testing.T, ix *schemaIndex, def *sectionDef, spec *SpecInfo, p *placed) *proto.IndividualSimSettings {
	settings := &proto.IndividualSimSettings{Player: &proto.Player{Class: spec.Class}}
	m := settings.ProtoReflect()
	settings.Player.ProtoReflect().Mutable(spec.Field)
	if def.Presence != nil {
		presence, _ := ix.specPath(def.Presence, spec)
		if holder := mutableAt(m, presence); def.Name == "settings" && spec.ClassOptions != nil {
			holder.Mutable(spec.ClassOptions)
		}
	}
	if def.Category == rotationCategory {
		settings.Player.Rotation = &proto.APLRotation{Type: proto.APLRotation_TypeAPL}
		if def.Name == sectionSimple {
			settings.Player.Rotation.Type = proto.APLRotation_TypeSimple
		}
	}
	holder := mutableAt(m, p.at)
	if p.encoded == nil {
		setSample(t, ix, holder, p.field, 0)
		return settings
	}
	r := newMessage(spec.Rotation)
	setSample(t, ix, r, p.field, 0)
	holder.Set(p.encoded, protoreflect.ValueOfString(ix.jsonText(r)))
	return settings
}

func TestEverySectionField(t *testing.T) {
	ix, err := loadIndex()
	if err != nil {
		t.Fatal(err)
	}
	hunter := ix.specs["hunter"]
	for _, def := range ix.sectionList {
		if slices.ContainsFunc(def.Body, func(b string) bool { return b == "rows" || b == "var" }) {
			continue
		}
		specs := []SpecInfo{hunter}
		if slices.ContainsFunc(def.Sources, func(src sectionSource) bool { return src.Flatten != "" || slices.Contains(src.Path, "*") }) {
			specs = Specs()
		}
		for _, spec := range specs {
			t.Run(def.Name+"/"+string(spec.Field.Name()), func(t *testing.T) {
				names := corpusNames(t)
				for _, p := range ix.sectionFields(def, &spec) {
					settings := sectionSample(t, ix, def, &spec, p)
					text := Print(Kind{}, spec.Spec, settings, nil, names)
					f, err := Parse("", text, names, nil)
					if err != nil {
						t.Fatalf("%s: %v\n%s", p.Siml, err, text)
					}
					if got, want := Normalize(f.Settings, names, spec.Spec), Normalize(settings, names, spec.Spec); !googleProto.Equal(got, want) {
						t.Errorf("%s reads back differently\n%s\ngot:  %s\nwant: %s", p.Siml, text, protojson.Format(got), protojson.Format(want))
					}
					if body, form := sectionBody(string(text), def.Name), entryForm(p.field); !form.MatchString(body) {
						t.Errorf("%s prints as %q, which is not of the form %s", p.Siml, body, form)
					}
				}
			})
		}
	}
}
