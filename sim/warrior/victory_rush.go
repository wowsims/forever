package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var victoryRushRank = spellData.VictoryRush.Highest()

var victoryRushAPCoef = victoryRushRank.EffectN(3).Percent()
var victoryRushHealPercent = victoryRushRank.EffectN(2).Percent()

var victoriousRank = spellData.VictoryRushTriggered.Highest()

// This spell works but the Sim never kills a target
func (warrior *Warrior) registerVictoryRush() {
	healthMetrics := warrior.NewHealthMetrics(core.ActionID{SpellID: victoryRushRank.ID})

	victoriousAura := warrior.RegisterAura(spelldata.AuraConfig(victoriousRank))

	config := spelldata.SpellConfig(&warrior.Unit, victoryRushRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return victoriousAura.IsActive()
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := victoryRushRank.DamageEffect().Average() + victoryRushAPCoef*spell.MeleeAttackPower(target)
		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		warrior.GainHealth(sim, warrior.MaxHealth()*victoryRushHealPercent, healthMetrics)
		victoriousAura.Deactivate(sim)
	}

	config.RelatedSelfBuff = victoriousAura

	warrior.VictoryRush = warrior.RegisterSpell(config)
}
