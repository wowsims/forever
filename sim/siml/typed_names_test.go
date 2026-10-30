package siml

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

// One name, x, in each of a hunter's spell-like namespaces, each meaning another spell id.
var namespaceNames = &Names{
	Classes: map[proto.Class]SpellNames{proto.Class_ClassHunter: {
		NamespaceSpell:      {"x": {SpellID: 1}, "multi_shot": {SpellID: 25294, Rank: 5}},
		NamespaceAura:       {"x": {SpellID: 2}, "only_aura": {SpellID: 7}},
		NamespaceTargetAura: {"x": {SpellID: 3}},
		NamespacePetAura:    {"x": {SpellID: 5}},
	}},
	Shared: SpellNames{
		NamespaceSpell:       {"potion_spell": {SpellID: 9}},
		NamespaceTargetSpell: {"x": {SpellID: 4}},
	},
	Items:    map[string]int32{"sapper": 10646},
	Enchants: map[string]EnchantName{"crusader": {EffectID: 1900, Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}}},
}

func rotationText(rows ...string) string {
	return "siml 1\nspec hunter\n\npriority:\n  " + strings.Join(rows, "\n  ") + "\n"
}

// A unit that does not read leaves an aura_id reading the player's auras, so the unit's error stands alone.
func TestUnreadUnitReadsThePlayersAuras(t *testing.T) {
	_, err := Parse("x.apl.siml", []byte(rotationText("cast x if aura_is_active(only_aura, source_unit=nobody)")), namespaceNames, nil)
	var diags Diagnostics
	if !errors.As(err, &diags) || len(diags) != 1 || diags[0].Code != "wrong-type" {
		t.Errorf("the parse gave %v, want one wrong-type at the unit", err)
	}
}

func TestUnknownNames(t *testing.T) {
	for _, c := range []struct {
		path, text string
		at         Position
		message    string
	}{
		{"x.apl.siml", rotationText("cast sapper"), Position{Line: 5, Col: 8}, "sapper is an item; write item.sapper"},
		{"x.apl.siml", rotationText("cast only_aura"), Position{Line: 5, Col: 8}, "only_aura is an aura; this position takes a spell"},
		{"x.apl.siml", rotationText("cast multi_shoot"), Position{Line: 5, Col: 8}, "no spell is named multi_shoot; did you mean multi_shot?"},
		{"x.apl.siml", rotationText("cast item.x"), Position{Line: 5, Col: 13}, "x is a spell; item. takes an item"},
		{"x.apl.siml", rotationText("cast x if target.aura(only_aura).is_active"), Position{Line: 5, Col: 25}, "only_aura is an aura; this position takes an aura on a target"},
		{"x.gear.siml", "siml 1\n\nmain_hand crusader\n", Position{Line: 3, Col: 11}, "crusader is an enchant; a slot takes an item"},
		{"x.gear.siml", "siml 1\n\nmain_hand sapper + sapper\n", Position{Line: 3, Col: 20}, "sapper is an item; + takes an enchant"},
		{"x.cooldowns.siml", "siml 1\n\ncooldowns(id=multi_shot)\n", Position{Line: 3, Col: 14}, "no spell is named multi_shot; with no spec, a file reads only the spells and auras every class shares"},
	} {
		_, err := Parse(c.path, []byte(c.text), namespaceNames, nil)
		var diags Diagnostics
		if !errors.As(err, &diags) || len(diags) != 1 || diags[0].Code != "unknown-name" || diags[0].Start != c.at || diags[0].Message != c.message {
			t.Errorf("%q gave %v, want unknown-name at %d:%d: %s", c.text, err, c.at.Line, c.at.Col, c.message)
		}
	}
	_, err := Parse("x.consumables.siml", []byte("siml 1\n\noffensive_pot item.sapper\n"), namespaceNames, nil)
	var diags Diagnostics
	if !errors.As(err, &diags) || len(diags) != 1 || diags[0].Code != "wrong-type" {
		t.Errorf("item.<name> outside an ActionID gave %v, want wrong-type", err)
	}
}

// A name prints by its spell id and tag whatever the rank, and reads back with its entry's rank; what no name writes
// prints as raw JSON that reads back exactly.
func TestPrintRanksAndRawRefs(t *testing.T) {
	cast := func(id *proto.ActionID) *proto.APLListItem {
		return &proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: id}}}}
	}
	r := &proto.APLRotation{Type: proto.APLRotation_TypeAPL, PriorityList: []*proto.APLListItem{
		cast(&proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: 25294}, Rank: 3}),
		cast(&proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: 99}, Rank: 2}),
		cast(&proto.ActionID{RawId: &proto.ActionID_ItemId{ItemId: 10646}, Rank: 1}),
		cast(&proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: 25294}, Tag: 1}),
	}}
	text := string(printRotation(proto.Spec_SpecHunter, r, namespaceNames))
	want := rotationText("cast multi_shot", "cast `{\"spellId\":99,\"rank\":2}`", "cast `{\"itemId\":10646,\"rank\":1}`", "cast `{\"spellId\":25294,\"tag\":1}`")
	if text != want {
		t.Fatalf("the rotation prints as\n%s\nwant\n%s", text, want)
	}
	f, err := Parse("x.apl.siml", []byte(text), namespaceNames, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.PriorityList[0].Action.GetCastSpell().SpellId.Rank = 5
	if !googleProto.Equal(f.Rotation(), r) {
		t.Errorf("the print reads back as %v, want %v", f.Rotation(), r)
	}
	if len(f.RawRefs) != 3 {
		t.Errorf("the raw refs are %v, want the three raw casts", f.RawRefs)
	}
}

// A class's own name comes before a shared one, and a file printed with no spec writes only shared names.
func TestPrintClassNames(t *testing.T) {
	names := &Names{
		Classes: map[proto.Class]SpellNames{proto.Class_ClassHunter: {NamespaceSpell: {"own": {SpellID: 10}, "class_only": {SpellID: 11}}}},
		Shared:  SpellNames{NamespaceSpell: {"shared": {SpellID: 10}}},
	}
	settings := &proto.IndividualSimSettings{Player: &proto.Player{
		Spec: &proto.Player_Hunter{Hunter: &proto.Hunter{}},
		Rotation: &proto.APLRotation{Type: proto.APLRotation_TypeAPL, PriorityList: []*proto.APLListItem{{Action: &proto.APLAction{
			Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: 10}}}},
		}}}},
		Cooldowns: &proto.Cooldowns{Cooldowns: []*proto.Cooldown{
			{Id: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: 10}}},
			{Id: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: 11}}},
		}},
	}}
	if got, want := string(Print(Kind{Section: "apl"}, proto.Spec_SpecHunter, settings, nil, names)), rotationText("cast own"); got != want {
		t.Errorf("the rotation prints as\n%s\nwant\n%s", got, want)
	}
	want := "siml 1\nspec hunter\n\ncooldowns(id=own)\ncooldowns(id=class_only)\n"
	if got := string(Print(Kind{Section: "cooldowns"}, proto.Spec_SpecHunter, settings, nil, names)); got != want {
		t.Errorf("the cooldowns print as\n%s\nwant\n%s", got, want)
	}
	settings.Player.Spec = nil
	want = "siml 1\n\ncooldowns(id=shared)\ncooldowns(id=`{\"spellId\":11}`)\n"
	if got := string(Print(Kind{Section: "cooldowns"}, proto.Spec_SpecUnknown, settings, nil, names)); got != want {
		t.Errorf("the cooldowns print with no spec as\n%s\nwant\n%s", got, want)
	}
}

// An enchant prints with the name whose slots hold its slot, else its first name; a slot row with an unnamed part
// prints as raw JSON.
func TestPrintSlotRows(t *testing.T) {
	names := &Names{
		Items: map[string]int32{"boots": 1, "bracers": 2, "helm": 3},
		Enchants: map[string]EnchantName{
			"boots_agility":  {EffectID: 904, Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotFeet}},
			"bracer_agility": {EffectID: 904, Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotWrist}},
		},
	}
	items := make([]*proto.ItemSpec, 17)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotHead] = &proto.ItemSpec{Id: 3, Enchant: 904}
	items[proto.ItemSlot_ItemSlotWrist] = &proto.ItemSpec{Id: 2, Enchant: 904}
	items[proto.ItemSlot_ItemSlotFeet] = &proto.ItemSpec{Id: 1, Enchant: 904, RandomSuffix: 7}
	want := "siml 1\n\nhead       helm + boots_agility\nwrist      bracers + bracer_agility\nfeet       `{\"id\":1,\"randomSuffix\":7,\"enchant\":904}`\n"
	if got := string(printGear(&proto.EquipmentSpec{Items: items}, names)); got != want {
		t.Errorf("the gear prints as\n%s\nwant\n%s", got, want)
	}
}

// Raw JSON that holds a spell, item, enchant or suffix id is raw-ref, and raw JSON that holds none is not.
func TestRawRefs(t *testing.T) {
	src := "siml 1\nspec hunter\n\ngear:\n  head       `{\"id\":3}`\n\nconsumables:\n  `{\"offensivePotId\":5}`\n  flask `0`\n\n" +
		"priority:\n  cast other.offensive_potion if aura_is_active(x, source_unit=`{\"type\":\"Pet\"}`)\n  cast `{\"spellId\":1}`\n"
	f, err := Parse("x.siml", []byte(src), namespaceNames, nil)
	if err != nil {
		t.Fatal(err)
	}
	var at []Position
	for _, d := range f.RawRefs {
		at = append(at, d.Start)
	}
	if want := []Position{{Line: 5, Col: 14}, {Line: 8, Col: 3}, {Line: 13, Col: 8}}; !slices.Equal(at, want) {
		t.Errorf("the raw refs are at %v, want %v", at, want)
	}

	// A file with an error still lists the names and raw refs that read.
	f, err = Parse("x.apl.siml", []byte(rotationText("cast x", "cast `{\"spellId\":1}`", "cast nothing")), namespaceNames, nil)
	if err == nil || f.Settings != nil || len(f.Names) != 1 || f.Names[0].Name != "x" || len(f.RawRefs) != 1 {
		t.Errorf("the file with an error is %+v (%v), want the name x, one raw ref and no settings", f, err)
	}
}
