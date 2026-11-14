package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var ExorcismRankMap = spellData.Exorcism

// Exorcism
// https://www.wowhead.com/forever/spell=10314
//
// Causes 491 to 545 Holy damage to an Undead or Demon target.
func (paladin *Paladin) registerExorcism(_ int32, rank *spelldata.Spell) {
	damage := rank.DamageEffect()

	config := spelldata.SpellConfig(&paladin.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = SpellMaskExorcism

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
	}

	paladin.RegisterSpell(config)
}
