package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// The debuff half is A_MOD_SPELL_DAMAGE_FROM_CASTER: while the channel is on the target, it takes 10%
// more damage from the caster's spells the effect's class flags name. Those are Corruption and Bane of
// Agony (0x2 | 0x400); Siphon Life, a leech rather than periodic damage, and the drains are not named.
func (warlock *Warlock) registerWrackSpell() {
	rankedSpell := spellData.Wrack.ByID(1316697)
	wrackTick := rankedSpell.PeriodicEffect()
	wrackAmplify := rankedSpell.Effect(dbcenums.A_MOD_SPELL_DAMAGE_FROM_CASTER, 0)

	ampMult := 1 + wrackAmplify.Percent()

	// Damage done by caster: this warlock's damage from the named spells, against the target Wrack is on.
	amplifyNamed := func(_ *core.Simulation, spell *core.Spell, _ *core.AttackTable) float64 {
		if spell.MatchesFlags(wrackAmplify.ClassFlags) {
			return ampMult
		}
		return 1
	}

	config := warlock.soulSiphonChannelConfig(rankedSpell, wrackTick, nil)
	config.ClassSpellMask = WarlockSpellWrack

	// While the channel is on the target, this warlock's damage from the named spells is raised there.
	config.Dot.Aura.OnGain = func(aura *core.Aura, _ *core.Simulation) {
		core.EnableDamageDoneByCaster(DDBC_Wrack, DDBC_Total, warlock.AttackTables[aura.Unit.UnitIndex], amplifyNamed)
	}
	config.Dot.Aura.OnExpire = func(aura *core.Aura, _ *core.Simulation) {
		core.DisableDamageDoneByCaster(DDBC_Wrack, warlock.AttackTables[aura.Unit.UnitIndex])
	}

	warlock.Wrack = warlock.RegisterSpell(config)
}
