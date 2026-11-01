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

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		extraRage := spell.Unit.CurrentRage()
		maxRage := warrior.MaximumRage() - spell.Cost.GetCurrentCost()
		if extraRage > maxRage {
			extraRage = maxRage
		}
		warrior.SpendRage(sim, extraRage, rageMetrics)
		rageMetrics.Events--

		baseDamage := executeBaseDamage + executeDamagePerRage*extraRage
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	spell := warrior.RegisterSpell(config)

	rageMetrics = spell.Cost.ResourceCostImpl.(*core.RageCost).ResourceMetrics

}
