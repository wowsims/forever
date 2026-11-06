//go:build with_db

package retribution

import (
	"math"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/paladin"
)

// Twenty casts of a spell on the same seed, with and without Judgement of the Crusader on the
// target and with the spell's done multiplier raised: the judgement's share of the damage, per hit
// and per crit, and whether the multiplier scales it.
func judgementOfTheCrusaderShare(t *testing.T, find func(*core.Character) *core.Spell, multiplier float64) (share float64, bonus float64) {
	t.Helper()
	run := func(judged bool) (damage, bonus float64, landed float64) {
		sim, character := paladinAgainstUndead(t, nil)
		character.AutoAttacks.CancelAutoSwing(sim)
		target := character.CurrentTarget
		if judged {
			for _, aura := range target.GetAurasWithTag(buffs.JudgementAuraTag) {
				if strings.HasPrefix(aura.Label, "Judgement of the Crusader") {
					aura.Activate(sim)
				}
			}
		}
		spell := find(character)
		if spell == nil {
			t.Fatal("the spell is not registered")
		}
		spell.DamageMultiplier *= multiplier
		for i := 0; i < 20; i++ {
			spell.SkipCastAndApplyEffects(sim, target)
			// A seal proc deals its damage a spell batch after the hit.
			for end := sim.CurrentTime + core.SpellBatchWindow; sim.CurrentTime <= end && !sim.Step(); {
			}
		}
		metrics := spell.SpellMetrics[target.UnitIndex]
		return metrics.TotalDamage, target.PseudoStats.SchoolBonusSpellDamage[stats.SchoolIndexHoly],
			float64(metrics.Hits) + float64(metrics.Crits)*spell.CritDamageMultiplier(character.AttackTables[target.UnitIndex])
	}
	without, _, landedWithout := run(false)
	with, bonus, landed := run(true)
	if bonus == 0 {
		t.Fatal("no Judgement of the Crusader debuff raised Holy damage taken")
	}
	if landed == 0 || landed != landedWithout {
		t.Fatalf("the casts landed %v and %v, want the same on the same seed", landedWithout, landed)
	}
	return (with - without) / landed, bonus
}

// Holy Strike takes Judgement of the Crusader at its coefficient after its done multipliers, so a
// crit doubles the judgement's share and Sacred Arbiter never scales it.
func TestHolyStrikeTakesJudgementOfTheCrusaderAfterDoneMultipliers(t *testing.T) {
	rank := paladin.HolyStrikeRankMap.Highest()
	find := func(character *core.Character) *core.Spell {
		return character.GetSpell(core.ActionID{SpellID: rank.ID})
	}
	share, bonus := judgementOfTheCrusaderShare(t, find, 1)
	if want := rank.DamageEffect().Coeff() * bonus; math.Abs(share-want) > 1e-6 {
		t.Errorf("a hit takes %v of the judgement, want %v", share, want)
	}
	if raised, _ := judgementOfTheCrusaderShare(t, find, 1.2); math.Abs(raised-share) > 1e-6 {
		t.Errorf("under a 1.2 done multiplier a hit takes %v of the judgement, want the %v it takes without", raised, share)
	}
}

// The Seal of Righteousness proc takes Judgement of the Crusader inside its base, with the spell
// power, so its done multipliers scale it.
func TestSealOfRighteousnessProcTakesJudgementOfTheCrusaderUnderDoneMultipliers(t *testing.T) {
	find := func(character *core.Character) *core.Spell {
		for _, spell := range character.Spellbook {
			if spell.Matches(paladin.SpellMaskSealOfRighteousnessProc) {
				return spell
			}
		}
		return nil
	}
	share, _ := judgementOfTheCrusaderShare(t, find, 1)
	if share == 0 {
		t.Fatal("the proc takes none of the judgement")
	}
	if raised, _ := judgementOfTheCrusaderShare(t, find, 1.2); math.Abs(raised-1.2*share) > 1e-6 {
		t.Errorf("under a 1.2 done multiplier a hit takes %v of the judgement, want %v", raised, 1.2*share)
	}
}

// Retribution Aura's hit takes the paladin's spell power at the shield's coefficient and none of the
// attacker's Judgement of the Crusader.
func TestRetributionAuraTakesNoJudgementOfTheCrusader(t *testing.T) {
	rank := paladin.RetributionAuraRankMap.Highest()
	const spellPower = 200
	run := func(judged bool, spellPower float64) (perHit float64) {
		sim, character := paladinAgainstUndead(t, nil)
		character.AutoAttacks.CancelAutoSwing(sim)
		target := character.CurrentTarget
		if judged {
			for _, aura := range target.GetAurasWithTag(buffs.JudgementAuraTag) {
				if strings.HasPrefix(aura.Label, "Judgement of the Crusader") {
					aura.Activate(sim)
				}
			}
			if target.PseudoStats.SchoolBonusSpellDamage[stats.SchoolIndexHoly] == 0 {
				t.Fatal("no Judgement of the Crusader debuff raised Holy damage taken")
			}
		}
		character.AddStatDynamic(sim, stats.SpellDamage, spellPower)
		aura := character.GetAuraByID(core.ActionID{SpellID: rank.ID})
		if aura == nil {
			t.Fatal("the paladin's own Retribution Aura is not registered")
		}
		aura.Activate(sim)
		shield := character.GetSpell(core.ActionID{SpellID: rank.ID, Tag: 2})
		if shield == nil {
			t.Fatal("Retribution Aura registered no spell to deal its damage with")
		}

		strike := target.RegisterSpell(core.SpellConfig{
			ActionID:         core.ActionID{OtherID: 1},
			SpellSchool:      core.SpellSchoolPhysical,
			ProcMask:         core.ProcMaskMeleeMHSpecial,
			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealDamage(sim, target, 1, spell.OutcomeAlwaysHit)
			},
		})
		for i := 0; i < 20; i++ {
			strike.SkipCastAndApplyEffects(sim, &character.Unit)
			for end := sim.CurrentTime + core.SpellBatchWindow; sim.CurrentTime <= end && !sim.Step(); {
			}
		}
		metrics := shield.SpellMetrics[target.UnitIndex]
		if metrics.Hits == 0 {
			t.Fatal("the shield landed no hit")
		}
		return metrics.TotalDamage / float64(metrics.Hits)
	}

	plain := run(false, 0)
	if judged := run(true, 0); math.Abs(judged-plain) > 1e-6 {
		t.Errorf("a hit deals %v under Judgement of the Crusader, want the %v it deals without", judged, plain)
	}
	want := plain + buffs.DamageShieldSpellPowerCoefficient*spellPower
	if powered := run(true, spellPower); math.Abs(powered-want) > 1e-6 {
		t.Errorf("a hit at %v spell power deals %v, want %v", spellPower, powered, want)
	}
}
