package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var naturesSwiftnessRank = spellData.NaturesSwiftness.Highest()

// The cooldown, the bar on Moonkin Form and the instant cast are the row's, and the next spell the
// mod reaches spends the charge. Of the spells the sim registers, its class flags reach Wrath and
// not Starfire.
func (druid *Druid) applyNaturesSwiftness() {
	if !druid.Talents.NaturesSwiftness {
		return
	}

	consumes := naturesSwiftnessRank.EffectN(1).ClassFlags

	config := spelldata.AuraConfig(naturesSwiftnessRank)
	config.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if consumes.Reaches(spell) {
			aura.Deactivate(sim)
		}
	}
	aura := druid.RegisterAura(config)
	spelldata.ParseEffects(&druid.Character, aura, naturesSwiftnessRank)

	spellConfig := spelldata.SpellConfig(&druid.Unit, naturesSwiftnessRank,
		spelldata.Flags(core.SpellFlagNoOnCastComplete))
	spellConfig.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		aura.Activate(sim)
	}
	spellConfig.RelatedSelfBuff = aura
	druid.NaturesSwiftness = druid.RegisterSpell(spellConfig)

	druid.AddMajorCooldown(core.MajorCooldown{
		Spell: druid.NaturesSwiftness.Spell,
		Type:  core.CooldownTypeDPS,
	})
}
