package shared

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// Flurry Axe 871's chance on hit, 18797, grants one extra attack. Its item effect states 1.9 PPM; the
// tests roll it at 100% instead.
const flurryAxe int32 = 18797

func withExtraAttackWeapon(t *testing.T, itemID int32) {
	t.Helper()
	withAuraItem(itemID, proto.ItemType_ItemTypeWeapon,
		&proto.ItemEffect{BuffId: flurryAxe, Effect: &proto.ItemEffect_Proc{Proc: &proto.ProcEffect{}}})
	everyHit(t, flurryAxe)
	NewSpellDataExtraAttackProc(SpellDataProc{TriggerSpellID: flurryAxe, IsWeaponProc: true},
		[]ItemVariant{{ItemID: itemID, ItemName: "Test Flurry Axe"}})
}

// Each proc swings the main hand once, under the granting spell's tag, and the swing it grants does
// not proc it again: two strikes grant two.
func TestExtraAttackProcSwingsTheMainHandOncePerProc(t *testing.T) {
	const itemID int32 = 992501
	withExtraAttackWeapon(t, itemID)

	sim := newAuraSim(shaman("Wielder", map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: itemID}))
	wielder := auraTesterAt(sim, 0)
	target := sim.Encounter.ActiveTargetUnits[0]

	extraAttack := wielder.GetSpell(core.ActionID{OtherID: proto.OtherAction_OtherActionAttack, Tag: flurryAxe})
	if extraAttack == nil {
		t.Fatalf("no extra attack is registered under the granting spell's tag")
	}

	strikeAndStep(t, sim, wielder, target)
	strikeAndStep(t, sim, wielder, target)
	if got := extraAttack.SpellMetrics[target.UnitIndex].Casts; got != 2 {
		t.Errorf("two strikes at 100%% chance granted %d extra attacks, want 2", got)
	}
	if got := wielder.AutoAttacks.MHAuto().SpellMetrics[target.UnitIndex].Casts; got != 0 {
		t.Errorf("the main-hand auto attack swung %d times, want 0: the extra attack is a spell of its own", got)
	}
}

// A character that never swings can still wear the weapon: the proc registers nothing to swing with
// and the sim builds.
func TestExtraAttackProcOnACharacterWithoutAutoAttacks(t *testing.T) {
	const itemID int32 = 992502
	withExtraAttackWeapon(t, itemID)

	sim := newTestCasterSim(testHands(&proto.ItemSpec{Id: itemID}, &proto.ItemSpec{}), nil)
	caster := sim.Raid.Parties[0].Players[0].(*testCaster)
	target := sim.Encounter.ActiveTargetUnits[0]

	landMainHandSwing(sim, caster, target)
	stepPast(t, sim, sim.CurrentTime+core.SpellBatchWindow+time.Millisecond)
	if extraAttack := caster.GetSpell(core.ActionID{OtherID: proto.OtherAction_OtherActionAttack, Tag: flurryAxe}); extraAttack != nil {
		t.Errorf("an extra attack is registered for a character with no main-hand swing to copy")
	}
}
