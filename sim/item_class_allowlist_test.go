//go:build with_db

package sim

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/forever/assets/database"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

// The class allowlist each item states in the database reaches the sim's item as it is, and names
// playable classes only, each once. The database loads an item id once, its first row.
func TestClassRestrictedItemsCarryTheirAllowlist(t *testing.T) {
	playable := []proto.Class{proto.Class_ClassWarrior, proto.Class_ClassPaladin, proto.Class_ClassHunter,
		proto.Class_ClassRogue, proto.Class_ClassPriest, proto.Class_ClassShaman, proto.Class_ClassMage,
		proto.Class_ClassWarlock, proto.Class_ClassDruid}
	restricted := 0
	loaded := map[int32]bool{}
	for _, uiItem := range database.Load().Items {
		if loaded[uiItem.Id] {
			continue
		}
		loaded[uiItem.Id] = true
		item := core.GetItemByID(uiItem.Id)
		if item == nil {
			t.Errorf("item %d is not in the sim's database", uiItem.Id)
			continue
		}
		if !slices.Equal(item.ClassAllowlist, uiItem.ClassAllowlist) {
			t.Errorf("item %d %s loads class allowlist %v, the database states %v", item.ID, item.Name, item.ClassAllowlist, uiItem.ClassAllowlist)
		}
		if len(item.ClassAllowlist) > 0 {
			restricted++
		}
		for i, class := range item.ClassAllowlist {
			if !slices.Contains(playable, class) {
				t.Errorf("item %d %s class allowlist %v names %v, which is not a playable class", item.ID, item.Name, item.ClassAllowlist, class)
			}
			if slices.Contains(item.ClassAllowlist[:i], class) {
				t.Errorf("item %d %s class allowlist %v names %v twice", item.ID, item.Name, item.ClassAllowlist, class)
			}
		}
	}
	if restricted == 0 {
		t.Error("no item carries a class allowlist")
	}
}

// Gri'lek's Charm of Valor 19952 may only be used by a paladin: its on-use registers for a paladin
// wearing it and not for a warrior.
func TestClassRestrictedOnUseRegistersOnlyForItsClass(t *testing.T) {
	const charmID int32 = 19952

	onUseFor := func(player *proto.Player) *core.Spell {
		items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotTrinket1+1)
		for i := range items {
			items[i] = &proto.ItemSpec{}
		}
		items[proto.ItemSlot_ItemSlotTrinket1] = &proto.ItemSpec{Id: charmID}
		player.Name = "Charm Tester"
		player.Race = proto.Race_RaceHuman
		player.Equipment = &proto.EquipmentSpec{Items: items}

		sim := core.NewSim(&proto.RaidSimRequest{
			SimOptions: &proto.SimOptions{RandomSeed: 1},
			Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}},
			Encounter:  STEncounter,
		}, simsignals.CreateSignals())
		return sim.Raid.Parties[0].Players[0].GetCharacter().GetSpell(core.ActionID{ItemID: charmID})
	}

	paladin := &proto.Player{Class: proto.Class_ClassPaladin, Spec: &proto.Player_RetributionPaladin{
		RetributionPaladin: &proto.RetributionPaladin{Options: &proto.RetributionPaladin_Options{ClassOptions: &proto.PaladinOptions{}}}}}
	if onUseFor(paladin) == nil {
		t.Errorf("a paladin wearing item %d has no on-use", charmID)
	}

	warrior := &proto.Player{Class: proto.Class_ClassWarrior, Spec: &proto.Player_DpsWarrior{
		DpsWarrior: &proto.DpsWarrior{Options: &proto.DpsWarrior_Options{ClassOptions: &proto.WarriorOptions{}}}}}
	if onUseFor(warrior) != nil {
		t.Errorf("a warrior wearing item %d has its on-use", charmID)
	}
}

// Every gear set and build the specs ship holds only items its class may use; the class is the folder
// the spec sits in, ui/specs/<class>/<spec>/gear_sets or ui/specs/<class>/<spec>/builds.
func TestGearSetsHoldOnlyTheirClassItems(t *testing.T) {
	gearFiles, err := filepath.Glob("../ui/specs/*/*/gear_sets/*.gear.json")
	if err != nil {
		t.Fatal(err)
	}
	buildFiles, err := filepath.Glob("../ui/specs/*/*/builds/*.build.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(gearFiles) == 0 || len(buildFiles) == 0 {
		t.Fatalf("found %d gear sets and %d builds under ../ui/specs", len(gearFiles), len(buildFiles))
	}
	for _, file := range slices.Concat(gearFiles, buildFiles) {
		className := strings.Split(filepath.ToSlash(file), "/")[3]
		classValue, ok := proto.Class_value["Class"+strings.ToUpper(className[:1])+className[1:]]
		if !ok {
			t.Errorf("%s: %s is not a class", file, className)
			continue
		}
		class := proto.Class(classValue)

		dir, name := filepath.Dir(file), filepath.Base(file)
		var equipment *proto.EquipmentSpec
		if build, ok := strings.CutSuffix(name, ".build.json"); ok {
			equipment = core.GetTestBuildFromJSON(class, dir, build, core.ItemFilter{}, nil, nil).GearSet.GearSet
		} else {
			equipment = core.GetGearSet(dir, strings.TrimSuffix(name, ".gear.json")).GearSet
		}
		for _, spec := range equipment.GetItems() {
			if spec.Id == 0 {
				continue
			}
			item := core.GetItemByID(spec.Id)
			if item == nil {
				t.Errorf("%s: item %d is not in the sim's database", file, spec.Id)
				continue
			}
			if !item.UsableBy(class) {
				t.Errorf("%s: item %d %s may only be used by %v, not %v", file, item.ID, item.Name, item.ClassAllowlist, class)
			}
		}
	}
}
