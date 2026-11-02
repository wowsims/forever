package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// spellData.FireBlast holds the seven trainer ranks, 2136 to 10199.
// The game files also carry 400616 to 400623, the copies the Season of Discovery rune passive
// Overheat (400615) swaps onto the action bar. Overheat is an Engrave grant with no place in
// Forever, and the generator drops its stand-ins.

// Only the highest rank is registered: the mage casts the rank configured here, not every rank
// spellData.FireBlast lists.
func (mage *Mage) registerFireBlastSpell() {
	mage.registerFireBlastRank(spellData.FireBlast.Len(), spellData.FireBlast.Highest())
}

func (mage *Mage) registerFireBlastRank(_ int32, rank *spelldata.Spell) {
	damage := rank.DamageEffect()

	config := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellFireBlast

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, damage.Average()/2, spell.OutcomeMagicHitAndCrit)
	}

	mage.RegisterSpell(config)
}
