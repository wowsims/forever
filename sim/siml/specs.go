package siml

import (
	"fmt"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// One field of Player's spec oneof, and what the spec line and the settings: and simple: sections read through it
// (SPEC.md §8.4).
type SpecInfo struct {
	Field        protoreflect.FieldDescriptor   // The oneof field the spec line names: hunter, dps_warrior.
	Spec         proto.Spec                     // Spec + the message name.
	Class        proto.Class                    // The class whose short name ends the message name in snake case.
	Options      protoreflect.MessageDescriptor // The type of the spec message's options field, which settings: flattens.
	ClassOptions protoreflect.FieldDescriptor   // The options' class_options field, flattened after the rest; nil when there is none.
	Rotation     protoreflect.MessageDescriptor // The spec message's nested Rotation, which simple: writes as spec_rotation_json.
}

// Every spec, in the oneof's declaration order. It panics if the protos break a rule of SPEC.md §8.4, which
// TestSpecs catches before any build ships.
func Specs() []SpecInfo {
	specs, err := loadSpecs()
	if err != nil {
		panic(err)
	}
	return slices.Clone(specs)
}

// The spec whose spec line writes field, such as hunter or dps_warrior.
func SpecNamed(field string) (SpecInfo, bool) {
	return specWhere(func(info SpecInfo) bool { return string(info.Field.Name()) == field })
}

// What the spec line of spec reads; false for SpecUnknown.
func SpecOf(spec proto.Spec) (SpecInfo, bool) {
	return specWhere(func(info SpecInfo) bool { return info.Spec == spec })
}

func specWhere(match func(SpecInfo) bool) (SpecInfo, bool) {
	specs, err := loadSpecs()
	if err != nil {
		panic(err)
	}
	for _, info := range specs {
		if match(info) {
			return info, true
		}
	}
	return SpecInfo{}, false
}

var specsValue lazy[[]SpecInfo]

func loadSpecs() ([]SpecInfo, error) { return specsValue.get(readSpecs) }

func readSpecs() ([]SpecInfo, error) {
	f, err := loadForms()
	if err != nil {
		return nil, err
	}
	classes := proto.Class_ClassUnknown.Descriptor()
	classShort := enumShortNames(classes, f.keywordSet())

	fields := playerSpecOneof().Fields()
	specs := make([]SpecInfo, 0, fields.Len())
	for i := range fields.Len() {
		fd := fields.Get(i)
		info, err := specInfo(fd, classes, classShort)
		if err != nil {
			return nil, fmt.Errorf("spec %s: %w", fd.Name(), err)
		}
		specs = append(specs, info)
	}
	return specs, nil
}

func specInfo(fd protoreflect.FieldDescriptor, classes protoreflect.EnumDescriptor, classShort []string) (SpecInfo, error) {
	md := fd.Message()
	if md == nil {
		return SpecInfo{}, fmt.Errorf("the field is not a message")
	}
	name := string(md.Name())

	spec, ok := proto.Spec_value["Spec"+name]
	if !ok {
		return SpecInfo{}, fmt.Errorf("the Spec enum has no value Spec%s", name)
	}

	snakeName := snake(name)
	class := -1
	for i, short := range classShort {
		if snakeName == short || strings.HasSuffix(snakeName, "_"+short) {
			if class >= 0 {
				return SpecInfo{}, fmt.Errorf("%s ends in two class names, %s and %s", snakeName, classShort[class], short)
			}
			class = i
		}
	}
	if class < 0 {
		return SpecInfo{}, fmt.Errorf("%s ends in no class name", snakeName)
	}

	options := md.Fields().ByName("options")
	if options == nil || options.Message() == nil || options.IsList() {
		return SpecInfo{}, fmt.Errorf("%s has no options message field", md.FullName())
	}
	rotation := md.Messages().ByName("Rotation")
	if rotation == nil {
		return SpecInfo{}, fmt.Errorf("%s has no nested Rotation message", md.FullName())
	}
	classOptions := options.Message().Fields().ByName("class_options")
	if classOptions != nil && (classOptions.Message() == nil || classOptions.IsList()) {
		return SpecInfo{}, fmt.Errorf("%s.class_options is not a message field", options.Message().FullName())
	}

	return SpecInfo{
		Field:        fd,
		Spec:         proto.Spec(spec),
		Class:        proto.Class(classes.Values().Get(class).Number()),
		Options:      options.Message(),
		ClassOptions: classOptions,
		Rotation:     rotation,
	}, nil
}

func playerSpecOneof() protoreflect.OneofDescriptor {
	return (&proto.Player{}).ProtoReflect().Descriptor().Oneofs().ByName("spec")
}
