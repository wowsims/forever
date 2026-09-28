package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var wrathRank = spellData.Wrath.Highest()
var wrathDamage = wrathRank.DamageEffect()

func (druid *Druid) registerWrathSpell() {
	config := spelldata.SpellConfig(&druid.Unit, wrathRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = DruidSpellWrath

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, wrathDamage.Roll(sim, core.CharacterLevel), spell.OutcomeMagicHitAndCrit)

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
		})
	}

	druid.Wrath = druid.RegisterSpell(config)
}
