package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var executeRank = spellData.Execute.Highest()

// TODO: The dummy effect carries the base damage, and no finder picks it out: the row states no
// damage effect, and both of its effects share the aura and misc values Effect() selects on.
var executeBaseDamage = executeRank.EffectN(1).Average()

// The tooltip's $*10;F1: the dummy's chain amplitude, times 10, per extra point of rage.
var executeDamagePerRage = float64(executeRank.EffectN(1).ChainAmp) * 10

func (warrior *Warrior) registerExecute() {

	var rageMetrics *core.ResourceMetrics

	// TODO: Manual review needed -- the game files state no threat coefficient; 1 until measured in game.
	config := spelldata.SpellConfig(&warrior.Unit, executeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	config.ExtraCastCondition = func(_ *core.Simulation, target *core.Unit) bool {
		return target.IsExecutePhase20()
	}

	// The cast makes an attack of its own before the damage: a miss, dodge and parry roll that deals
	// nothing, after which the damage rolls the table again. Each one that lands rolls Windfury Totem.
	// Logs show a strike after 31.6% of landed Executes (267 of 846) and 18.9% of dodged ones (35 of
	// 185). Only the listeners on a cast result hear the cast's attack.
	castAttack := warrior.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: executeRank.ID, Tag: 1},
		SpellSchool:      core.SpellSchoolPhysical,
		ProcMask:         core.ProcMaskMeleeMHSpecial,
		DefenseType:      core.DefenseTypeMelee,
		Flags:            core.SpellFlagNoMetrics | core.SpellFlagNoLogs | core.SpellFlagPassiveSpell,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
	})

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		castResult := castAttack.CalcOutcome(sim, target, castAttack.OutcomeMeleeSpecialHit)
		castAttack.DealCastResult(sim, castResult)
		castAttack.DisposeResult(castResult)

		extraRage := spell.Unit.CurrentRage()
		maxRage := warrior.MaximumRage() - spell.Cost.GetCurrentCost()
		if extraRage > maxRage {
			extraRage = maxRage
		}

		baseDamage := executeBaseDamage + executeDamagePerRage*extraRage
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

		// A miss, dodge or parry costs the base rage alone: the extra rage stays, and none of the base
		// cost comes back.
		if result.Landed() {
			warrior.SpendRage(sim, extraRage, rageMetrics)
			rageMetrics.Events--
		}
		spell.DealDamage(sim, result)
	}

	spell := warrior.RegisterSpell(config)

	rageMetrics = spell.Cost.ResourceCostImpl.(*core.RageCost).ResourceMetrics

}
