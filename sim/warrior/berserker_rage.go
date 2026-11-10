package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// TODO: Ingame test needed -- the game files state no amount for the extra rage a hit taken
// generates while Berserker Rage is up; doubled.
const berserkerRageDamageTakenRageMultiplier = 2.0

var berserkerRageRank = spellData.BerserkerRage.Highest()

func (warrior *Warrior) registerBerserkerRage() {
	rageMetrics := warrior.NewRageMetrics(core.ActionID{SpellID: berserkerRageRank.ID})
	rageGain := spellData.ImprovedBerserkerRage.EffectAt(1).TenthsAt(warrior.Talents.ImprovedBerserkerRage)

	auraConfig := spelldata.AuraConfig(berserkerRageRank)
	auraConfig.OnGain = func(aura *core.Aura, sim *core.Simulation) {
		warrior.MultiplyDamageTakenRageGen(berserkerRageDamageTakenRageMultiplier)
	}
	auraConfig.OnExpire = func(aura *core.Aura, sim *core.Simulation) {
		warrior.MultiplyDamageTakenRageGen(1 / berserkerRageDamageTakenRageMultiplier)
	}
	aura := warrior.RegisterAura(auraConfig).AttachFearImmunity()

	config := spelldata.SpellConfig(&warrior.Unit, berserkerRageRank,
		spelldata.Flags(core.SpellFlagAPL|core.SpellFlagCastWhileIncapacitated))

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		if rageGain > 0 {
			warrior.AddRage(sim, rageGain, rageMetrics)
		}
		aura.Activate(sim)
	}

	config.RelatedSelfBuff = aura

	spell := warrior.RegisterSpell(config)

	warrior.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		// It goes on its own for the rage it gains, which a DPS warrior wants too.
		Type: core.CooldownTypeSurvival | core.CooldownTypeDPS,
		ShouldActivate: func(s *core.Simulation, c *core.Character) bool {
			return rageGain > 0 && warrior.CurrentRage()+rageGain <= warrior.MaximumRage()
		},
	})
}
