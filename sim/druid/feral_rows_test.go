package druid

import (
	"testing"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func TestBerserkLeavesJumpTargetsUnparsed(t *testing.T) {
	parsed := spelldata.DryRun(spellData.Berserk.Highest())
	var jump bool
	var crit bool
	for _, skipped := range parsed.Skipped {
		if skipped.Aura == dbcenums.A_ADD_FLAT_MODIFIER && dbcenums.SpellModOp(skipped.Misc) == dbcenums.SPELLMOD_JUMP_TARGETS {
			jump = true
			if skipped.BasePoints != 3 {
				t.Errorf("jump targets base points = %v, want 3", skipped.BasePoints)
			}
		}
	}
	for _, applied := range parsed.Applied {
		if applied.Kind == "SpellMod_BonusCrit_Percent" {
			crit = true
			if applied.Value != 100 {
				t.Errorf("berserk crit mod = %v, want 100", applied.Value)
			}
		}
	}
	if !jump {
		t.Errorf("SPELLMOD_JUMP_TARGETS was parsed or missing; skipped %#v", parsed.Skipped)
	}
	if !crit {
		t.Errorf("berserk crit mod was not parsed; applied %#v", parsed.Applied)
	}
}

func TestOmenOfClarityUsesThePlaceholderRate(t *testing.T) {
	if got := spellData.OmenOfClarity.Highest().RPPM; got != 2 {
		t.Errorf("Omen of Clarity RPPM = %v, want the placeholder 2", got)
	}
}

func TestImprovedShiftingPowerReadsTheTalentCurve(t *testing.T) {
	for rank, want := range map[int32]float64{1: -4000, 2: -8000} {
		if got := spellData.ImprovedShiftingPower.Rank(rank).EffectN(1).BaseValue(); got != want {
			t.Errorf("improved shifting power rank %d cooldown = %v ms, want %v", rank, got, want)
		}
	}
}

func TestFinisherAttackPowerPlaceholders(t *testing.T) {
	if got := spellData.Rip.Highest().PeriodicDamageEffect().APCoeff(); got != 0.01 {
		t.Errorf("Rip tick attack power per combo point = %v, want the placeholder 0.01", got)
	}
	if got := spellData.FerociousBite.ByID(22829).DamageEffect().APCoeff(); got != 0.03 {
		t.Errorf("Ferocious Bite attack power per combo point = %v, want the placeholder 0.03", got)
	}
}
