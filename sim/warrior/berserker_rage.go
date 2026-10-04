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

	// Marks the talent for the APL's auraIsKnown, tagged with its points; the rage it adds is rageGain.
	if warrior.Talents.ImprovedBerserkerRage > 0 {
		core.MakePermanent(warrior.RegisterAura(core.Aura{
			Label:    "Improved Berserker Rage",
			ActionID: core.ActionID{SpellID: spellData.ImprovedBerserkerRage.Highest().ID}.WithTag(warrior.Talents.ImprovedBerserkerRage),
			Source:   core.SourceTalent,
		}))
	}

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
		Type:  core.CooldownTypeSurvival,
		ShouldActivate: func(s *core.Simulation, c *core.Character) bool {
			return rageGain > 0 && warrior.CurrentRage()+rageGain <= warrior.MaximumRage()
		},
	})
}
