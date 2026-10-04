package shaman

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The tag of the copy Lightning Overload casts, which shares its parent's spell id.
const CastTagLightningOverload int32 = 6

// Lightning Bolt and Chain Lightning, and the copy of either that Lightning Overload casts: the same
// row, with no cast or cost, for "half damage and no threat" as the tooltip of 408438 reads.
func (shaman *Shaman) newElectricSpellConfig(rank *spelldata.Spell, isOverload bool) core.SpellConfig {
	opts := []spelldata.SpellOpt{
		spelldata.Magic(core.ProcMaskSpellDamage),
		spelldata.Flags(SpellFlagShamanSpell | SpellFlagFocusable),
	}
	if isOverload {
		opts = append(opts, spelldata.Proc(), spelldata.Tag(CastTagLightningOverload))
	}

	config := spelldata.SpellConfig(&shaman.Unit, rank, opts...)

	if isOverload {
		config.DamageMultiplier = 0.5
		config.ThreatMultiplier = 0
		return config
	}

	// A cast holds the melee swing until it finishes. Core applies the cast speed after this.
	config.Cast.ModifyCast = func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
		castTime := shaman.ApplyCastSpeedForSpell(cast.CastTime, spell)
		if sim.CurrentTime+castTime > shaman.AutoAttacks.NextAttackAt() {
			shaman.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+castTime)
		}
	}

	return config
}
