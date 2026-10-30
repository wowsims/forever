package siml

import (
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

// A table in the form of SPEC.md §7.4, as Encode writes it.
const namesFixture = `{
  "classes": {
    "warrior": {
      "aura": {
        "battle_shout": {"spellId": 25289, "registeredBy": ["dps_warrior"]},
        "battle_shout_external": {"spellId": 25289, "tag": -1, "registeredBy": ["dps_warrior", "protection_warrior"]},
        "overpower": {"spellId": 1282733, "registeredBy": []}
      },
      "spell": {
        "heroic_strike": {"spellId": 25286, "tag": 1, "registeredBy": ["dps_warrior"]},
        "heroic_strike_tag0": {"spellId": 25286, "rank": 9, "registeredBy": ["dps_warrior"]},
        "overpower": {"spellId": 11585, "rank": 4, "registeredBy": ["dps_warrior"]}
      }
    }
  },
  "enchants": {
    "boots_agility": {"effectId": 904, "slots": ["feet"]},
    "bracer_agility": {"effectId": 904, "slots": ["wrist"]},
    "necklace_agility": {"effectId": 904, "slots": ["neck"]}
  },
  "items": {
    "goblin_sapper_charge": 10646,
    "jom_gabbar": 23570
  },
  "shared": {
    "spell": {
      "jom_gabbar": {"spellId": 29602, "registeredBy": ["dps_warrior", "hunter"]}
    }
  },
  "suffixes": {
    "of_the_tiger": 12727
  }
}
`

func fixtureNames(t *testing.T) *Names {
	t.Helper()
	names, err := ParseNames([]byte(namesFixture))
	if err != nil {
		t.Fatal(err)
	}
	return names
}

func spellID(id, tag, rank int32) *proto.ActionID {
	return &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}, Tag: tag, Rank: rank}
}

func TestNamesEncodeIsCanonical(t *testing.T) {
	if got := string(fixtureNames(t).Encode()); got != namesFixture {
		t.Errorf("Encode wrote\n%s\nwant\n%s", got, namesFixture)
	}
	if got := string((&Names{}).Encode()); got != "{\n}\n" {
		t.Errorf("Encode of an empty table wrote %q, want %q", got, "{\n}\n")
	}
}

func TestNamesSpellLookup(t *testing.T) {
	names := fixtureNames(t)
	for _, c := range []struct {
		class proto.Class
		ns    Namespace
		name  string
		want  *proto.ActionID
	}{
		{proto.Class_ClassWarrior, NamespaceSpell, "heroic_strike", spellID(25286, 1, 0)},
		{proto.Class_ClassWarrior, NamespaceSpell, "heroic_strike_tag0", spellID(25286, 0, 9)},
		{proto.Class_ClassWarrior, NamespaceSpell, "overpower", spellID(11585, 0, 4)},
		{proto.Class_ClassWarrior, NamespaceAura, "overpower", spellID(1282733, 0, 0)},
		{proto.Class_ClassWarrior, NamespaceSpell, "jom_gabbar", spellID(29602, 0, 0)},
		{proto.Class_ClassHunter, NamespaceSpell, "jom_gabbar", spellID(29602, 0, 0)},
		{proto.Class_ClassUnknown, NamespaceSpell, "jom_gabbar", spellID(29602, 0, 0)},
		{proto.Class_ClassHunter, NamespaceSpell, "heroic_strike", nil},
		{proto.Class_ClassUnknown, NamespaceSpell, "heroic_strike", nil},
		{proto.Class_ClassWarrior, NamespaceTargetAura, "overpower", nil},
	} {
		got, ok := names.Spell(c.class, c.ns, c.name)
		if ok != (c.want != nil) || ok && !googleProto.Equal(got, c.want) {
			t.Errorf("Spell(%s, %s, %s) = %v, %v; want %v", c.class, c.ns, c.name, got, ok, c.want)
		}
	}
}

func TestNamesSpellNameIgnoresTheRank(t *testing.T) {
	names := fixtureNames(t)
	for _, c := range []struct {
		class proto.Class
		ns    Namespace
		id    *proto.ActionID
		want  string
	}{
		{proto.Class_ClassWarrior, NamespaceSpell, spellID(25286, 1, 0), "heroic_strike"},
		{proto.Class_ClassWarrior, NamespaceSpell, spellID(25286, 1, 9), "heroic_strike"},
		{proto.Class_ClassWarrior, NamespaceSpell, spellID(25286, 0, 0), "heroic_strike_tag0"},
		{proto.Class_ClassWarrior, NamespaceAura, spellID(25289, -1, 7), "battle_shout_external"},
		{proto.Class_ClassWarrior, NamespaceSpell, spellID(29602, 0, 0), "jom_gabbar"},
		{proto.Class_ClassHunter, NamespaceSpell, spellID(29602, 0, 0), "jom_gabbar"},
		{proto.Class_ClassHunter, NamespaceSpell, spellID(25286, 1, 0), ""},
		{proto.Class_ClassWarrior, NamespaceSpell, spellID(25286, 2, 0), ""},
		{proto.Class_ClassWarrior, NamespaceSpell, &proto.ActionID{RawId: &proto.ActionID_ItemId{ItemId: 10646}}, ""},
	} {
		got, ok := names.SpellName(c.class, c.ns, c.id)
		if ok != (c.want != "") || got != c.want {
			t.Errorf("SpellName(%s, %s, %v) = %q, %v; want %q", c.class, c.ns, c.id, got, ok, c.want)
		}
	}
}

func TestNamesItemsEnchantsSuffixes(t *testing.T) {
	names := fixtureNames(t)
	if id, ok := names.Item("goblin_sapper_charge"); !ok || id != 10646 {
		t.Errorf("Item(goblin_sapper_charge) = %d, %v; want 10646", id, ok)
	}
	if name, ok := names.ItemName(23570); !ok || name != "jom_gabbar" {
		t.Errorf("ItemName(23570) = %q, %v; want jom_gabbar", name, ok)
	}
	if id, ok := names.Suffix("of_the_tiger"); !ok || id != 12727 {
		t.Errorf("Suffix(of_the_tiger) = %d, %v; want 12727", id, ok)
	}
	if name, ok := names.SuffixName(12727); !ok || name != "of_the_tiger" {
		t.Errorf("SuffixName(12727) = %q, %v; want of_the_tiger", name, ok)
	}
	if id, ok := names.Enchant("bracer_agility"); !ok || id != 904 {
		t.Errorf("Enchant(bracer_agility) = %d, %v; want 904", id, ok)
	}
	for _, c := range []struct {
		slot proto.ItemSlot
		want string
	}{
		{proto.ItemSlot_ItemSlotFeet, "boots_agility"},
		{proto.ItemSlot_ItemSlotWrist, "bracer_agility"},
		{proto.ItemSlot_ItemSlotNeck, "necklace_agility"},
		{proto.ItemSlot_ItemSlotHead, "boots_agility"}, // No name fits the head: the effect's first name.
	} {
		if got, ok := names.EnchantName(904, c.slot); !ok || got != c.want {
			t.Errorf("EnchantName(904, %s) = %q, %v; want %q", c.slot, got, ok, c.want)
		}
	}
	if got, ok := names.EnchantName(2586, proto.ItemSlot_ItemSlotHead); ok {
		t.Errorf("EnchantName(2586, head) = %q, want no name", got)
	}
}

// An effect a two-hand and a one-hand weapon row share prints with the one-hand row's name in either hand, since that
// row fits the one-handed Assassination Blade and a two-hander alike.
func TestEnchantNameFitsEitherWeapon(t *testing.T) {
	names := &Names{
		Items: map[string]int32{"assassination_blade": 13036, "heartseeker": 12783},
		Enchants: map[string]EnchantName{
			"enchant_2h_weapon_lesser_impact": {EffectID: 1897, Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}},
			"weapon_superior_striking": {EffectID: 1897,
				Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand}},
		},
	}
	items := make([]*proto.ItemSpec, 17)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 13036, Enchant: 1897}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: 12783, Enchant: 1897}
	want := "siml 1\n\nmain_hand  assassination_blade + weapon_superior_striking\noff_hand   heartseeker + weapon_superior_striking\n"
	if got := string(printGear(&proto.EquipmentSpec{Items: items}, names)); got != want {
		t.Errorf("the gear prints as\n%s\nwant\n%s", got, want)
	}
}

func TestNamesVisibleAndHolding(t *testing.T) {
	names := fixtureNames(t)
	if got, want := names.Visible(proto.Class_ClassWarrior, NamespaceSpell), []string{"heroic_strike", "heroic_strike_tag0", "jom_gabbar", "overpower"}; !slices.Equal(got, want) {
		t.Errorf("Visible(warrior, spell) = %v, want %v", got, want)
	}
	if got, want := names.Visible(proto.Class_ClassHunter, NamespaceSpell), []string{"jom_gabbar"}; !slices.Equal(got, want) {
		t.Errorf("Visible(hunter, spell) = %v, want %v", got, want)
	}
	if got, want := names.Visible(proto.Class_ClassUnknown, NamespaceItem), []string{"goblin_sapper_charge", "jom_gabbar"}; !slices.Equal(got, want) {
		t.Errorf("Visible(item) = %v, want %v", got, want)
	}
	if got, want := names.Holding(proto.Class_ClassWarrior, "overpower"), []Namespace{NamespaceSpell, NamespaceAura}; !slices.Equal(got, want) {
		t.Errorf("Holding(warrior, overpower) = %v, want %v", got, want)
	}
	if got, want := names.Holding(proto.Class_ClassHunter, "jom_gabbar"), []Namespace{NamespaceSpell, NamespaceItem}; !slices.Equal(got, want) {
		t.Errorf("Holding(hunter, jom_gabbar) = %v, want %v", got, want)
	}
	if got := names.Holding(proto.Class_ClassHunter, "overpower"); len(got) != 0 {
		t.Errorf("Holding(hunter, overpower) = %v, want none", got)
	}
}

func TestParseNamesRejects(t *testing.T) {
	for _, c := range []struct{ table, want string }{
		{`{"classes": {"warrior": {"spell": {"a": {"spellId": 1, "registeredBy": []}, "b": {"spellId": 1, "rank": 2, "registeredBy": []}}}}}`, "names spell 1 twice, as a and b"},
		{`{"classes": {"warrior": {"spell": {"a": {"spellId": 1, "registeredBy": []}}}}, "shared": {"spell": {"a": {"spellId": 2, "registeredBy": []}}}}`, "a means spell 1 in warrior.spell and spell 2 in shared.spell"},
		{`{"classes": {"knight": {}}}`, "knight is no class"},
		{`{"shared": {"item": {}}}`, "shared.item is no spell-like namespace"},
		{`{"items": {"Sapper": 1}}`, "items.Sapper is no name of a positive id"},
		{`{"items": {"a": 1, "b": 1}}`, "items names 1 twice, as a and b"},
		{`{"suffixes": {"a": 0}}`, "suffixes.a is no name of a positive id"},
		{`{"enchants": {"a": {"effectId": 1, "slots": ["hat"]}}}`, "enchants.a: hat is no slot"},
		{`{"spells": {}}`, `the table has no key spells`},
		{`{"Items": {"a": 1}}`, `the table has no key Items`},
		{`{"items": {"a": 1.5}}`, `items.a is no name of a positive id`},
		{`{"shared": {"spell": {"a": {"spellId": 1, "tag": "1", "registeredBy": []}}}}`, `shared.spell.a has a tag or rank that is no int32`},
		{`{"shared": {"spell": {"a": {"spellId": 1, "registeredBy": ["knight"]}}}}`, `shared.spell.a is registered by "knight", which is no spec`},
		{`null`, "the table is null"},
		{`{} {}`, "text after the table"},
	} {
		_, err := ParseNames([]byte(c.table))
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("ParseNames(%s) gave %v, want an error with %q", c.table, err, c.want)
		}
	}
}

func TestNamesReserved(t *testing.T) {
	for _, word := range []string{"cast", "none", "aura", "boss_spell", "target", "target2", "pet1", "self", "other", "item", "spell"} {
		if !Reserved(word) {
			t.Errorf("Reserved(%s) = false, want true", word)
		}
	}
	for _, word := range []string{"heroic_strike", "targets", "pet", "enchant", "suffix"} {
		if Reserved(word) {
			t.Errorf("Reserved(%s) = true, want false", word)
		}
	}
}

func TestClassName(t *testing.T) {
	if got := ClassName(proto.Class_ClassWarrior); got != "warrior" {
		t.Errorf("ClassName(ClassWarrior) = %q, want warrior", got)
	}
}
