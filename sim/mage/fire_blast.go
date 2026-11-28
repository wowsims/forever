package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The game files carry 400616 to 400623, the copies the rune passive Overheat (400615) swaps onto
// the action bar. Overheat is an Engrave grant with no place in Forever, and the generator
// drops its stand-ins.

// Only the highest rank is registered.
func (mage *Mage) registerFireBlastSpell() {
	mage.registerFireBlastRank(spellData.FireBlast.Len(), spellData.FireBlast.Highest())
}

func (mage *Mage) registerFireBlastRank(_ int32, rank *spelldata.Spell) {
	damage := rank.DamageEffect()

	config := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellFireBlast

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
	}

	mage.RegisterSpell(config)
}
