//go:build with_db

package forever_test

import (
	"maps"
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/common/forever"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
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

// One cast of spell that took castTime, and the spell batch window the listener resolves after.
func completeCast(t *testing.T, sim *core.Simulation, character *core.Character, spell *core.Spell, castTime time.Duration) {
	t.Helper()
	spell.CurCast.CastTime = castTime
	character.OnCastComplete(sim, spell)
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
}

// Fails unless procs is within four standard deviations of what casts casts of castTime give at the row's
// procs per minute, a cast priced on its time and at least castTimeFloor.
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
// and at least castTimeFloor, and a spell's hit or a heal's landing rolls nothing. A buff cast that costs
// its caster a resource rolls as a spell does, off the global cooldown too; a cast that costs nothing and a
// melee ability roll nothing.
func TestSwingAndCastSetProcsPriceASwingOnItsWeaponAndACastOnItsTime(t *testing.T) {
	for _, tc := range []struct {
		name   string
		row    int32
		pieces map[proto.ItemSlot]int32
	}{
		{"Lightforge Armor 5P", 450625, map[proto.ItemSlot]int32{
			proto.ItemSlot_ItemSlotHead: 16727, proto.ItemSlot_ItemSlotShoulder: 16729, proto.ItemSlot_ItemSlotChest: 16726,
			proto.ItemSlot_ItemSlotWrist: 16722, proto.ItemSlot_ItemSlotHands: 16724, proto.ItemSlot_ItemSlotMainHand: northernShortsword,
		}},
		{"Soulforge Armor 4P", 450625, map[proto.ItemSlot]int32{
			proto.ItemSlot_ItemSlotShoulder: 226969, proto.ItemSlot_ItemSlotWrist: 226970, proto.ItemSlot_ItemSlotWaist: 226971,
			proto.ItemSlot_ItemSlotLegs: 226972, proto.ItemSlot_ItemSlotMainHand: northernShortsword,
		}},
		{"The Elements 5P", 450626, map[proto.ItemSlot]int32{
			proto.ItemSlot_ItemSlotHead: 16667, proto.ItemSlot_ItemSlotShoulder: 16669, proto.ItemSlot_ItemSlotChest: 16666,
			proto.ItemSlot_ItemSlotWrist: 16671, proto.ItemSlot_ItemSlotHands: 16672, proto.ItemSlot_ItemSlotMainHand: northernShortsword,
		}},
		{"The Five Thunders 4P", 450626, map[proto.ItemSlot]int32{
			proto.ItemSlot_ItemSlotShoulder: 227003, proto.ItemSlot_ItemSlotWrist: 227001, proto.ItemSlot_ItemSlotWaist: 227008,
			proto.ItemSlot_ItemSlotLegs: 227005, proto.ItemSlot_ItemSlotMainHand: northernShortsword,
		}},
	} {
		row := spelldata.MustFind(tc.row)
		// The warrior stand-in swings its main hand, which a healer's stand-in does not.
		sim, wearer := itemWearerWithEquipment(t, proto.Class_ClassWarrior, setPieces(tc.pieces))
		proc := wearer.GetAura(tc.name)
		buff := wearer.GetAuraByID(core.ActionID{SpellID: row.FiredSpell().ID})
		if proc == nil || proc.Dpm == nil || buff == nil || row.RPPM <= 0 {
			t.Fatalf("%s registers no proc on a procs per minute rate or no buff, or %d carries no rate", tc.name, tc.row)
		}
		if got, want := proc.Dpm.Chance(core.ProcMaskMeleeMHAuto, sim), float64(row.RPPM)*core.GetItemByID(northernShortsword).SwingSpeed/60; math.Abs(got-want) > 1e-9 {
			t.Errorf("%s rolls %v on a main-hand swing, want the row's procs per minute on the weapon's speed, %v", tc.name, got, want)
		}

		// With the swing listener's manager certain on every hit it could hear, a spell's hit and a
		// heal's landing still roll nothing and a white swing puts the buff up.
		*proc.Dpm = *wearer.NewFixedProcChanceManager(1, core.ProcMaskMeleeWhiteHit|core.ProcMaskSpellDamage|core.ProcMaskSpellHealing)
		dealHit(t, sim, wearer, fireSpell(&wearer.Unit))
		dealHeal(t, sim, wearer, healSpell(&wearer.Unit))
		if buff.IsActive() {
			t.Errorf("%s: a spell's hit or a heal's landing put %d up, want only its cast to roll", tc.name, buff.ActionID.SpellID)
		}
		dealHit(t, sim, wearer, meleeSwing(&wearer.Unit))
		if !buff.IsActive() {
			t.Errorf("%s: a landed white swing on a certain roll put no %d up", tc.name, buff.ActionID.SpellID)
		}
		buff.Deactivate(sim)

		// Seven rounds of casts, one spell batch window each, fit the stand-in's fight.
		const casts = 2000
		castProcs := func(spell *core.Spell, castTime time.Duration) float64 {
			procs := 0.0
			for range casts {
				completeCast(t, sim, wearer, spell, castTime)
				if buff.IsActive() {
					procs++
					buff.Deactivate(sim)
				}
			}
			return procs
		}
		onGCD, costly := core.Cast{GCD: core.GCDDefault}, &core.SpellCost{BaseCost: 1}
		for _, cast := range []struct {
			spell    *core.Spell
			castTime time.Duration
		}{
			{fireSpell(&wearer.Unit), 0},
			{fireSpell(&wearer.Unit), 2 * castTimeFloor},
			{healSpell(&wearer.Unit), 0},
			{&core.Spell{ProcMask: core.ProcMaskEmpty, Cost: costly, DefaultCast: onGCD, Unit: &wearer.Unit}, 0},
			{&core.Spell{ProcMask: core.ProcMaskEmpty, Cost: costly, Unit: &wearer.Unit}, 0},
		} {
			wantCastTimeRate(t, tc.name, tc.row, cast.castTime, casts, castProcs(cast.spell, cast.castTime))
		}
		for _, cast := range []struct {
			name  string
			spell *core.Spell
		}{
			{"a cast that costs nothing", &core.Spell{ProcMask: core.ProcMaskEmpty, DefaultCast: onGCD, Unit: &wearer.Unit}},
			{"a melee ability", &core.Spell{ProcMask: core.ProcMaskMeleeMHSpecial, SpellSchool: core.SpellSchoolPhysical, Cost: costly, DefaultCast: onGCD, Unit: &wearer.Unit}},
		} {
			if procs := castProcs(cast.spell, 0); procs != 0 {
				t.Errorf("%s: %s procced it %v times in %d casts, want none", tc.name, cast.name, procs, casts)
			}
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
		spell := fireSpell(&druid.Unit)
		for range casts {
			completeCast(t, sim, druid, spell, castTime)
		}
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

// The pieces in the order the tests put them on.
var kroldokPieceList = []struct {
	slot proto.ItemSlot
	id   int32
}{
	{proto.ItemSlot_ItemSlotHead, 275006}, {proto.ItemSlot_ItemSlotShoulder, 287014},
	{proto.ItemSlot_ItemSlotChest, 275005}, {proto.ItemSlot_ItemSlotHands, 287012},
	{proto.ItemSlot_ItemSlotLegs, 287013},
}

// The first n Krol'dok Battlegear pieces.
func kroldokPieces(n int) map[proto.ItemSlot]int32 {
	pieces := map[proto.ItemSlot]int32{}
	for _, p := range kroldokPieceList[:n] {
		pieces[p.slot] = p.id
	}
	return pieces
}

// Krol'dok Battlegear's 4-piece 1324298 casts 1324296's heal on the wearer on a blocked melee or ranged
// hit, one the block leaves no damage too, and on no other hit; it then rests the row's cooldown. The heal
// rolls the row's spread, and in Strongholds and Cities is the flat amount of its second effect.
func TestKroldokBattlegearHealsOnBlocks(t *testing.T) {
	proc, healRow := spelldata.MustFind(1324298), spelldata.MustFind(1324296)
	heal, stronghold := healRow.HealEffect(), healRow.EffectN(2).BasePoints
	spread := heal.Average() * heal.DirectVariance() / 2
	if proc.ICD() <= 0 || spread <= 0 || stronghold <= heal.Average()+spread {
		t.Fatalf("1324298 rests %v and 1324296 heals %v to %v, %v in Strongholds and Cities; want a cooldown, a spread and a larger flat amount",
			proc.ICD(), heal.Average()-spread, heal.Average()+spread, stronghold)
	}

	for _, tc := range []struct {
		name     string
		areas    []proto.AreaType
		min, max float64
	}{
		{"outside Strongholds and Cities", nil, heal.Average() - spread, heal.Average() + spread},
		{"in Strongholds and Cities", []proto.AreaType{proto.AreaType_AreaTypeStrongholdsCities}, stronghold, stronghold},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sim, warrior := itemWearerIn(t, proto.Class_ClassWarrior, setPieces(kroldokPieces(4)), tc.areas...)
			warrior.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: -warrior.GetStat(stats.SpellCritPercent)})
			spell := warrior.GetSpell(core.ActionID{SpellID: healRow.ID})
			if spell == nil {
				t.Fatal("four pieces register no 1324296 heal")
			}
			metrics := &spell.SpellMetrics[warrior.UnitIndex]
			enemy := sim.Encounter.ActiveTargetUnits[0]

			takeHit(t, sim, warrior, meleeSwing(enemy))
			dodgeSwing(t, sim, warrior, enemy)
			if metrics.Casts != 0 {
				t.Fatalf("a hit and a dodge not blocked healed %d times, want none", metrics.Casts)
			}

			takeOutcome(t, sim, warrior, meleeSwing(enemy), core.OutcomeHit|core.OutcomeBlock, 0)
			if metrics.Casts != 1 || metrics.TotalHealing < tc.min || metrics.TotalHealing > tc.max {
				t.Errorf("a block that left no damage healed %d times for %v, want once for %v to %v", metrics.Casts, metrics.TotalHealing, tc.min, tc.max)
			}
			takeOutcome(t, sim, warrior, meleeSwing(enemy), core.OutcomeHit|core.OutcomeBlock, 50)
			if metrics.Casts != 1 {
				t.Errorf("a second block inside the cooldown healed %d times in all, want 1", metrics.Casts)
			}
			stepTo(t, sim, sim.CurrentTime+proc.ICD())
			takeOutcome(t, sim, warrior, rangedShot(enemy), core.OutcomeHit|core.OutcomeBlock, 50)
			if metrics.Casts != 2 {
				t.Errorf("a blocked ranged hit after the cooldown healed %d times in all, want 2", metrics.Casts)
			}
		})
	}

	_, three := itemWearerWithEquipment(t, proto.Class_ClassWarrior, setPieces(kroldokPieces(3)))
	if three.GetSpell(core.ActionID{SpellID: healRow.ID}) != nil {
		t.Error("three pieces register the 1324296 heal, want four needed")
	}
}

// Krol'dok Battlegear's 5-piece 1324309 adds a stack of 1324308's Stamina on a landed melee or ranged hit
// taken, one that deals no damage too, and refreshes it; a dodge and a spell hit add none. It then rests
// the row's cooldown, and the stacks stop at the row's most.
func TestKroldokBattlegearStacksStaminaOnHitsTaken(t *testing.T) {
	proc, buffRow := spelldata.MustFind(1324309), spelldata.MustFind(1324308)
	stamina := buffRow.FirstAura(dbcenums.A_MOD_STAT).BasePoints
	if proc.ICD() <= 0 || buffRow.MaxStack < 2 || stamina <= 0 {
		t.Fatalf("1324309 rests %v and 1324308 stacks %d times for %v Stamina, want a cooldown, stacks and Stamina", proc.ICD(), buffRow.MaxStack, stamina)
	}

	_, four := itemWearerWithEquipment(t, proto.Class_ClassWarrior, setPieces(kroldokPieces(4)))
	if four.GetAuraByID(core.ActionID{SpellID: buffRow.ID}) != nil {
		t.Error("four pieces register the 1324308 buff, want five needed")
	}

	sim, warrior := itemWearerWithEquipment(t, proto.Class_ClassWarrior, setPieces(kroldokPieces(5)))
	buff := warrior.GetAuraByID(core.ActionID{SpellID: buffRow.ID})
	if buff == nil {
		t.Fatal("five pieces register no 1324308 buff")
	}
	enemy := sim.Encounter.ActiveTargetUnits[0]
	before := warrior.GetStat(stats.Stamina)

	dodgeSwing(t, sim, warrior, enemy)
	takeHit(t, sim, warrior, fireSpell(enemy))
	if buff.IsActive() {
		t.Fatal("a dodge and a spell hit taken put the buff up, want neither")
	}

	takeOutcome(t, sim, warrior, meleeSwing(enemy), core.OutcomeHit, 0)
	if buff.GetStacks() != 1 || buff.Duration != buffRow.Duration() {
		t.Fatalf("a hit taken that dealt no damage left %d stacks lasting %v, want 1 for the row's %v", buff.GetStacks(), buff.Duration, buffRow.Duration())
	}
	if got := warrior.GetStat(stats.Stamina) - before; got != stamina {
		t.Errorf("one stack adds %v Stamina, want the row's %v", got, stamina)
	}
	takeHit(t, sim, warrior, meleeSwing(enemy))
	if buff.GetStacks() != 1 {
		t.Errorf("a second hit inside the cooldown left %d stacks, want 1", buff.GetStacks())
	}

	for range buffRow.MaxStack + 2 {
		stepTo(t, sim, sim.CurrentTime+proc.ICD())
		takeHit(t, sim, warrior, rangedShot(enemy))
		if got, want := buff.ExpiresAt(), sim.CurrentTime+buff.Duration; got != want {
			t.Fatalf("the buff after a hit taken by %v expires at %v, want the row's duration later, %v", sim.CurrentTime, got, want)
		}
	}
	if got := buff.GetStacks(); got != int32(buffRow.MaxStack) {
		t.Errorf("hits past the most stacks left %d, want the row's %d", got, buffRow.MaxStack)
	}
	if got, want := warrior.GetStat(stats.Stamina)-before, stamina*float64(buffRow.MaxStack); got != want {
		t.Errorf("the most stacks add %v Stamina, want %v", got, want)
	}
}

// Teachings of the Furbolgs' 2-piece 1294087 puts 1294071 up off a landed melee or ranged hit, for the
// attack power and ranged attack power its effects state. A heal puts nothing up, and one trinket carries
// no proc.
func TestTeachingsOfTheFurbolgsRaisesAttackPowerOnHits(t *testing.T) {
	row := spelldata.MustFind(1294071)
	melee, ranged := row.EffectN(1).BaseValue(), row.EffectN(2).BaseValue()
	if melee <= 0 || ranged <= 0 {
		t.Fatal("1294071 states no attack power or no ranged attack power")
	}
	sim, warrior := itemWearerWithEquipment(t, proto.Class_ClassWarrior, setPieces(map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotTrinket1: 21326, proto.ItemSlot_ItemSlotTrinket2: 268873,
	}))
	warrior.AutoAttacks.CancelAutoSwing(sim)
	buff := warrior.GetAuraByID(core.ActionID{SpellID: row.ID})
	if buff == nil {
		t.Fatal("two trinkets register no 1294071")
	}

	for range 200 {
		dealHeal(t, sim, warrior, healSpell(&warrior.Unit))
	}
	if buff.IsActive() {
		t.Error("a heal put the buff up, want only melee and ranged hits")
	}

	attackPower, rangedAttackPower := warrior.GetStat(stats.AttackPower), warrior.GetStat(stats.RangedAttackPower)
	for _, hit := range []*core.Spell{meleeSwing(&warrior.Unit), rangedShot(&warrior.Unit)} {
		hitUntil(t, func() { dealHit(t, sim, warrior, hit) }, buff.IsActive)
		if got := warrior.GetStat(stats.AttackPower) - attackPower; got != melee {
			t.Errorf("a proc on a %v hit adds %v attack power, want the row's %v", hit.ProcMask, got, melee)
		}
		if got := warrior.GetStat(stats.RangedAttackPower) - rangedAttackPower; got != ranged {
			t.Errorf("a proc on a %v hit adds %v ranged attack power, want the row's %v", hit.ProcMask, got, ranged)
		}
		buff.Deactivate(sim)
	}

	_, one := itemWearerWithEquipment(t, proto.Class_ClassWarrior, setPieces(map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotTrinket1: 21326}))
	if one.GetAuraByID(core.ActionID{SpellID: row.ID}) != nil {
		t.Error("one trinket registers 1294071, want two needed")
	}
}

// Defias Leather's 4-piece 1292028 casts the bleed 1292029 at the target off a landed melee swing, only
// while the wearer is not in front of the target.
func TestDefiasLeatherBleedsOnlyFromBehind(t *testing.T) {
	sim, rogue := itemWearerWithEquipment(t, proto.Class_ClassRogue, setPieces(map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotChest: 10399, proto.ItemSlot_ItemSlotLegs: 10400, proto.ItemSlot_ItemSlotHands: 10401,
		proto.ItemSlot_ItemSlotFeet: 10402,
	}))
	target := rogue.CurrentTarget
	bleed := rogue.GetSpell(core.ActionID{SpellID: 1292029})
	if bleed == nil || bleed.Dot(target) == nil || bleed.SpellSchool != core.SpellSchoolPhysical {
		t.Fatal("the 4-piece registers no 1292029 bleed of Physical damage")
	}
	casts := &bleed.SpellMetrics[target.UnitIndex].Casts
	swing := func() { dealHit(t, sim, rogue, meleeSwing(&rogue.Unit)) }

	rogue.PseudoStats.InFrontOfTarget = true
	for range 400 {
		swing()
	}
	if *casts != 0 {
		t.Errorf("400 swings in front of the target cast the bleed %d times, want none", *casts)
	}

	rogue.PseudoStats.InFrontOfTarget = false
	hitUntil(t, swing, bleed.Dot(target).IsActive)
}

// Chain of the Scarlet Crusade's 4-piece 1293674 puts 1293673 up off a hit taken. While it is up every
// landed melee hit, and no spell hit, casts 1293675's Holy damage at the target, which deals 1293674's
// first effect times as much to an Undead target.
func TestChainOfTheScarletCrusadeLightsMeleeHits(t *testing.T) {
	factor := spelldata.MustFind(1293674).EffectN(1).BaseValue()
	if factor <= 1 {
		t.Fatalf("1293674 states a factor of %v against Undead, want one above 1", factor)
	}
	pieces := map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotChest: 10328, proto.ItemSlot_ItemSlotWaist: 10329, proto.ItemSlot_ItemSlotLegs: 10330,
		proto.ItemSlot_ItemSlotHands: 10331,
	}
	const hits = 20
	dealt := map[proto.MobType]float64{}
	for _, mobType := range []proto.MobType{proto.MobType_MobTypeUnknown, proto.MobType_MobTypeUndead} {
		sim, warrior := itemWearerFacing(t, proto.Class_ClassWarrior, setPieces(pieces), mobType)
		warrior.AutoAttacks.CancelAutoSwing(sim)
		enemy := sim.Encounter.ActiveTargetUnits[0]
		buff := warrior.GetAuraByID(core.ActionID{SpellID: 1293673})
		light := warrior.GetSpell(core.ActionID{SpellID: 1293675})
		if buff == nil || light == nil || light.SpellSchool != core.SpellSchoolHoly {
			t.Fatal("the 4-piece registers no 1293673 buff or no 1293675 of Holy damage")
		}
		metrics := &light.SpellMetrics[enemy.UnitIndex]

		dealHit(t, sim, warrior, meleeSwing(&warrior.Unit))
		if metrics.Casts != 0 {
			t.Fatalf("a melee hit without the buff cast 1293675 %d times, want none", metrics.Casts)
		}
		hitUntil(t, func() { takeHit(t, sim, warrior, meleeSwing(enemy)) }, buff.IsActive)
		dealHit(t, sim, warrior, fireSpell(&warrior.Unit))
		if metrics.Casts != 0 {
			t.Errorf("a spell hit under the buff cast 1293675 %d times, want none", metrics.Casts)
		}
		for range hits {
			dealHit(t, sim, warrior, meleeSwing(&warrior.Unit))
		}
		if metrics.Casts != hits {
			t.Errorf("%d melee hits under the buff cast 1293675 %d times, want each once", hits, metrics.Casts)
		}
		dealt[mobType] = metrics.TotalDamage
	}

	// Both sims draw the same rolls, so only the factor tells the two totals apart.
	plain, undead := dealt[proto.MobType_MobTypeUnknown], dealt[proto.MobType_MobTypeUndead]
	if plain <= 0 || math.Abs(undead/plain-factor) > 1e-6 {
		t.Errorf("1293675 dealt %v to an Undead target and %v to another, want the row's %v times", undead, plain, factor)
	}
}

// Deathbone Guardian's 5-piece 1213289 lowers the chance for the wearer's attacks to be dodged and to be
// parried by the percent each of its effects states.
func TestDeathboneGuardianLowersDodgeAndParry(t *testing.T) {
	row := spelldata.MustFind(1213289)
	pieceList := []struct {
		slot proto.ItemSlot
		id   int32
	}{
		{proto.ItemSlot_ItemSlotWaist, 14620}, {proto.ItemSlot_ItemSlotFeet, 14621},
		{proto.ItemSlot_ItemSlotHands, 14622}, {proto.ItemSlot_ItemSlotLegs, 14623},
		{proto.ItemSlot_ItemSlotChest, 14624},
	}
	expertise := func(n int) float64 {
		pieces := map[proto.ItemSlot]int32{}
		for _, p := range pieceList[:n] {
			pieces[p.slot] = p.id
		}
		_, warrior := itemWearerWithEquipment(t, proto.Class_ClassWarrior, setPieces(pieces))
		return warrior.GetStat(stats.ExpertisePercent)
	}
	got := expertise(5) - expertise(4)
	for _, i := range []int{1, 2} {
		want := -row.EffectN(i).BaseValue()
		if want <= 0 {
			t.Fatalf("effect %d of 1213289 lowers nothing", i)
		}
		if math.Abs(got-want) > 1e-9 {
			t.Errorf("the fifth piece lowers the chance to be dodged and parried by %v%%, want effect %d's %v%%", got, i, want)
		}
	}
}

// Rotmender's Raiment's 4-piece 1321560 casts 1321558 off a cast that leaves the wearer below the share of
// mana its first effect states, which restores the row's mana every tick, and then rests the row's
// cooldown. A cast above that share restores nothing.
func TestRotmendersRaimentRestoresManaWhenLow(t *testing.T) {
	proc, row := spelldata.MustFind(1321560), spelldata.MustFind(1321558)
	tick := row.ProcEnergizeEffect()
	threshold := proc.EffectN(1).Percent()
	ticks := int32(row.Duration() / tick.Period())
	if proc.ICD() <= row.Duration() || threshold <= 0 || ticks < 2 || tick.Average() <= 0 {
		t.Fatalf("1321560 rests %v below %v of mana and 1321558 restores %v over %d ticks; want a cooldown, a share and a gain over time",
			proc.ICD(), threshold, tick.Average(), ticks)
	}
	sim, paladin := itemWearerWithEquipment(t, proto.Class_ClassPaladin, setPieces(map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotChest: 286978, proto.ItemSlot_ItemSlotHands: 286979, proto.ItemSlot_ItemSlotWaist: 286980,
		proto.ItemSlot_ItemSlotLegs: 271207,
	}))
	spell := fireSpell(&paladin.Unit)
	runOut := func() {
		completeCast(t, sim, paladin, spell, 0)
		stepTo(t, sim, sim.CurrentTime+row.Duration()+core.SpellBatchWindow)
	}
	drainMetrics := paladin.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionManaRegen})
	drain := func() {
		paladin.SpendMana(sim, paladin.CurrentMana()-paladin.MaxMana()*threshold/2, drainMetrics)
	}

	runOut()
	if got := spellGain(paladin, row.ID); got.Events != 0 {
		t.Fatalf("a cast at full mana restored mana %d times, want none", got.Events)
	}

	drain()
	runOut()
	got := spellGain(paladin, row.ID)
	if got.Events != ticks || math.Abs(got.Gain-float64(ticks)*tick.Average()) > 1e-6 {
		t.Errorf("a cast below the share restored %v mana in %d ticks, want the row's %v in each of %d", got.Gain, got.Events, tick.Average(), ticks)
	}

	drain()
	runOut()
	if again := spellGain(paladin, row.ID); again.Events != ticks {
		t.Errorf("a second cast below the share inside the cooldown left %d ticks in all, want %d", again.Events, ticks)
	}
}

// Spider's Kiss' 2-piece 17332 puts 17333 on the target off a landed melee hit, lowering its armor by the
// row's second effect until it runs out.
func TestSpidersKissLowersTheTargetsArmor(t *testing.T) {
	row := spelldata.MustFind(17333)
	cut := row.EffectN(2).BaseValue()
	if cut >= 0 {
		t.Fatal("17333's second effect lowers no armor")
	}
	sim, warrior := itemWearerAgainst(t, proto.Class_ClassWarrior, setPieces(map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotMainHand: 13183, proto.ItemSlot_ItemSlotOffHand: 13218,
	}), &proto.Target{Name: "target", Level: 60, Stats: stats.Stats{stats.Armor: 2000}.ToProtoArray()})
	warrior.AutoAttacks.CancelAutoSwing(sim)
	target := warrior.CurrentTarget
	debuff := target.GetAuraByID(core.ActionID{SpellID: row.ID})
	if debuff == nil {
		t.Fatal("the 2-piece registers no 17333 on the target")
	}
	before := target.GetStat(stats.Armor)

	hitUntil(t, func() { dealHit(t, sim, warrior, meleeSwing(&warrior.Unit)) }, debuff.IsActive)
	if got := target.GetStat(stats.Armor) - before; got != cut {
		t.Errorf("the debuff changes the target's armor by %v, want the row's %v", got, cut)
	}
	stepTo(t, sim, debuff.ExpiresAt()+core.SpellBatchWindow)
	if got := target.GetStat(stats.Armor); debuff.IsActive() || got != before {
		t.Errorf("the target's armor after the debuff ran out is %v, want %v", got, before)
	}
}

// Partners in Crime's 2-piece 1317038 adds the Strength its row states.
func TestPartnersInCrimeAddsStrength(t *testing.T) {
	want := spelldata.MustFind(1317038).EffectN(1).BaseValue()
	if want <= 0 {
		t.Fatal("1317038 adds no Strength")
	}
	sim, warrior := itemWearerWithEquipment(t, proto.Class_ClassWarrior, setPieces(map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotFinger1: 284715, proto.ItemSlot_ItemSlotFinger2: 284716,
	}))
	bonus := warrior.GetAura("Partners in Crime 2P")
	if bonus == nil || !bonus.IsActive() {
		t.Fatal("two rings put up no 2-piece bonus")
	}
	before := warrior.GetStat(stats.Strength)
	bonus.Deactivate(sim)
	if got := before - warrior.GetStat(stats.Strength); got != want {
		t.Errorf("the 2-piece bonus adds %v Strength, want the row's %v", got, want)
	}
}
