package core

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

func TestOnUseBuffDropsWhenItsItemIsSwappedOff(t *testing.T) {
	const onUseMaceID, plainMaceID, hasteSpellID int32 = 990911, 990912, 990913

	// A swap in combat changes only the weapon slots, so the on-use is on a weapon.
	twoHander := func(id int32) *proto.SimItem {
		return &proto.SimItem{
			Id:             id,
			Name:           "Test Swap Mace",
			Type:           proto.ItemType_ItemTypeWeapon,
			WeaponType:     proto.WeaponType_WeaponTypeMace,
			HandType:       proto.HandType_HandTypeTwoHand,
			WeaponSpeed:    3.5,
			ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {WeaponDamageMin: 100, WeaponDamageMax: 150}},
		}
	}
	addToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{twoHander(onUseMaceID), twoHander(plainMaceID)}})

	// The spell is keyed by the item and its aura by the haste spell, as a speed on-use from the game
	// files registers them.
	pseudoStats := make([]float64, stats.PseudoStatsLen)
	pseudoStats[proto.PseudoStat_PseudoStatMeleeHastePercent] = 20
	NewItemEffect(onUseMaceID, func(agent Agent) {
		character := agent.GetCharacter()
		aura := character.RegisterAura(Aura{
			Label:    "Test Swap Haste",
			ActionID: ActionID{SpellID: hasteSpellID},
			Duration: time.Second * 20,
		}).AttachHastePseudoStats(pseudoStats)
		spell := character.RegisterSpell(SpellConfig{
			ActionID:        ActionID{ItemID: onUseMaceID},
			Flags:           SpellFlagNoOnCastComplete,
			RelatedSelfBuff: aura,
			Cast:            CastConfig{CD: Cooldown{Timer: character.NewTimer(), Duration: time.Minute * 3}},
			ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
				aura.Activate(sim)
			},
		})
		character.AddMajorCooldown(MajorCooldown{Spell: spell, Type: CooldownTypeDPS})
	})

	withMainHand := func(id int32) []*proto.ItemSpec {
		items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
		for i := range items {
			items[i] = &proto.ItemSpec{}
		}
		items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: id}
		return items
	}
	sim := NewSim(&proto.RaidSimRequest{
		Raid: SinglePlayerRaidProto(&proto.Player{
			Name:           "Swap Tester",
			Class:          proto.Class_ClassWarrior,
			Race:           proto.Race_RaceHuman,
			Buffs:          &proto.IndividualBuffs{},
			Spec:           &proto.Player_DpsWarrior{},
			Equipment:      &proto.EquipmentSpec{Items: withMainHand(onUseMaceID)},
			EnableItemSwap: true,
			ItemSwap:       &proto.ItemSwap{Items: withMainHand(plainMaceID)},
			Rotation:       &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
		}, &proto.PartyBuffs{}, nil, nil),
		Encounter:  &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 63}}, Duration: 180},
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	character := sim.Raid.Parties[0].Players[0].GetCharacter()

	use := character.GetSpell(ActionID{ItemID: onUseMaceID})
	if use == nil {
		t.Fatal("the on-use mace has no use")
	}
	before := character.TotalMeleeHasteMultiplier()
	if !use.Cast(sim, character.CurrentTarget) {
		t.Fatal("the on-use did not cast")
	}
	haste := character.GetAuraByID(ActionID{SpellID: hasteSpellID})
	if !haste.IsActive() {
		t.Fatal("haste aura is not up after the use")
	}
	if got, want := character.TotalMeleeHasteMultiplier(), before*1.2; math.Abs(got-want) > 1e-9 {
		t.Fatalf("melee haste after the use = %v, want %v", got, want)
	}

	character.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false)
	if haste.IsActive() {
		t.Fatal("haste aura stayed up after its mace was swapped off")
	}
	if got := character.TotalMeleeHasteMultiplier(); math.Abs(got-before) > 1e-9 {
		t.Fatalf("melee haste after the swap = %v, want %v", got, before)
	}
}

// A set swapped in before the pull stays on until an Item Swap: Main runs. The prepull swap warns when nothing swaps
// back.
func TestPrepullSwapStaysOnIntoTheFight(t *testing.T) {
	const mainMaceID, swapMaceID int32 = 990921, 990922
	mace := func(id int32) *proto.SimItem {
		return &proto.SimItem{
			Id:             id,
			Name:           "Test Prepull Swap Mace",
			Type:           proto.ItemType_ItemTypeWeapon,
			WeaponType:     proto.WeaponType_WeaponTypeMace,
			HandType:       proto.HandType_HandTypeTwoHand,
			WeaponSpeed:    3.5,
			ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {WeaponDamageMin: 100, WeaponDamageMax: 150}},
		}
	}
	addToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{mace(mainMaceID), mace(swapMaceID)}})
	withMainHand := func(id int32) []*proto.ItemSpec {
		items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
		for i := range items {
			items[i] = &proto.ItemSpec{}
		}
		items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: id}
		return items
	}
	swapTo := func(set proto.APLActionItemSwap_SwapSet) *proto.APLAction {
		return &proto.APLAction{Action: &proto.APLAction_ItemSwap{ItemSwap: &proto.APLActionItemSwap{SwapSet: set}}}
	}
	prepullSwapTo := func(set proto.APLActionItemSwap_SwapSet, doAt string) *proto.APLPrepullAction {
		return &proto.APLPrepullAction{DoAtValue: constValue(doAt), Action: swapTo(set)}
	}
	const warning = "Nothing swaps back to the main set"

	cases := []struct {
		name        string
		rotation    *proto.APLRotation
		wantWarning bool
	}{
		{"no swap back", &proto.APLRotation{
			PrepullActions: []*proto.APLPrepullAction{prepullSwapTo(proto.APLActionItemSwap_Swap1, "-5s")},
		}, true},
		{"swap back in the priority list", &proto.APLRotation{
			PrepullActions: []*proto.APLPrepullAction{prepullSwapTo(proto.APLActionItemSwap_Swap1, "-5s")},
			PriorityList:   []*proto.APLListItem{{Action: swapTo(proto.APLActionItemSwap_Main)}},
		}, false},
		{"swap back in a group", &proto.APLRotation{
			PrepullActions: []*proto.APLPrepullAction{prepullSwapTo(proto.APLActionItemSwap_Swap1, "-5s")},
			PriorityList: []*proto.APLListItem{{Action: &proto.APLAction{Action: &proto.APLAction_GroupReference{
				GroupReference: &proto.APLActionGroupReference{GroupName: "swap back"},
			}}}},
			Groups: []*proto.APLGroup{{Name: "swap back", Actions: []*proto.APLListItem{{Action: swapTo(proto.APLActionItemSwap_Main)}}}},
		}, false},
		{"swap back only in a hidden group", &proto.APLRotation{
			PrepullActions: []*proto.APLPrepullAction{prepullSwapTo(proto.APLActionItemSwap_Swap1, "-5s")},
			PriorityList: []*proto.APLListItem{{Hide: true, Action: &proto.APLAction{Action: &proto.APLAction_GroupReference{
				GroupReference: &proto.APLActionGroupReference{GroupName: "swap back"},
			}}}},
			Groups: []*proto.APLGroup{{Name: "swap back", Actions: []*proto.APLListItem{{Action: swapTo(proto.APLActionItemSwap_Main)}}}},
		}, true},
		{"swap back in the prepull", &proto.APLRotation{
			PrepullActions: []*proto.APLPrepullAction{
				prepullSwapTo(proto.APLActionItemSwap_Swap1, "-5s"),
				prepullSwapTo(proto.APLActionItemSwap_Main, "-1s"),
			},
		}, false},
	}
	for _, tc := range cases {
		tc.rotation.Type = proto.APLRotation_TypeAPL
		sim := NewSim(&proto.RaidSimRequest{
			Raid: SinglePlayerRaidProto(&proto.Player{
				Name:           "Prepull Swap Tester",
				Class:          proto.Class_ClassWarrior,
				Race:           proto.Race_RaceHuman,
				Buffs:          &proto.IndividualBuffs{},
				Spec:           &proto.Player_DpsWarrior{},
				Equipment:      &proto.EquipmentSpec{Items: withMainHand(mainMaceID)},
				EnableItemSwap: true,
				ItemSwap:       &proto.ItemSwap{Items: withMainHand(swapMaceID)},
				Rotation:       tc.rotation,
			}, &proto.PartyBuffs{}, nil, nil),
			Encounter:  &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 63}}, Duration: 180},
			SimOptions: &proto.SimOptions{RandomSeed: 1},
		}, simsignals.CreateSignals())
		character := sim.Raid.Parties[0].Players[0].GetCharacter()

		var want []string
		if tc.wantWarning {
			want = []string{warning}
		}
		wantWarnings(t, tc.name, character.Rotation.prepullValidations[0], want...)

		sim.Reset()
		sim.PrePull()
		stepSwingSim(t, sim, 0)
		wantSwapped := len(tc.rotation.PrepullActions) == 1
		if got := character.ItemSwap.IsSwapped(); got != wantSwapped {
			t.Errorf("%s: swapped at the pull = %v, want %v", tc.name, got, wantSwapped)
		}
		if got, want := character.MainHand().ID, Ternary(wantSwapped, swapMaceID, mainMaceID); got != want {
			t.Errorf("%s: main hand at the pull = %d, want %d", tc.name, got, want)
		}
	}
}
