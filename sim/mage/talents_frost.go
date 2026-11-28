package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (mage *Mage) registerFrostTalents() {
	// Tier 1
	mage.registerFrostWarding()
	mage.registerImprovedFrostbolt()
	mage.registerPrecision()

	// Tier 2
	mage.registerIceShards()
	mage.registerPermafrost()
	mage.registerImprovedFrostNova()
	mage.registerFrostbite()

	// Tier 3
	mage.registerFrostChanneling()
	mage.registerIceLance()
	mage.registerImprovedBlizzard()

	// Tier 4
	mage.registerArcticReach()
	mage.registerIceBlock()
	mage.registerShatter()

	// Tier 5
	mage.registerImprovedConeOfCold()
	mage.registerColdSnapTalent()
	mage.registerFingersOfFrost()

	// Tier 6
	mage.registerWinterChill()

	// Tier 7
	mage.registerIceBarrier()
}

func (mage *Mage) registerColdSnapTalent() {
	if !mage.Talents.ColdSnap {
		return
	}

	mage.registerColdSnapSpell()
}

// "Increases the Armor and resistances granted by your Frost Armor and Ice Armor spells by $s1%, gives
// your Frost Ward a $s2% chance to negate the warded damage spell, and increases the amount absorbed
// by your Ice Barrier and Frost Ward by $s3%."
//
// The armor is applied in armors.go and the absorb in ice_barrier.go and defensives.go. The negate
// needs an enemy casting Frost at the mage, so it is not modelled.
func (mage *Mage) registerFrostWarding() {
	if mage.Talents.FrostWarding == 0 {
		return
	}
}

// Improved Frostbolt: Frostbolt's casting time is reduced by 0.1 sec per point.
func (mage *Mage) registerImprovedFrostbolt() {
	if mage.Talents.ImprovedFrostbolt == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ImprovedFrostbolt.Rank(mage.Talents.ImprovedFrostbolt))
}

// "Improves your chance to hit with all spells by $s1%."
func (mage *Mage) registerPrecision() {
	if mage.Talents.Precision == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.Precision.Rank(mage.Talents.Precision))
}

func (mage *Mage) registerIceShards() {
	if mage.Talents.IceShards == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.IceShards.Rank(mage.Talents.IceShards))
}

// "Increases the duration of your Chill effects by 33% and reduces the target's speed by an
// additional 10%."
//
// Not modelled: both effects act on the slow alone, which does nothing to a boss the sim fights. A
// Chill rolls for Frostbite or Fingers of Frost per hit, not per second chilled.
func (mage *Mage) registerPermafrost() {
	if mage.Talents.Permafrost == 0 {
		return
	}
}

func (mage *Mage) registerImprovedFrostNova() {
	if mage.Talents.ImprovedFrostNova == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ImprovedFrostNova.Rank(mage.Talents.ImprovedFrostNova))
}

// "Gives your Chill effects a 15% chance to Freeze the target for 5 sec."
//
// The freeze (12494) is not modelled: it is a root, which a boss the sim fights is immune to. This
// talent's roll grants Fingers of Frost charges (registerFingersOfFrost).
func (mage *Mage) registerFrostbite() {
	if mage.Talents.Frostbite == 0 {
		return
	}
}

// Reduces the mana cost of your Frost spells by 5/10/15% and their threat by 10/20/30%.
func (mage *Mage) registerFrostChanneling() {
	if mage.Talents.FrostChanneling == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.FrostChanneling.Rank(mage.Talents.FrostChanneling))
}

func (mage *Mage) registerIceLance() {
	if !mage.Talents.IceLance {
		return
	}

	mage.registerIceLanceSpell()
}

// "Adds a Chill effect to your Blizzard spell. This effect lowers the target's movement speed by 40%
// for 1.5 sec."
//
// The slow (12484) does nothing to a boss; the talent makes Blizzard a Chill effect, which isChill reads.
func (mage *Mage) registerImprovedBlizzard() {
	if mage.Talents.ImprovedBlizzard == 0 {
		return
	}
}

// "Increases the range of your Frostbolt, Frost Blast, Ice Lance, and Blizzard spells and the radius of
// your Frost Nova and Cone of Cold spells by 20%." The tooltip names Frost Blast, which the game files
// do not ship.
func (mage *Mage) registerArcticReach() {
	if mage.Talents.ArcticReach == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ArcticReach.Rank(mage.Talents.ArcticReach))
}

func (mage *Mage) registerIceBlock() {
	if !mage.Talents.IceBlock {
		return
	}

	mage.registerIceBlockSpell()
}

// "Increases the critical strike chance of all your spells against Frozen targets by 50%."
//
// Fingers of Frost's charges give Blizzard's ticks no bonus (EV-0029). They are the only way a target is
// Frozen here, so the ticks are left out, and Flamestrike's ground ticks with them.
func (mage *Mage) registerShatter() {
	if mage.Talents.Shatter == 0 {
		return
	}

	critBonus := spellData.Shatter.ValueAt(mage.Talents.Shatter)

	mage.addTargetCritBonus(
		func(spell *core.Spell) bool {
			return spell.ProcMask.Matches(core.ProcMaskSpellDamage) && !spell.Matches(MageSpellBlizzard|MageSpellFlamestrikeDot)
		},
		func(target *core.Unit) float64 {
			if mage.isFrozen(target) {
				return critBonus
			}
			return 0
		})
}

func (mage *Mage) registerImprovedConeOfCold() {
	if mage.Talents.ImprovedConeOfCold == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ImprovedConeOfCold.Rank(mage.Talents.ImprovedConeOfCold))
}

// "Gives your Chill effects a 15% chance to grant you the Fingers of Frost effect, which treats your
// next 2 spells cast as if the target were Frozen. Lasts 15 sec."
//
// The buff's row (400669) names only Deep Freeze, which is not in Forever, so what "as if Frozen" reaches
// is the tooltip's: every spell reading isFrozen, Shatter and Ice Lance.
func (mage *Mage) registerFingersOfFrost() {
	if mage.Talents.FingersOfFrost == 0 {
		return
	}

	talentRank := spellData.FingersOfFrost.Rank(mage.Talents.FingersOfFrost)
	charges := int32(talentRank.EffectN(1).BaseValue())

	config := spelldata.AuraConfig(spellData.FingersOfFrostTriggered.Highest())
	config.MaxStacks = charges

	// A cast that granted the charges from inside its own effects, which spends none of them.
	var grantedBy *core.Spell
	// Counts the grants, so a spend can tell whether charges were granted after it was cast.
	var grants int

	// Charge spending in EV-0030. The charge goes a spell batch after the cast: a spell queued in that
	// batch spends the same charge, and a grant inside that batch leaves the charges full.
	config.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if spell == grantedBy {
			grantedBy = nil
			return
		}
		if spell.Flags.Matches(core.SpellFlagAPL) && !spell.Flags.Matches(core.SpellFlagChanneled) &&
			spell.ProcMask.Matches(core.ProcMaskSpellDamage) {
			grantsAtCast := grants
			pa := sim.GetConsumedPendingActionFromPool()
			pa.NextActionAt = sim.CurrentTime + core.SpellBatchWindow
			pa.Priority = core.ActionPriorityDOT
			pa.OnAction = func(sim *core.Simulation) {
				if grants == grantsAtCast {
					aura.RemoveStack(sim)
				}
			}
			sim.AddPendingAction(pa)
		}
	}
	// Each Arcane Missiles missile that lands spends a charge, and a miss keeps it. The charge goes on
	// the cast result, so the roll that read the target as Frozen spends it, whatever the travel time.
	config.OnCastResult = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		if spell.Matches(MageSpellArcaneMissilesTick) && result.Landed() {
			aura.RemoveStack(sim)
		}
	}

	mage.FingersOfFrostAura = mage.RegisterAura(config)

	grant := func(atCast bool) core.ProcHandler {
		return func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			// Charges granted inside a rotation spell's own cast spend none on it.
			if spell.Flags.Matches(core.SpellFlagAPL) && (atCast || spell.MissileSpeed == 0) {
				grantedBy = spell
			}
			mage.FingersOfFrostAura.Activate(sim)
			mage.FingersOfFrostAura.SetStacks(sim, charges)
			grants++
		}
	}

	// The talent's own roll, on the landed hit, with the charges up at once.
	trigger := spelldata.ProcTrigger(&mage.Character, talentRank, grant(false))
	trigger.Name = "Fingers of Frost - Trigger"
	trigger.Callback = core.CallbackOnSpellHitDealt
	trigger.Outcome = core.OutcomeLanded
	trigger.TriggerImmediately = true
	trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		return mage.isChill(spell)
	}

	// Manaflare Regalia's 5pc adds effect 3's points to Frostfire Bolt's chance: an independent roll on the
	// landed hit, sized so the two grant at their sum.
	if mage.manaflareAura != nil {
		manaflareTrigger := trigger
		manaflareTrigger.Name = "Fingers of Frost - Manaflare Trigger"
		manaflareTrigger.ProcChance = manaflareFrostfireBolt.EffectN(3).Percent() / (1 - trigger.ProcChance)
		manaflareTrigger.ClassSpellMask = MageSpellFrostfireBolt
		manaflareTrigger.ExtraCondition = nil
		mage.manaflareAura.AttachProcTrigger(manaflareTrigger)
	}

	// With Frostbite a Chill rolls twice. Frostbite's roll grants the charges at its rank's chance whether
	// or not the freeze lands, on the cast result, so a missile grants at its cast. The talent's own roll
	// keeps what is left of its chance, so a Chill still grants at least once at that chance.
	if mage.Talents.Frostbite > 0 {
		chance := trigger.ProcChance
		freezeChance := spellData.Frostbite.FractionAt(mage.Talents.Frostbite)

		frostbiteTrigger := trigger
		frostbiteTrigger.Name = "Fingers of Frost - Frostbite Trigger"
		frostbiteTrigger.Callback = core.CallbackOnCastResult
		frostbiteTrigger.ProcChance = freezeChance
		frostbiteTrigger.Handler = grant(true)
		mage.MakeProcTriggerAura(frostbiteTrigger)

		trigger.ProcChance = (chance - freezeChance) / (1 - freezeChance)
	}

	// A trigger with no chance would fire on every hit.
	if trigger.ProcChance > 1e-9 {
		mage.MakeProcTriggerAura(trigger)
	}
}

// "Gives your Frostbolt, Frost Blast, and Fingers of Frost empowered Ice Lance spells a 20/40/60/80/100%
// chance to apply Winter's Chill, increasing the target's damage taken from your next Cone of Cold by
// 20% and it gains an additional 7% of your spell damage, stacking up to 5 times."
//
// The tooltip names Frost Blast, which the game files do not ship. The talent (11180) names no class
// mask, so the spells are checked here.
func (mage *Mage) registerWinterChill() {
	if mage.Talents.WintersChill == 0 {
		return
	}

	talentRank := spellData.WintersChill.Rank(mage.Talents.WintersChill)
	debuff := spellData.WintersChillTriggered.Highest()
	coneFlags := debuff.EffectN(1).ClassFlags

	mage.WintersChillAuras = mage.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		config := spelldata.AuraConfig(debuff, spelldata.Label(debuff.Name+mage.Label))
		config.MaxStacks = int32(talentRank.EffectN(1).BaseValue())
		aura := target.RegisterAura(config)
		spelldata.ParseEffects(nil, aura, debuff, spelldata.FromCaster(&mage.Unit, DDBC_WintersChill, DDBC_Total))
		return aura
	})

	mage.Env.RegisterPreFinalizeEffect(func() {
		for _, spell := range mage.Spellbook {
			if spell.MatchesFlags(coneFlags) {
				spell.RelatedAuraArrays = mage.WintersChillAuras.ToMap()
			}
		}
	})

	addStack := func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		aura := mage.WintersChillAuras.Get(result.Target)
		aura.Activate(sim)
		aura.AddStack(sim)
	}

	// PROC_FLAG_DEAL_HARMFUL_SPELL decodes to the cast, which names no target; the debuff goes on
	// the one the bolt landed on, when it lands.
	trigger := spelldata.ProcTrigger(&mage.Character, talentRank, addStack)
	trigger.Name = "Winter's Chill - Trigger"
	trigger.Callback = core.CallbackOnSpellHitDealt
	trigger.Outcome = core.OutcomeLanded
	trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		return spellData.Frostbolt.Contains(spell)
	}
	mage.MakeProcTriggerAura(trigger)

	// Ice Lance rolls on its cast result, where its damage read the target as Frozen: Fingers of Frost
	// may have no charge left by the time the lance lands.
	iceLanceTrigger := trigger
	iceLanceTrigger.Name = "Winter's Chill - Ice Lance Trigger"
	iceLanceTrigger.Callback = core.CallbackOnCastResult
	iceLanceTrigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
		return spellData.IceLance.Contains(spell) && mage.isFrozen(result.Target)
	}
	mage.MakeProcTriggerAura(iceLanceTrigger)
}

// The spell power coefficient Winter's Chill adds to this mage's Cone of Cold on the target: the talent's
// "additional 7% of your spell damage" per stack, a share added to the coefficient.
func (mage *Mage) wintersChillCoefficient(target *core.Unit) float64 {
	stacks := mage.WintersChillAuras.Get(target).GetStacks()
	return float64(stacks) * spellData.WintersChillTriggered.Highest().EffectN(2).Percent()
}

// Cone of Cold spends every stack on a target it lands on.
func (mage *Mage) consumeWintersChill(sim *core.Simulation, target *core.Unit) {
	mage.WintersChillAuras.Get(target).Deactivate(sim)
}

func (mage *Mage) registerIceBarrier() {
	if !mage.Talents.IceBarrier {
		return
	}

	mage.registerIceBarrierSpell()
}

// The Chill effects Permafrost's mask names: Frostbolt, Cone of Cold, Frostfire Bolt and the armors' Chilled.
var chillFlags = spellData.Permafrost.Highest().EffectN(1).ClassFlags

// Whether the spell applies a Chill effect: one the Chill mask names, or Blizzard with Improved
// Blizzard, which adds a Chill the mask does not carry.
func (mage *Mage) isChill(spell *core.Spell) bool {
	if spell.MatchesFlags(chillFlags) {
		return true
	}
	return mage.Talents.ImprovedBlizzard > 0 && spell.Matches(MageSpellBlizzard)
}

// Whether this mage's spells treat the target as Frozen. A boss the sim fights is immune to Frost Nova's
// and Frostbite's roots, so the only way there is Fingers of Frost, which does not look at the target.
func (mage *Mage) isFrozen(_ *core.Unit) bool {
	return mage.FingersOfFrostAura.IsActive()
}

// Raises matching spells' crit by bonus for the target during their effects, where the crit rolls whatever
// the travel time; Shatter's bonus is a dummy effect with no aura on the target to parse onto.
func (mage *Mage) addTargetCritBonus(matches func(*core.Spell) bool, bonus func(target *core.Unit) float64) {
	mage.Env.RegisterPreFinalizeEffect(func() {
		for _, spell := range mage.Spellbook {
			if spell.ApplyEffects == nil || !matches(spell) {
				continue
			}

			applyEffects := spell.ApplyEffects
			spell.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				critBonus := bonus(target)
				spell.BonusCritPercent += critBonus
				applyEffects(sim, target, spell)
				spell.BonusCritPercent -= critBonus
			}
		}
	})
}
