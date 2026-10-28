package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (hunter *Hunter) registerSurvivalTalents() {
	// Tier 1
	hunter.registerImprovedTracking()
	hunter.registerDeflection()

	// Tier 2
	// Entrapment (19184) roots the targets a trap springs on, and the sim's targets do not move.
	hunter.registerSavageStrikes()
	hunter.registerSurvivalist()
	// Improved Wing Clip (19228) is a root chance on Wing Clip, which a boss shrugs off.

	// Tier 3
	hunter.registerCleverTraps()
	hunter.registerSurefooted()
	// Deterrence (19263) is parry and dodge the sim's targets do not test.

	// Tier 4
	hunter.registerSurvivalTactics()
	hunter.registerPredatorsEdge()
	// Counterattack (19306) needs a parry, which takes a target attacking the hunter.

	// Tier 5
	hunter.registerResourcefulness()
	// Expose Prey: mongoose_bite.go
	hunter.registerSurvivalistsDiscipline()
	hunter.registerStriderKick()

	// Tier 6
	hunter.registerLightningReflexes()

	// Tier 7
	hunter.registerLaceratingStrikes()
}

// The talent raises effect 2 of the Track spells, A_MOD_DAMAGE_DONE_VERSUS the tracked creature
// type, by 1% a rank. The sim does not track; a hunter is taken to track the target's type, so the
// bonus is on every damage done.
func (hunter *Hunter) registerImprovedTracking() {
	if hunter.Talents.ImprovedTracking == 0 {
		return
	}

	hunter.PseudoStats.DamageDealtMultiplier *= spellData.ImprovedTracking.MultiplierAt(hunter.Talents.ImprovedTracking)
}

func (hunter *Hunter) registerDeflection() {
	if hunter.Talents.Deflection == 0 {
		return
	}

	spelldata.ParseStatic(&hunter.Character, spellData.Deflection.Rank(hunter.Talents.Deflection))
}

func (hunter *Hunter) registerSavageStrikes() {
	if hunter.Talents.SavageStrikes == 0 {
		return
	}

	spelldata.ParseStatic(&hunter.Character, spellData.SavageStrikes.Rank(hunter.Talents.SavageStrikes))
}

func (hunter *Hunter) registerSurvivalist() {
	if hunter.Talents.Survivalist == 0 {
		return
	}

	spelldata.ParseStatic(&hunter.Character, spellData.Survivalist.Rank(hunter.Talents.Survivalist))
}

func (hunter *Hunter) registerSurefooted() {
	if hunter.Talents.Surefooted == 0 {
		return
	}

	spelldata.ParseStatic(&hunter.Character, spellData.Surefooted.Rank(hunter.Talents.Surefooted))
}

func (hunter *Hunter) registerPredatorsEdge() {
	if hunter.Talents.PredatorsEdge == 0 {
		return
	}

	spelldata.ParseStatic(&hunter.Character, spellData.PredatorsEdge.Rank(hunter.Talents.PredatorsEdge))
}

var resourcefulnessBuff = spellData.ResourcefulnessTriggered.Highest()

func (hunter *Hunter) registerResourcefulness() {
	if hunter.Talents.Resourcefulness == 0 {
		return
	}

	rank := spellData.Resourcefulness.Rank(hunter.Talents.Resourcefulness)
	spelldata.ParseStatic(&hunter.Character, rank, spelldata.Effects(1))

	buff := hunter.RegisterAura(spelldata.AuraConfig(resourcefulnessBuff))
	spelldata.ParseEffects(&hunter.Character, buff, resourcefulnessBuff)

	// The row states the chance on effect 2. Its proc flags (0x110) state melee and ranged abilities
	// and no crit; the tooltip limits the trigger to those abilities' crits.
	trigger := spelldata.ProcTrigger(&hunter.Character, rank,
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			buff.Activate(sim)
		})
	trigger.Name = "Resourcefulness - Trigger"
	trigger.Outcome = core.OutcomeCrit

	hunter.MakeProcTriggerAura(trigger)
}

// The cooldown mod matches every trap's placement spell and Deterrence; the sim registers only the
// Immolation and Explosive Trap placements, and their shared category cooldown shortens with them.
func (hunter *Hunter) registerSurvivalistsDiscipline() {
	if hunter.Talents.SurvivalistsDiscipline == 0 {
		return
	}

	spelldata.ParseStatic(&hunter.Character, spellData.SurvivalistsDiscipline.Rank(hunter.Talents.SurvivalistsDiscipline))
}

var striderKickRank = spellData.StriderKick.Highest()

// Effect 1 is normalised weapon damage with no bonus; effect 2 is the share of it, 100%.
var striderKickWeaponShare = striderKickRank.EffectN(2).Percent()

func (hunter *Hunter) registerStriderKick() {
	if !hunter.Talents.StriderKick {
		return
	}

	config := spelldata.SpellConfig(&hunter.Unit, striderKickRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	// Effect 3 is the hunter's own movement speed for the row's duration.
	// TODO: In-game testing required. The speed is taken to come with a kick that misses too.
	speed := hunter.RegisterAura(spelldata.AuraConfig(striderKickRank))
	speed.NewActiveMovementSpeedEffect(striderKickRank.EffectN(3).Percent())

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := striderKickWeaponShare * hunter.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		speed.Activate(sim)
	}

	hunter.RegisterSpell(config)
}

func (hunter *Hunter) registerLightningReflexes() {
	if hunter.Talents.LightningReflexes == 0 {
		return
	}

	spelldata.ParseStatic(&hunter.Character, spellData.LightningReflexes.Rank(hunter.Talents.LightningReflexes))
}

var laceratingStrikesBleed = spellData.LaceratingStrikesTriggered.Highest()

func (hunter *Hunter) registerLaceratingStrikes() {
	if !hunter.Talents.LaceratingStrikes {
		return
	}

	// The talent's dummy holds 40: the share of the bite's damage the bleed deals over its duration.
	share := spellData.LaceratingStrikes.FractionAt(1)
	tick := laceratingStrikesBleed.PeriodicEffect()

	config := spelldata.SpellConfig(&hunter.Unit, laceratingStrikesBleed, spelldata.Proc())
	config.ProcMask = core.ProcMaskEmpty
	// A share of the damage the bite dealt, which has every modifier on it already.
	config.Flags |= core.SpellFlagIgnoreModifiers

	// The tick is a share of the bite that landed, not the row's amount: 1310536's periodic effect
	// states one base point, so the amount is snapshot from the bite and the resolver's tick replaced.
	var bleedDamage float64
	config.Dot = spelldata.DotConfig(laceratingStrikesBleed, tick)
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, dot.SnapshotBaseDamage, laceratingStrikesBleed.TickOutcome(dot))
	}

	// The refresh Deep Wounds has: a bite while the bleed is up adds the unticked damage to the new
	// application and spreads the sum over a full set of ticks; the ticks keep the first application's timer.
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		dot := spell.Dot(target)
		totalDamage := dot.OutstandingDmg() + bleedDamage

		if dot.IsActive() {
			dot.DurationExtend(sim, laceratingStrikesBleed.Duration())
		} else {
			dot.Apply(sim)
		}
		dot.SnapshotBaseDamage = totalDamage / float64(dot.RemainingTicks())
	}

	bleed := hunter.RegisterSpell(config)

	// The proc shape with no roll: 1310533 fires on a melee ability and names Mongoose Bite only in
	// its tooltip, so the listener is narrowed to it by hand. Raptor Strike carries the bite's class
	// flags, so the flags alone would let it start the bleed too.
	trigger := spelldata.ProcTrigger(&hunter.Character, spellData.LaceratingStrikes.Rank(1),
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			bleedDamage = result.Damage * share
			bleed.Cast(sim, result.Target)
		})
	trigger.Name = "Lacerating Strikes - Trigger"
	trigger.ClassFlags = mongooseBiteRank.ClassFlags
	trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		return spell.ActionID.SpellID == mongooseBiteRank.ID
	}

	hunter.MakeProcTriggerAura(trigger)
}

// Effect 1 lengthens Freezing and Frost Trap, which the sim does not place; effect 2 raises the
// Immolation and Explosive Trap damage.
func (hunter *Hunter) registerCleverTraps() {
	if hunter.Talents.CleverTraps == 0 {
		return
	}

	spelldata.ParseStatic(&hunter.Character, spellData.CleverTraps.Rank(hunter.Talents.CleverTraps), spelldata.Effects(2))
}

// Hit with Immolation and Explosive Trap. The talent also names Freezing Trap, Frost Trap and Feign
// Death, which the sim does not cast.
func (hunter *Hunter) registerSurvivalTactics() {
	if hunter.Talents.SurvivalTactics == 0 {
		return
	}

	spelldata.ParseStatic(&hunter.Character, spellData.SurvivalTactics.Rank(hunter.Talents.SurvivalTactics))
}
