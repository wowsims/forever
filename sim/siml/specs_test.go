package siml

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestSpecs(t *testing.T) {
	if _, err := loadSpecs(); err != nil {
		t.Fatalf("a field of Player's spec oneof breaks a rule of sim/siml/SPEC.md §8.4: %v", err)
	}
	specs := Specs()
	if want := playerSpecOneof().Fields().Len(); len(specs) != want {
		t.Fatalf("Specs() lists %d specs, want one per field of Player's spec oneof, %d", len(specs), want)
	}

	seen := map[proto.Spec]string{}
	for _, s := range specs {
		if prev, ok := seen[s.Spec]; ok {
			t.Errorf("%s and %s are both %s", prev, s.Field.Name(), s.Spec)
		}
		seen[s.Spec] = string(s.Field.Name())
	}
	for number, name := range proto.Spec_name {
		if spec := proto.Spec(number); spec != proto.Spec_SpecUnknown && seen[spec] == "" {
			t.Errorf("no field of Player's spec oneof is %s", name)
		}
	}

	for _, want := range []struct {
		field    string
		spec     proto.Spec
		class    proto.Class
		options  string
		rotation string
	}{
		{"hunter", proto.Spec_SpecHunter, proto.Class_ClassHunter, "proto.Hunter.Options", "proto.Hunter.Rotation"},
		{"dps_warrior", proto.Spec_SpecDpsWarrior, proto.Class_ClassWarrior, "proto.DpsWarrior.Options", "proto.DpsWarrior.Rotation"},
		{"feral_bear_druid", proto.Spec_SpecFeralBearDruid, proto.Class_ClassDruid, "proto.FeralBearDruid.Options", "proto.FeralBearDruid.Rotation"},
		{"healer_priest", proto.Spec_SpecHealerPriest, proto.Class_ClassPriest, "proto.HealerPriest.Options", "proto.HealerPriest.Rotation"},
	} {
		var got *SpecInfo
		for i := range specs {
			if string(specs[i].Field.Name()) == want.field {
				got = &specs[i]
			}
		}
		if got == nil {
			t.Errorf("Specs() has no %s", want.field)
			continue
		}
		if got.Spec != want.spec || got.Class != want.class || string(got.Options.FullName()) != want.options || string(got.Rotation.FullName()) != want.rotation {
			t.Errorf("%s: got %s, %s, %s, %s; want %s, %s, %s, %s", want.field,
				got.Spec, got.Class, got.Options.FullName(), got.Rotation.FullName(),
				want.spec, want.class, want.options, want.rotation)
		}
		if got.ClassOptions == nil || got.ClassOptions.Name() != "class_options" {
			t.Errorf("%s: ClassOptions is %v, want the options' class_options field", want.field, got.ClassOptions)
		}
	}
}
