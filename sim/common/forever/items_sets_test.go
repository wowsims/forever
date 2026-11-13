//go:build with_db

package forever_test

import (
	"testing"

	"github.com/wowsims/forever/sim/common/forever"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func init() {
	forever.RegisterAllSetBonuses()
}

func setPieces(pieces map[proto.ItemSlot]int32) *proto.EquipmentSpec {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	for slot, id := range pieces {
		items[slot].Id = id
	}
	return &proto.EquipmentSpec{Items: items}
}

// Lands hits until fired reports the proc. The seed decides which hit the row's chance lands on.
func hitUntil(t *testing.T, hit func(), fired func() bool) {
	t.Helper()
	for range 2000 {
		if fired() {
			return
		}
		hit()
	}
	t.Fatal("2000 hits never fired the proc")
}

func spellGain(character *core.Character, spellID int32) *proto.ResourceMetrics {
	for _, r := range character.Metrics.ToProto().Resources {
		if r.Id.GetSpellId() == spellID {
			return r
		}
	}
	return &proto.ResourceMetrics{}
}

// Stormshroud Armor's 2-piece 18979 and 3-piece 23863 are proc triggers the generated set builds: a melee
// hit casts 18980's Nature damage at the target at the row's chance, and restores 23864's energy at its own,
// which the bear's energy bar holds for Cat Form.
func TestStormshroudArmorProcsItsLightningAndItsEnergy(t *testing.T) {
	items := bearItemSlots()
	items[proto.ItemSlot_ItemSlotChest].Id = 15056
	items[proto.ItemSlot_ItemSlotLegs].Id = 15057
	items[proto.ItemSlot_ItemSlotShoulder].Id = 15058
	sim, bear := equippedBear(t, items)
	wearer := &bear.Character
	target := wearer.CurrentTarget
	lightning := wearer.GetSpell(core.ActionID{SpellID: 18980})
	if lightning == nil || lightning.SpellSchool != core.SpellSchoolNature {
		t.Fatal("the 2-piece registers no 18980 of Nature damage")
	}
	swing := func() { dealHit(t, sim, wearer, meleeSwing(&wearer.Unit)) }

	hitUntil(t, swing, func() bool { return lightning.SpellMetrics[target.UnitIndex].Casts > 0 })

	energy := spelldata.MustFind(23864).EnergizeEffect().Average()
	hitUntil(t, swing, func() bool { return spellGain(wearer, 23863).Events > 0 })
	if got := spellGain(wearer, 23863); got.Gain != float64(got.Events)*energy {
		t.Errorf("the 3-piece restored %v energy over %d procs, want the row's %v each", got.Gain, got.Events, energy)
	}
}

// Volcanic Armor's 3-piece 9233 casts 9057's Fire damage at the target off a melee hit.
func TestVolcanicArmorProcsItsFirebolt(t *testing.T) {
	sim, wearer := itemWearerWithEquipment(t, proto.Class_ClassRogue, setPieces(map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotChest: 15053, proto.ItemSlot_ItemSlotLegs: 15054, proto.ItemSlot_ItemSlotShoulder: 15055,
	}))
	target := wearer.CurrentTarget
	firebolt := wearer.GetSpell(core.ActionID{SpellID: 9057})
	if firebolt == nil || firebolt.SpellSchool != core.SpellSchoolFire {
		t.Fatal("the 3-piece registers no 9057 of Fire damage")
	}

	hitUntil(t, func() { dealHit(t, sim, wearer, meleeSwing(&wearer.Unit)) },
		func() bool { return firebolt.SpellMetrics[target.UnitIndex].Casts > 0 })
}

// Deathbone Guardian's 3-piece 1299743 puts 1299746 up on the wearer off a melee hit taken: spell damage
// and healing for the row's duration.
func TestDeathboneGuardianSurgesWhenStruck(t *testing.T) {
	sim, warrior := itemWearerWithEquipment(t, proto.Class_ClassWarrior, setPieces(map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotWaist: 14620, proto.ItemSlot_ItemSlotFeet: 14621, proto.ItemSlot_ItemSlotHands: 14622,
	}))
	enemy := sim.Encounter.ActiveTargetUnits[0]
	row := spelldata.MustFind(1299746)
	before := warrior.GetStat(stats.SpellDamage)

	hitUntil(t, func() { takeHit(t, sim, warrior, meleeSwing(enemy)) }, func() bool {
		surge := warrior.GetAuraByID(core.ActionID{SpellID: 1299746})
		return surge != nil && surge.IsActive()
	})

	if got, want := warrior.GetStat(stats.SpellDamage)-before, row.EffectN(1).Average(); got != want {
		t.Errorf("the surge adds %v spell damage, want the row's %v", got, want)
	}
	if got, want := warrior.GetAuraByID(core.ActionID{SpellID: 1299746}).Duration, row.Duration(); got != want {
		t.Errorf("the surge lasts %v, want the row's %v", got, want)
	}
}
