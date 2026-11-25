package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerInnervateCD() {
	rank := spellData.Innervate.Highest()
	// Every raid player can receive Innervate; RegisterAura cannot run after finalize.
	for _, unit := range druid.Env.Raid.AllPlayerUnits {
		buffs.NewSpellDataInnervatesAura(unit, true, 0)
	}
	// Dreamstate gives the druid its share of an Innervate cast on another player.
	var dreamstate *core.Aura
	if druid.Talents.Dreamstate > 0 {
		percent := spellData.Dreamstate.Rank(druid.Talents.Dreamstate).EffectN(1).BaseValue()
		dreamstate = buffs.NewSpellDataInnervatesAuraShare(&druid.Unit, int32(percent))
	}
	config := spelldata.SpellConfig(&druid.Unit, rank)
	config.Flags |= core.SpellFlagAPL | core.SpellFlagHelpful
	config.ExtraCastCondition = func(_ *core.Simulation, castTarget *core.Unit) bool {
		return !druid.innervateRecipient(castTarget).HasActiveAuraWithTag(buffs.InnervatesCategory)
	}
	config.ApplyEffects = func(sim *core.Simulation, castTarget *core.Unit, _ *core.Spell) {
		recipient := druid.innervateRecipient(castTarget)
		buffs.NewSpellDataInnervatesAura(recipient, true, 0).Activate(sim)
		if dreamstate != nil && recipient != &druid.Unit {
			dreamstate.Activate(sim)
		}
	}
	druid.Innervate = druid.RegisterSpell(config)
	druid.AddMajorCooldown(core.MajorCooldown{
		Spell:           druid.Innervate.Spell,
		Type:            core.CooldownTypeMana,
		DisableAutocast: true,
	})
}

// A friendly APL target receives the mana. Cast Spell aims at the encounter, so the dest is InnervateTarget or the caster.
func (druid *Druid) innervateRecipient(castTarget *core.Unit) *core.Unit {
	if castTarget != nil && castTarget.Type == core.PlayerUnit {
		return castTarget
	}
	if target := druid.GetUnit(druid.SelfBuffs.InnervateTarget); target != nil {
		return target
	}
	return &druid.Unit
}
