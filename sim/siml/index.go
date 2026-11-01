package siml

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// The schema's lookups that reading uses: the derivations of buildSchema, keyed for the parser and tied to the proto
// descriptors they describe.
type schemaIndex struct {
	sections       map[string]*sectionDef
	sectionList    []*sectionDef                  // In the order of SPEC.md §5.1, which the printer writes.
	rotationPath   []protoreflect.FieldDescriptor // From IndividualSimSettings to the APLRotation the rotation sections write.
	actionKinds    map[string]*aplKind
	valueKinds     map[string]*aplKind
	valueKindList  []*aplKind          // In APLValue's declaration order.
	accessors      map[string]*aplKind // Keyed "head.prop".
	heads          []string
	specs          map[string]SpecInfo // Keyed by the spec oneof field's name.
	keywords       map[string]bool
	slots          map[string]bool // The ItemSlot short names, which start slot rows.
	unitWords      map[string]*proto.UnitReference
	unitPrefixes   map[string]proto.UnitReference_Type // target → Target: the prefixes of targetN, playerN and petN.
	messages       map[string]*message
	enums          map[string]*enum
	sequenceFields map[string]bool // The string fields that name a sequence: the sequence head's subject.
}

// An action or value kind, with its field in APLAction's or APLValue's oneof.
type aplKind struct {
	*kindForm
	field   protoreflect.FieldDescriptor
	message *message
}

type message struct {
	desc    protoreflect.MessageDescriptor
	fields  []*field          // In declaration order, the skipped ones left out.
	byName  map[string]*field // By the name a file writes, and by its aliases.
	byProto map[string]*field
	byJSON  map[string]*field
}

type field struct {
	*schemaField
	desc protoreflect.FieldDescriptor
}

type enum struct {
	values   []enumValue
	byName   map[string]*enumValue // By short name and by proto name.
	byNumber map[int32]*enumValue  // The first value of each number.
	flag     string                // The value a bare field name sets.
}

const (
	settingsName = "proto.IndividualSimSettings"
	slotEnum     = "proto.ItemSlot"    // The enum whose short names are the slot words of slot rows.
	statEnum     = "proto.Stat"        // The enum stat fields (printed "stat") hold a value of.
	otherEnum    = "proto.OtherAction" // The enum other.<short> names a value of.
	rotationName = "proto.APLRotation"
	itemSpecName = "proto.ItemSpec"
)

var indexValue lazy[*schemaIndex]

func loadIndex() (*schemaIndex, error) { return indexValue.get(buildIndex) }

func buildIndex() (*schemaIndex, error) {
	s, err := buildSchema()
	if err != nil {
		return nil, err
	}
	specs, err := loadSpecs()
	if err != nil {
		return nil, err
	}
	ix := &schemaIndex{
		sections:       map[string]*sectionDef{},
		actionKinds:    map[string]*aplKind{},
		valueKinds:     map[string]*aplKind{},
		accessors:      map[string]*aplKind{},
		heads:          s.Heads,
		specs:          map[string]SpecInfo{},
		keywords:       map[string]bool{},
		slots:          map[string]bool{},
		unitWords:      map[string]*proto.UnitReference{},
		unitPrefixes:   map[string]proto.UnitReference_Type{},
		messages:       map[string]*message{},
		enums:          map[string]*enum{},
		sequenceFields: map[string]bool{},
	}
	for name, m := range s.Messages {
		d, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(name))
		if err != nil {
			return nil, err
		}
		md := d.(protoreflect.MessageDescriptor)
		msg := &message{desc: md, byName: map[string]*field{}, byProto: map[string]*field{}, byJSON: map[string]*field{}}
		for i := range m.Fields {
			f := &field{schemaField: &m.Fields[i], desc: md.Fields().ByName(protoreflect.Name(m.Fields[i].Name))}
			msg.fields = append(msg.fields, f)
			msg.byProto[f.Name] = f
			msg.byJSON[f.JSON] = f
			msg.byName[f.Siml] = f
			for _, alias := range f.Aliases {
				msg.byName[alias] = f
			}
		}
		ix.messages[name] = msg
	}
	for name, e := range s.Enums {
		en := &enum{values: e.Values, byName: map[string]*enumValue{}, byNumber: map[int32]*enumValue{}, flag: e.Flag}
		for i := range en.values {
			v := &en.values[i]
			en.byName[v.Short] = v
			en.byName[v.Name] = v
			if _, ok := en.byNumber[v.Number]; !ok {
				en.byNumber[v.Number] = v
			}
		}
		ix.enums[name] = en
	}
	if err := ix.addKinds(s.Kinds.Action, actionMessage.Oneofs().ByName("action"), ix.actionKinds); err != nil {
		return nil, err
	}
	if err := ix.addKinds(s.Kinds.Value, valueMessage.Oneofs().ByName("value"), ix.valueKinds); err != nil {
		return nil, err
	}
	for i := range s.Kinds.Value {
		k := ix.valueKinds[s.Kinds.Value[i].Kind]
		ix.valueKindList = append(ix.valueKindList, k)
		if a := k.Accessor; a != nil {
			ix.accessors[a.Head+"."+a.Prop] = k
			if k.message.byProto[a.Subject].Printed == "string" {
				ix.sequenceFields[a.Subject] = true
			}
		}
	}
	for i := range s.Sections {
		ix.sections[s.Sections[i].Name] = &s.Sections[i]
		ix.sectionList = append(ix.sectionList, &s.Sections[i])
	}
	rotationPath, ok := ix.specPath(ix.sections[sectionAPL].Sources[0].Path, nil)
	if !ok {
		return nil, fmt.Errorf("the apl section's source names no message")
	}
	ix.rotationPath = rotationPath
	for _, spec := range specs {
		ix.specs[string(spec.Field.Name())] = spec
	}
	for _, k := range s.Keywords {
		ix.keywords[k] = true
	}
	for _, v := range s.Enums[slotEnum].Values {
		ix.slots[v.Short] = true
	}
	for word, raw := range s.Units.Words {
		u := &proto.UnitReference{}
		if err := protojson.Unmarshal(raw, u); err != nil {
			return nil, fmt.Errorf("forms.json unit word %s: %w", word, err)
		}
		ix.unitWords[word] = u
	}
	for prefix, typ := range s.Units.Indexed {
		ix.unitPrefixes[prefix] = proto.UnitReference_Type(proto.UnitReference_Type_value[typ])
	}
	return ix, nil
}

func (ix *schemaIndex) addKinds(forms []kindForm, oneof protoreflect.OneofDescriptor, into map[string]*aplKind) error {
	for i := range forms {
		k := &forms[i]
		msg := ix.messages[k.Message]
		if msg == nil {
			return fmt.Errorf("the schema has no message %s for kind %s", k.Message, k.Kind)
		}
		into[k.Kind] = &aplKind{kindForm: k, field: oneof.Fields().ByName(protoreflect.Name(k.Kind)), message: msg}
	}
	return nil
}

func (m *message) fieldNames() []string {
	names := make([]string, len(m.fields))
	for i, f := range m.fields {
		names[i] = f.Siml
	}
	return names
}

func (e *enum) shortNames() []string {
	names := make([]string, len(e.values))
	for i, v := range e.values {
		names[i] = v.Short
	}
	return names
}

// A field a section's entries set (SPEC.md §5.1), with the message fields that lead to it.
type placed struct {
	*field
	at      []protoreflect.FieldDescriptor // From IndividualSimSettings to the message that holds the field.
	encoded protoreflect.FieldDescriptor   // For a field of the spec's Rotation: the string field, in the message at leads to, that holds the Rotation as JSON.
	path    string                         // The field's JSON path from IndividualSimSettings, which tells two placed fields apart.
}

// The message fields a section source's path of JSON names reads from IndividualSimSettings, "*" being the spec's
// field; false when the path names the spec and there is none.
func (ix *schemaIndex) specPath(path []string, spec *SpecInfo) ([]protoreflect.FieldDescriptor, bool) {
	m := ix.messages[settingsName]
	out := make([]protoreflect.FieldDescriptor, 0, len(path))
	for _, segment := range path {
		var fd protoreflect.FieldDescriptor
		switch f := m.byJSON[segment]; {
		case segment == "*" && spec != nil:
			fd = spec.Field
		case f != nil:
			fd = f.desc
		}
		if fd == nil || fd.Message() == nil {
			return nil, false
		}
		m = ix.messages[string(fd.Message().FullName())]
		out = append(out, fd)
	}
	return out, true
}

// The message the path of message fields reaches from IndividualSimSettings.
func (ix *schemaIndex) messageAt(at []protoreflect.FieldDescriptor) *message {
	if len(at) == 0 {
		return ix.messages[settingsName]
	}
	return ix.messages[string(at[len(at)-1].Message().FullName())]
}

// The fields of a section in its order (SPEC.md §5.1): each source's fields in declaration order, with the spec's
// Options flattened (its own fields, then its class options') and the spec's Rotation in place of the string
// simple: encodes it into. Without a spec, the fields the spec decides are left out.
func (ix *schemaIndex) sectionFields(def *sectionDef, spec *SpecInfo) []*placed {
	var out []*placed
	for _, src := range def.Sources {
		at, ok := ix.specPath(src.Path, spec)
		if !ok {
			continue
		}
		m := ix.messageAt(at)
		chosen := m.fields
		if src.Fields != nil {
			chosen = nil
			for _, name := range src.Fields {
				chosen = append(chosen, m.byJSON[name])
			}
		}
		switch src.Flatten {
		case "spec_options":
			var classOptions *field
			if spec != nil && spec.ClassOptions != nil {
				classOptions = m.byProto[string(spec.ClassOptions.Name())]
			}
			for _, f := range m.fields {
				if f != classOptions {
					out = append(out, newPlaced(f, at, nil))
				}
			}
			if classOptions != nil {
				inner := slices.Concat(at, []protoreflect.FieldDescriptor{classOptions.desc})
				for _, f := range ix.messages[classOptions.Type].fields {
					out = append(out, newPlaced(f, inner, nil))
				}
			}
		case "spec_rotation_json":
			if spec == nil {
				continue
			}
			for _, holder := range chosen {
				for _, f := range ix.messages[string(spec.Rotation.FullName())].fields {
					out = append(out, newPlaced(f, at, holder.desc))
				}
			}
		default:
			for _, f := range chosen {
				out = append(out, newPlaced(f, at, nil))
			}
		}
	}
	return out
}

func newPlaced(f *field, at []protoreflect.FieldDescriptor, encoded protoreflect.FieldDescriptor) *placed {
	path := jsonPath(at)
	if encoded != nil {
		path = jsonPath(at, encoded)
	}
	return &placed{field: f, at: at, encoded: encoded, path: path + "." + f.JSON}
}

// The JSON names of a path of message fields and a last field, joined by dots.
func jsonPath(at []protoreflect.FieldDescriptor, last ...protoreflect.FieldDescriptor) string {
	var names []string
	for _, fd := range slices.Concat(at, last) {
		names = append(names, fd.JSONName())
	}
	return strings.Join(names, ".")
}

// The field a section's slot rows fill: its list of ItemSpecs.
func slotField(fields []*placed) *placed {
	for _, p := range fields {
		if p.Type == itemSpecName && p.isList() {
			return p
		}
	}
	return nil
}

// The length of an array an enum indexes: one more than its largest number.
func (ix *schemaIndex) enumLength(name string) int {
	length := 0
	for _, v := range ix.enums[name].values {
		length = max(length, int(v.Number)+1)
	}
	return length
}

// The UnitReference a unit word writes (SPEC.md §6.6), or nil for a word that is no unit.
func (ix *schemaIndex) unitReference(word string) *proto.UnitReference {
	if u, ok := ix.unitWords[word]; ok {
		return googleProto.Clone(u).(*proto.UnitReference)
	}
	// targetN, playerN and petN, with N ≥ 1.
	split := 0
	for split < len(word) && (isLower(word[split]) || word[split] == '_') {
		split++
	}
	typ, ok := ix.unitPrefixes[word[:split]]
	digits := word[split:]
	if split == 0 || !ok || !isInt(digits) || digits[0] == '0' {
		return nil
	}
	n, err := strconv.ParseInt(digits, 10, 64)
	if err != nil || n-1 > math.MaxInt32 {
		return nil
	}
	u := &proto.UnitReference{Type: typ, Index: int32(n - 1)}
	if typ == proto.UnitReference_Pet {
		u.Owner = ix.unitReference("self")
	}
	return u
}

func (ix *schemaIndex) unitWordList() []string {
	words := make([]string, 0, len(ix.unitWords))
	for w := range ix.unitWords {
		words = append(words, w)
	}
	sort.Strings(words)
	return words
}

// Whether name is a reserved word, which the name table never holds (SPEC.md §4 note 6).
func (ix *schemaIndex) isReserved(name string) bool {
	return slices.Contains(ix.heads, name) || ix.keywords[name] || ix.unitReference(name) != nil
}

// A new, empty message of the type md describes.
func newMessage(md protoreflect.MessageDescriptor) protoreflect.Message {
	mt, err := protoregistry.GlobalTypes.FindMessageByName(md.FullName())
	if err != nil {
		panic(fmt.Sprintf("siml: %s is not registered: %v", md.FullName(), err))
	}
	return mt.New()
}
