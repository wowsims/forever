package shared

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Thrash Blade's chance on hit, Thrash 21919, grants extra attacks. The tests roll it at 100%.
const thrash int32 = 21919

func withExtraAttackWeapon(t *testing.T, itemID int32) {
	t.Helper()
	withAuraItem(itemID, proto.ItemType_ItemTypeWeapon,
		&proto.ItemEffect{BuffId: thrash, Effect: &proto.ItemEffect_Proc{Proc: &proto.ProcEffect{}}})
	everyHit(t, thrash)
	NewSpellDataExtraAttackProc(SpellDataProc{TriggerSpellID: thrash, IsWeaponProc: true},
		[]ItemVariant{{ItemID: itemID, ItemName: "Test Thrash Blade"}})
}

// Each proc swings the main hand as often as the row grants, under the granting spell's tag, and the
// swing it grants does not proc it again: two strikes grant twice the row's count.
func TestExtraAttackProcSwingsTheMainHandOncePerProc(t *testing.T) {
	const itemID int32 = 992501
	withExtraAttackWeapon(t, itemID)

	sim := newAuraSim(shaman("Wielder", map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: itemID}))
	wielder := auraTesterAt(sim, 0)
	target := sim.Encounter.ActiveTargetUnits[0]

	extraAttack := wielder.GetSpell(core.ActionID{OtherID: proto.OtherAction_OtherActionAttack, Tag: thrash})
	if extraAttack == nil {
		t.Fatalf("no extra attack is registered under the granting spell's tag")
	}

	strikeAndStep(t, sim, wielder, target)
	strikeAndStep(t, sim, wielder, target)
	want := 2 * int32(spelldata.MustFind(thrash).ExtraAttacksEffect().BasePoints)
	if want == 0 {
		t.Fatalf("21919 grants no extra attack, nothing to check")
	}
	if got := extraAttack.SpellMetrics[target.UnitIndex].Casts; got != want {
		t.Errorf("two strikes at 100%% chance granted %d extra attacks, want %d", got, want)
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
	if extraAttack := caster.GetSpell(core.ActionID{OtherID: proto.OtherAction_OtherActionAttack, Tag: thrash}); extraAttack != nil {
		t.Errorf("an extra attack is registered for a character with no main-hand swing to copy")
	}
}

// The extra attack swings the main hand, so a weapon swap moves it to the new weapon's school.
func TestExtraAttackProcFollowsTheMainHandsSchool(t *testing.T) {
	const itemID int32 = 992503
	withExtraAttackWeapon(t, itemID)

	sim := newAuraSim(shaman("Wielder", map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: itemID}))
	wielder := auraTesterAt(sim, 0)
	target := sim.Encounter.ActiveTargetUnits[0]
	extraAttack := wielder.GetSpell(core.ActionID{OtherID: proto.OtherAction_OtherActionAttack, Tag: thrash})

	shadowWeapon := *wielder.AutoAttacks.MH()
	shadowWeapon.SpellSchool = core.SpellSchoolShadow
	wielder.AutoAttacks.SetMH(shadowWeapon)
	strikeAndStep(t, sim, wielder, target)
	if extraAttack.SpellSchool != core.SpellSchoolShadow || extraAttack.BonusCoefficient != 0 {
		t.Errorf("after a swap to a shadow weapon the extra attack is %v with bonus coefficient %v, want shadow with 0", extraAttack.SpellSchool, extraAttack.BonusCoefficient)
	}
}
