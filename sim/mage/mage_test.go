package mage

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/common" // imported to get item effects included.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	goproto "google.golang.org/protobuf/proto"
)

// Enchant Weapon - Revelation.
const revelationEnchant = 8217

func init() {
	RegisterMage()
	common.RegisterAllEffects()
}

func TestArcane(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{arcaneSuite()}))
}

func TestFFB(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{ffbSuite()}))
}

func TestFire(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{fireSuite()}))
}

func TestFrost(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{frostSuite()}))
}

func TestFingersOfFrostGivesBlizzardTicksNoShatterBonus(t *testing.T) {
	const noShatter, shatter = "--000000001000000200", "--000000001000300200"
	ticksWithout, directWithout, _ := shatterCrits(noShatter)
	ticksWith, directWith, uptimeWith := shatterCrits(shatter)

	if uptimeWith == 0 {
		t.Fatal("Fingers of Frost was never up")
	}
	if ticksWithout == 0 {
		t.Fatal("no Blizzard tick crit")
	}
	if directWith <= directWithout {
		t.Fatalf("Fire Blast gained no crits with Shatter: %d without, %d with", directWithout, directWith)
	}
	if ticksWith != ticksWithout {
		t.Errorf("Blizzard tick crits moved with Shatter: %d without, %d with", ticksWithout, ticksWith)
	}
}

// With no crit chance of their own, Flamestrike's ground ticks never crit while Fingers of Frost makes the
// target Frozen for Shatter.
func TestFingersOfFrostGivesFlamestrikeTicksNoShatterBonus(t *testing.T) {
	const shatter = "--000000001000300200"
	sim, mage := newMageSimAgainst(shatter, longEncounter(1))
	target := mage.CurrentTarget
	burn, tick := flamestrikeBurn(t, mage), flamestrikeTick(t, mage)
	mage.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: -mage.GetStat(stats.SpellCritPercent)})
	if chance := tick.SpellCritChance(target); chance != 0 {
		t.Fatalf("a ground tick's crit chance is %v with no crit, want 0", chance)
	}
	mage.FingersOfFrostAura.Activate(sim)
	mage.FingersOfFrostAura.SetStacks(sim, mage.FingersOfFrostAura.MaxStacks)

	const ticks = 200
	for range ticks {
		burn.AOEDot().TickOnce(sim)
	}
	if !mage.isFrozen(target) {
		t.Fatal("the target is no longer Frozen after the ticks")
	}
	if metrics := tick.SpellMetrics[target.UnitIndex]; metrics.Hits == 0 || metrics.Crits != 0 {
		t.Errorf("%d ground ticks landed %d hits and %d crits on a Frozen target, want hits and no crit",
			ticks, metrics.Hits, metrics.Crits)
	}
}

// Each book rank registers beside the trainer's top rank, Arcane Missiles with its own missile. Both ranks of Frost
// Ward run Fire Ward's one category cooldown and keep a shield each, and the higher rank's Fireball burn holds a target.
func TestBookRanksBesideTrainerRanks(t *testing.T) {
	sim, mage := newMageSim("")
	target := mage.CurrentTarget

	spellOf := func(id int32, tag int32) *core.Spell {
		spell := mage.GetSpell(core.ActionID{SpellID: id, Tag: tag})
		if spell == nil {
			t.Fatalf("spell %d tag %d is not registered", id, tag)
		}
		return spell
	}
	for _, rank := range []*spelldata.Spell{frostboltTrainerRank, frostboltBookRank, arcaneMissilesTrainerRank, arcaneMissilesBookRank} {
		spellOf(rank.ID, 0)
	}
	spellOf(arcaneMissilesTrainerRank.PeriodicEffect().TriggerID, 0)
	spellOf(arcaneMissilesBookRank.PeriodicEffect().TriggerID, 0)

	trainerWard, bookWard := spellOf(frostWardTrainerRank.ID, 0), spellOf(frostWardBookRank.ID, 0)
	if !bookWard.Cast(sim, target) {
		t.Fatal("the book rank of Frost Ward did not cast")
	}
	if trainerWard.CD.IsReady(sim) || spellOf(spellData.FireWard.Highest().ID, 0).CD.IsReady(sim) {
		t.Error("the book rank of Frost Ward left the trainer's rank or Fire Ward ready, want one category cooldown")
	}
	if !bookWard.RelatedSelfBuff.IsActive() || trainerWard.RelatedSelfBuff.IsActive() {
		t.Error("the book rank of Frost Ward shares its shield with the trainer's rank, want a shield each")
	}

	spellOf(fireballTrainerRank.ID, 0)
	spellOf(fireballBookRank.ID, 0)
	trainerBurn := spellOf(fireballTrainerRank.ID, 1).Dot(target)
	bookBurn := spellOf(fireballBookRank.ID, 1).Dot(target)
	trainerBurn.Apply(sim)
	bookBurn.Apply(sim)
	if trainerBurn.IsActive() || !bookBurn.IsActive() {
		t.Error("the book rank's Fireball burn left the trainer's up, want it to replace the trainer's")
	}
	trainerBurn.Apply(sim)
	if trainerBurn.IsActive() || !bookBurn.IsActive() {
		t.Error("the trainer's Fireball burn landed under the book rank, want it refused")
	}
}

// An Arcane Missiles that Missile Barrage pays for keeps Clearcasting up; one it does not pay for uses it.
func TestMissileBarrageArcaneMissilesKeepsClearcasting(t *testing.T) {
	for _, withBarrage := range []bool{true, false} {
		sim, mage := newMageSim(ArcaneTalents)
		mage.ClearCasting.Activate(sim)
		if withBarrage {
			mage.MissileBarrageAura.Activate(sim)
		}
		if !mage.GetSpell(core.ActionID{SpellID: arcaneMissilesBookRank.ID}).Cast(sim, mage.CurrentTarget) {
			t.Fatalf("Arcane Missiles did not cast, Missile Barrage up = %v", withBarrage)
		}
		if mage.MissileBarrageAura.IsActive() {
			t.Error("Arcane Missiles left Missile Barrage up")
		}
		if mage.ClearCasting.IsActive() != withBarrage {
			t.Errorf("Clearcasting up after Arcane Missiles = %v with Missile Barrage up = %v, want them equal",
				mage.ClearCasting.IsActive(), withBarrage)
		}
	}
}

// An Arcane Blast hits with the Arcane Charge it adds.
func TestArcaneBlastHitsWithTheChargeItAdds(t *testing.T) {
	sim, mage := newMageSimAgainst(ArcaneTalents, longEncounter(1))
	target := mage.CurrentTarget
	arcaneBlast := mage.GetSpell(core.ActionID{SpellID: spellData.ArcaneBlast.Highest().ID})
	hit := arcaneBlast.ActionID.String() + " "
	manaMetrics := mage.NewManaMetrics(arcaneBlast.ActionID)
	var atHit float64
	hits := 0
	sim.Log = func(message string, vals ...any) {
		line := fmt.Sprintf(message, vals...)
		if strings.Contains(line, hit) && strings.Contains(line, " damage") {
			atHit = mage.GetStat(stats.SpellDamage)
			hits++
		}
	}

	belowCap, atCap := 0, 0
	for range 6 {
		mage.AddMana(sim, mage.MaxMana(), manaMetrics)
		stacks := mage.ArcaneChargesAura.GetStacks()
		before := mage.GetStat(stats.SpellDamage)
		hits = 0
		if !arcaneBlast.Cast(sim, target) {
			t.Fatalf("Arcane Blast did not cast at %v", sim.CurrentTime)
		}
		stepFor(sim, arcaneBlast.CurCast.CastTime)
		if hits == 0 {
			continue
		}
		if stacks < mage.ArcaneChargesAura.MaxStacks {
			belowCap++
			if atHit <= before {
				t.Errorf("Arcane Blast cast at %d charges hit at %.1f spell damage, want above the %.1f before the cast", stacks, atHit, before)
			}
		} else {
			atCap++
			if atHit != before {
				t.Errorf("Arcane Blast cast at the cap hit at %.1f spell damage, want the %.1f before the cast", atHit, before)
			}
		}
	}
	if belowCap == 0 || atCap == 0 {
		t.Fatalf("%d Arcane Blasts hit below the charge cap and %d at it, want both", belowCap, atCap)
	}
}

// A rotation of only Arcane Missiles procs Clearcasting at about the talent's chance per channel.
func TestArcaneMissilesSplitsClearcastingChance(t *testing.T) {
	_, mage := newMageSim(ArcaneTalents)
	chance := arcaneConcentrationChance(mage)
	metrics := runMageRotation(arcaneSuite(), ArcaneTalents, 300, arcaneMissilesBookRank.ID)
	var channels, procs float64
	for _, action := range metrics.Actions {
		if action.Id.GetSpellId() == arcaneMissilesBookRank.ID {
			for _, target := range action.Targets {
				channels += float64(target.Casts) / float64(core.DefaultSimTestOptions.Iterations)
			}
		}
	}
	for _, aura := range metrics.Auras {
		if aura.Id.GetSpellId() == spellData.Clearcasting.Highest().ID {
			procs = aura.ProcsAvg
		}
	}
	if channels == 0 || procs < chance*channels/2 || procs > chance*channels*3/2 {
		t.Fatalf("Clearcasting procced %.2f times over %.2f channels, want about the talent's %.3f a channel", procs, channels, chance)
	}
}

// The Arcane Missiles channel rolls Arcane Concentration once as it starts, at a sixth of the talent's chance.
func TestArcaneMissilesChannelStartProcsClearcasting(t *testing.T) {
	sim, mage := newMageSimAgainst(ArcaneTalents, longEncounter(1))
	target := mage.CurrentTarget
	channel := mage.GetSpell(core.ActionID{SpellID: arcaneMissilesBookRank.ID})
	chance := arcaneConcentrationChance(mage)
	manaMetrics := mage.NewManaMetrics(core.ActionID{SpellID: arcaneMissilesBookRank.ID})
	gains := 0
	mage.ClearCasting.ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) { gains++ })

	const channels = 2300
	for range channels {
		mage.AddMana(sim, mage.MaxMana(), manaMetrics)
		if !channel.Cast(sim, target) {
			t.Fatalf("Arcane Missiles did not cast at %v", sim.CurrentTime)
		}
		stepFor(sim, 100*time.Millisecond)
		dot := channel.Dot(target)
		dot.Deactivate(sim)
		if dot.TickCount() != 0 {
			t.Fatal("a missile flew before the channel was ended")
		}
		mage.ClearCasting.Deactivate(sim)
		stepFor(sim, 1400*time.Millisecond)
	}

	if rate, want := float64(gains)/channels, chance/6; rate < want/2 || rate > want*3/2 {
		t.Errorf("Clearcasting procced on %d of %d channel starts (%.4f), want about a sixth of the talent's chance, %.4f", gains, channels, rate, want)
	}
}

// Missile partial resists in EV-0142.
func TestArcaneMissilesMeetTheTargetsFullResistance(t *testing.T) {
	sim, mage := newMageSimAgainst(ArcaneTalents, longEncounter(1))
	target := mage.CurrentTarget
	// An average resist of 300 / (300 + 8 * 60 + 19) leaves no unresisted step at the full share, and
	// about seven in ten unresisted at a tenth.
	target.AddStatDynamic(sim, stats.ArcaneResistance, 300)
	channel := mage.GetSpell(core.ActionID{SpellID: arcaneMissilesBookRank.ID})
	missile := core.ActionID{SpellID: arcaneMissilesBookRank.PeriodicEffect().TriggerID}.String() + " "
	manaMetrics := mage.NewManaMetrics(core.ActionID{SpellID: arcaneMissilesBookRank.ID})
	landed, unresisted := 0, 0
	sim.Log = func(message string, vals ...any) {
		line := fmt.Sprintf(message, vals...)
		if !strings.Contains(line, missile) || !strings.Contains(line, " damage") {
			return
		}
		landed++
		if !strings.Contains(line, "% Resist)") {
			unresisted++
		}
	}

	for range 40 {
		mage.AddMana(sim, mage.MaxMana(), manaMetrics)
		if !channel.Cast(sim, target) {
			t.Fatalf("Arcane Missiles did not cast at %v", sim.CurrentTime)
		}
		stepFor(sim, 6*time.Second)
	}

	if landed < 100 || unresisted != 0 {
		t.Errorf("%d of %d landed missiles took no partial resist, want none of at least 100", unresisted, landed)
	}
}

// One Flamestrike ground tick rolls Arcane Concentration once, however many enemies stand in it.
func TestFlamestrikeGroundTickProcsClearcastingOnce(t *testing.T) {
	sim, mage := newMageSimAgainst(ArcaneTalents, longEncounter(3))
	chance := arcaneConcentrationChance(mage)
	burn := flamestrikeBurn(t, mage)
	gains := 0
	mage.ClearCasting.ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) { gains++ })

	const ticks = 2000
	for range ticks {
		burn.AOEDot().TickOnce(sim)
		stepFor(sim, 1100*time.Millisecond)
		mage.ClearCasting.Deactivate(sim)
	}

	if rate := float64(gains) / ticks; rate < chance/2 || rate > chance*3/2 {
		t.Errorf("Clearcasting procced on %d of %d ticks on 3 targets (%.3f), want about the talent's %.3f once per tick", gains, ticks, rate, chance)
	}
}

// A Blizzard tick is cast by the channel, so it rolls Arcane Concentration only when the tick's spell is
// flagged Not a Proc or the talent can proc from procs.
func TestBlizzardTickRollsArcaneConcentrationAsTheRowsAllow(t *testing.T) {
	sim, mage := newMageSimAgainst(ArcaneTalents, longEncounter(1))
	var tick *core.Spell
	for _, spell := range mage.Spellbook {
		if spell.Matches(MageSpellBlizzard) && spell.Flags.Matches(core.SpellFlagPassiveSpell) {
			tick = spell
		}
	}
	if tick == nil {
		t.Fatal("the mage has no Blizzard tick spell")
	}
	row := spelldata.MustFind(tick.ActionID.SpellID)
	want := !row.IsAProc() || spellData.ArcaneConcentration.Rank(mage.Talents.ArcaneConcentration).CanProcFromProcs()
	mage.AddStatDynamic(sim, stats.SpellHitPercent, 100)
	gains := 0
	mage.ClearCasting.ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) { gains++ })

	for range 500 {
		tick.Cast(sim, mage.CurrentTarget)
		stepFor(sim, 1100*time.Millisecond)
		mage.ClearCasting.Deactivate(sim)
	}

	if landed := tick.SpellMetrics[mage.CurrentTarget.UnitIndex]; landed.Hits+landed.Crits == 0 {
		t.Fatal("no Blizzard tick landed")
	}
	if got := gains > 0; got != want {
		t.Errorf("Blizzard ticks procced Clearcasting %d times in 500, want any: %v", gains, want)
	}
}

// A Blizzard cast rolls Arcane Concentration once as the channel starts, however many enemies stand in the
// area, after it uses Clearcasting, so a cast that uses the buff can come out with a new one.
func TestBlizzardCastRollsArcaneConcentrationOncePerCast(t *testing.T) {
	const casts = 400
	regainsAgainst := func(targets int) (regains int) {
		sim, mage := newMageSimAgainst(ArcaneTalents, longEncounter(targets))
		blizzard := mage.GetSpell(core.ActionID{SpellID: spellData.Blizzard.Highest().ID})
		manaMetrics := mage.NewManaMetrics(blizzard.ActionID)
		for range casts {
			mage.AddMana(sim, mage.MaxMana(), manaMetrics)
			mage.ClearCasting.Activate(sim)
			if !blizzard.Cast(sim, mage.CurrentTarget) {
				t.Fatalf("Blizzard did not cast at %v", sim.CurrentTime)
			}
			stepFor(sim, 50*time.Millisecond)
			if mage.ClearCasting.IsActive() {
				regains++
			}
			blizzard.AOEDot().Deactivate(sim)
			mage.ClearCasting.Deactivate(sim)
			stepFor(sim, 1500*time.Millisecond)
		}
		return regains
	}

	one, five := regainsAgainst(1), regainsAgainst(5)
	if one == 0 || one > casts/2 {
		t.Fatalf("Blizzard spent Clearcasting and came out with a new one on %d of %d casts on one target, want some but not most", one, casts)
	}
	if five > 2*one {
		t.Errorf("Blizzard came out with Clearcasting on %d of %d casts on five targets, %d on one; want about as often", five, casts, one)
	}
}

// Rolls per enemy in EV-0115.
func TestBlizzardCastRollsDraconicInfusedEmblemPerEnemy(t *testing.T) {
	const casts, enemies = 2000, 5
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotTrinket1+1)
	for slot := range items {
		items[slot] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotTrinket1].Id = 22268
	encounter := longEncounter(enemies)
	encounter.Duration = casts * 9
	sim, mage := newMageSimWearing(FrostTalents, encounter, &proto.EquipmentSpec{Items: items})
	buff := mage.GetAuraByID(core.ActionID{SpellID: 1318930})
	if buff == nil {
		t.Fatal("22268 registers no 1318930")
	}
	blizzard := mage.GetSpell(core.ActionID{SpellID: spellData.Blizzard.Highest().ID})

	// The proc is deferred by the spell batch window, one pending action each, so each step deals at most
	// one, and the buff comes down before the next roll can refresh it.
	var atCast, onTicks int
	for range casts {
		start := sim.CurrentTime
		blizzard.SkipCastAndApplyEffects(sim, mage.CurrentTarget)
		for blizzard.AOEDot().IsActive() || sim.CurrentTime < start+100*time.Millisecond {
			if sim.Step() {
				t.Fatal("the encounter ended before the casts did")
			}
			if buff.IsActive() {
				if sim.CurrentTime < start+100*time.Millisecond {
					atCast++
				} else {
					onTicks++
				}
				buff.Deactivate(sim)
			}
		}
	}

	chance := spelldata.MustFind(1318931).StatedChance()
	if chance <= 0 || chance >= 1 {
		t.Fatalf("1318931 states a chance of %v, want one between none and certain", chance)
	}
	rolls := float64(casts * enemies)
	if mean, deviation := chance*rolls, math.Sqrt(chance*(1-chance)*rolls); math.Abs(float64(atCast)-mean) > 4*deviation {
		t.Errorf("%d Blizzard casts on %d enemies procced it %d times at the cast, want about %.0f", casts, enemies, atCast, mean)
	}
	if onTicks != 0 {
		t.Errorf("Blizzard ticks procced it %d times, want none", onTicks)
	}
}

// Rolls per tick in EV-0149.
func TestFlamestrikeGroundTickRollsRevelationOnce(t *testing.T) {
	const ticks, manyTargets = 1500, 8
	one, many := flamestrikeTickRevelations(t, 1, ticks), flamestrikeTickRevelations(t, manyTargets, ticks)

	if one == 0 || many > 2*one {
		t.Errorf("Revelation procced on %d of %d ticks on one target and %d on %d targets, want about as many on %d",
			one, ticks, many, manyTargets, manyTargets)
	}
}

// With Clearcasting and Revelation up, Flamestrike's ground ticks use neither and take no crit from
// Revelation, while the cast they come from takes the crit.
func TestFlamestrikeGroundTickKeepsClearcastingAndRevelation(t *testing.T) {
	sim, mage, buff := newRevelationMage(t, ArcaneTalents, 3)
	target := mage.CurrentTarget
	burn := flamestrikeBurn(t, mage)
	tick := flamestrikeTick(t, mage)
	cast := mage.GetSpell(core.ActionID{SpellID: spellData.Flamestrike.Highest().ID})
	expirations := 0
	buff.ApplyOnExpire(func(_ *core.Aura, _ *core.Simulation) { expirations++ })

	tickCrit, castCrit := tick.SpellCritChance(target), cast.SpellCritChance(target)
	buff.Activate(sim)
	buff.SetStacks(sim, buff.MaxStacks)
	if got := tick.SpellCritChance(target); got != tickCrit {
		t.Errorf("a ground tick's crit chance is %v with Revelation up and %v without", got, tickCrit)
	}
	if got := cast.SpellCritChance(target); got <= castCrit {
		t.Errorf("Flamestrike's crit chance is %v with Revelation up and %v without", got, castCrit)
	}

	mage.ClearCasting.Activate(sim)
	if !burn.Cast(sim, target) {
		t.Fatal("the Flamestrike burn did not cast")
	}
	for tickCount := range int32(4) {
		stepFor(sim, 2100*time.Millisecond)
		if ticks := burn.AOEDot().TickCount(); ticks <= tickCount {
			t.Fatalf("the burn ticked %d times by %v", ticks, sim.CurrentTime)
		}
		if !mage.ClearCasting.IsActive() {
			t.Fatalf("Clearcasting fell off at %v while only Flamestrike's ground ticks dealt damage", sim.CurrentTime)
		}
	}
	if expirations != 0 || !buff.IsActive() {
		t.Errorf("Revelation fell off %d times while only Flamestrike's ground ticks dealt damage", expirations)
	}
}

// Fireball, Pyroblast and Frostfire Bolt have a periodic part and Revelation's mage buff does not name them.
// Each uses the charge as its cast completes and crits on its direct hit (EV-0149).
func TestRevelationCritsTheDirectHitOfFireballPyroblastAndFrostfireBolt(t *testing.T) {
	for _, test := range []struct {
		name    string
		talents string
		id      int32
	}{
		{"Fireball", FireTalents, spellData.Fireball.Highest().ID},
		{"Pyroblast", FireTalents, spellData.Pyroblast.Highest().ID},
		{"Frostfire Bolt", FFBTalents, spellData.FrostfireBolt.Highest().ID},
	} {
		sim, mage, buff := newRevelationMage(t, test.talents, 1)
		target := mage.CurrentTarget
		spell := mage.GetSpell(core.ActionID{SpellID: test.id})
		if spell == nil {
			t.Fatalf("%s is not registered", test.name)
		}
		metrics := &spell.SpellMetrics[target.UnitIndex]

		// A cast that misses uses the charge too, so cast until one lands.
		for tries := 0; metrics.Hits+metrics.Crits == 0; tries++ {
			if tries == 10 {
				t.Fatalf("%s missed 10 times", test.name)
			}
			buff.Activate(sim)
			buff.SetStacks(sim, buff.MaxStacks)
			if !spell.Cast(sim, target) {
				t.Fatalf("%s did not cast", test.name)
			}
			stepFor(sim, spell.CurCast.CastTime)
			if buff.IsActive() {
				t.Errorf("%s was cast and Revelation's charge is still up", test.name)
			}
			stepFor(sim, 3*time.Second)
			buff.Deactivate(sim)
		}
		if metrics.Crits != 1 {
			t.Errorf("%s cast with Revelation up landed without a crit", test.name)
		}
	}
}

// Charge use in EV-0149.
func TestArcaneMissilesGiveRevelationsChargeToTheNextMissileLaunched(t *testing.T) {
	for _, critsInFlight := range []bool{false, true} {
		sim, mage, buff := newRevelationMage(t, ArcaneTalents, 1)
		target := mage.CurrentTarget
		cast := mage.GetSpell(core.ActionID{SpellID: arcaneMissilesBookRank.ID})
		missile := mage.GetSpell(core.ActionID{SpellID: arcaneMissilesBookRank.PeriodicEffect().TriggerID})
		if cast == nil || missile == nil {
			t.Fatalf("Arcane Missiles %v or its missile %v is not registered", cast, missile)
		}
		// A missile flies 1.25 s, past the next one's launch a second later.
		mage.DistanceFromTarget = 25
		tick := cast.Dot(target).TickPeriod()
		if missile.TravelTime() <= tick {
			t.Fatalf("a missile flies %v and the channel ticks every %v, want the flight longer", missile.TravelTime(), tick)
		}
		mage.AddStatsDynamic(sim, stats.Stats{
			stats.SpellCritPercent: -mage.GetStat(stats.SpellCritPercent),
			stats.SpellHitPercent:  100,
		})
		missile.BonusCritPercent -= 100 * missile.SpellCritChance(target)
		metrics := &missile.SpellMetrics[target.UnitIndex]

		if critsInFlight {
			missile.BonusCritPercent += 100
		}
		if !cast.Cast(sim, target) {
			t.Fatal("Arcane Missiles did not cast")
		}
		stepFor(sim, tick+tick/2)
		if critsInFlight {
			missile.BonusCritPercent -= 100
		}
		if metrics.Hits+metrics.Crits != 1 || (metrics.Crits == 1) != critsInFlight {
			t.Fatalf("the first missile launched %d hits and %d crits, want one crit %v", metrics.Hits, metrics.Crits, critsInFlight)
		}
		buff.Activate(sim)
		buff.SetStacks(sim, buff.MaxStacks)

		// The first missile has landed and the second launched.
		stepFor(sim, missile.TravelTime()-tick/2+2*core.SpellBatchWindow)
		if critsInFlight {
			if buff.IsActive() {
				t.Error("the missile in flight at the gain landed with its own crit and Revelation's charge is still up")
			}
			continue
		}
		if !buff.IsActive() {
			t.Fatal("the missile in flight at the gain landed without a crit and spent Revelation's charge")
		}
		if metrics.Crits != 1 {
			t.Fatalf("the missile launched after the gain launched %d crits, want it to crit", metrics.Crits)
		}
		stepFor(sim, tick)
		if buff.IsActive() {
			t.Error("the missile launched after the gain landed with a crit and Revelation's charge is still up")
		}
	}
}

// The suite's gear loses Enchant Weapon - Revelation: it rolls on each Fire Blast that does not crit and its
// charge crits the next Blizzard, so Shatter's Fire Blast crits would move Blizzard's.
func shatterCrits(talents string) (tickCrits, directCrits int32, fingersOfFrostUptime float64) {
	tickID := spellData.BlizzardTriggered.Highest().ID
	directID := spellData.FireBlast.Highest().ID
	suite := frostSuite()
	suite.GearSet.GearSet = goproto.Clone(suite.GearSet.GearSet).(*proto.EquipmentSpec)
	for _, item := range suite.GearSet.GearSet.Items {
		if item.Enchant == revelationEnchant {
			item.Enchant = 0
		}
	}
	metrics := runMageRotation(suite, talents, 40, directID, spellData.Blizzard.Highest().ID)
	for _, action := range metrics.Actions {
		for _, target := range action.Targets {
			switch action.Id.GetSpellId() {
			case tickID:
				tickCrits += target.Crits
			case directID:
				directCrits += target.Crits
			}
		}
	}
	for _, aura := range metrics.Auras {
		if aura.Id.GetSpellId() == spellData.FingersOfFrostTriggered.Highest().ID {
			fingersOfFrostUptime = aura.UptimeSecondsAvg
		}
	}
	return tickCrits, directCrits, fingersOfFrostUptime
}

// A mage with no gear and no rotation, ready to cast by hand.
func newMageSim(talents string) (*core.Simulation, *Mage) {
	return newMageSimAgainst(talents, core.MakeSingleTargetEncounter(0))
}

func newMageSimAgainst(talents string, encounter *proto.Encounter) (*core.Simulation, *Mage) {
	return newMageSimWearing(talents, encounter, &proto.EquipmentSpec{})
}

func newMageSimWearing(talents string, encounter *proto.Encounter, equipment *proto.EquipmentSpec) (*core.Simulation, *Mage) {
	suite := arcaneSuite()
	player := core.WithSpec(&proto.Player{
		Class:         suite.Class,
		Race:          suite.Race,
		Equipment:     equipment,
		TalentsString: talents,
		Rotation:      &proto.APLRotation{},
	}, suite.SpecOptions.SpecOptions)
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*Mage)
}

// An hour against the given number of default targets, long enough for a test to cast by hand a thousand times.
func longEncounter(targets int) *proto.Encounter {
	encounter := core.MakeSingleTargetEncounter(0)
	encounter.Duration = 3600
	for range targets - 1 {
		encounter.Targets = append(encounter.Targets, core.NewDefaultTarget())
	}
	return encounter
}

// The burn Flamestrike's cast leaves on the ground, the first spell registered under the cast's id with
// tag 1.
func flamestrikeBurn(t *testing.T, mage *Mage) *core.Spell {
	burn := mage.GetSpell(core.ActionID{SpellID: spellData.Flamestrike.Highest().ID, Tag: 1})
	if burn == nil {
		t.Fatal("the mage has no Flamestrike burn spell")
	}
	return burn
}

// The spell each Flamestrike ground tick casts, which reports under the burn's id.
func flamestrikeTick(t *testing.T, mage *Mage) *core.Spell {
	burn := flamestrikeBurn(t, mage)
	for _, spell := range mage.Spellbook {
		if spell != burn && spell.ActionID.SameAction(burn.ActionID) {
			return spell
		}
	}
	t.Fatal("the mage has no Flamestrike ground tick spell")
	return nil
}

// A mage with no gear but a staff enchanted with Revelation, and the mage's Revelation buff.
func newRevelationMage(t *testing.T, talents string, targets int) (*core.Simulation, *Mage, *core.Aura) {
	const staff = 1613
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for slot := range items {
		items[slot] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: staff, Enchant: revelationEnchant}
	sim, mage := newMageSimWearing(talents, longEncounter(targets), &proto.EquipmentSpec{Items: items})
	buff := mage.GetAura("Revelation")
	if buff == nil {
		t.Fatal("the mage has no Revelation buff")
	}
	return sim, mage, buff
}

// How many of the given Flamestrike ground ticks proc Revelation, the buff taken down after every tick.
func flamestrikeTickRevelations(t *testing.T, targets, ticks int) (revelations int) {
	sim, mage, buff := newRevelationMage(t, ArcaneTalents, targets)
	burn := flamestrikeBurn(t, mage)
	buff.ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) { revelations++ })

	for range ticks {
		burn.AOEDot().TickOnce(sim)
		stepFor(sim, 1100*time.Millisecond)
		buff.Deactivate(sim)
	}
	return revelations
}

func arcaneConcentrationChance(mage *Mage) float64 {
	return spellData.ArcaneConcentration.Rank(mage.Talents.ArcaneConcentration).StatedChance()
}

func stepFor(sim *core.Simulation, d time.Duration) {
	for end := sim.CurrentTime + d; sim.CurrentTime < end && !sim.Step(); {
	}
}

// The suite's mage in its gear, casting the spells in priority order against one target.
func runMageRotation(suite core.CharacterSuiteConfig, talents string, duration float64, spellIDs ...int32) *proto.UnitMetrics {
	actions := make([]string, len(spellIDs))
	for i, id := range spellIDs {
		actions[i] = fmt.Sprintf(`{"action":{"castSpell":{"spellId":{"spellId":%d}}}}`, id)
	}
	rotation := fmt.Sprintf(`{"type":"TypeAPL","priorityList":[%s]}`, strings.Join(actions, ","))
	player := core.WithSpec(&proto.Player{
		Class:         suite.Class,
		Race:          suite.Race,
		Equipment:     suite.GearSet.GearSet,
		TalentsString: talents,
		Rotation:      core.APLRotationFromJsonString(rotation),
	}, suite.SpecOptions.SpecOptions)
	result := core.RunRaidSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: duration, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: core.DefaultSimTestOptions,
	})
	return result.RaidMetrics.Parties[0].Players[0]
}

func BenchmarkMage(b *testing.B) {
	b.Run("Arcane", func(b *testing.B) { core.CharacterBenchmark(b, arcaneSuite()) })
	b.Run("FFB", func(b *testing.B) { core.CharacterBenchmark(b, ffbSuite()) })
	b.Run("Fire", func(b *testing.B) { core.CharacterBenchmark(b, fireSuite()) })
	b.Run("Frost", func(b *testing.B) { core.CharacterBenchmark(b, frostSuite()) })
}

func arcaneSuite() core.CharacterSuiteConfig {
	config := mageSuite(ArcaneTalents, "arcane")
	config.OtherRotations = []core.RotationCombo{
		core.RotationWithSpellIds("arcane_book_ranks", config.Rotation, map[int32]int32{10212: 25345}),
	}
	return config
}

func ffbSuite() core.CharacterSuiteConfig {
	return mageSuite(FFBTalents, "ffb")
}

func fireSuite() core.CharacterSuiteConfig {
	config := mageSuite(FireTalents, "deepFire")
	config.OtherRotations = []core.RotationCombo{
		core.RotationWithSpellIds("deepFire_book_ranks", config.Rotation, map[int32]int32{10151: 25306, 10212: 25345}),
	}
	return config
}

func frostSuite() core.CharacterSuiteConfig {
	config := mageSuite(FrostTalents, "frost")
	config.StartingDistance = 30
	config.OtherRotations = []core.RotationCombo{
		core.RotationWithSpellIds("frost_book_ranks", config.Rotation, map[int32]int32{10181: 25304}),
	}
	return config
}

func mageSuite(talents string, apl string) core.CharacterSuiteConfig {
	return core.CharacterSuiteConfig{
		Class:      proto.Class_ClassMage,
		Race:       proto.Race_RaceHighOrderSkyborne,
		OtherRaces: []proto.Race{proto.Race_RaceGnome},
		SpecOptions: core.SpecOptionsCombo{Label: "Default", SpecOptions: &proto.Player_Mage{
			Mage: &proto.Mage{
				Options: &proto.Mage_Options{
					ClassOptions: &proto.MageOptions{
						DefaultMageArmor: proto.MageArmor_MageArmorMageArmor,
					},
				},
			},
		}},
		Consumables: DefaultConsumables,
		Profession1: proto.Profession_Engineering,
		Profession2: proto.Profession_Tailoring,
		GearSet:     core.GetGearSet("../../ui/specs/mage/dps/gear_sets", "preraid"),
		Talents:     talents,
		Rotation:    core.GetAplRotation("../../ui/specs/mage/dps/apls", apl),
		ItemFilter: core.ItemFilter{
			WeaponTypes: []proto.WeaponType{
				proto.WeaponType_WeaponTypeDagger,
				proto.WeaponType_WeaponTypeSword,
				proto.WeaponType_WeaponTypeOffHand,
				proto.WeaponType_WeaponTypeStaff,
			},
			ArmorType: proto.ArmorType_ArmorTypeCloth,
			RangedWeaponTypes: []proto.RangedWeaponType{
				proto.RangedWeaponType_RangedWeaponTypeWand,
			},
			EnchantBlacklist: []int32{2673, 3225, 3273},
		},
	}
}

var ArcaneTalents = "050205223101311531-5500003-003"
var FFBTalents = "-5550000123011303-2035010311013002"
var FireTalents = "0502052001001-55500001230133031-003"
var FrostTalents = "203205-5-053500011121331251"

var DefaultConsumables = &proto.ConsumesSpec{
	FlaskId: 13512, // Flask of Supreme Power
	FoodId:  13931, // Nightfin Soup
	ElixirIds: []int32{
		13454,  // Greater Arcane Elixir
		250348, // Elixir of the Whale
		250341, // Greater Mageblood Elixir
	},
	MhImbueId:       20749,  // Brilliant Wizard Oil
	MhWeaponImbueId: 277502, // Scroll of Imbue Precision
	ResourcePotId:   13444,  // Major Mana Potion
}

// A Counterspell that lands deals its outcome, so a listener for landed harmful spells hears it:
// Blade of Eternal Darkness 17780 procs off it.
func TestCounterspellLandingIsHeard(t *testing.T) {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 17780
	encounter := core.MakeSingleTargetEncounter(0)
	encounter.Duration = 36000
	sim, mage := newMageSimWearing("", encounter, &proto.EquipmentSpec{Items: items})
	counterspell := mage.GetSpell(core.ActionID{SpellID: counterspellRank.ID})
	shadows := mage.GetSpell(core.ActionID{SpellID: 27860})
	target := sim.Encounter.ActiveTargetUnits[0]
	mage.AddDynamicMod(core.SpellModConfig{Kind: core.SpellMod_PowerCost_Pct, FloatValue: -1}).Activate()

	for range 1000 {
		stepFor(sim, max(counterspell.TimeToReady(sim), mage.GCD.TimeToReady(sim), core.SpellBatchWindow))
		counterspell.Cast(sim, target)
	}
	stepFor(sim, core.SpellBatchWindow)

	if casts := counterspell.SpellMetrics[target.UnitIndex].Casts; casts != 1000 {
		t.Fatalf("%d Counterspells cast, want 1000", casts)
	}
	if shadows == nil || shadows.SpellMetrics[target.UnitIndex].Casts == 0 {
		t.Error("no landed Counterspell procced Blade of Eternal Darkness")
	}
}

// Frostfire Bolt's damage over time ticks meet the target's full resistance, so Spell Piercing beyond it
// gives them its bonus.
func TestFrostfireBoltTicksTakeTheSpellPiercingBonus(t *testing.T) {
	sim, mage := newMageSimAgainst(FFBTalents, longEncounter(1))
	target := mage.CurrentTarget
	rank := spellData.FrostfireBolt.Highest()
	bolt := mage.GetSpell(core.ActionID{SpellID: rank.ID})
	dot := mage.GetSpell(core.ActionID{SpellID: rank.ID, Tag: 1})
	if bolt == nil || dot == nil {
		t.Fatal("Frostfire Bolt or its damage over time is not registered")
	}
	mage.AddStatsDynamic(sim, stats.Stats{stats.SpellPiercing: 100})
	mage.AddDynamicMod(core.SpellModConfig{Kind: core.SpellMod_PowerCost_Pct, FloatValue: -1}).Activate()

	for range 100 {
		stepFor(sim, max(bolt.TimeToReady(sim), mage.GCD.TimeToReady(sim), core.SpellBatchWindow))
		bolt.Cast(sim, target)
		stepFor(sim, bolt.CurCast.CastTime+15*time.Second)
	}

	metrics := dot.SpellMetrics[target.UnitIndex]
	if metrics.Ticks+metrics.CritTicks == 0 {
		t.Fatal("no Frostfire Bolt tick landed")
	}
	if metrics.PiercedTicks+metrics.PiercedCritTicks == 0 {
		t.Errorf("none of %d Frostfire Bolt ticks took a Spell Piercing bonus", metrics.Ticks+metrics.CritTicks)
	}
}

// Each enemy a Flamestrike ground tick lands on adds a stack of Combustion's crit, and each crit uses one of
// its charges.
func TestCombustionCountsEachFlamestrikeTickHit(t *testing.T) {
	sim, mage := newMageSimAgainst(FireTalents, longEncounter(3))
	charges, crit := mage.GetAura("Combustion"), mage.GetAura("Combustion Crit")
	tick := flamestrikeTick(t, mage)

	charges.Activate(sim)
	landed, crits := 0, 0
	for _, result := range tick.CalcAndDealAoeDamage(sim, 100, tick.OutcomeMagicHitAndCrit) {
		if result.Landed() {
			landed++
		}
		if result.DidCrit() {
			crits++
		}
	}
	if crit.GetStacks() != int32(1+landed) {
		t.Errorf("a tick that landed on %d enemies left %d stacks of Combustion's crit, want %d", landed, crit.GetStacks(), 1+landed)
	}
	if charges.GetStacks() != int32(3-crits) {
		t.Errorf("a tick that crit %d enemies left %d of Combustion's charges, want %d", crits, charges.GetStacks(), 3-crits)
	}
}
