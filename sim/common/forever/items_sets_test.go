//go:build with_db

package forever_test

import (
	"maps"
	"math"
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

// Battlegear of Valor's 5-piece 450587, and the Battlegear of Heroism reissue's 4-piece that carries the same
// row, roll the row's procs per minute on landed white swings of either hand, each on its own weapon's speed,
// and on Windfury Totem's strike as a main-hand swing, and never on a main-hand special such as Heroic Strike
// or Cleave. Each proc heals the wearer for 450589's heal, and restores its rage to a wearer with a rage bar.
// Four Valor pieces beside a Heroism piece carry no proc.
func TestWarriorsResolveHealsAndRestoresRageOnSwings(t *testing.T) {
	row := spelldata.MustFind(450587)
	resolve := spelldata.MustFind(450589)
	heal, rage := resolve.HealEffect(), resolve.EnergizeEffect()
	const mainHand, offHand = northernShortsword, warKnife
	valor := map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotChest: 16730, proto.ItemSlot_ItemSlotHead: 16731, proto.ItemSlot_ItemSlotLegs: 16732,
		proto.ItemSlot_ItemSlotShoulder: 16733, proto.ItemSlot_ItemSlotFeet: 16734,
	}
	heroism := map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotFeet: 226857, proto.ItemSlot_ItemSlotShoulder: 226858, proto.ItemSlot_ItemSlotLegs: 226859,
		proto.ItemSlot_ItemSlotHead: 226860,
	}
	for _, tc := range []struct {
		name   string
		pieces map[proto.ItemSlot]int32
	}{
		{"Battlegear of Valor 5P", valor},
		{"Battlegear of Heroism 4P", heroism},
	} {
		pieces := maps.Clone(tc.pieces)
		pieces[proto.ItemSlot_ItemSlotMainHand] = mainHand
		pieces[proto.ItemSlot_ItemSlotOffHand] = offHand
		sim, warrior := itemWearerWithEquipment(t, proto.Class_ClassWarrior, setPieces(pieces))
		proc := warrior.GetAura(tc.name)
		if proc == nil || proc.Dpm == nil || row.RPPM <= 0 {
			t.Fatalf("%s registers no proc on a procs per minute rate, or 450587 carries none", tc.name)
		}
		for _, hand := range []struct {
			mask  core.ProcMask
			speed float64
		}{
			{core.ProcMaskMeleeMHAuto, core.GetItemByID(mainHand).SwingSpeed},
			{core.ProcMaskMeleeOHAuto, core.GetItemByID(offHand).SwingSpeed},
		} {
			if got, want := proc.Dpm.Chance(hand.mask, sim), float64(row.RPPM)*hand.speed/60; math.Abs(got-want) > 1e-9 {
				t.Errorf("%s rolls %v on a %v swing, want the row's procs per minute on its weapon's speed, %v", tc.name, got, hand.mask, want)
			}
		}

		// Each hit now procs it wherever the manager hears it, and Windfury Totem's strike reaches the
		// manager as a main-hand swing, never as an off-hand one. The warrior stand-in has no rage bar,
		// so the proc heals alone.
		strike := &core.Spell{ActionID: core.ActionID{SpellID: 10610}, ProcMask: core.ProcMaskMeleeMHSpecial, Flags: core.SpellFlagCountsAsSwing, SpellSchool: core.SpellSchoolPhysical, Unit: &warrior.Unit}
		special := &core.Spell{ActionID: core.ActionID{SpellID: 11567}, ProcMask: core.ProcMaskMeleeMHSpecial, SpellSchool: core.SpellSchoolPhysical, Unit: &warrior.Unit}
		healed := &warrior.GetSpell(core.ActionID{SpellID: resolve.ID}).SpellMetrics[warrior.UnitIndex]
		for _, hit := range []struct {
			spell   *core.Spell
			manager core.ProcMask
			casts   int32
		}{
			{meleeSwing(&warrior.Unit), core.ProcMaskMeleeWhiteHit, 1},
			{special, core.ProcMaskMeleeWhiteHit, 1},
			{strike, core.ProcMaskMeleeOHAuto, 1},
			{strike, core.ProcMaskMeleeMHAuto, 2},
		} {
			*proc.Dpm = *warrior.NewFixedProcChanceManager(1, hit.manager)
			dealHit(t, sim, warrior, hit.spell)
			if healed.Casts != hit.casts {
				t.Errorf("%s: after a %v hit it healed %d times, want %d", tc.name, hit.spell.ActionID, healed.Casts, hit.casts)
			}
		}
		if healed.TotalHealing < 2*heal.Min() {
			t.Errorf("%s healed %v over two procs, want at least the row's %v each", tc.name, healed.TotalHealing, heal.Min())
		}
	}

	// The bear's rage bar takes the gain, which the row states in tenths.
	items := bearItemSlots()
	for slot, id := range valor {
		items[slot].Id = id
	}
	sim, bear := equippedBear(t, items)
	proc := bear.GetAura("Battlegear of Valor 5P")
	if proc == nil || proc.Dpm == nil {
		t.Fatal("the bear's Battlegear of Valor 5P registers no proc on a procs per minute rate")
	}
	*proc.Dpm = *bear.NewFixedProcChanceManager(1, core.ProcMaskMeleeWhiteHit)
	before := bear.CurrentRage()
	dealHit(t, sim, &bear.Character, meleeSwing(&bear.Unit))
	var restored float64
	for _, r := range bear.Metrics.ToProto().Resources {
		if r.Id.GetSpellId() == resolve.ID && r.Type == proto.ResourceType_ResourceTypeRage {
			restored += r.Gain
		}
	}
	if want := rage.Average() / 10; restored != want || bear.CurrentRage()-before < want {
		t.Errorf("a proc restored %v rage to the bear, want the row's %v", restored, want)
	}

	mixed := maps.Clone(valor)
	mixed[proto.ItemSlot_ItemSlotFeet] = heroism[proto.ItemSlot_ItemSlotFeet]
	_, wearer := itemWearerWithEquipment(t, proto.Class_ClassWarrior, setPieces(mixed))
	if wearer.GetAura("Battlegear of Valor 5P") != nil || wearer.GetAura("Battlegear of Heroism 4P") != nil {
		t.Error("four Valor pieces and a Heroism piece carry a proc, want none")
	}
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
