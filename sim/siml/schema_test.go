package siml

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
	"testing"

	googleProto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

func testSchema(t *testing.T) (*schema, *forms) {
	t.Helper()
	s, err := buildSchema()
	if err != nil {
		t.Fatal(err)
	}
	f, err := loadForms()
	if err != nil {
		t.Fatal(err)
	}
	return s, f
}

func messageDescriptor(t *testing.T, name string) protoreflect.MessageDescriptor {
	t.Helper()
	d, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(name))
	if err != nil {
		t.Fatal(err)
	}
	return d.(protoreflect.MessageDescriptor)
}

// A kind named like a keyword, head or unit word could not be told apart from it in a row. The short-form kinds
// are the exception: the keyword is how they are written (wait, and, sequence).
func TestKindNamesAreNotWords(t *testing.T) {
	s, f := testSchema(t)
	words := map[string]string{}
	for _, k := range f.Keywords {
		words[k] = "keyword"
	}
	for _, h := range f.Heads {
		words[h.Name] = "accessor head"
	}
	for w := range f.Units.Words {
		words[w] = "unit word"
	}
	indexed := regexp.MustCompile(`^(` + strings.Join(slices.Sorted(maps.Keys(f.Units.Indexed)), "|") + `)[1-9][0-9]*$`)

	for _, group := range []struct {
		message string
		kinds   []kindForm
	}{{"APLAction", s.Kinds.Action}, {"APLValue", s.Kinds.Value}} {
		for _, k := range group.kinds {
			if k.Form == "short" {
				continue
			}
			what, ok := words[k.Kind]
			if !ok && indexed.MatchString(k.Kind) {
				what, ok = "unit word", true
			}
			if ok {
				t.Errorf("the %s kind %s (the oneof field %s.%s in proto/apl.proto) is also the SIML %s %q, so a row could not tell them apart. Name the oneof field differently.",
					group.message, k.Kind, group.message, k.Kind, what, k.Kind)
			}
		}
	}
}

// §8.1: an integer field named like an id holds a game files id, a Stat or a plain number, and only forms.json
// says which.
func TestIDFieldsAreClassified(t *testing.T) {
	s, f := testSchema(t)
	for _, name := range slices.Sorted(maps.Keys(s.Messages)) {
		md := messageDescriptor(t, name)
		for _, sf := range s.Messages[name].Fields {
			if !slices.Contains([]string{"int32", "uint32", "int64"}, sf.Scalar) || sf.Printed == "map" {
				continue
			}
			if sf.Name != "id" && !strings.HasSuffix(sf.Name, "_id") && !strings.HasSuffix(sf.Name, "_ids") && !strings.HasPrefix(sf.Name, "stat_type") {
				continue
			}
			ref, err := matchField(f.Refs, md.Fields().ByName(protoreflect.Name(sf.Name)))
			if err != nil {
				t.Error(err)
			} else if ref == "" {
				t.Errorf(`%s.%s is an integer field named like an id, so SIML has to know what it holds. Add it to "refs" in sim/siml/forms.json: "item", "enchant" or "suffix" when it holds that game files id (a file then writes its name), "stat" when it holds a Stat value, or "number" for a plain number.`,
					name, sf.Name)
			}
		}
	}
}

// Every field of a message a section reads only part of, which is IndividualSimSettings, Player, each spec
// message, the rotation and the simple rotation, is held by a section, set by a top-level line, or skipped. A new
// field then cannot go unwritten: whoever adds it chooses its section, and with it its import category.
func TestSectionsMapEveryField(t *testing.T) {
	_, f := testSchema(t)
	b := &builder{forms: f}
	lines := map[protoreflect.FullName]string{ // SPEC.md §3.3
		"proto.Player.class":          "spec",
		"proto.Player.talents_string": "talents",
	}
	unheld := map[protoreflect.FullName]bool{}
	for _, spec := range Specs() {
		held := map[string]map[protoreflect.Name]bool{} // Concrete path → the fields of its message some source holds.
		whole := map[string]bool{}                      // Paths a source holds every field of.
		hold := func(p string, name protoreflect.Name) {
			if held[p] == nil {
				held[p] = map[protoreflect.Name]bool{}
			}
			held[p][name] = true
		}
		for _, sec := range f.Sections {
			for _, src := range sec.Sources {
				p := ""
				for seg := range strings.SplitSeq(src.Path, ".") {
					if seg == "" {
						break
					}
					if seg == "*" {
						fields := spec.Field.ContainingOneof().Fields()
						for i := range fields.Len() {
							hold(p, fields.Get(i).Name())
						}
						seg = string(spec.Field.Name())
					}
					hold(p, protoreflect.Name(seg))
					p = strings.TrimPrefix(p+"."+seg, ".")
				}
				if len(src.Fields) == 0 {
					whole[p] = true
				}
				for _, name := range src.Fields {
					hold(p, protoreflect.Name(name))
				}
			}
		}
		for _, p := range slices.Sorted(maps.Keys(held)) {
			if whole[p] {
				continue
			}
			md, _, err := b.resolvePath(p, spec)
			if err != nil {
				t.Fatal(err)
			}
			fields := md.Fields()
			for i := range fields.Len() {
				fd := fields.Get(i)
				if !held[p][fd.Name()] && lines[fd.FullName()] == "" && !f.skips(fd) {
					unheld[fd.FullName()] = true
				}
			}
		}
	}
	for _, name := range slices.Sorted(maps.Keys(unheld)) {
		t.Errorf(`%s is in no SIML section, so no .siml file can write or read it. In sim/siml/forms.json, add it to the "fields" of the section whose import category it belongs to (and to that section's row of the table in sim/siml/SPEC.md §5.1), or add it to "skip" if no file should hold it.`,
			name)
	}
}

// Within one section, every entry name, slot word included, names one field, for every spec: settings flattens
// the spec's options and class options into Player's fields, and simple writes each spec's Rotation.
// The messages SPEC.md §6 writes by hand, as a row, a prepull row, a group header or a var line, hold only the fields
// that syntax writes. A field it does not write would be lost from every print, which no read-back check sees.
func TestHandWrittenMessagesHoldOnlyWhatTheirSyntaxWrites(t *testing.T) {
	_, f := testSchema(t)
	for _, m := range []struct {
		message protoreflect.FullName
		syntax  string
		fields  []protoreflect.Name
	}{
		{"proto.APLListItem", "a row", []protoreflect.Name{"hide", "notes", "action"}},
		{"proto.APLPrepullAction", "a prepull row", []protoreflect.Name{"action", "do_at_value", "hide"}},
		{"proto.APLAction", "a row", []protoreflect.Name{"condition"}},
		{"proto.APLGroup", "a group header", []protoreflect.Name{"name", "actions", "variables"}},
		{"proto.APLValueVariable", "a var line and a group reference's with", []protoreflect.Name{"name", "value"}},
	} {
		fields := messageDescriptor(t, string(m.message)).Fields()
		for i := range fields.Len() {
			fd := fields.Get(i)
			if action := fd.ContainingOneof(); slices.Contains(m.fields, fd.Name()) || action != nil && action.Name() == "action" || f.skips(fd) {
				continue
			}
			t.Errorf(`%s has no SIML form: %s writes only %v (sim/siml/SPEC.md §6), so every print would lose it. Give it a form in SPEC.md §6 and in sim/siml, with a corpus case, or add it to "skip" in sim/siml/forms.json if no file should hold it.`,
				fd.FullName(), m.syntax, m.fields)
		}
	}
}

func TestSectionNamesAreUniquePerSpec(t *testing.T) {
	s, f := testSchema(t)
	b := &builder{forms: f}
	var slots []string
	for _, v := range s.Enums["proto.ItemSlot"].Values {
		slots = append(slots, v.Short)
	}
	type collision struct{ section, name, first, second string }
	var collisions []collision
	specsOf := map[collision][]string{}
	specs := Specs()
	for _, spec := range specs {
		for _, sec := range f.Sections {
			named := map[string]string{}
			add := func(name, what string) {
				if prev, ok := named[name]; ok && prev != what {
					c := collision{sec.Name, name, prev, what}
					if specsOf[c] == nil {
						collisions = append(collisions, c)
					}
					specsOf[c] = append(specsOf[c], string(spec.Field.Name()))
				}
				named[name] = what
			}
			for _, src := range sec.Sources {
				md, _, err := b.resolvePath(src.Path, spec)
				if err != nil {
					t.Fatal(err)
				}
				type owned struct {
					message string
					field   schemaField
				}
				var fields []owned
				collect := func(message string, keep func(schemaField) bool) {
					for _, sf := range s.Messages[message].Fields {
						if keep(sf) {
							fields = append(fields, owned{message, sf})
						}
					}
				}
				switch src.Flatten {
				case "spec_options":
					collect(string(spec.Options.FullName()), func(sf schemaField) bool { return sf.Name != "class_options" })
					if spec.ClassOptions != nil {
						collect(string(spec.ClassOptions.Message().FullName()), func(schemaField) bool { return true })
					}
				case "spec_rotation_json":
					collect(string(spec.Rotation.FullName()), func(schemaField) bool { return true })
				default:
					collect(string(md.FullName()), func(sf schemaField) bool {
						return len(src.Fields) == 0 || slices.Contains(src.Fields, sf.Name)
					})
				}
				for _, o := range fields {
					sf := o.field
					what := fmt.Sprintf("%s.%s", o.message, sf.Name)
					if slices.Contains(sec.Body, "slot_rows") && sf.Repeated && sf.Type == "proto.ItemSpec" {
						for _, slot := range slots {
							add(slot, "the slot row "+slot)
						}
						continue
					}
					for _, name := range append([]string{sf.Siml}, sf.Aliases...) {
						add(name, what)
					}
				}
			}
		}
	}
	for _, c := range collisions {
		which := "every spec"
		if len(specsOf[c]) < len(specs) {
			which = strings.Join(specsOf[c], ", ")
		}
		t.Errorf("section %s, for %s: %q names both %s and %s, so an entry line could not tell which it sets. Rename one of the two proto fields, or move one to another section in sim/siml/forms.json.",
			c.section, which, c.name, c.first, c.second)
	}
}

func TestFieldNamesAreUnique(t *testing.T) {
	s, _ := testSchema(t)
	for _, name := range slices.Sorted(maps.Keys(s.Messages)) {
		named := map[string]string{}
		for _, sf := range s.Messages[name].Fields {
			for _, n := range append([]string{sf.Siml}, sf.Aliases...) {
				if prev, ok := named[n]; ok {
					t.Errorf("%s: %q names both the field %s and the field %s (sim/siml/SPEC.md §8.2). Rename one of them in the proto.", name, n, prev, sf.Name)
				}
				named[n] = sf.Name
			}
		}
	}
}

func TestEnumShortNamesAreUnique(t *testing.T) {
	s, f := testSchema(t)
	keywords := f.keywordSet()
	for _, name := range slices.Sorted(maps.Keys(s.Enums)) {
		values := s.Enums[name].Values
		byName := map[string]string{}
		for _, v := range values {
			byName[v.Name] = v.Name
		}
		short := map[string]string{}
		for _, v := range values {
			if prev, ok := short[v.Short]; ok {
				t.Errorf("%s: the values %s and %s both read as %q (sim/siml/SPEC.md §8.2), so a file could not tell them apart. Rename one of the values in the proto.", name, prev, v.Name, v.Short)
			}
			short[v.Short] = v.Name
			if other, ok := byName[v.Short]; ok && other != v.Name {
				t.Errorf("%s: the value %s reads as %q, which is also the proto name of the value %s. Rename one of the values in the proto.", name, v.Name, v.Short, other)
			}
			if keywords[v.Short] {
				t.Errorf("%s: the value %s reads as the SIML keyword %q. Rename the value in the proto.", name, v.Name, v.Short)
			}
		}
	}
}

// Every forms.json pattern still names a field, so a renamed or removed proto field cannot leave a dead entry.
func TestFormsPatternsMatchFields(t *testing.T) {
	s, f := testSchema(t)
	var fields []protoreflect.FieldDescriptor
	for name := range s.Messages {
		fds := messageDescriptor(t, name).Fields()
		for i := range fds.Len() {
			fields = append(fields, fds.Get(i))
		}
	}
	for _, table := range []struct {
		key      string
		patterns []string
	}{
		{"refs", slices.Sorted(maps.Keys(f.Refs))},
		{"indexed", slices.Sorted(maps.Keys(f.Indexed))},
		{"skip", f.Skip},
	} {
		for _, p := range table.patterns {
			if !slices.ContainsFunc(fields, func(fd protoreflect.FieldDescriptor) bool { return fieldMatches(p, fd) }) {
				t.Errorf("sim/siml/forms.json %s: %q matches no field reachable from IndividualSimSettings. Remove the entry, or name the field as the proto does now.", table.key, p)
			}
		}
	}
}

// The derivations reproduce the examples sim/siml/SPEC.md gives.
func TestSchemaDerivations(t *testing.T) {
	s, _ := testSchema(t)

	shortOf := func(enum, value string) string {
		for _, v := range s.Enums[enum].Values {
			if v.Name == value {
				return v.Short
			}
		}
		return "(no such value)"
	}
	for _, c := range []struct{ enum, value, short string }{
		{"proto.Race", "RaceOrc", "orc"},
		{"proto.APLRotation.Type", "TypeAPL", "apl"},
		{"proto.ItemSlot", "ItemSlotMainHand", "main_hand"},
		{"proto.TristateEffect", "TristateEffectRegular", "regular"},
		{"proto.APLValueAutoAttackType", "MainHandAuto", "main_hand_auto"},
		{"proto.APLValueIsExecutePhase.ExecutePhaseThreshold", "E20", "e20"},
		{"proto.HunterOptions.PetAttackSpeed", "PetAttackSpeedNone", "pet_attack_speed_none"},
		{"proto.HunterOptions.PetAttackSpeed", "FasterAttackII", "faster_attack_ii"},
		{"proto.OtherAction", "OtherActionOffensivePotion", "offensive_potion"},
		{"proto.MageArmor", "MageArmorFrostArmor", "frost_armor"},
		{"proto.HunterOptions.PetType", "PetNone", "pet_none"},
	} {
		if got := shortOf(c.enum, c.value); got != c.short {
			t.Errorf("%s.%s reads as %q, want %q", c.enum, c.value, got, c.short)
		}
	}
	if flag := s.Enums["proto.TristateEffect"].Flag; flag != "TristateEffectRegular" {
		t.Errorf("TristateEffect's bare value is %q, want TristateEffectRegular", flag)
	}

	field := func(message, name string) schemaField {
		for _, sf := range s.Messages[message].Fields {
			if sf.Name == name {
				return sf
			}
		}
		t.Fatalf("%s has no field %s in the schema", message, name)
		return schemaField{}
	}
	for _, c := range []struct {
		message, name, siml, alias, printed, ref string
	}{
		{"proto.ConsumesSpec", "offensive_pot_id", "offensive_pot", "offensive_pot_id", "ref", "item"},
		{"proto.ConsumesSpec", "extra_ids", "extra", "extra_ids", "ref", "item"},
		{"proto.ConsumesSpec", "mhImbue_id", "mh_imbue", "mh_imbue_id", "ref", "item"},
		{"proto.ConsumesSpec", "potions", "potions", "", "ref", "item"},
		{"proto.ItemSpec", "id", "id", "", "ref", "item"},
		{"proto.ItemSpec", "enchant", "enchant", "", "ref", "enchant"},
		{"proto.ItemSpec", "random_suffix", "random_suffix", "", "ref", "suffix"},
		{"proto.APLActionCastAllStatBuffCooldowns", "stat_type2", "stat_type2", "", "stat", ""},
		{"proto.Target", "id", "id", "", "number", ""},
		{"proto.RaidBuffs", "battle_shout", "battle_shout", "", "flag_enum", ""},
		{"proto.Player", "distance_from_target", "distance_from_target", "", "number", ""},
		{"proto.Player", "race", "race", "", "enum", ""},
		{"proto.Cooldown", "id", "id", "", "action_id", ""},
		{"proto.IndividualSimSettings", "tanks", "tanks", "", "unit", ""},
		{"proto.Mage.Rotation", "conserveStart", "conserve_start", "", "number", ""},
	} {
		sf := field(c.message, c.name)
		alias := strings.Join(sf.Aliases, ",")
		if sf.Siml != c.siml || alias != c.alias || sf.Printed != c.printed || sf.Ref != c.ref {
			t.Errorf("%s.%s: siml %q, aliases %q, printed %q, ref %q; want %q, %q, %q, %q",
				c.message, c.name, sf.Siml, alias, sf.Printed, sf.Ref, c.siml, c.alias, c.printed, c.ref)
		}
	}
	if sf := field("proto.UnitStats", "pseudo_stats"); sf.Printed != "indexed" || sf.Index != "proto.PseudoStat" {
		t.Errorf("UnitStats.pseudo_stats: printed %q, index %q; want indexed by proto.PseudoStat", sf.Printed, sf.Index)
	}
	for _, skipped := range []struct{ message, name string }{
		{"proto.Player", "database"}, {"proto.Player", "api_version"}, {"proto.APLValue", "uuid"}, {"proto.FeralCatDruid", "rotation"},
	} {
		for _, sf := range s.Messages[skipped.message].Fields {
			if sf.Name == skipped.name {
				t.Errorf("%s.%s is skipped, but the schema lists it", skipped.message, skipped.name)
			}
		}
	}

	kinds := map[string]kindForm{}
	for _, k := range slices.Concat(s.Kinds.Action, s.Kinds.Value) {
		kinds[k.Kind] = k
	}
	// A few long-standing properties of each head (SPEC.md §6.7); the wowsims-siml skill lists them all.
	accessors := map[string][]string{
		"aura":       {"is_known", "is_active", "remaining_time", "num_stacks", "duration"},
		"dot":        {"is_active", "remaining_time", "tick_frequency", "percent_increase"},
		"spell":      {"is_known", "is_ready", "time_to_ready", "cast_time", "current_cost"},
		"boss_spell": {"time_to_ready", "is_casting"},
		"sequence":   {"is_complete", "is_ready", "time_to_ready"},
	}
	subjects := map[string]string{"aura": "aura_id", "dot": "spell_id", "spell": "spell_id", "boss_spell": "spell_id", "sequence": "sequence_name"}
	for head, props := range accessors {
		for _, prop := range props {
			k := kinds[head+"_"+prop]
			if a := k.Accessor; k.Form != "accessor" || a == nil || a.Head != head || a.Prop != prop || a.Subject != subjects[head] {
				t.Errorf("%s_%s: form %q, accessor %+v; want %s(%s).%s", head, prop, k.Form, a, head, subjects[head], prop)
			}
		}
	}
	for kind, unit := range map[string]string{"aura_is_active": "source_unit", "dot_remaining_time": "target_unit", "boss_spell_is_casting": "target_unit", "spell_cast_time": ""} {
		if a := kinds[kind].Accessor; a == nil || a.Unit != unit {
			t.Errorf("%s: accessor %+v, want the unit prefix to fill %q", kind, a, unit)
		}
	}

	forms := map[string][]string{ // SPEC.md §6.4, §6.8
		"short": {"cast_spell", "channel_spell", "wait", "wait_until", "move", "move_duration", "group_reference", "schedule", "sequence",
			"const", "and", "or", "not", "cmp", "math", "variable_ref", "variable_placeholder"},
		// A sample of core kinds; the wowsims-siml skill lists every kind with its form.
		"bare":        {"autocast_other_cooldowns", "current_time", "remaining_time", "current_energy", "gcd_is_ready", "input_delay"},
		"unit_prefix": {"current_health", "current_health_percent", "current_mana", "current_mana_percent", "unit_is_moving", "unit_distance"},
		"list_args":   {"max", "min", "strict_sequence"},
	}
	for form, names := range forms {
		for _, name := range names {
			if got := kinds[name].Form; got != form {
				t.Errorf("%s: form %q, want %q", name, got, form)
			}
		}
	}
	for kind, listField := range map[string]string{"and": "vals", "or": "vals", "max": "vals", "strict_sequence": "actions"} {
		if got := kinds[kind].ListField; got != listField {
			t.Errorf("%s: list field %q, want %q", kind, got, listField)
		}
	}
	for kind, positional := range map[string]string{
		"is_execute_phase":             "threshold",
		"auto_time_to_next":            "auto_type",
		"auto_time_since_last":         "auto_type",
		"action_group_used":            "name",
		"multiple_cd_usages":           "base_spell_id",
		"selected_potion":              "potion_id",
		"energy_time_to_target":        "target_energy",
		"totem_remaining_time":         "totem_type",
		"reset_sequence":               "sequence_name",
		"cast_all_stat_buff_cooldowns": "stat_type1",
		"multidot":                     "spell_id",
		"warlock_active_pet":           "summon",
		"change_target":                "",
	} {
		if k := kinds[kind]; k.Form != "generic" || k.Positional != positional {
			t.Errorf("%s: form %q, positional %q; want generic with %q", kind, k.Form, k.Positional, positional)
		}
	}
	if k := kinds["aura_is_active"]; k.Positional != "aura_id" {
		t.Errorf("aura_is_active: positional %q, want aura_id, so the generic form aura_is_active(x) reads", k.Positional)
	}

	var names []string
	for _, sec := range s.Sections {
		names = append(names, sec.Name)
	}
	want := []string{"settings", "gear", "item_swap", "consumables", "buffs", "party_buffs", "raid_buffs", "debuffs", "encounter", "sim",
		"apl", "simple", "cooldowns", "prepull", "priority", "group"}
	if !slices.Equal(names, want) {
		t.Errorf("sections %v, want %v (SPEC.md §5.1, §12.1)", names, want)
	}
}

// Each enum short-name rule of SPEC.md §8.2, on enums built for it, since the protos need not hold a value that
// takes each rule.
func TestEnumShortNameRules(t *testing.T) {
	keywords := map[string]bool{"none": true}
	for _, c := range []struct {
		enum   string
		values []string
		want   []string
	}{
		{"SwipeUsage", []string{"SwipeUsage_Never", "SwipeUsageAlways"}, []string{"never", "always"}}, // E and the _ after it go.
		{"Mode", []string{"Mode_", "ModeAuto"}, []string{"mode_", "auto"}},                            // Nothing left after E and _ is snake(V).
		{"E", []string{"E20", "E21"}, []string{"e20", "e21"}},                                         // A digit first is snake(V).
		{"PetAttackSpeed", []string{"PetAttackSpeedNone"}, []string{"pet_attack_speed_none"}},         // A keyword is snake(V).
		{"FinishingMove", []string{"None", "Rip"}, []string{"finishing_move_none", "rip"}},            // snake(V) a keyword too takes snake(E) in front.
		{"Dup", []string{"DupFoo", "Foo"}, []string{"dup_foo", "foo"}},                                // Two values of one short name take snake(V).
	} {
		values := make([]*descriptorpb.EnumValueDescriptorProto, len(c.values))
		for i, v := range c.values {
			values[i] = &descriptorpb.EnumValueDescriptorProto{Name: googleProto.String(v), Number: googleProto.Int32(int32(i))}
		}
		fd, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{
			Name:     googleProto.String(c.enum + ".proto"),
			Package:  googleProto.String("short_names"),
			EnumType: []*descriptorpb.EnumDescriptorProto{{Name: googleProto.String(c.enum), Value: values}},
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got := enumShortNames(fd.Enums().Get(0), keywords); !slices.Equal(got, c.want) {
			t.Errorf("%s %v read as %v, want %v", c.enum, c.values, got, c.want)
		}
	}
}
