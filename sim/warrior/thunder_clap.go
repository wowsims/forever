package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var thunderClapRank = spellData.ThunderClap.Highest()

// The Attack Power share (0.07) is measured in game on Ranks 1 and 4 (Rank 4: floor(0.99 * (55 + 0.07 * AP))).
// TODO: Ranks 2, 3, 5 and 6 carry the same share unmeasured. Verify it on them once logs at their level exist.
var thunderClapBaseDamage = thunderClapRank.DamageEffect().Average()
var thunderClapAPCoeff = thunderClapRank.DamageEffect().APCoeff()
var thunderClapSlow = thunderClapRank.EffectN(2).BaseValue()

// The melee speed factor the warrior's clap leaves on the target, the Conqueror's set included.
func (warrior *Warrior) thunderClapSpeed() float64 {
	return core.SymmetricMultiplierFromPercent(thunderClapSlow * (1 + warrior.thunderClapEffectBonus))
}

func (warrior *Warrior) registerThunderClap() {
	auras := warrior.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		// The clap bids its whole slow in the attack-speed category, so the strongest single slow
		// on the target is the only one applied. The generated aura holds the category at the
		// game files' amount; the clap sets its bid before it lands, and the effect applies what the
		// bid is worth.
		aura := buffs.NewSpellDataThunderClapAura(target, true, 0)
		speed := 1.0
		bid := aura.ExclusiveEffects[0]
		bid.OnGain = func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			speed = 1 - ee.Priority
			ee.Aura.Unit.MultiplyMeleeSpeed(sim, speed)
		}
		bid.OnExpire = func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.MultiplyMeleeSpeed(sim, 1/speed)
		}
		return aura
	})

	// Thunder Clap is Physical but Magic in SpellCategories: it rolls on the spell hit table
	// (logs show full resists next to armor mitigation) and crits on spell crit chance for
	// 1.5x. Warriors have no base spell crit, so logs without Totem of Wrath show none
	// (0 of 799 landed hits from 6 prot warriors on fresh.warcraftlogs.com, 2026-09-14). A magic
	// hit spends no Sweeping Strikes charge, which melee swings and abilities do. Its ranged mask
	// reaches equip procs that hear ranged abilities, and core's canProcFrom decides which weapon procs
	// hear it.
	config := spelldata.SpellConfig(&warrior.Unit, thunderClapRank,
		spelldata.Flags(core.SpellFlagAPL|core.SpellFlagBinary))
	config.ProcMask = core.ProcMaskRangedSpecial
	config.DamageMultiplier = 1
	// Threat is 1 times the damage with no flat bonus, at 1 to 4 targets, Improved Thunder Clap included.
	config.ThreatMultiplier = 1

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := thunderClapBaseDamage + thunderClapAPCoeff*spell.MeleeAttackPower(target)
		results := spell.CalcCleaveDamage(sim, target, int32(thunderClapRank.MaxTargets), baseDamage, spell.OutcomeMagicHitAndCrit)

		for _, result := range results {
			if result.Landed() {
				aura := auras.Get(result.Target)
				if bid := aura.ExclusiveEffects[0]; bid.Priority != 1-warrior.thunderClapSpeed() {
					bid.SetPriority(sim, 1-warrior.thunderClapSpeed())
				}
				aura.Activate(sim)
			}
			spell.DealDamage(sim, result)
		}
	}

	config.RelatedAuraArrays = auras.ToMap()

	warrior.RegisterSpell(config)
}
