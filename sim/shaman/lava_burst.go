package shaman

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var lavaBurstRank = spellData.LavaBurst.Highest()

// Lightning Overload names Lightning Bolt and Chain Lightning only, so Lava Burst has no overload.
func (shaman *Shaman) registerLavaBurstSpell() {
	config := spelldata.SpellConfig(&shaman.Unit, lavaBurstRank,
		spelldata.Magic(core.ProcMaskSpellDamage),
		spelldata.Flags(SpellFlagShamanSpell|SpellFlagFocusable))

	// "If your Flame Shock is on the target, Lava Burst deals $s2% increased damage."
	damage := lavaBurstRank.DamageEffect()
	flameShockBonus := 1 + lavaBurstRank.EffectN(2).Average()/100

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		multiplier := 1.0
		if shaman.FlameShock != nil && shaman.FlameShock.Dot(target).IsActive() {
			multiplier = flameShockBonus
		}

		spell.DamageMultiplier *= multiplier
		result := spell.CalcDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
		spell.DamageMultiplier /= multiplier

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
		})
	}

	shaman.LavaBurst = shaman.RegisterSpell(config)
}
