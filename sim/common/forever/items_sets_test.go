//go:build with_db

package forever_test

import (
	"maps"
	"math"
	"testing"
	"time"

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

// How many of casts spell casts of castTime fired the proc. fired reports a proc since it was last asked.
func castProcs(t *testing.T, sim *core.Simulation, character *core.Character, castTime time.Duration, casts int, fired func() bool) float64 {
	t.Helper()
	spell := fireSpell(&character.Unit)
	procs := 0.0
	for range casts {
		spell.CurCast.CastTime = castTime
		character.OnCastComplete(sim, spell)
		stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
		if fired() {
			procs++
		}
	}
	return procs
}

// Fails unless procs of casts spell casts of castTime is about the row's procs per minute on that time, at
// least castTimeFloor.
func wantCastTimeRate(t *testing.T, name string, rowID int32, castTime time.Duration, casts int, procs float64) {
	t.Helper()
	want := castTimeChance(t, rowID, castTime)
	mean, deviation := want*float64(casts), math.Sqrt(want*(1-want)*float64(casts))
	if math.Abs(procs-mean) > 4*deviation {
		t.Errorf("%s: %d casts of %v procced it %v times, want about %.0f", name, casts, castTime, procs, mean)
	}
}

// Lightforge Armor's 5-piece and the Soulforge Armor reissue's 4-piece share 450625, The Elements' 5-piece
// and The Five Thunders reissue's 4-piece 450626. Each row hears melee swings and spell casts at its procs
// per minute: a landed white swing rolls the rate on its weapon's speed, a spell cast on the time it took
// and at least castTimeFloor, and a spell's hit rolls nothing.
func TestSwingAndCastSetProcsPriceASwingOnItsWeaponAndACastOnItsTime(t *testing.T) {
	const mainHand = northernShortsword
	for _, tc := range []struct {
		name   string
		row    int32
		pieces map[proto.ItemSlot]int32
	}{
		{"Lightforge Armor 5P", 450625, map[proto.ItemSlot]int32{
			proto.ItemSlot_ItemSlotHead: 16727, proto.ItemSlot_ItemSlotShoulder: 16729, proto.ItemSlot_ItemSlotChest: 16726,
			proto.ItemSlot_ItemSlotWrist: 16722, proto.ItemSlot_ItemSlotHands: 16724,
		}},
		{"Soulforge Armor 4P", 450625, map[proto.ItemSlot]int32{
			proto.ItemSlot_ItemSlotShoulder: 226969, proto.ItemSlot_ItemSlotWrist: 226970, proto.ItemSlot_ItemSlotWaist: 226971,
			proto.ItemSlot_ItemSlotLegs: 226972,
		}},
		{"The Elements 5P", 450626, map[proto.ItemSlot]int32{
			proto.ItemSlot_ItemSlotHead: 16667, proto.ItemSlot_ItemSlotShoulder: 16669, proto.ItemSlot_ItemSlotChest: 16666,
			proto.ItemSlot_ItemSlotWrist: 16671, proto.ItemSlot_ItemSlotHands: 16672,
		}},
		{"The Five Thunders 4P", 450626, map[proto.ItemSlot]int32{
			proto.ItemSlot_ItemSlotShoulder: 227003, proto.ItemSlot_ItemSlotWrist: 227001, proto.ItemSlot_ItemSlotWaist: 227008,
			proto.ItemSlot_ItemSlotLegs: 227005,
		}},
	} {
		row := spelldata.MustFind(tc.row)
		pieces := maps.Clone(tc.pieces)
		pieces[proto.ItemSlot_ItemSlotMainHand] = mainHand
		// The warrior stand-in swings its main hand, which a healer's stand-in does not.
		sim, wearer := itemWearerWithEquipment(t, proto.Class_ClassWarrior, setPieces(pieces))
		proc := wearer.GetAura(tc.name)
		buff := wearer.GetAuraByID(core.ActionID{SpellID: row.FiredSpell().ID})
		if proc == nil || proc.Dpm == nil || buff == nil || row.RPPM <= 0 {
			t.Fatalf("%s registers no proc on a procs per minute rate or no buff, or %d carries no rate", tc.name, tc.row)
		}
		if got, want := proc.Dpm.Chance(core.ProcMaskMeleeMHAuto, sim), float64(row.RPPM)*core.GetItemByID(mainHand).SwingSpeed/60; math.Abs(got-want) > 1e-9 {
			t.Errorf("%s rolls %v on a main-hand swing, want the row's procs per minute on the weapon's speed, %v", tc.name, got, want)
		}

		// With the swing listener's manager certain on every hit it could hear, a spell's hit still
		// rolls nothing and a white swing puts the buff up.
		*proc.Dpm = *wearer.NewFixedProcChanceManager(1, core.ProcMaskMeleeWhiteHit|core.ProcMaskSpellDamage)
		dealHit(t, sim, wearer, fireSpell(&wearer.Unit))
		if buff.IsActive() {
			t.Errorf("%s: a spell's hit put %d up, want only its cast to roll", tc.name, buff.ActionID.SpellID)
		}
		dealHit(t, sim, wearer, meleeSwing(&wearer.Unit))
		if !buff.IsActive() {
			t.Errorf("%s: a landed white swing on a certain roll put no %d up", tc.name, buff.ActionID.SpellID)
		}

		const casts = 6000
		for _, castTime := range []time.Duration{0, 2 * castTimeFloor} {
			buff.Deactivate(sim)
			procs := castProcs(t, sim, wearer, castTime, casts, func() bool {
				up := buff.IsActive()
				buff.Deactivate(sim)
				return up
			})
			wantCastTimeRate(t, tc.name, tc.row, castTime, casts, procs)
		}
	}
}

// Wildheart Raiment's 5-piece 450608 restores 27782's mana on spell casts at the row's procs per minute, a
// cast priced on the time it took and at least castTimeFloor.
func TestWildheartRaimentRestoresManaOnCastsPricedOnTheirTime(t *testing.T) {
	mana := spelldata.MustFind(27782).EnergizeEffect().Average()
	if mana <= 0 {
		t.Fatal("27782 restores no mana, nothing to check")
	}
	sim, druid := itemWearerWithEquipment(t, proto.Class_ClassDruid, setPieces(map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotHead: 16720, proto.ItemSlot_ItemSlotShoulder: 16718, proto.ItemSlot_ItemSlotChest: 16706,
		proto.ItemSlot_ItemSlotWrist: 16714, proto.ItemSlot_ItemSlotHands: 16717,
	}))

	const casts = 6000
	for _, castTime := range []time.Duration{0, 2 * castTimeFloor} {
		before := spellGain(druid, 27782)
		castProcs(t, sim, druid, castTime, casts, func() bool { return false })
		after := spellGain(druid, 27782)
		procs := float64(after.Events - before.Events)
		wantCastTimeRate(t, "Wildheart Raiment 5P", 450608, castTime, casts, procs)
		if got := after.Gain - before.Gain; got != procs*mana {
			t.Errorf("%v procs restored %v mana, want the row's %v each", procs, got, mana)
		}
	}
}

// The reissued Beastmaster Armor's 4-piece is Beaststalker Armor's 5-piece proc 450577: a landed melee or
// ranged autoattack restores 450554's mana at the row's chance.
func TestBeastmasterArmorReissueRestoresManaOnAutoattacks(t *testing.T) {
	row := spelldata.MustFind(450577)
	mana := spelldata.MustFind(450554).EnergizeEffect().Average()
	sim, hunter := itemWearerWithEquipment(t, proto.Class_ClassHunter, setPieces(map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotFeet: 226881, proto.ItemSlot_ItemSlotShoulder: 226884, proto.ItemSlot_ItemSlotWrist: 226885,
		proto.ItemSlot_ItemSlotWaist: 226888,
	}))
	proc := hunter.GetAura("Beastmaster Armor 4P")
	if proc == nil || mana <= 0 || row.StatedChance() <= 0 || row.StatedChance() >= 1 {
		t.Fatal("four reissued Beastmaster pieces register no 4-piece proc, or 450577 states no chance to roll")
	}
	for _, hit := range []*core.Spell{meleeSwing(&hunter.Unit), rangedShot(&hunter.Unit)} {
		before := spellGain(hunter, 450554)
		hitUntil(t, func() { dealHit(t, sim, hunter, hit) }, func() bool { return spellGain(hunter, 450554).Events > before.Events })
		if got := spellGain(hunter, 450554).Gain - before.Gain; got != mana {
			t.Errorf("a proc on a %v hit restored %v mana, want the row's %v", hit.ProcMask, got, mana)
		}
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
