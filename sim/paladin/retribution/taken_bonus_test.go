//go:build with_db

package retribution

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/paladin"
)

var holyStrikeAndSealOfFuryHits = map[string]int64{
	"Holy Strike":       paladin.SpellMaskHolyStrike,
	"Seal of Fury":      paladin.SpellMaskSealOfFuryProc,
	"Judgement of Fury": paladin.SpellMaskJudgementOfFury,
}

func spellWithMask(mask int64) func(*core.Character) *core.Spell {
	return func(character *core.Character) *core.Spell {
		for _, spell := range character.Spellbook {
			if spell.Matches(mask) {
				return spell
			}
		}
		return nil
	}
}

func judgeCrusader(sim *core.Simulation, target *core.Unit) {
	for _, aura := range target.GetAurasWithTag(buffs.JudgementAuraTag) {
		if strings.HasPrefix(aura.Label, "Judgement of the Crusader") {
			aura.Activate(sim)
		}
	}
}

// Twenty casts of a spell on the same seed after setup has changed the paladin, the spell or the
// target: the damage they deal, and the hits and crits they land, a crit counting as its multiplier.
func twentyCasts(t *testing.T, find func(*core.Character) *core.Spell, setup func(*core.Simulation, *core.Character, *core.Spell)) (damage, landed float64) {
	t.Helper()
	sim, character := paladinAgainstUndead(t, nil)
	character.AutoAttacks.CancelAutoSwing(sim)
	target := character.CurrentTarget
	spell := find(character)
	if spell == nil {
		t.Fatal("the spell is not registered")
	}
	setup(sim, character, spell)
	for range 20 {
		spell.SkipCastAndApplyEffects(sim, target)
		// A seal proc deals its damage a spell batch after the hit.
		for end := sim.CurrentTime + core.SpellBatchWindow; sim.CurrentTime <= end && !sim.Step(); {
		}
	}
	metrics := spell.SpellMetrics[target.UnitIndex]
	return metrics.TotalDamage, float64(metrics.Hits) + float64(metrics.Crits)*spell.CritDamageMultiplier(character.AttackTables[target.UnitIndex])
}

// Twenty casts of a spell on the same seed, with and without Judgement of the Crusader on the
// target and with the spell's done multiplier raised: the judgement's share of the damage, per hit
// and per crit, and whether the multiplier scales it.
func judgementOfTheCrusaderShare(t *testing.T, find func(*core.Character) *core.Spell, multiplier float64) (share float64, bonus float64) {
	t.Helper()
	run := func(judged bool) (damage, landed float64) {
		return twentyCasts(t, find, func(sim *core.Simulation, character *core.Character, spell *core.Spell) {
			if judged {
				judgeCrusader(sim, character.CurrentTarget)
				bonus = character.CurrentTarget.PseudoStats.SchoolBonusSpellDamage[stats.SchoolIndexHoly]
			}
			spell.DamageMultiplier *= multiplier
		})
	}
	without, landedWithout := run(false)
	with, landed := run(true)
	if bonus == 0 {
		t.Fatal("no Judgement of the Crusader debuff raised Holy damage taken")
	}
	if landed == 0 || landed != landedWithout {
		t.Fatalf("the casts landed %v and %v, want the same on the same seed", landedWithout, landed)
	}
	return (with - without) / landed, bonus
}

// The Seal of Righteousness proc takes Judgement of the Crusader inside its base, with the spell
// power, so its done multipliers scale it.
func TestSealOfRighteousnessProcTakesJudgementOfTheCrusaderUnderDoneMultipliers(t *testing.T) {
	find := spellWithMask(paladin.SpellMaskSealOfRighteousnessProc)
	share, _ := judgementOfTheCrusaderShare(t, find, 1)
	if share == 0 {
		t.Fatal("the proc takes none of the judgement")
	}
	if raised, _ := judgementOfTheCrusaderShare(t, find, 1.2); math.Abs(raised-1.2*share) > 1e-6 {
		t.Errorf("under a 1.2 done multiplier a hit takes %v of the judgement, want %v", raised, 1.2*share)
	}
}

// Holy Strike, Seal of Fury's hit and its judgement take Judgement of the Crusader at their
// coefficient after their done multipliers, so a crit doubles its share and no multiplier scales it.
func TestJudgementOfTheCrusaderLandsAfterDoneMultipliers(t *testing.T) {
	for name, mask := range holyStrikeAndSealOfFuryHits {
		t.Run(name, func(t *testing.T) {
			var coefficient float64
			find := func(character *core.Character) *core.Spell {
				spell := spellWithMask(mask)(character)
				if spell != nil {
					coefficient = spelldata.MustFind(spell.ActionID.SpellID).DamageEffect().Coeff()
				}
				return spell
			}
			share, bonus := judgementOfTheCrusaderShare(t, find, 1)
			if coefficient == 0 {
				t.Fatal("the spell's row states no spell power coefficient")
			}
			if want := coefficient * bonus; math.Abs(share-want) > 1e-6 {
				t.Errorf("a hit takes %v of the judgement, want %v", share, want)
			}
			if raised, _ := judgementOfTheCrusaderShare(t, find, 1.2); math.Abs(raised-share) > 1e-6 {
				t.Errorf("under a 1.2 done multiplier a hit takes %v of the judgement, want the %v it takes without", raised, share)
			}
		})
	}
}

// One Consecration on one enemy more than the ground marks, on the same seed, after setup has
// changed the paladin, the ticks' spell or the enemies: per enemy, the damage its ticks deal and the
// ticks that land, a crit counting as its multiplier. The last enemy is the one left unmarked.
func consecrationTicks(t *testing.T, setup func(*core.Simulation, *core.Character, *core.Spell)) (damage, landed []float64) {
	t.Helper()
	rank := paladin.ConsecrationRankMap.Highest()
	marked := int(rank.Effect(dbcenums.A_PERIODIC_DUMMY, 0).Average())
	sim, character := paladinAgainstUndeadTargets(t, nil, marked+1)
	character.AutoAttacks.CancelAutoSwing(sim)
	spell := character.GetSpell(core.ActionID{SpellID: rank.ID})
	if spell == nil {
		t.Fatal("Consecration is not registered")
	}
	// The ticks are the pulse row's spell, which holds the ground.
	ticks := spell.RelatedDotSpell
	if ticks == nil {
		t.Fatal("Consecration has no spell for its ticks")
	}
	setup(sim, character, ticks)
	spell.SkipCastAndApplyEffects(sim, character.CurrentTarget)
	for end := sim.CurrentTime + rank.Duration() + time.Second; sim.CurrentTime <= end && !sim.Step(); {
	}
	for _, target := range sim.Encounter.AllTargetUnits {
		metrics := ticks.SpellMetrics[target.UnitIndex]
		if metrics.ResistedTicks+metrics.ResistedCritTicks > 0 {
			t.Fatalf("%d of the ticks on an enemy were partly resisted, want none", metrics.ResistedTicks+metrics.ResistedCritTicks)
		}
		damage = append(damage, metrics.TotalDamage)
		landed = append(landed, float64(metrics.Ticks)+float64(metrics.CritTicks)*ticks.CritDamageMultiplier(character.AttackTables[target.UnitIndex]))
	}
	return damage, landed
}

// Consecration's marked tick takes Judgement of the Crusader and the spell damage against the
// target's creature type at the bonus damage's coefficient: the spell damage under the done
// multipliers, the judgement after them, which no done multiplier scales (EV-0143). The tick on an
// unmarked enemy takes neither.
func TestConsecrationMarkedTickTakesJudgementOfTheCrusaderAfterDoneMultipliers(t *testing.T) {
	coefficient := paladin.ConsecrationRankMap.Highest().Refs()[0].EffectN(2).Coeff()
	if coefficient == 0 {
		t.Fatal("the tick's bonus damage states no spell power coefficient")
	}
	const versus = 100.0
	var bonus float64
	run := func(judged bool, spellDamage, multiplier float64) (damage, landed []float64) {
		return consecrationTicks(t, func(sim *core.Simulation, character *core.Character, spell *core.Spell) {
			if judged {
				for _, target := range sim.Encounter.AllTargetUnits {
					judgeCrusader(sim, target)
				}
				bonus = character.CurrentTarget.PseudoStats.SchoolBonusSpellDamage[stats.SchoolIndexHoly]
			}
			character.AddMobTypeBonusStats(stats.Stats{stats.SpellDamage: spellDamage}, proto.MobType_MobTypeUndead)
			spell.DamageMultiplier *= multiplier
		})
	}

	for _, multiplier := range []float64{1, 1.2} {
		without, landed := run(false, 0, multiplier)
		judged, judgedLanded := run(true, 0, multiplier)
		versused, _ := run(false, versus, multiplier)
		if bonus == 0 {
			t.Fatal("no Judgement of the Crusader debuff raised Holy damage taken")
		}
		for i := range landed {
			if landed[i] == 0 || landed[i] != judgedLanded[i] {
				t.Fatalf("the ticks on enemy %d landed %v and %v, want the same on the same seed", i, landed[i], judgedLanded[i])
			}
			wantJudgement, wantVersus := coefficient*bonus, coefficient*versus*multiplier
			if i == len(landed)-1 {
				wantJudgement, wantVersus = 0, 0
			}
			if got := (judged[i] - without[i]) / landed[i]; math.Abs(got-wantJudgement) > 1e-6 {
				t.Errorf("under a %v done multiplier a tick on enemy %d takes %v of the judgement, want %v", multiplier, i, got, wantJudgement)
			}
			if got := (versused[i] - without[i]) / landed[i]; math.Abs(got-wantVersus) > 1e-6 {
				t.Errorf("under %v spell power against Undead and a %v done multiplier a tick on enemy %d takes %v of it, want %v", versus, multiplier, i, got, wantVersus)
			}
		}
	}
}
