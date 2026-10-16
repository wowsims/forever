package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The top rank, 7, and the cheaper rank 6, which an APL can cast in its place.
var StarfireRankMap = spelldata.Ranked(spellData.Starfire.Rank(6).ID, spellData.Starfire.Rank(7).ID)

func (druid *Druid) registerStarfireSpell(rankConfig *spelldata.Spell) {
	damage := rankConfig.DamageEffect()

	config := spelldata.SpellConfig(&druid.Unit, rankConfig, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
	}

	druid.Starfire = append(druid.Starfire, druid.RegisterSpell(config))
}
