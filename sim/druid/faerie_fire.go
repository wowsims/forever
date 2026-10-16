package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// TODO: Faerie Fire does not reset the swing timer.
// TODO: Ingame research needed: Faerie Fire's threat in Bear and Dire Bear Form; its E_THREAT effect states 0.
func (druid *Druid) registerFaerieFireSpell() {
	rank := spellData.FaerieFire.Highest()
	druid.FaerieFireAuras = druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return buffs.NewSpellDataFaerieFireAura(target, true, 0)
	})
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	// The timer must exist before the form passive adds its recovery time.
	config.Cast.CD = core.Cooldown{Timer: druid.NewTimer(), Duration: spellData.CatFormTriggered.Highest().EffectN(4).TimeValue()}
	// Cost, cooldown and GCD in cat or bear come from the form passive's spell mods. A cast on a target
	// that has the debuff refreshes it; when to refresh is the rotation's.
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
		if result.Landed() {
			druid.FaerieFireAuras.Get(target).Activate(sim)
		}
	}
	config.RelatedAuraArrays = druid.FaerieFireAuras.ToMap()
	druid.FaerieFire = druid.RegisterSpell(config)
	druid.FaerieFire.CD.Duration = rank.Cooldown()
}
