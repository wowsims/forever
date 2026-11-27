package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var HolyWrathRankMap = spellData.HolyWrath

// Holy Wrath
// https://www.wowhead.com/forever/spell=10318
//
// Sends bolts of holy power in all directions, causing 498 to 584 Holy damage to all Undead and Demon
// targets within 20 yds and stunning them for 2 sec.
func (paladin *Paladin) registerHolyWrath(_ int32, rank *spelldata.Spell) {
	damage := rank.DamageEffect()
	targets := rank.TargetCreatureTypes()

	config := spelldata.SpellConfig(&paladin.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = SpellMaskHolyWrath

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		results := []*core.SpellResult{}
		for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
			if targets.Matches(aoeTarget) {
				results = append(results, spell.CalcDamage(sim, aoeTarget, damage.Average(), spell.OutcomeMagicHitAndCrit))
			}
		}
		for _, result := range results {
			spell.DealCastResult(sim, result)
		}

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			for _, result := range results {
				spell.DealDamage(sim, result)
			}
		})
	}

	paladin.RegisterSpell(config)
}
