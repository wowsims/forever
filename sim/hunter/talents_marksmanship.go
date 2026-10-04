package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func (hunter *Hunter) registerMarksmanshipTalents() {
	// Tier 1
	hunter.registerHawkEye()
	// Improved Concussive Shot (19407) is a stun chance on a slow the sim does not cast.
	hunter.registerLethalAttacks()

	// Tier 2
	hunter.registerImprovedStings()
	hunter.registerEfficiency()
	hunter.registerCarefulAim()

	// Tier 3
	hunter.registerRapidKilling()
	hunter.registerImprovedArcaneShot()
	hunter.registerLoneWolf()

	// Tier 4
	// Trueshot Aura: a party buff, hunter.go
	hunter.registerMortalShots()
	hunter.registerImprovedSerpentSting()

	// Tier 5
	hunter.registerRapidRecuperation()
	hunter.registerBarrage()
	// Scatter Shot (19503) is a disorient the sim does not cast.

	// Tier 6
	hunter.registerRangedWeaponSpecialization()

	// Tier 7
	hunter.registerSniperShot()
}

func (hunter *Hunter) registerHawkEye() {
	if hunter.Talents.HawkEye == 0 {
		return
	}

	spelldata.ParseStatic(&hunter.Character, spellData.HawkEye.Rank(hunter.Talents.HawkEye))
}

func (hunter *Hunter) registerLethalAttacks() {
	if hunter.Talents.LethalAttacks == 0 {
		return
	}

	spelldata.ParseStatic(&hunter.Character, spellData.LethalAttacks.Rank(hunter.Talents.LethalAttacks))
}

func (hunter *Hunter) registerImprovedStings() {
	if hunter.Talents.ImprovedStings == 0 {
		return
	}

	spelldata.ParseStatic(&hunter.Character, spellData.ImprovedStings.Rank(hunter.Talents.ImprovedStings))
}

func (hunter *Hunter) registerEfficiency() {
	if hunter.Talents.Efficiency == 0 {
		return
	}

	spelldata.ParseStatic(&hunter.Character, spellData.Efficiency.Rank(hunter.Talents.Efficiency))
}

// The two effects are auras the parse table has no row for, each stating a share of Intellect - 20%
// a rank - as attack power and as ranged attack power.
func (hunter *Hunter) registerCarefulAim() {
	if hunter.Talents.CarefulAim == 0 {
		return
	}

	rank := spellData.CarefulAim.Rank(hunter.Talents.CarefulAim)
	hunter.AddStatDependency(stats.Intellect, stats.AttackPower, rank.EffectN(1).Percent())
	hunter.AddStatDependency(stats.Intellect, stats.RangedAttackPower, rank.EffectN(2).Percent())
}

// Effect 2 is the damage bonus the next shot takes after a kill, which the sim does not model.
func (hunter *Hunter) registerRapidKilling() {
	if hunter.Talents.RapidKilling == 0 {
		return
	}

	spelldata.ParseStatic(&hunter.Character, spellData.RapidKilling.Rank(hunter.Talents.RapidKilling),
		spelldata.Effects(1))
}

func (hunter *Hunter) registerImprovedArcaneShot() {
	if hunter.Talents.ImprovedArcaneShot == 0 {
		return
	}

	spelldata.ParseStatic(&hunter.Character, spellData.ImprovedArcaneShot.Rank(hunter.Talents.ImprovedArcaneShot))
}

// Effect 2 is a dummy restating the bonus. The row states a stack of one, which would keep the
// parse off a pseudo-stat multiplier.
func (hunter *Hunter) registerLoneWolf() {
	if !hunter.Talents.LoneWolf || hunter.Pet != nil {
		return
	}

	spelldata.ParseStatic(&hunter.Character, spellData.LoneWolf.Rank(1), spelldata.Effects(1), spelldata.IgnoreStacks())
}

func (hunter *Hunter) registerMortalShots() {
	if hunter.Talents.MortalShots == 0 {
		return
	}

	spelldata.ParseStatic(&hunter.Character, spellData.MortalShots.Rank(hunter.Talents.MortalShots))
}

func (hunter *Hunter) registerImprovedSerpentSting() {
	if hunter.Talents.ImprovedSerpentSting == 0 {
		return
	}

	spelldata.ParseStatic(&hunter.Character, spellData.ImprovedSerpentSting.Rank(hunter.Talents.ImprovedSerpentSting))
}

var rapidRecuperationBuff = spellData.RapidRecuperationTriggered.Highest()

func (hunter *Hunter) registerRapidRecuperation() {
	if hunter.Talents.RapidRecuperation == 0 {
		return
	}

	// The regeneration share is the talent's own ladder - 25/50 by rank - where the buff row states a
	// flat 50, so the row is scaled by the ranks taken.
	buff := hunter.RegisterAura(spelldata.AuraConfig(rapidRecuperationBuff))
	spelldata.ParseEffects(&hunter.Character, buff, rapidRecuperationBuff,
		spelldata.ScaledByRanks(hunter.Talents.RapidRecuperation, 2))

	// The proc shape with no roll: 1223987 fires on a ranged ability and names Serpent Sting only
	// in its tooltip, so the listener is narrowed to it by hand. The sting's application carries no
	// proc mask and deals no damage of its own, which the listener has to accept. The Rapid Killing
	// consumption the tooltip also names is not modelled.
	trigger := spelldata.ProcTrigger(&hunter.Character, spellData.RapidRecuperation.Rank(hunter.Talents.RapidRecuperation),
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			buff.Activate(sim)
		})
	trigger.Name = "Rapid Recuperation - Trigger"
	trigger.ProcMask = core.ProcMaskEmpty
	trigger.ClassFlags = serpentStingRank.ClassFlags
	trigger.RequireDamageDealt = false

	hunter.MakeProcTriggerAura(trigger)
}

// Effect 2 states the same bonus on the ticks of every hunter spell, with no mask. The tooltip
// names Multi-Shot, Volley and Aimed Shot, which effect 1 masks, so effect 2 is left out.
func (hunter *Hunter) registerBarrage() {
	if hunter.Talents.Barrage == 0 {
		return
	}

	spelldata.ParseStatic(&hunter.Character, spellData.Barrage.Rank(hunter.Talents.Barrage), spelldata.Effects(1))
}

func (hunter *Hunter) registerRangedWeaponSpecialization() {
	if hunter.Talents.RangedWeaponSpecialization == 0 {
		return
	}

	spelldata.ParseStatic(&hunter.Character,
		spellData.RangedWeaponSpecialization.Rank(hunter.Talents.RangedWeaponSpecialization))
}

var sniperShotRank = spellData.SniperShot.Highest()

func (hunter *Hunter) registerSniperShot() {
	if !hunter.Talents.SniperShot {
		return
	}

	// The buff is the row's own aura: its duration, its charges and the range its second effect adds
	// to the Shots its class flags name, Auto Shot among them. Sniper Shot is not one of them.
	buff := hunter.RegisterAura(spelldata.AuraConfig(sniperShotRank))
	spelldata.ParseEffects(&hunter.Character, buff, sniperShotRank)

	// One charge per Shot the range mod names, however many targets it hits. The row's proc flags hear
	// every ranged hit, Sniper Shot's own among them, so the listener takes the mod's class flags.
	// TODO: Ingame test needed: a Shot that misses keeps the charge (the row's proc flags ask for damage
	// dealt), and Multi-Shot spends one charge per cast.
	spelldata.AttachChargeSpender(&hunter.Character, buff, sniperShotRank, "Sniper Shot - Spend")

	config := spelldata.SpellConfig(&hunter.Unit, sniperShotRank, spelldata.Melee(core.ProcMaskRangedSpecial))
	hunter.hasteRangedCast(&config)
	config.RelatedSelfBuff = buff

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		buff.Activate(sim)
		buff.SetStacks(sim, buff.MaxStacks)

		baseDamage := sniperShotRank.WeaponStrike(sim, spell, hunter.RangedNormalizedWeaponDamage(sim, spell.RangedAttackPower(target)), 0)
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeRangedHitAndCrit)

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
		})
	}

	hunter.RegisterSpell(config)
}
