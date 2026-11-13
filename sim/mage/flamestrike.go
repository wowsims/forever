package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Only the highest rank is registered: the mage casts the rank configured here, not every rank
// spellData.Flamestrike lists.
func (mage *Mage) registerFlamestrikeSpells() {
	mage.registerFlamestrikeSpell(spellData.Flamestrike.Len(), spellData.Flamestrike.Highest())
}

// The burn is its own spell (MageSpellFlamestrikeDot) on the cast's row, tagged like Fireball's DoT so
// the two share a header: every 2 sec it casts the same rank of spellData.FlamestrikeTriggered, which
// hits each enemy in the area and rolls a crit unless the tick's row bars it. The cast's row names no
// tick spell, only an area trigger, so the two pair by rank. Each tick is a direct hit of the tick
// spell, so a hit listener hears every enemy it lands on: Revelation rolls per enemy, Ignite takes
// each crit, and Arcane Concentration rolls once a tick. The tick's row is outside the class flags of
// Revelation's crit and charge, and it costs nothing, so it neither takes nor spends Revelation or
// Clearcasting. A tick takes no partial resist: in game, mobs that partially resist about a quarter of
// Scorch and Fire Blast hits resist almost no ticks.
// TODO: Verify in game whether a tick rolls a partial resist, and whether Spell Piercing beyond a
// target's resistance adds to it.
func (mage *Mage) registerFlamestrikeSpell(rank int32, castRank *spelldata.Spell) {
	tickRank := spellData.FlamestrikeTriggered.Rank(rank)
	cannotCrit := tickRank.CannotCrit()

	tickConfig := spelldata.SpellConfig(&mage.Unit, tickRank, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc(),
		spelldata.Flags(core.SpellFlagNoPartialResist))
	tickConfig.ClassSpellMask = MageSpellFlamestrikeDot
	tickConfig.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		outcome := spell.OutcomeMagicHitAndCrit
		if cannotCrit {
			outcome = spell.OutcomeMagicHit
		}
		spell.CalcAndDealAoeDamageWithVariance(sim, outcome, rollDamage(tickRank.DamageEffect()))
	}
	tickSpell := mage.RegisterSpell(tickConfig)

	dotConfig := spelldata.SpellConfig(&mage.Unit, castRank, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc(), spelldata.Tag(1))
	dotConfig.ClassSpellMask = MageSpellFlamestrikeDot
	dotConfig.Dot = spelldata.DotConfig(castRank, castRank.Effect(dbcenums.A_PERIODIC_DUMMY, 0),
		spelldata.RankLabel(castRank, &mage.Unit))
	dotConfig.Dot.IsAOE = true
	dotConfig.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, _ *core.Dot) {
		tickSpell.Cast(sim, target)
	}
	dotConfig.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.AOEDot().Apply(sim)
	}
	dotSpell := mage.RegisterSpell(dotConfig)

	config := spelldata.SpellConfig(&mage.Unit, castRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellFlamestrike
	config.RelatedDotSpell = dotSpell

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeMagicHitAndCrit, rollDamage(castRank.DamageEffect()))
		dotSpell.Cast(sim, target)
	}

	mage.RegisterSpell(config)
}
