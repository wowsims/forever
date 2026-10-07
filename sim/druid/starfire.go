package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// TODO: was Ranks(6, 8); Forever's Starfire tops out at rank 7, so the max-rank entry
// moves down rather than naming a rank the table does not hold.
var StarfireRankMap = spelldata.Ranked(spellData.Starfire.Rank(6).ID, spellData.Starfire.Rank(7).ID)

func (druid *Druid) registerStarfireSpell(rankConfig *spelldata.Spell) {
	damage := rankConfig.DamageEffect()

	config := spelldata.SpellConfig(&druid.Unit, rankConfig, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = DruidSpellStarfire

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
	}

	druid.Starfire = append(druid.Starfire, druid.RegisterSpell(config))
}
