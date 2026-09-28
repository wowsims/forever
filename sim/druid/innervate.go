package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var innervateRank = spellData.Innervate.Highest()

// The generated aura carries Innervate's spirit regen, so the druid's own cast needs no more than
// the aura it activates.
func (druid *Druid) registerInnervateCD() {
	innervateTarget := druid.GetUnit(druid.SelfBuffs.InnervateTarget)
	if innervateTarget == nil {
		innervateTarget = &druid.Unit
	}

	innervateAura := buffs.NewSpellDataInnervatesAura(innervateTarget, true, 0)

	config := spelldata.SpellConfig(&druid.Unit, innervateRank, spelldata.Flags(core.SpellFlagAPL))
	config.ActionID.Tag = druid.Index
	config.ClassSpellMask = DruidSpellInnervate
	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		// If target already has another innervate, don't cast.
		return !innervateTarget.HasActiveAuraWithTag(buffs.InnervatesCategory)
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		innervateAura.Activate(sim)
	}

	druid.Innervate = druid.RegisterSpell(config)

	druid.AddMajorCooldown(core.MajorCooldown{
		Spell: druid.Innervate.Spell,
		Type:  core.CooldownTypeMana,
		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
			// Require manual APL usage
			return false
		},
	})
}
