package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var victoryRushRank = spellData.VictoryRush.Highest()

var victoryRushAPCoef = victoryRushRank.EffectN(3).Percent()
var victoryRushHealPercent = victoryRushRank.EffectN(2).Percent()

var victoriousRank = spellData.VictoryRushTriggered.Highest()
var victoriousStateRank = spellData.VictoriousState.Highest()

// Victorious State hears the killing blow and casts Victorious, which Victory Rush needs and spends.
// The kill roll stands for the game's experience or honor check, so an add counts like any enemy.
func (warrior *Warrior) registerVictoryRush() {
	// The heal makes healing threat as any heal does: half the health it restores, times the stance,
	// split between the enemies in combat. Whether healing done and taken modifiers change it is not
	// measured, so it restores the row's share of maximum health with none of them.
	healConfig := spelldata.SpellConfig(&warrior.Unit, victoryRushRank, spelldata.Proc(), spelldata.Tag(1),
		spelldata.Flags(core.SpellFlagIgnoreModifiers))
	healConfig.ProcMask = core.ProcMaskSpellHealing
	healConfig.DamageMultiplier = 1
	healConfig.ThreatMultiplier = 1
	healConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealHealing(sim, target, warrior.MaxHealth()*victoryRushHealPercent, spell.OutcomeAlwaysHit)
	}
	heal := warrior.RegisterSpell(healConfig)

	victoriousAura := warrior.RegisterAura(spelldata.AuraConfig(victoriousRank))
	warrior.MakeProcTriggerAura(spelldata.ProcTrigger(&warrior.Character, victoriousStateRank,
		func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			victoriousAura.Activate(sim)
		}))

	config := spelldata.SpellConfig(&warrior.Unit, victoryRushRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return victoriousAura.IsActive()
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := victoryRushRank.DamageEffect().Average() + victoryRushAPCoef*spell.MeleeAttackPower(target)
		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		heal.Cast(sim, &warrior.Unit)
		victoriousAura.Deactivate(sim)
	}

	config.RelatedSelfBuff = victoriousAura

	warrior.VictoryRush = warrior.RegisterSpell(config)
}
