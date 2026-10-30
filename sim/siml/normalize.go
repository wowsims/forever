package siml

import (
	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// A copy of m normalised as SPEC.md §12.4 compares a proto with the parse of its print: uuid and api_version cleared,
// the items of an EquipmentSpec or ItemSwap padded to 17, every array an enum indexes padded to its enum's length, an
// absent one too, spec_rotation_json rewritten as the canonical JSON of the spec's Rotation it holds (an empty string,
// or one that does not read as the Rotation, as {}, which is what its print reads back as), and the rank cleared of
// each ActionID whose spell id and tag a name of the spec's class names, since the name reads back with the rank of
// its entry. The spec is spec, or for SpecUnknown the one IndividualSimSettings sets. It panics if the protos break a
// rule of SPEC.md §8, as Print does.
func Normalize(m googleProto.Message, names *Names, spec proto.Spec) googleProto.Message {
	ix, err := loadIndex()
	if err != nil {
		panic(err)
	}
	out := googleProto.Clone(m)
	settings, _ := out.(*proto.IndividualSimSettings)
	var rotation protoreflect.MessageDescriptor
	class := proto.Class_ClassUnknown
	if info := specFor(ix, spec, settings); info != nil {
		rotation, class = info.Rotation, info.Class
	}
	var pad func(m protoreflect.Message)
	pad = func(m protoreflect.Message) {
		msg := ix.messages[string(m.Descriptor().FullName())]
		fields := m.Descriptor().Fields()
		for i := range fields.Len() {
			fd := fields.Get(i)
			f := msg.byProto[string(fd.Name())]
			switch {
			case f == nil:
				m.Clear(fd)
			case f.Printed == "indexed":
				l := m.Mutable(fd).List()
				for l.Len() < ix.enumLength(f.Index) {
					l.Append(l.NewElement())
				}
			case f.Type == itemSpecName && f.Repeated:
				l := m.Mutable(fd).List()
				for l.Len() < ix.enumLength(slotEnum) {
					l.Append(l.NewElement())
				}
			case f.Name == "spec_rotation_json" && rotation != nil:
				m.Set(fd, protoreflect.ValueOfString(ix.jsonText(ix.decodedRotation(m.Get(fd).String(), rotation))))
			case fd.Message() != nil && fd.IsList():
				for i := range m.Get(fd).List().Len() {
					pad(m.Mutable(fd).List().Get(i).Message())
				}
			case fd.Message() != nil && !fd.IsMap() && m.Has(fd):
				pad(m.Mutable(fd).Message())
			}
		}
	}
	pad(out.ProtoReflect())
	walkMessages(out.ProtoReflect(), func(m protoreflect.Message) {
		id, ok := m.Interface().(*proto.ActionID)
		if !ok {
			return
		}
		for _, ns := range AllNamespaces() {
			if _, named := names.SpellName(class, ns, id); named {
				id.Rank = 0
			}
		}
	})
	return out
}

// The Rotation a spec_rotation_json text holds. A text that does not read as one, keys that name no field aside,
// gives an empty Rotation.
func (ix *schemaIndex) decodedRotation(text string, md protoreflect.MessageDescriptor) protoreflect.Message {
	r := newMessage(md)
	if ix.readJSON(text, r, protojson.UnmarshalOptions{DiscardUnknown: true}) != "" {
		return newMessage(md)
	}
	return r
}
