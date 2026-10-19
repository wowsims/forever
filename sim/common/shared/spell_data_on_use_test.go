package shared

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Infernal Lasso 219345's on-use: ItemEffect 183857 casts 443265, A_PERIODIC_DAMAGE on the target. The
// test's on-use puts it on a 5 min cooldown in category 1141 for 12 s.
const (
	lassoSpell    int32 = 443265
	lassoCategory int32 = 1141
)

func lassoOnUse() *proto.ItemEffect {
	return testOnUse(lassoSpell, 300000, lassoCategory, 12000)
}

func testOnUse(spellID int32, cooldownMs int32, categoryID int32, categoryCooldownMs int32) *proto.ItemEffect {
	return &proto.ItemEffect{BuffId: spellID, Effect: &proto.ItemEffect_OnUse{OnUse: &proto.OnUseEffect{
		CooldownMs: cooldownMs, CategoryId: categoryID, CategoryCooldownMs: categoryCooldownMs}}}
}

// A caster wearing test trinkets carrying the on-use effects, registered through the constructor, at
// half health, with its spells certain to hit and never to crit.
func newOnUseSim(t *testing.T, register func(int32), trinkets map[int32]*proto.ItemEffect) (*core.Simulation, *testCaster) {
	t.Helper()
	return newOnUseSimWearing(t, register, trinkets, &proto.ItemSpec{})
}

func newOnUseSimWearing(t *testing.T, register func(int32), trinkets map[int32]*proto.ItemEffect, neck *proto.ItemSpec) (*core.Simulation, *testCaster) {
	t.Helper()
	return newOnUseSimAgainst(t, register, trinkets, neck, 1)
}

func newOnUseSimAgainst(t *testing.T, register func(int32), trinkets map[int32]*proto.ItemEffect, neck *proto.ItemSpec, targetCount int) (*core.Simulation, *testCaster) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotTrinket2+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotNeck] = neck

	slot := proto.ItemSlot_ItemSlotTrinket1
	for id, effect := range trinkets {
		core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{{Id: id, Name: "Test Trinket",
			Type: proto.ItemType_ItemTypeTrinket, ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}},
			ItemEffects: []*proto.ItemEffect{effect}}}})
		register(id)
		items[slot] = &proto.ItemSpec{Id: id}
		slot++
	}

	sim := newTestCasterSimAgainst(items, nil, targetCount)
	caster := sim.Raid.Parties[0].Players[0].(*testCaster)
	caster.AddStatsDynamic(sim, stats.Stats{
		stats.SpellHitPercent:  100,
		stats.SpellCritPercent: -caster.GetStat(stats.SpellCritPercent),
	})
	caster.RemoveHealth(sim, caster.MaxHealth()/2)
	return sim, caster
}

func onUseSpell(t *testing.T, caster *testCaster, itemID int32) *core.Spell {
	t.Helper()
	spell := caster.GetSpell(core.ActionID{ItemID: itemID})
	if spell == nil {
		t.Fatalf("item %d registered no on-use spell", itemID)
	}
	return spell
}

// What the spell deals to the target between now and the given time after start, stepping just past
// it so a tick landing on it is counted.
func dealtUntil(t *testing.T, sim *core.Simulation, spell *core.Spell, start, after time.Duration) float64 {
	t.Helper()
	target := sim.Encounter.ActiveTargetUnits[0]
	before := spell.SpellMetrics[target.UnitIndex].TotalDamage
	stepPast(t, sim, start+after+time.Millisecond)
	return spell.SpellMetrics[target.UnitIndex].TotalDamage - before
}

// The amount a row's ticking effect lands each tick before power, how often it ticks, and how many
// ticks its duration holds. A row that ticks nothing leaves the test nothing to check, and fails it.
func ticksOf(t *testing.T, row *spelldata.Spell) (float64, time.Duration, int) {
	t.Helper()
	e := row.PeriodicEffect()
	if e.Period() <= 0 || row.Duration() < e.Period() || e.Average() == 0 {
		t.Fatalf("%d ticks %v every %v for %v, nothing to check", row.ID, e.Average(), e.Period(), row.Duration())
	}
	return e.Average(), e.Period(), int(row.Duration() / e.Period())
}

// Fails the test where a row read gives it no amount to check.
func needAmount(t *testing.T, what string, amount float64) {
	t.Helper()
	if amount == 0 {
		t.Fatalf("%s reads 0, nothing to check", what)
	}
}

// The lasso lands nothing when it is used, then deals its row's amount every period until its
// duration runs out, on the target it was used on.
func TestOnUseDotTicksItsAmountOnTheTarget(t *testing.T) {
	const itemID int32 = 991001
	sim, caster := newOnUseSim(t, NewSpellDataDamageOnUse, map[int32]*proto.ItemEffect{itemID: lassoOnUse()})
	lasso := onUseSpell(t, caster, itemID)
	target := sim.Encounter.ActiveTargetUnits[0]
	per, period, ticks := ticksOf(t, spelldata.MustFind(lassoSpell))

	if !lasso.Cast(sim, target) {
		t.Fatalf("the lasso could not be used")
	}
	start := sim.CurrentTime
	if !lasso.Dot(target).IsActive() {
		t.Fatalf("using the lasso applied no damage over time to the target")
	}
	if got := lasso.SpellMetrics[target.UnitIndex].TotalDamage; got != 0 {
		t.Errorf("using the lasso dealt %v at once, want nothing until the first tick", got)
	}

	for tick := 1; tick <= ticks; tick++ {
		if got := dealtUntil(t, sim, lasso, start, time.Duration(tick)*period); got != per {
			t.Errorf("tick %d dealt %v, want %v", tick, got, per)
		}
	}
	if lasso.Dot(target).IsActive() {
		t.Errorf("the damage over time is still up after its %d ticks", ticks)
	}
	if got := dealtUntil(t, sim, lasso, start, time.Duration(ticks+1)*period); got != 0 {
		t.Errorf("the damage over time dealt %v after it ran out, want nothing", got)
	}

	metrics := lasso.SpellMetrics[target.UnitIndex]
	if want := per * float64(ticks); metrics.Casts != 1 || metrics.Ticks != int32(ticks) || metrics.TotalDamage != want {
		t.Errorf("metrics = %d casts, %d ticks, %v damage; want 1, %d, %v", metrics.Casts, metrics.Ticks, metrics.TotalDamage, ticks, want)
	}
}

// Each tick reads the spell power the wearer has when it lands. 443265 states no spell power share, so
// a copy stating 1 stands in for the rows that do.
func TestOnUseDotTicksOnTheSpellPowerOfTheTick(t *testing.T) {
	const itemID int32 = 991002
	editRow(t, lassoSpell, func(s *spelldata.Spell) { s.Effects[0].SPCoef = 1 })
	sim, caster := newOnUseSim(t, NewSpellDataDamageOnUse, map[int32]*proto.ItemEffect{itemID: lassoOnUse()})
	caster.AddStatsDynamic(sim, stats.Stats{stats.SpellDamage: -caster.GetStat(stats.SpellDamage)})
	lasso := onUseSpell(t, caster, itemID)
	per, period, _ := ticksOf(t, spelldata.MustFind(lassoSpell))

	lasso.Cast(sim, sim.Encounter.ActiveTargetUnits[0])
	start := sim.CurrentTime
	if got := dealtUntil(t, sim, lasso, start, period); got != per {
		t.Errorf("the first tick at no spell power dealt %v, want %v", got, per)
	}

	caster.AddStatsDynamic(sim, stats.Stats{stats.SpellDamage: 100})
	if got := dealtUntil(t, sim, lasso, start, 2*period); got != per+100 {
		t.Errorf("the tick after gaining 100 spell power dealt %v, want %v", got, per+100)
	}
}

// The test lasso's own 5 min cooldown, and its category's 12 s that a second item in category 1141
// waits out as well.
func TestOnUseRunsOnTheItemsCooldownAndCategory(t *testing.T) {
	const lassoID, otherID int32 = 991003, 991004
	sim, caster := newOnUseSim(t, NewSpellDataDamageOnUse, map[int32]*proto.ItemEffect{
		lassoID: lassoOnUse(),
		otherID: testOnUse(lassoSpell, 60000, lassoCategory, 12000),
	})
	lasso, other := onUseSpell(t, caster, lassoID), onUseSpell(t, caster, otherID)
	target := sim.Encounter.ActiveTargetUnits[0]

	for _, mcd := range []*core.Spell{lasso, other} {
		if got := caster.GetInitialMajorCooldown(mcd.ActionID); !got.Type.Matches(core.CooldownTypeDPS) {
			t.Errorf("%v is a major cooldown of type %v, want a DPS one", mcd.ActionID, got.Type)
		}
	}

	lasso.Cast(sim, target)
	start := sim.CurrentTime
	if offCooldown(sim, other) {
		t.Errorf("an item sharing category %d could be used while the lasso's category cooldown runs", lassoCategory)
	}

	stepPast(t, sim, start+12*time.Second)
	if !offCooldown(sim, other) {
		t.Errorf("an item sharing category %d could not be used once the 12 s ran out", lassoCategory)
	}
	if offCooldown(sim, lasso) {
		t.Errorf("the lasso could be used again after 12 s, inside its 5 min cooldown")
	}
	if got := lasso.CD.ReadyAt(); got != start+5*time.Minute {
		t.Errorf("the lasso is ready again at %v, want 5 min after its use at %v", got, start)
	}
}

func offCooldown(sim *core.Simulation, spell *core.Spell) bool {
	return core.BothTimersReady(spell.CD.Timer, spell.SharedCD.Timer, sim)
}

// Gem-studded Leather Belt 4262's on-use, 9163, heals the wearer for an amount rolled over a spread.
func TestOnUseHealsTheWearer(t *testing.T) {
	const itemID int32 = 991005
	const heal int32 = 9163
	sim, caster := newOnUseSim(t, NewSpellDataHealOnUse, map[int32]*proto.ItemEffect{itemID: testOnUse(heal, 300000, 0, 0)})
	spell := onUseSpell(t, caster, itemID)

	if got := caster.GetInitialMajorCooldown(spell.ActionID); !got.Type.Matches(core.CooldownTypeSurvival) || !spell.Flags.Matches(core.SpellFlagHelpful) {
		t.Errorf("the heal is a major cooldown of type %v, helpful %v; want a survival one cast on the wearer",
			got.Type, spell.Flags.Matches(core.SpellFlagHelpful))
	}

	effect := spelldata.MustFind(heal).ProcHealEffect()
	before := caster.CurrentHealth()
	spell.Cast(sim, &caster.Unit)
	if got, low, high := caster.CurrentHealth()-before, effect.Min(), effect.Max(); got < low || got > high {
		t.Errorf("the heal was %v, want %v to %v", got, low, high)
	}
	if offCooldown(sim, spell) {
		t.Errorf("the heal could be used again inside its 5 min cooldown")
	}
}

// Furbolg Medicine Pouch 16768's on-use, 20631, heals the wearer its row's amount every period for
// its duration.
func TestOnUseHotHealsTheWearerOverTime(t *testing.T) {
	const itemID int32 = 991006
	sim, caster := newOnUseSim(t, NewSpellDataHealOnUse, map[int32]*proto.ItemEffect{itemID: testOnUse(20631, 1200000, 0, 0)})
	spell := onUseSpell(t, caster, itemID)
	per, period, ticks := ticksOf(t, spelldata.MustFind(20631))

	spell.Cast(sim, &caster.Unit)
	start := sim.CurrentTime
	for tick := 1; tick <= ticks; tick++ {
		if got := healedUntil(t, sim, caster, start, time.Duration(tick)*period); got != per {
			t.Errorf("tick %d healed %v, want %v", tick, got, per)
		}
	}
	if spell.SelfHot().IsActive() {
		t.Errorf("the heal over time is still up after its %d ticks", ticks)
	}
	if got, want := spell.SpellMetrics[caster.UnitIndex].TotalHealing, per*float64(ticks); got != want {
		t.Errorf("healing metrics = %v, want the %d ticks' %v", got, ticks, want)
	}
}

// Helm of Fire 8348's on-use, 10578, deals its direct damage with a spread and applies a damage over
// time with it.
func TestOnUseDealsItsDirectDamageAndItsDamageOverTime(t *testing.T) {
	const itemID int32 = 991007
	const fireball int32 = 10578
	sim, caster := newOnUseSim(t, NewSpellDataDamageOnUse, map[int32]*proto.ItemEffect{itemID: testOnUse(fireball, 300000, lassoCategory, 10000)})
	spell := onUseSpell(t, caster, itemID)
	target := sim.Encounter.ActiveTargetUnits[0]

	direct := spelldata.MustFind(fireball).DamageEffect()
	spell.Cast(sim, target)
	start := sim.CurrentTime
	if got, low, high := spell.SpellMetrics[target.UnitIndex].TotalDamage, direct.Min(), direct.Max(); got < low || got > high {
		t.Errorf("the direct damage was %v, want %v to %v", got, low, high)
	}
	per, period, ticks := ticksOf(t, spelldata.MustFind(fireball))
	if got, want := dealtUntil(t, sim, spell, start, time.Duration(ticks)*period), per*float64(ticks); math.Abs(got-want) > 1e-6 {
		t.Errorf("the damage over time dealt %v over its %d ticks, want %d ticks of %v", got, ticks, ticks, per)
	}
}

// At no spell hit, against a target it misses half the time, a use of the lasso that misses applies no
// damage over time, and one that lands ticks every time: the hit is rolled once, when it is applied.
func TestOnUseDotRollsItsHitOnceWhenApplied(t *testing.T) {
	const itemID int32 = 991016
	sim, caster := newOnUseSim(t, NewSpellDataDamageOnUse, map[int32]*proto.ItemEffect{itemID: lassoOnUse()})
	caster.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: -caster.GetStat(stats.SpellHitPercent)})
	lasso := onUseSpell(t, caster, itemID)
	target := sim.Encounter.ActiveTargetUnits[0]
	caster.AttackTables[target.UnitIndex].BaseSpellMissChance = 0.5
	metrics := &lasso.SpellMetrics[target.UnitIndex]
	_, period, rowTicks := ticksOf(t, spelldata.MustFind(lassoSpell))

	missedUses, landedUses := 0, 0
	for use := 1; use <= 14; use++ {
		lasso.CD.Reset()
		lasso.SharedCD.Reset()
		misses, ticks := metrics.Misses, metrics.Ticks

		lasso.Cast(sim, target)
		missed := metrics.Misses > misses
		if missed == lasso.Dot(target).IsActive() {
			t.Errorf("use %d: missed %v, damage over time up %v", use, missed, lasso.Dot(target).IsActive())
		}

		stepPast(t, sim, sim.CurrentTime+time.Duration(rowTicks)*period+time.Millisecond)
		wantTicks := int32(rowTicks)
		if missed {
			missedUses++
			wantTicks = 0
		} else {
			landedUses++
		}
		if got := metrics.Ticks - ticks; got != wantTicks {
			t.Errorf("use %d (missed %v) ticked %d times, want %d", use, missed, got, wantTicks)
		}
		if got := metrics.Misses - misses; missed && got != 1 || !missed && got != 0 {
			t.Errorf("use %d (missed %v) counted %d misses, want only the application's", use, missed, got)
		}
	}
	if missedUses == 0 || landedUses == 0 {
		t.Fatalf("%d of 14 uses missed and %d landed; the test needs both", missedUses, landedUses)
	}
}

// Where the row states Periodic Can Crit, a landed lasso's ticks roll the spell crit and still no hit.
// 443265 states no such attribute, so a copy stating it stands in for the rows that do.
func TestOnUseDotTicksCritWithoutAHitRoll(t *testing.T) {
	const itemID int32 = 991017
	editRow(t, lassoSpell, func(s *spelldata.Spell) { s.Attr[dbcenums.ATTR_INDEX_EX_8] |= dbcenums.ATTR_EX_8_PERIODIC_CAN_CRIT })
	sim, caster := newOnUseSim(t, NewSpellDataDamageOnUse, map[int32]*proto.ItemEffect{itemID: lassoOnUse()})
	caster.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: 100})
	lasso := onUseSpell(t, caster, itemID)
	target := sim.Encounter.ActiveTargetUnits[0]
	_, period, ticks := ticksOf(t, spelldata.MustFind(lassoSpell))

	lasso.Cast(sim, target)
	caster.AttackTables[target.UnitIndex].BaseSpellMissChance = 1
	caster.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: -caster.GetStat(stats.SpellHitPercent)})
	stepPast(t, sim, sim.CurrentTime+time.Duration(ticks)*period+time.Millisecond)

	metrics := lasso.SpellMetrics[target.UnitIndex]
	if metrics.CritTicks != int32(ticks) || metrics.Ticks != 0 || metrics.Misses != 0 {
		t.Errorf("metrics = %d crit ticks, %d ticks, %d misses; want %d, 0, 0", metrics.CritTicks, metrics.Ticks, metrics.Misses, ticks)
	}
}

// A copy of Linken's Boomerang 11905's on-use, 15712, thrown from 20 yards, states a cast time
// and no global cooldown: it lands when the cast completes. A copy of the lasso's 443265 states
// neither and lands at once.
func TestOnUseCastsForTheCastTimeItsRowStates(t *testing.T) {
	const boomerangID, lassoID int32 = 991018, 991019
	editRow(t, 15712, func(s *spelldata.Spell) { s.CastTimeMs, s.GCDMs = 500, 0 })
	editRow(t, lassoSpell, func(s *spelldata.Spell) { s.CastTimeMs, s.GCDMs = 0, 0 })
	cast := spelldata.MustFind(15712).CastTime()
	sim, caster := newOnUseSim(t, NewSpellDataDamageOnUse, map[int32]*proto.ItemEffect{
		boomerangID: testOnUse(15712, 180000, 0, 0),
		lassoID:     lassoOnUse(),
	})
	target := sim.Encounter.ActiveTargetUnits[0]

	lasso := onUseSpell(t, caster, lassoID)
	start := sim.CurrentTime
	lasso.Cast(sim, target)
	if lasso.SpellMetrics[target.UnitIndex].Casts != 1 || caster.Hardcast.Expires > start || caster.NextGCDAt() > start {
		t.Errorf("the lasso cast %d times, casting until %v, GCD until %v; want it used at once at %v",
			lasso.SpellMetrics[target.UnitIndex].Casts, caster.Hardcast.Expires, caster.NextGCDAt(), start)
	}

	boomerang := onUseSpell(t, caster, boomerangID)
	caster.DistanceFromTarget = 20
	boomerang.Cast(sim, target)
	if caster.Hardcast.Expires != start+cast || boomerang.SpellMetrics[target.UnitIndex].Casts != 0 {
		t.Errorf("the boomerang casts until %v with %d casts done, want a %v cast from %v",
			caster.Hardcast.Expires, boomerang.SpellMetrics[target.UnitIndex].Casts, cast, start)
	}
	stepPast(t, sim, start+cast+time.Millisecond)
	if boomerang.SpellMetrics[target.UnitIndex].Casts != 1 {
		t.Errorf("the boomerang's cast had not completed %v after it began", cast)
	}
}

// A copy of the heal of 9163 stating a 1.5 s global cooldown spends it on use.
func TestOnUseSpendsTheGlobalCooldownItsRowStates(t *testing.T) {
	const itemID int32 = 991020
	editRow(t, 9163, func(s *spelldata.Spell) { s.GCDMs = 1500 })
	sim, caster := newOnUseSim(t, NewSpellDataHealOnUse, map[int32]*proto.ItemEffect{itemID: testOnUse(9163, 300000, 0, 0)})

	start := sim.CurrentTime
	onUseSpell(t, caster, itemID).Cast(sim, &caster.Unit)
	if got, gcd := caster.NextGCDAt(), spelldata.MustFind(9163).GCD(); got != start+gcd {
		t.Errorf("the global cooldown runs until %v, want the row's %v after the use at %v", got, gcd, start)
	}
}

// A neck whose effect puts a listener on the wearer for the callback, counting what it hears, and
// registers the spell a damage proc of procRow would cast, where procRow is not 0.
func listeningNeck(neckID int32, callback core.AuraCallback, procRow int32) (*proto.ItemSpec, *int) {
	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{{Id: neckID, Name: "Test Neck",
		Type: proto.ItemType_ItemTypeNeck, ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}}}})

	heard := new(int)
	core.NewItemEffect(neckID, func(agent core.Agent) {
		character := agent.GetCharacter()
		character.MakeProcTriggerAura(core.ProcTrigger{
			Name:               "Test Listener",
			Callback:           callback,
			Outcome:            core.OutcomeLanded,
			TriggerImmediately: true,
			Handler:            func(*core.Simulation, *core.Spell, *core.SpellResult) { *heard++ },
		})
		if procRow != 0 {
			character.RegisterSpell(spellDataProcDamageSpell(character, spelldata.MustFind(procRow), true))
		}
	})
	return &proto.ItemSpec{Id: neckID}, heard
}

// Each of the lasso's ticks reaches a listener on the damage over time the wearer deals.
func TestOnUseDotTicksReachTheDamageDealtListeners(t *testing.T) {
	const neckID, itemID int32 = 991010, 991011
	neck, heard := listeningNeck(neckID, core.CallbackOnPeriodicDamageDealt, 0)
	sim, caster := newOnUseSimWearing(t, NewSpellDataDamageOnUse, map[int32]*proto.ItemEffect{itemID: lassoOnUse()}, neck)
	lasso := onUseSpell(t, caster, itemID)
	_, period, ticks := ticksOf(t, spelldata.MustFind(lassoSpell))

	lasso.Cast(sim, sim.Encounter.ActiveTargetUnits[0])
	dealtUntil(t, sim, lasso, sim.CurrentTime, time.Duration(ticks)*period)
	if *heard != ticks {
		t.Errorf("the listener heard %d of the lasso's ticks, want all %d", *heard, ticks)
	}
}

// Shard of the Fallen Star 21891's on-use, 26789, reaches a listener on the hits the wearer deals.
// The same row cast as a damage proc's spell does not.
func TestOnUseHitReachesTheDamageDealtListenersAndAProcsHitDoesNot(t *testing.T) {
	const neckID, itemID int32 = 991012, 991013
	const shard int32 = 26789
	neck, heard := listeningNeck(neckID, core.CallbackOnSpellHitDealt, shard)
	sim, caster := newOnUseSimWearing(t, NewSpellDataDamageOnUse, map[int32]*proto.ItemEffect{itemID: testOnUse(shard, 600000, 0, 0)}, neck)
	target := sim.Encounter.ActiveTargetUnits[0]

	onUse := onUseSpell(t, caster, itemID)
	onUse.Cast(sim, target)
	if onUse.SpellMetrics[target.UnitIndex].TotalDamage == 0 || *heard != 1 {
		t.Errorf("the on-use dealt %v and the listener heard %d hits, want its hit heard once",
			onUse.SpellMetrics[target.UnitIndex].TotalDamage, *heard)
	}

	proc := caster.GetSpell(core.ActionID{SpellID: shard})
	proc.Cast(sim, target)
	if proc.SpellMetrics[target.UnitIndex].TotalDamage == 0 || *heard != 1 {
		t.Errorf("the proc's spell dealt %v and the listener heard %d hits in all, want the proc's hit unheard",
			proc.SpellMetrics[target.UnitIndex].TotalDamage, *heard)
	}
}

// Gem-studded Leather Belt 4262's heal, 9163, reaches a listener on the healing the wearer does.
func TestOnUseHealReachesTheHealingListeners(t *testing.T) {
	const neckID, itemID int32 = 991014, 991015
	neck, heard := listeningNeck(neckID, core.CallbackOnHealDealt, 0)
	sim, caster := newOnUseSimWearing(t, NewSpellDataHealOnUse, map[int32]*proto.ItemEffect{itemID: testOnUse(9163, 300000, 0, 0)}, neck)

	onUseSpell(t, caster, itemID).Cast(sim, &caster.Unit)
	if *heard != 1 {
		t.Errorf("the listener heard %d heals, want the on-use's heal once", *heard)
	}
}

// Dragonfire Potion's 1252554 states Always Hit. Used from an item it lands for a wearer who cannot
// hit anything, and its damage over time goes on with it; the same row cast as a proc's spell rolls
// its hit and misses.
func TestAnAlwaysHitRowUsedFromAnItemNeverMisses(t *testing.T) {
	const neckID, itemID int32 = 991021, 991022
	const dragonfire int32 = 1252554
	if !spelldata.MustFind(dragonfire).AlwaysHits() {
		t.Fatalf("%d states no Always Hit, nothing to check", dragonfire)
	}
	neck, _ := listeningNeck(neckID, core.CallbackOnSpellHitDealt, dragonfire)
	sim, caster := newOnUseSimWearing(t, NewSpellDataDamageOnUse, map[int32]*proto.ItemEffect{itemID: testOnUse(dragonfire, 600000, 0, 0)}, neck)
	caster.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: -1000})
	target := sim.Encounter.ActiveTargetUnits[0]

	onUse := onUseSpell(t, caster, itemID)
	onUse.Cast(sim, target)
	if metrics := onUse.SpellMetrics[target.UnitIndex]; metrics.Hits != 1 || metrics.Misses != 0 || !onUse.Dot(target).IsActive() {
		t.Errorf("the on-use took %d hits and %d misses, damage over time up %v; want the hit and its damage over time",
			metrics.Hits, metrics.Misses, onUse.Dot(target).IsActive())
	}

	proc := caster.GetSpell(core.ActionID{SpellID: dragonfire})
	proc.Cast(sim, target)
	if metrics := proc.SpellMetrics[target.UnitIndex]; metrics.Misses != 1 || metrics.Hits != 0 {
		t.Errorf("the proc's spell took %d hits and %d misses, want it to miss", metrics.Hits, metrics.Misses)
	}
}
