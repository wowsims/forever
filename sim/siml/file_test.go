package siml

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

func TestKindOf(t *testing.T) {
	for _, c := range []struct {
		path    string
		section string
		fails   bool
	}{
		{path: "ui/specs/hunter/dps/apls/default.apl.siml", section: "apl"},
		{path: `ui\specs\hunter\dps\gear_sets\p1.gear.siml`, section: "gear"},
		{path: "x.consumables.siml", section: "consumables"},
		{path: "defaults.siml"},
		{path: "builds.d/testing.siml"},
		{path: "x..siml"},
		{path: "x.build.siml", fails: true},
		{path: "default.apl.json", fails: true},
	} {
		k, err := KindOf(c.path)
		switch {
		case c.fails && err == nil:
			t.Errorf("KindOf(%q) = %+v, want an error", c.path, k)
		case !c.fails && err != nil:
			t.Errorf("KindOf(%q): %v", c.path, err)
		case !c.fails && k.Section != c.section:
			t.Errorf("KindOf(%q) = %q, want %q", c.path, k.Section, c.section)
		}
	}
}

// Columns count UTF-16 code units: the emoji is two, the é one.
func TestDiagnosticsError(t *testing.T) {
	src := "siml 1\nspec hunter\n\npriority:\n  cast x if \"😀é\" == @\n"
	_, err := Parse("rotation.apl.siml", []byte(src), &Names{}, nil)
	var diags Diagnostics
	if !errors.As(err, &diags) {
		t.Fatalf("Parse gave %v, want Diagnostics", err)
	}
	want := "rotation.apl.siml:5:22 bad-character '@' starts nothing SIML reads"
	if err.Error() != want {
		t.Errorf("the error reads\n%s\nwant\n%s", err, want)
	}
	if d := diags[0]; d.Severity != Error || d.End != (Position{Line: 5, Col: 23}) {
		t.Errorf("the diagnostic is %+v, want an error ending at 5:23", d)
	}
}

// A byte order mark at the start is skipped, so columns count from the character after it; a second one is a
// bad character, written as its escape.
func TestByteOrderMark(t *testing.T) {
	src := "\uFEFF\uFEFFsiml 1\nspec hunter\n"
	_, err := Parse("rotation.apl.siml", []byte(src), &Names{}, nil)
	var diags Diagnostics
	if !errors.As(err, &diags) {
		t.Fatalf("Parse gave %v, want Diagnostics", err)
	}
	if want := `rotation.apl.siml:1:1 bad-character '\uFEFF' starts nothing SIML reads`; diags[0].String() != want {
		t.Errorf("the first diagnostic reads\n%s\nwant\n%s", diags[0], want)
	}
	if f, err := Parse("rotation.apl.siml", []byte(src[3:]), &Names{}, nil); err != nil || f.Spec != proto.Spec_SpecHunter {
		t.Errorf("with one byte order mark, Parse gave %v, %v; want a hunter rotation", f, err)
	}
}

// A warrior's table: a class spell and target aura, and an item, enchant and suffix.
var warriorNames = &Names{
	Classes: map[proto.Class]SpellNames{proto.Class_ClassWarrior: {
		NamespaceSpell:      {"heroic_strike": {SpellID: 25286, Tag: 1, Rank: 9}},
		NamespaceTargetAura: {"sunder_armor": {SpellID: 11597, Rank: 5}},
	}},
	Items:    map[string]int32{"sword": 19352, "sapper": 10646},
	Enchants: map[string]EnchantName{"crusader": {EffectID: 1900, Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}}},
	Suffixes: map[string]int32{"of_agility": 5},
}

// The file gives what each name means, where it is written, and the raw JSON that holds an id.
func TestParseFile(t *testing.T) {
	src := "siml 1\nspec dps_warrior\n\ngear:\n  main_hand  sword of_agility + crusader\n\npriority:\n" +
		"  cast heroic_strike if target.aura(sunder_armor).is_active\n  cast item.sapper\n  cast `{\"spellId\":7}`\n"
	f, err := Parse("arms.siml", []byte(src), warriorNames, nil)
	if err != nil {
		t.Fatal(err)
	}
	if f.Kind.Section != "" || f.Spec != proto.Spec_SpecDpsWarrior || !slices.Equal(f.Categories, []string{"g", "r"}) {
		t.Errorf("the file is %+v", f)
	}
	player := f.Settings.GetPlayer()
	if player.GetClass() != proto.Class_ClassWarrior || player.GetDpsWarrior() == nil {
		t.Errorf("the player is %v, want a warrior with the dps_warrior spec set", player)
	}
	strike := &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: 25286}, Tag: 1, Rank: 9}
	if id := f.Rotation().GetPriorityList()[0].GetAction().GetCastSpell().GetSpellId(); !googleProto.Equal(id, strike) {
		t.Errorf("the row casts %v, want %v", id, strike)
	}
	if item := player.GetEquipment().GetItems()[proto.ItemSlot_ItemSlotMainHand]; item.GetId() != 19352 || item.GetRandomSuffix() != 5 || item.GetEnchant() != 1900 {
		t.Errorf("the main hand is %v", item)
	}
	row := &SlotRow{Slot: proto.ItemSlot_ItemSlotMainHand, Item: 19352}
	at := func(line, col, length int) (Position, Position) {
		return Position{Line: line, Col: col}, Position{Line: line, Col: col + length}
	}
	ref := func(name string, ns Namespace, id int32, row *SlotRow, line, col int) NameRef {
		start, end := at(line, col, len(name))
		return NameRef{Name: name, Namespace: ns, Start: start, End: end, ID: id, Row: row}
	}
	want := []NameRef{
		ref("sword", NamespaceItem, 19352, row, 5, 14),
		ref("of_agility", NamespaceSuffix, 5, row, 5, 20),
		ref("crusader", NamespaceEnchant, 1900, row, 5, 33),
		ref("heroic_strike", NamespaceSpell, 25286, nil, 8, 8),
		ref("sunder_armor", NamespaceTargetAura, 11597, nil, 8, 37),
		ref("sapper", NamespaceItem, 10646, nil, 9, 13),
	}
	if !slices.EqualFunc(f.Names, want, func(a, b NameRef) bool {
		return a.Name == b.Name && a.Namespace == b.Namespace && a.Start == b.Start && a.End == b.End && a.ID == b.ID &&
			(a.Row == nil) == (b.Row == nil) && (a.Row == nil || *a.Row == *b.Row)
	}) {
		t.Errorf("the names are\n%+v\nwant\n%+v", f.Names, want)
	}
	if len(f.RawRefs) != 1 || f.RawRefs[0].Code != "raw-ref" || f.RawRefs[0].Start != (Position{Line: 10, Col: 8}) {
		t.Errorf("the raw refs are %v, want raw-ref at 10:8", f.RawRefs)
	}
	if len(f.Warnings) != 0 {
		t.Errorf("the warnings are %v, want none", f.Warnings)
	}
}

// ReadFile reads with the committed table, and a build without it reads only a file that writes no name.
func TestReadFile(t *testing.T) {
	dir := t.TempDir()
	write := func(name, text string) string {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	plain := write("plain.settings.siml", "siml 1\nrace orc\n")
	write("gear/p1.gear.siml", "siml 1\n\nhead lieutenant_commanders_chain_helmet + falcons_call\n")
	write("rotation.apl.siml", "siml 1\nspec hunter\n\npriority:\n  cast multi_shot\n")
	main := write("main.siml", "siml 1\nspec hunter\nuse \"gear/p1.gear.siml\"\nuse \"rotation.apl.siml\"\n")

	if f, err := ReadFile(plain); err != nil || f.Settings.GetPlayer().GetRace() != proto.Race_RaceOrc {
		t.Errorf("ReadFile(%s) = %v, %v; want an orc", plain, f, err)
	}
	f, err := ReadFile(main)
	names, noTable := DefaultNames()
	if noTable != nil {
		if err == nil || !strings.Contains(err.Error(), "with_db") {
			t.Errorf("without the table, ReadFile(%s) = %v, want an error naming the with_db tag", main, err)
		}
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"g", "r"}; !slices.Equal(f.Categories, want) {
		t.Errorf("the categories are %v, want %v", f.Categories, want)
	}
	helmet, falcons := names.Items["lieutenant_commanders_chain_helmet"], names.Enchants["falcons_call"].EffectID
	if items := f.Equipment().GetItems(); len(items) != 17 || items[0].GetId() != helmet || items[0].GetEnchant() != falcons {
		t.Errorf("the equipment is %v, want 17 items from gear/p1.gear.siml", items)
	}
	if swap := f.Settings.GetPlayer().GetItemSwap(); f.Rotation().GetType() != proto.APLRotation_TypeAPL || len(f.Rotation().GetPriorityList()) != 1 || swap != nil {
		t.Errorf("the rotation is %v and the item swap %v, want the used rotation and no item swap", f.Rotation(), swap)
	}
}

// A file with an error is not formatted; its diagnostics are the error.
func TestFormatError(t *testing.T) {
	_, err := Format("x.apl.siml", []byte("siml 1\nspec hunter\n\npriority:\n  cast x\n"), &Names{}, nil)
	var diags Diagnostics
	if !errors.As(err, &diags) || len(diags) != 1 || diags[0].Code != "unknown-name" {
		t.Errorf("Format gave %v, want unknown-name", err)
	}
}

// A rotation nested deeper than a statement holds prints without running out of stack, and its print reads as the
// one error that says so.
func TestPrintDeepNesting(t *testing.T) {
	condition := &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: "1"}}}
	for range 1500 {
		condition = &proto.APLValue{Value: &proto.APLValue_Not{Not: &proto.APLValueNot{Val: condition}}}
	}
	rotation := &proto.APLRotation{Type: proto.APLRotation_TypeAPL, PriorityList: []*proto.APLListItem{{Action: &proto.APLAction{
		Condition: condition,
		Action:    &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: 1}}}},
	}}}}
	settings := &proto.IndividualSimSettings{Player: &proto.Player{Rotation: rotation}}
	text := Print(Kind{Section: sectionAPL}, proto.Spec_SpecHunter, settings, nil, &Names{})
	_, err := Parse("x.apl.siml", text, &Names{}, nil)
	var diags Diagnostics
	if !errors.As(err, &diags) || len(diags) != 1 || diags[0].Message != "the statement nests deeper than 256" {
		t.Errorf("the print reads as %v, want the one error that the statement nests deeper than 256", err)
	}
}
