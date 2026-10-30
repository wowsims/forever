package siml

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode"

	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// The words of SIML that the editor's grammar colours (SPEC.md §2-§4): the keywords, the accessor heads, the unit
// words, the prefixes of the indexed units (target1, player2), and every section name, in the schema's order.
type Words struct {
	Keywords, Heads, Units, UnitPrefixes, Sections []string
	DurationUnits                                  []string // The units a duration literal takes (SPEC.md §2).
}

func GrammarWords() (Words, error) {
	s, err := buildSchema()
	if err != nil {
		return Words{}, err
	}
	w := Words{
		Keywords:      s.Keywords,
		Heads:         s.Heads,
		Units:         slices.Sorted(maps.Keys(s.Units.Words)),
		UnitPrefixes:  slices.Sorted(maps.Keys(s.Units.Indexed)),
		DurationUnits: slices.Clone(durationUnits),
	}
	for _, def := range s.Sections {
		w.Sections = append(w.Sections, def.Name)
	}
	return w, nil
}

// Every name and form SPEC.md §8 derives from the protos, with forms.json folded in. loadIndex builds its lookups from
// it.
type schema struct {
	Keywords []string
	Heads    []string
	Units    unitForms
	Messages map[string]schemaMessage // Every message reachable from IndividualSimSettings, and each spec's Rotation.
	Enums    map[string]schemaEnum
	Kinds    struct {
		Action []kindForm
		Value  []kindForm
	}
	Sections []sectionDef
}

type schemaMessage struct {
	Fields []schemaField // Declaration order, skipped fields left out.
}

type schemaField struct {
	Name     string
	Siml     string   // The name a file writes (§8.2).
	Aliases  []string // Other names a file may write.
	JSON     string
	Number   int32
	Scalar   string // For a map, the value's.
	Type     string // For a map, the value's.
	Repeated bool
	Printed  string
	Ref      string
	Index    string
}

type schemaEnum struct {
	Values []enumValue
	Flag   string // The value a bare field name sets.
}

type enumValue struct {
	Name   string
	Number int32
	Short  string
}

type kindForm struct {
	Kind       string
	JSON       string
	Message    string
	Form       string
	Positional string // The generic form's positional field, which every kind accepts.
	Accessor   *accessor
	UnitField  string
	ListField  string
}

type accessor struct {
	Head    string
	Prop    string
	Subject string
	Unit    string
}

type sectionDef struct {
	Name     string
	Category string
	Body     []string
	Sources  []sectionSource
	Presence []string
}

// Paths are JSON names from IndividualSimSettings, with "*" for the spec field.
type sectionSource struct {
	Path    []string
	Fields  []string // JSON names; all of the message's fields when absent.
	Flatten string
}

var (
	settingsMessage = (&proto.IndividualSimSettings{}).ProtoReflect().Descriptor()
	actionMessage   = (&proto.APLAction{}).ProtoReflect().Descriptor()
	valueMessage    = (&proto.APLValue{}).ProtoReflect().Descriptor()
	actionIDMessage = (&proto.ActionID{}).ProtoReflect().Descriptor()
	unitMessage     = (&proto.UnitReference{}).ProtoReflect().Descriptor()
)

type builder struct {
	forms    *forms
	keywords map[string]bool
	messages map[string]schemaMessage
	enums    map[string]schemaEnum
}

func buildSchema() (*schema, error) {
	f, err := loadForms()
	if err != nil {
		return nil, err
	}
	specs, err := loadSpecs()
	if err != nil {
		return nil, err
	}
	b := &builder{forms: f, keywords: f.keywordSet(), messages: map[string]schemaMessage{}, enums: map[string]schemaEnum{}}

	if err := b.addMessage(settingsMessage); err != nil {
		return nil, err
	}
	for _, spec := range specs {
		if err := b.addMessage(spec.Rotation); err != nil {
			return nil, err
		}
	}
	for enum := range f.FlagValues {
		if _, ok := b.enums[enum]; !ok {
			return nil, fmt.Errorf("forms.json flag_values: no field of IndividualSimSettings reaches %s", enum)
		}
	}

	s := &schema{
		Keywords: f.Keywords,
		Units:    f.Units,
		Messages: b.messages,
		Enums:    b.enums,
	}
	for _, h := range f.Heads {
		s.Heads = append(s.Heads, h.Name)
	}
	if err := checkUnits(f.Units); err != nil {
		return nil, err
	}
	if s.Kinds.Action, err = b.kinds(actionMessage.Oneofs().ByName("action"), f.Short.Action, false); err != nil {
		return nil, err
	}
	if s.Kinds.Value, err = b.kinds(valueMessage.Oneofs().ByName("value"), f.Short.Value, true); err != nil {
		return nil, err
	}
	if s.Sections, err = b.sections(specs); err != nil {
		return nil, err
	}
	return s, nil
}

func checkUnits(u unitForms) error {
	for word, raw := range u.Words {
		if err := protojson.Unmarshal(raw, &proto.UnitReference{}); err != nil {
			return fmt.Errorf("forms.json unit word %s: %w", word, err)
		}
	}
	for prefix, typ := range u.Indexed {
		if _, ok := proto.UnitReference_Type_value[typ]; !ok {
			return fmt.Errorf("forms.json indexed unit %s: no UnitReference type %s", prefix, typ)
		}
	}
	return nil
}

func (b *builder) addMessage(md protoreflect.MessageDescriptor) error {
	name := string(md.FullName())
	if _, ok := b.messages[name]; ok {
		return nil
	}
	b.messages[name] = schemaMessage{} // Marks it while its fields, which may lead back to it, are added.
	fields := []schemaField{}
	fds := md.Fields()
	for i := range fds.Len() {
		fd := fds.Get(i)
		if b.forms.skips(fd) {
			continue
		}
		sf, err := b.field(fd)
		if err != nil {
			return err
		}
		fields = append(fields, sf)
	}
	nameFields(fields)
	b.messages[name] = schemaMessage{Fields: fields}
	return nil
}

func (b *builder) addEnum(ed protoreflect.EnumDescriptor) error {
	name := string(ed.FullName())
	if _, ok := b.enums[name]; ok {
		return nil
	}
	short := enumShortNames(ed, b.keywords)
	e := schemaEnum{Values: make([]enumValue, len(short))}
	for i := range short {
		v := ed.Values().Get(i)
		e.Values[i] = enumValue{Name: string(v.Name()), Number: int32(v.Number()), Short: short[i]}
	}
	if flag, ok := b.forms.FlagValues[name]; ok {
		if ed.Values().ByName(protoreflect.Name(flag)) == nil {
			return fmt.Errorf("forms.json flag_values: %s has no value %s", name, flag)
		}
		e.Flag = flag
	}
	b.enums[name] = e
	return nil
}

// Adds the message or enum a field holds, if any.
func (b *builder) addType(fd protoreflect.FieldDescriptor) error {
	switch {
	case fd.Message() != nil:
		return b.addMessage(fd.Message())
	case fd.Enum() != nil:
		return b.addEnum(fd.Enum())
	}
	return nil
}

func (b *builder) field(fd protoreflect.FieldDescriptor) (schemaField, error) {
	sf := schemaField{
		Name:     string(fd.Name()),
		JSON:     fd.JSONName(),
		Number:   int32(fd.Number()),
		Repeated: fd.IsList(),
	}
	if fd.IsMap() {
		sf.Printed = "map"
		sf.Scalar, sf.Type = valueType(fd.MapValue())
		return sf, b.addType(fd.MapValue())
	}
	sf.Scalar, sf.Type = valueType(fd)
	if err := b.addType(fd); err != nil {
		return sf, err
	}

	index, err := matchField(b.forms.Indexed, fd)
	if err != nil {
		return sf, err
	}
	ref, err := matchField(b.forms.Refs, fd)
	if err != nil {
		return sf, err
	}
	switch {
	case index != "" && ref != "":
		return sf, fmt.Errorf("%s is in both refs and indexed in forms.json", fd.FullName())
	case index != "":
		if !fd.IsList() || !isNumber(fd.Kind()) {
			return sf, fmt.Errorf("forms.json indexed: %s is not a repeated number", fd.FullName())
		}
		ed, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(index))
		if err != nil {
			return sf, fmt.Errorf("forms.json indexed %s: %w", fd.FullName(), err)
		}
		enum, ok := ed.(protoreflect.EnumDescriptor)
		if !ok {
			return sf, fmt.Errorf("forms.json indexed %s: %s is not an enum", fd.FullName(), index)
		}
		sf.Printed, sf.Index = "indexed", index
		return sf, b.addEnum(enum)
	case ref == "number":
		if !isNumber(fd.Kind()) {
			return sf, fmt.Errorf("forms.json refs: %s is not a number", fd.FullName())
		}
		sf.Printed = "number"
	case ref != "":
		if !isInteger(fd.Kind()) {
			return sf, fmt.Errorf("forms.json refs: %s holds a %s, which is not an integer", fd.FullName(), ref)
		}
		if ref == "stat" {
			sf.Printed = "stat"
		} else {
			sf.Printed, sf.Ref = "ref", ref
		}
	default:
		sf.Printed = b.printedByType(fd)
	}
	return sf, nil
}

func (b *builder) printedByType(fd protoreflect.FieldDescriptor) string {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.StringKind, protoreflect.BytesKind:
		return "string"
	case protoreflect.EnumKind:
		if _, ok := b.forms.FlagValues[string(fd.Enum().FullName())]; ok {
			return "flag_enum"
		}
		return "enum"
	case protoreflect.MessageKind, protoreflect.GroupKind:
		switch fd.Message().FullName() {
		case actionIDMessage.FullName():
			return "action_id"
		case unitMessage.FullName():
			return "unit"
		case valueMessage.FullName():
			return "value"
		case actionMessage.FullName():
			return "action"
		}
		return "message"
	}
	return "number"
}

// The printed names and aliases of a message's fields (§5.2, §8.2). An item ref field drops a trailing _id or
// _ids when the shorter name is no other field's.
func nameFields(fields []schemaField) {
	base := make([]string, len(fields))
	taken := map[string]bool{}
	for i, f := range fields {
		base[i] = f.Name
		if strings.ContainsFunc(f.Name, unicode.IsUpper) {
			base[i] = snake(f.JSON)
		}
		taken[base[i]] = true
	}
	short := make([]string, len(fields))
	shortCount := map[string]int{}
	for i, f := range fields {
		if f.Printed != "ref" || f.Ref != "item" {
			continue
		}
		for _, suffix := range []string{"_ids", "_id"} {
			if s, ok := strings.CutSuffix(base[i], suffix); ok && s != "" {
				short[i] = s
				shortCount[s]++
				break
			}
		}
	}
	for i := range fields {
		fields[i].Siml, fields[i].Aliases = base[i], []string{}
		if s := short[i]; s != "" && !taken[s] && shortCount[s] == 1 {
			fields[i].Siml, fields[i].Aliases = s, []string{base[i]}
		}
	}
}

// The short names of an enum's values, in declaration order (§8.2): the value name without the enum's own name
// in front, in snake case. A short name that is empty, starts with a digit, is a keyword or is shared with
// another value is the whole value name in snake case instead, and a keyword even then takes the enum's name in
// front: FinishingMove.None is finishing_move_none.
func enumShortNames(ed protoreflect.EnumDescriptor, keywords map[string]bool) []string {
	enum := string(ed.Name())
	values := ed.Values()
	candidate := make([]string, values.Len())
	count := map[string]int{}
	for i := range values.Len() {
		name := string(values.Get(i).Name())
		if rest, ok := strings.CutPrefix(name, enum); ok && rest != "" {
			name = strings.TrimLeft(rest, "_")
		}
		candidate[i] = snake(name)
		count[candidate[i]]++
	}
	short := make([]string, values.Len())
	for i, c := range candidate {
		if c == "" || isDigit(c[0]) || keywords[c] || count[c] > 1 {
			c = snake(string(values.Get(i).Name()))
			if keywords[c] {
				c = snake(enum) + "_" + c
			}
		}
		short[i] = c
	}
	return short
}

// The kinds of APLAction's or APLValue's oneof, each with the form §8.3 derives for it.
func (b *builder) kinds(oneof protoreflect.OneofDescriptor, short []string, value bool) ([]kindForm, error) {
	fds := oneof.Fields()
	for _, s := range short {
		if fds.ByName(protoreflect.Name(s)) == nil {
			return nil, fmt.Errorf("forms.json short: %s has no kind %s", oneof.FullName(), s)
		}
	}
	kinds := make([]kindForm, 0, fds.Len())
	for i := range fds.Len() {
		fd := fds.Get(i)
		md := fd.Message()
		if md == nil {
			return nil, fmt.Errorf("%s is not a message", fd.FullName())
		}
		fields := b.messages[string(md.FullName())].Fields
		k := kindForm{Kind: string(fd.Name()), JSON: fd.JSONName(), Message: string(md.FullName())}
		k.Positional = positionalField(fields)
		if value && len(fields) == 1 && fields[0].Printed == "unit" && !fields[0].Repeated {
			k.UnitField = fields[0].Name
		}
		if len(fields) == 1 && fields[0].Repeated {
			k.ListField = fields[0].Name
		}
		if value {
			k.Accessor = b.accessorOf(k.Kind, fields)
		}
		switch {
		case slices.Contains(short, k.Kind):
			k.Form = "short"
		case k.Accessor != nil:
			k.Form = "accessor"
		case len(fields) == 0:
			k.Form = "bare"
		case k.UnitField != "":
			k.Form = "unit_prefix"
		case k.ListField != "":
			k.Form = "list_args"
		default:
			k.Form = "generic"
		}
		kinds = append(kinds, k)
	}
	return kinds, nil
}

// The first field whose printed type is a ref, an enum, a stat, a string, a value, an action or a list (§8.3
// rule 6); units, numbers and bools are always named.
func positionalField(fields []schemaField) string {
	for _, f := range fields {
		if f.Repeated || slices.Contains(positionalTypes, f.Printed) {
			return f.Name
		}
	}
	return ""
}

var positionalTypes = []string{"ref", "action_id", "enum", "flag_enum", "stat", "string", "value", "action"}

// The accessor a value kind named <head>_<prop> reads as, when its message has exactly one subject field of the
// head's type and at most one unit field; the longest head that fits wins. Nil for any other kind.
func (b *builder) accessorOf(kind string, fields []schemaField) *accessor {
	var best *accessor
	for _, h := range b.forms.Heads {
		prop, ok := strings.CutPrefix(kind, h.Name+"_")
		if !ok || prop == "" || best != nil && len(best.Head) >= len(h.Name) {
			continue
		}
		var subjects, units []string
		for _, f := range fields {
			switch {
			case f.Repeated:
			case f.Printed == h.Subject:
				subjects = append(subjects, f.Name)
			case f.Printed == "unit":
				units = append(units, f.Name)
			}
		}
		if len(subjects) != 1 || len(units) > 1 {
			continue
		}
		best = &accessor{Head: h.Name, Prop: prop, Subject: subjects[0]}
		if len(units) == 1 {
			best.Unit = units[0]
		}
	}
	return best
}

func (b *builder) sections(specs []SpecInfo) ([]sectionDef, error) {
	defs := make([]sectionDef, 0, len(b.forms.Sections))
	for _, fs := range b.forms.Sections {
		def := sectionDef{Name: fs.Name, Category: fs.Category, Body: fs.Body, Sources: []sectionSource{}}
		for _, src := range fs.Sources {
			out, err := b.source(src, specs)
			if err != nil {
				return nil, fmt.Errorf("forms.json section %s: %w", fs.Name, err)
			}
			def.Sources = append(def.Sources, out)
		}
		if fs.Presence != "" {
			for _, spec := range specs {
				_, jsonPath, err := b.resolvePath(fs.Presence, spec)
				if err != nil {
					return nil, fmt.Errorf("forms.json section %s presence: %w", fs.Name, err)
				}
				def.Presence = jsonPath
			}
		}
		defs = append(defs, def)
	}
	return defs, nil
}

// A section source with JSON names, checked against every spec, since "*" reaches a different message for each.
func (b *builder) source(src formSource, specs []SpecInfo) (sectionSource, error) {
	out := sectionSource{Flatten: src.Flatten}
	for _, spec := range specs {
		md, jsonPath, err := b.resolvePath(src.Path, spec)
		if err != nil {
			return out, err
		}
		out.Path = jsonPath
		if src.Flatten == "spec_options" && (md != spec.Options || len(src.Fields) > 0) {
			return out, fmt.Errorf("source %q: spec_options flattens the whole of a spec's options", src.Path)
		}
		out.Fields = nil
		for _, name := range src.Fields {
			fd := md.Fields().ByName(protoreflect.Name(name))
			if fd == nil || b.forms.skips(fd) {
				return out, fmt.Errorf("source %q: %s has no field %s that a file may hold", src.Path, md.FullName(), name)
			}
			out.Fields = append(out.Fields, fd.JSONName())
		}
		if src.Flatten == "spec_rotation_json" {
			if len(src.Fields) != 1 || md.Fields().ByName(protoreflect.Name(src.Fields[0])).Kind() != protoreflect.StringKind {
				return out, fmt.Errorf("source %q: spec_rotation_json flattens exactly one string field", src.Path)
			}
		}
	}
	return out, nil
}

// The message a dotted path of proto field names reaches from IndividualSimSettings for one spec, and the path in
// JSON names, with "*" kept for the spec field.
func (b *builder) resolvePath(p string, spec SpecInfo) (protoreflect.MessageDescriptor, []string, error) {
	md := settingsMessage
	jsonPath := []string{}
	if p == "" {
		return md, jsonPath, nil
	}
	for _, seg := range strings.Split(p, ".") {
		fd := md.Fields().ByName(protoreflect.Name(seg))
		name := ""
		if seg == "*" {
			if spec.Field.ContainingMessage() != md {
				return nil, nil, fmt.Errorf("path %q: * stands for the spec field, which %s does not hold", p, md.FullName())
			}
			fd, name = spec.Field, seg
		} else if fd != nil {
			name = fd.JSONName()
		}
		if fd == nil || fd.Message() == nil || fd.IsList() || fd.IsMap() || b.forms.skips(fd) {
			return nil, nil, fmt.Errorf("path %q: %s has no message field %s that a file may hold", p, md.FullName(), seg)
		}
		md = fd.Message()
		jsonPath = append(jsonPath, name)
	}
	return md, jsonPath, nil
}

func valueType(fd protoreflect.FieldDescriptor) (scalar, typeName string) {
	if typeName = fieldTypeName(fd); typeName != "" {
		return "", typeName
	}
	return scalarName(fd.Kind()), ""
}

func scalarName(k protoreflect.Kind) string {
	switch k {
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	}
	return k.String() // bool, double, float, string, bytes
}

func isInteger(k protoreflect.Kind) bool {
	return isNumber(k) && k != protoreflect.FloatKind && k != protoreflect.DoubleKind
}

func isNumber(k protoreflect.Kind) bool {
	switch k {
	case protoreflect.BoolKind, protoreflect.EnumKind, protoreflect.StringKind, protoreflect.BytesKind,
		protoreflect.MessageKind, protoreflect.GroupKind:
		return false
	}
	return true
}
