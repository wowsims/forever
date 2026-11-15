package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var HammerOfWrathRankMap = spellData.HammerOfWrath

// Hammer of Wrath
// https://www.wowhead.com/forever/spell=24239
//
// Hurls a hammer that strikes an enemy for 489 to 538 Holy damage. Only usable on enemies that have
// 20% or less health.
func (paladin *Paladin) registerHammerOfWrath(_ int32, rank *spelldata.Spell) {
	damage := rank.DamageEffect()

	config := spelldata.SpellConfig(&paladin.Unit, rank, spelldata.Magic(core.ProcMaskRangedSpecial))
	config.ClassSpellMask = SpellMaskHammerOfWrath
	// The game files' 1s GCD sits at core's floor, so it is named as the floor too or GCDTime clamps
	// it.
	config.Cast.DefaultCast.GCDMin = rank.GCD()

	config.ExtraCastCondition = func(_ *core.Simulation, target *core.Unit) bool {
		return target.IsExecutePhase20()
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, damage.Average(), spell.OutcomeRangedHitAndCrit)
		spell.DealDamageAfterTravel(sim, result)
	}

	paladin.RegisterSpell(config)
}
