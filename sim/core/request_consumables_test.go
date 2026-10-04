//go:build with_db

package core_test

import (
	"maps"
	"slices"
	"sync"
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var registerAllEffects = sync.OnceFunc(common.RegisterAllEffects)

// The consumables a browser request carries for the spec, the way the UI's
// extendPlayerProtoWithMissingEffects collects them: every id field and every id list.
func requestConsumableIDs(consumables *proto.ConsumesSpec) []int32 {
	var ids []int32
	consumables.ProtoReflect().Range(func(field protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		switch {
		case field.IsList():
			for i := range value.List().Len() {
				ids = append(ids, int32(value.List().Get(i).Int()))
			}
		case field.Kind() == protoreflect.Int32Kind:
			ids = append(ids, int32(value.Int()))
		}
		return true
	})
	slices.Sort(ids)
	return slices.Compact(ids)
}

var requestPreset = &proto.ConsumesSpec{
	PotId:       13442, // Mighty Rage Potion
	Potions:     []int32{13446},
	FlaskId:     13510,                // Flask of the Titans
	ElixirIds:   []int32{9187, 13445}, // Elixir of Greater Agility, Elixir of Superior Defense
	FoodId:      20452,                // Smoked Desert Dumplings
	ExplosiveId: 18588,                // Ez-Thro Dynamite II
	MhImbueId:   12404,                // Dense Sharpening Stone
	OhImbueId:   12404,
	// Goblin Sapper Charge, Scrolls of Strength, Intellect, Spirit and Stamina IV, Bogling Root
	ExtraIds: []int32{10646, 10310, 10308, 10306, 10307, 5206},
}

type appliedConsumables struct {
	stats    []float64
	auras    map[string]core.ActionID
	spells   []core.ActionID
	enchants [2]int32
}

func applyRequestPreset(t *testing.T) appliedConsumables {
	t.Helper()
	gear := core.GetGearSet("../../ui/specs/warrior/dps/gear_sets", "p1_fury").GearSet
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Requester", Class: proto.Class_ClassWarrior, Race: proto.Race_RaceOrc,
				Spec: &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{Options: &proto.DpsWarrior_Options{
					ClassOptions: &proto.WarriorOptions{DefaultStance: proto.WarriorStance_WarriorStanceBerserker}}}},
				Equipment:   gear,
				Consumables: requestPreset,
				Profession1: proto.Profession_Engineering,
				Buffs:       &proto.IndividualBuffs{},
				Rotation:    &proto.APLRotation{},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 63}}, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()

	character := sim.Raid.Parties[0].Players[0].GetCharacter()
	applied := appliedConsumables{
		stats:    character.GetStats().ToProtoArray(),
		auras:    map[string]core.ActionID{},
		enchants: [2]int32{character.Equipment[proto.ItemSlot_ItemSlotMainHand].TempEnchant, character.Equipment[proto.ItemSlot_ItemSlotOffHand].TempEnchant},
	}
	for _, aura := range character.GetAuras() {
		applied.auras[aura.Label] = aura.ActionID
	}
	for _, spell := range character.Spellbook {
		applied.spells = append(applied.spells, spell.ActionID)
	}
	return applied
}

// A browser runs the sim on the consumables its request carries, not the whole database. Every
// consumable the preset selects applies there as it does with every consumable
// loaded.
func TestRequestCarriesEveryConsumableThePresetSelects(t *testing.T) {
	registerAllEffects()
	whole := applyRequestPreset(t)

	restore := core.KeepOnlyConsumables(requestConsumableIDs(requestPreset))
	requested := applyRequestPreset(t)
	restore()

	for _, id := range []int32{13442, 13446, 13510, 9187, 13445, 20452, 18588, 10646, 10310, 10308, 10306, 10307, 5206} {
		found := slices.ContainsFunc(whole.spells, func(actionID core.ActionID) bool { return actionID.ItemID == id }) ||
			slices.ContainsFunc(slices.Collect(maps.Values(whole.auras)), func(actionID core.ActionID) bool { return actionID.ItemID == id })
		if !found {
			t.Errorf("%d applies nothing even with every consumable loaded", id)
		}
	}
	if whole.enchants[0] == 0 || whole.enchants[1] == 0 {
		t.Errorf("12404 enchants neither weapon with every consumable loaded: %v", whole.enchants)
	}

	if !slices.Equal(requested.stats, whole.stats) {
		t.Errorf("the request's stats differ from the whole database's:\n%v\n%v", requested.stats, whole.stats)
	}
	if !maps.Equal(requested.auras, whole.auras) {
		for label := range whole.auras {
			if _, ok := requested.auras[label]; !ok {
				t.Errorf("the request is missing the aura %q", label)
			}
		}
	}
	for _, actionID := range whole.spells {
		if !slices.Contains(requested.spells, actionID) {
			t.Errorf("the request is missing the spell %v", actionID)
		}
	}
	if requested.enchants != whole.enchants {
		t.Errorf("the request enchants the weapons with %v, the whole database with %v", requested.enchants, whole.enchants)
	}
}
