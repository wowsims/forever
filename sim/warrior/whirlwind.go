package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var whirlwindRank = spellData.Whirlwind.Highest()

func (warrior *Warrior) registerWhirlwind() {
	// The tooltip of 1680 strikes with both melee weapons, and the row states only the one weapon
	// damage effect, so the off-hand strike is a sub-spell on the same row.
	ohConfig := spelldata.SpellConfig(&warrior.Unit, whirlwindRank,
		spelldata.Melee(core.ProcMaskMeleeOHSpecial), spelldata.Proc())
	ohConfig.ActionID = ohConfig.ActionID.WithHand(proto.ItemSlot_ItemSlotOffHand)

	ohConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := warrior.OHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
		spell.CalcCleaveDamage(sim, target, int32(whirlwindRank.MaxTargets), baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		spell.DealBatchedAoeDamage(sim)
	}

	whirlwindOH := warrior.RegisterSpell(ohConfig)

	// TODO: In-game testing required for threat multiplier / threat bonus.
	// The cast is the main-hand strike, so a rotation casts the main-hand copy.
	config := spelldata.SpellConfig(&warrior.Unit, whirlwindRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ActionID = config.ActionID.WithHand(proto.ItemSlot_ItemSlotMainHand)

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := warrior.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
		results := spell.CalcCleaveDamage(sim, target, int32(whirlwindRank.MaxTargets), baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		warrior.CastNormalizedSweepingStrikesAttack(results, sim)
		spell.DealBatchedAoeDamage(sim)

		if warrior.HasOHWeapon() {
			whirlwindOH.Cast(sim, target)
		}
	}

	warrior.RegisterSpell(config)
}
