package shared

import (
	"math"
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Two weapons stating their own procs-per-minute rate on the same trigger spell each get a manager
// at their rate: Forging Strike at 3 PPM on one, 1 PPM on the other, read off the item's effect
// entry when the shape passes none.
func TestWeaponProcRateComesFromTheItemEffect(t *testing.T) {
	const trigger = forgingStrike
	for _, tc := range []struct {
		itemID int32
		ppm    float64
	}{
		{992601, 3},
		{992602, 1},
	} {
		withAuraItem(tc.itemID, proto.ItemType_ItemTypeWeapon, &proto.ItemEffect{BuffId: trigger,
			Effect: &proto.ItemEffect_Proc{Proc: &proto.ProcEffect{ProcRate: &proto.ProcEffect_Ppm{Ppm: tc.ppm}}}})

		agent := newTestAgent()
		character := agent.character
		character.Equipment = core.ProtoToEquipment(&proto.EquipmentSpec{Items: testHands(&proto.ItemSpec{Id: tc.itemID}, &proto.ItemSpec{})})
		character.EnableAutoAttacks(agent, core.AutoAttackOptions{
			MainHand:       character.WeaponFromMainHand(),
			AutoSwingMelee: true,
		})

		cfg := SpellDataProc{Name: "Test Forging Strike Weapon", ItemID: tc.itemID, TriggerSpellID: trigger, BuffSpellID: trigger, IsWeaponProc: true}
		config := spellDataProcListener(character, cfg, cfg.effectSource(), spelldata.MustFind(trigger))
		if config.ProcChance != 0 || config.DPM == nil {
			t.Fatalf("item %d: chance %v, manager %v; want a manager and no flat chance", tc.itemID, config.ProcChance, config.DPM)
		}
		want := testOneHander(tc.itemID).WeaponSpeed * tc.ppm / 60
		if got := config.DPM.Chance(core.ProcMaskMeleeMHAuto, nil); math.Abs(got-want) > 1e-9 {
			t.Errorf("item %d at %v PPM rolls %v per main-hand swing, want %v", tc.itemID, tc.ppm, got, want)
		}
	}
}

// A weapon whose trigger's row states a flat chance in its column, and whose effect entry states no
// procs-per-minute rate, rolls that chance on the hits of the hand that carries it and on no other's.
// The same item in both hands registers once, so its one listener hears both hands.
func TestFlatChanceWeaponProcRollsOnItsOwnHand(t *testing.T) {
	const flatID, plainID, trigger int32 = 992611, 992612, forgingStrike
	editRow(t, trigger, func(s *spelldata.Spell) {
		s.RPPM, s.ProcChance, s.ProcChanceSource = 0, 50, spelldata.ProcChanceColumn
	})
	stated := spelldata.MustFind(trigger).StatedChance()
	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{testOneHander(flatID), testOneHander(plainID)}})

	mainHand := []core.ProcMask{core.ProcMaskMeleeMHAuto, core.ProcMaskMeleeMHSpecial}
	offHand := []core.ProcMask{core.ProcMaskMeleeOHAuto, core.ProcMaskMeleeOHSpecial}
	for _, tc := range []struct {
		name              string
		mainHand, offHand int32
		heard, unheard    []core.ProcMask
	}{
		{"in the main hand", flatID, plainID, mainHand, offHand},
		{"in the off hand", plainID, flatID, offHand, mainHand},
		{"in both hands", flatID, flatID, append(slices.Clone(mainHand), offHand...), nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			agent := newTestAgent()
			character := agent.character
			character.Equipment = core.ProtoToEquipment(&proto.EquipmentSpec{
				Items: testHands(&proto.ItemSpec{Id: tc.mainHand}, &proto.ItemSpec{Id: tc.offHand})})
			character.EnableAutoAttacks(agent, core.AutoAttackOptions{
				MainHand:       character.WeaponFromMainHand(),
				OffHand:        character.WeaponFromOffHand(),
				AutoSwingMelee: true,
			})

			cfg := SpellDataProc{Name: "Test Flat Strike Weapon", ItemID: flatID, TriggerSpellID: trigger, BuffSpellID: trigger, IsWeaponProc: true}
			config := spellDataProcListener(character, cfg, cfg.effectSource(), spelldata.MustFind(trigger))
			if config.ProcChance != 0 || config.DPM == nil {
				t.Fatalf("chance %v, manager %v; want a manager bound to the weapon and no flat chance", config.ProcChance, config.DPM)
			}
			for _, mask := range tc.heard {
				if got := config.DPM.Chance(mask, nil); got != stated {
					t.Errorf("a hit of mask %v rolls %v, want the row's %v", mask, got, stated)
				}
			}
			for _, mask := range tc.unheard {
				if got := config.DPM.Chance(mask, nil); got != 0 {
					t.Errorf("a hit of mask %v from the other hand rolls %v, want none", mask, got)
				}
			}
		})
	}
}

// An item swap that moves a flat-chance weapon from the main hand to the off hand moves its rolls
// with it: before the swap it rolls on main-hand hits only, after it on off-hand hits only.
func TestFlatChanceWeaponProcFollowsItsItemThroughASwap(t *testing.T) {
	const flatID, plainID, otherPlainID, trigger int32 = 992613, 992614, 992615, forgingStrike
	editRow(t, trigger, func(s *spelldata.Spell) {
		s.RPPM, s.ProcChance, s.ProcChanceSource = 0, 50, spelldata.ProcChanceColumn
	})
	stated := spelldata.MustFind(trigger).StatedChance()
	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{testOneHander(flatID), testOneHander(plainID), testOneHander(otherPlainID)}})
	registerSpellDataDamageProc(SpellDataProc{Name: "Test Flat Strike Weapon", ItemID: flatID,
		TriggerSpellID: trigger, BuffSpellID: trigger, IsWeaponProc: true})

	sim := newTestCasterSimAgainst(
		testHands(&proto.ItemSpec{Id: flatID}, &proto.ItemSpec{Id: plainID}),
		testHands(&proto.ItemSpec{Id: otherPlainID}, &proto.ItemSpec{Id: flatID}), 1,
		func(player *proto.Player) { player.Spec = &proto.Player_EnhancementShaman{} })
	caster := sim.Raid.Parties[0].Players[0].(*testCaster)
	listener := caster.GetAura("Test Flat Strike Weapon")
	if listener == nil || listener.Dpm == nil {
		t.Fatal("no listener with a manager is registered for the weapon")
	}

	expect := func(when string, mainHand, offHand float64) {
		t.Helper()
		if got := listener.Dpm.Chance(core.ProcMaskMeleeMHAuto, sim); got != mainHand {
			t.Errorf("%s: a main-hand hit rolls %v, want %v", when, got, mainHand)
		}
		if got := listener.Dpm.Chance(core.ProcMaskMeleeOHAuto, sim); got != offHand {
			t.Errorf("%s: an off-hand hit rolls %v, want %v", when, got, offHand)
		}
	}
	expect("in the main hand", stated, 0)

	caster.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false)
	if caster.MainHand().ID != otherPlainID || caster.OffHand().ID != flatID {
		t.Fatalf("hands after the swap = %d / %d, want %d / %d", caster.MainHand().ID, caster.OffHand().ID, otherPlainID, flatID)
	}
	expect("swapped into the off hand", 0, stated)
}
