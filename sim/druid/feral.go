package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) level() int32 {
	return druid.Level
}

func (druid *Druid) rendAndTearMultiplier(target *core.Unit) float64 {
	if druid.Talents.RendAndTear == 0 || !druid.targetIsBleeding(target) {
		return 1
	}
	// Melee special attacks and bleed ticks against a bleeding target; white swings are not raised. A
	// tester's table on PR #149 shows Rake and Rip ticks, the Rake hit and Claw rising by the talent's
	// percent with a bleed on the target, and the Rake hit staying put without one.
	return 1 + spellData.RendAndTear.Rank(druid.Talents.RendAndTear).EffectN(1).Percent()
}

// The row's weapon strike on the main-hand swing: the flat bonus and the swing, both scaled by the
// weapon percent.
func (druid *Druid) weaponSpecialDamage(sim *core.Simulation, spell *core.Spell, target *core.Unit, rank *spelldata.Spell) float64 {
	weapon := spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))
	return rank.WeaponStrike(sim, spell, weapon, 0) * druid.rendAndTearMultiplier(target)
}

func (druid *Druid) awardComboPoints(sim *core.Simulation, spell *core.Spell, rank *spelldata.Spell) {
	e := rank.EnergizeEffect()
	if e == spelldata.NilEffect || dbcenums.PowerType(e.Misc) != dbcenums.POWER_COMBO_POINTS {
		return
	}
	druid.AddComboPoints(sim, int32(e.BaseValue()), spell.ComboPointMetrics())
}

// Spends every combo point. The row's combo-point cost is the minimum the cast requires; the
// per-resource damage is priced on the points actually spent.
func (druid *Druid) spendComboPoints(sim *core.Simulation, spell *core.Spell) {
	if druid.ComboPoints() > 0 {
		druid.SpendComboPoints(sim, spell.ComboPointMetrics())
	}
}

func (druid *Druid) requiresComboPoints(rank *spelldata.Spell) core.CanCastCondition {
	need := rank.Power(dbcenums.POWER_COMBO_POINTS).Cost
	if need == 0 {
		return nil
	}
	return func(_ *core.Simulation, _ *core.Unit) bool {
		return druid.ComboPoints() >= need
	}
}

// Issues the row's miss refund, which SpellConfig set on the cost, when the hit did not land.
func (druid *Druid) refundOnMiss(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	if !result.Landed() {
		spell.IssueRefund(sim)
	}
}

func (druid *Druid) isComboBuilder(spell *core.Spell) bool {
	return spellData.Claw.Contains(spell) || spellData.Shred.Contains(spell) || spellData.Rake.Contains(spell) ||
		spellData.Ravage.Contains(spell) || spellData.Pounce.Contains(spell)
}

// Whether Shred, Ravage and Pounce can land: the druid is behind the target and the target can be shredded.
func (druid *Druid) BehindTarget() bool {
	return !druid.PseudoStats.InFrontOfTarget && !druid.CannotShredTarget
}

// Whether the target is bleeding, for Rend and Tear and the Bleed Is Active APL value: one of the
// druid's bleeds is up on it, or the druid assumes another raid member's.
func (druid *Druid) targetIsBleeding(target *core.Unit) bool {
	if druid.AssumeBleedActive {
		return true
	}
	for _, bleed := range [...]*DruidSpell{druid.Rake, druid.Rip, druid.Lacerate, druid.PounceBleed} {
		if bleed != nil && bleed.Dot(target).IsActive() {
			return true
		}
	}
	return false
}

func (druid *Druid) breakProwl(sim *core.Simulation) {
	if druid.ProwlAura != nil && druid.ProwlAura.IsActive() {
		druid.ProwlAura.Deactivate(sim)
	}
	if druid.ShadowmeldAura != nil && druid.ShadowmeldAura.IsActive() {
		druid.ShadowmeldAura.Deactivate(sim)
	}
}

func (druid *Druid) registerProwlBreakingCasts() {
	druid.OnSpellRegistered(func(spell *core.Spell) {
		if spell.ApplyEffects == nil || spell.Flags.Matches(core.SpellFlagPassiveSpell|core.SpellFlagProc) {
			return
		}
		if spell.DefaultCast == (core.Cast{}) && spell.CD.Timer == nil && spell.SharedCD.Timer == nil && !spell.Flags.Matches(core.SpellFlagAPL) {
			return
		}
		row := spelldata.Find(spell.SpellID)
		if spell.ItemID != 0 {
			row = itemOnUseRow(spell.ItemID)
		}
		if row.IsPassive() || row.HasAttr(dbcenums.ATTR_INDEX_EX_1, dbcenums.ATTR_EX_1_NO_STEALTH_BREAK) || spellData.Prowl.Contains(spell) || spell.SpellID == shadowmeldSpellID {
			return
		}
		apply := spell.ApplyEffects
		spell.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			druid.breakProwl(sim)
			apply(sim, target, spell)
		}
	})
}

// The row of the spell an item casts when used: a consumable's first on-use, or a worn item's, such
// as a trinket.
func itemOnUseRow(itemID int32) *spelldata.Spell {
	if effects := core.GetConsumableByID(itemID).OnUseEffects(); len(effects) > 0 {
		return spelldata.Find(effects[0].GetBuffId())
	}
	if item := core.GetItemByID(itemID); item != nil {
		for _, effect := range item.ItemEffects {
			if effect.GetOnUse() != nil {
				return spelldata.Find(effect.GetBuffId())
			}
		}
	}
	return spelldata.Nil
}

// The number of targets Primal Bite strikes under Berserk, the primary included:
// the tooltip's "strike up to $s3 targets" is effect 3, a SPELLMOD_JUMP_TARGETS the parser does not
// apply. 0 outside Berserk.
func (druid *Druid) berserkTargets() int32 {
	if druid.BerserkAura == nil || !druid.BerserkAura.IsActive() {
		return 0
	}
	e := spellData.Berserk.Highest().FindEffect(dbcenums.E_APPLY_AURA, dbcenums.A_ADD_FLAT_MODIFIER, int32(dbcenums.SPELLMOD_JUMP_TARGETS))
	if e == spelldata.NilEffect {
		return 0
	}
	return int32(e.BaseValue())
}

// Hits up to count active targets starting from target, prices each hit on its own target, and deals
// them together once all are rolled. It returns the primary target's result. results is the
// caller's own buffer, reused across casts.
func cleave(sim *core.Simulation, target *core.Unit, spell *core.Spell, count int32, results *core.SpellResultSlice, hit func(target *core.Unit) *core.SpellResult) *core.SpellResult {
	*results = (*results)[:0]
	for range min(count, sim.Environment.ActiveTargetCount()) {
		*results = append(*results, hit(target))
		target = sim.Environment.NextActiveTargetUnit(target)
	}
	for _, result := range *results {
		spell.DealDamage(sim, result)
	}
	return (*results)[0]
}
