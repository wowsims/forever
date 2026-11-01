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
	// Trueshot Aura: a raid buff, hunter.go
	hunter.registerMortalShots()

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

var rapidKillingBuff = spellData.RapidKillingTriggered.Highest()

// Effect 1 shortens Rapid Fire's cooldown. Effect 2 is the damage bonus 415407 gives the next Shot,
// 10/20% by rank, where the buff row states a flat 20, so the row is scaled by the ranks taken. The
// talent row hears the killing blow, and a credited kill grants the buff. The tooltip also grants it
// when the enemy dies with the hunter's Serpent Sting on it, which the row's proc flags do not state;
// a stung enemy grants it once, whether or not the raid gets the kill.
//
// The Shots its damage mod names spend its one charge at the cast's result, a miss included.
func (hunter *Hunter) registerRapidKilling() {
	if hunter.Talents.RapidKilling == 0 {
		return
	}

	rank := spellData.RapidKilling.Rank(hunter.Talents.RapidKilling)
	spelldata.ParseStatic(&hunter.Character, rank, spelldata.Effects(1))

	buff := hunter.RegisterAura(spelldata.AuraConfig(rapidKillingBuff))
	spelldata.ParseEffects(&hunter.Character, buff, rapidKillingBuff,
		spelldata.ScaledByRanks(hunter.Talents.RapidKilling, 2), spelldata.IgnoreStacks())
	buff.ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
		aura.SetStacks(sim, aura.MaxStacks)
	})

	spelldata.AttachChargeSpender(&hunter.Character, buff, rapidKillingBuff, "Rapid Killing - Spend",
		func(_ *core.Character, trigger *core.ProcTrigger) {
			trigger.Callback = core.CallbackOnCastResult
			trigger.Outcome = core.OutcomeEmpty
			trigger.RequireDamageDealt = false
			trigger.RequireDamageComponent = false
			trigger.TriggerImmediately = true

			spend := trigger.Handler
			trigger.Handler = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				spend(sim, spell, result)
				hunter.gainRapidRecuperation(sim, hunter.rapidRecuperationSpend)
			}
		})

	grant := func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		buff.Activate(sim)
	}
	// Serpent Sting registers after the talents. Both checks run while the dying target still carries
	// its dots.
	stung := func(_ *core.Simulation, _ *core.Spell, result *core.SpellResult) bool {
		return hunter.GetSpell(core.ActionID{SpellID: serpentStingRank.ID}).Dot(result.Target).IsActive()
	}

	trigger := spelldata.ProcTrigger(&hunter.Character, rank, grant)
	trigger.Name = "Rapid Killing - Trigger"
	trigger.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
		return !stung(sim, spell, result)
	}
	hunter.MakeProcTriggerAura(trigger)

	hunter.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Rapid Killing - Serpent Sting",
		Callback:       core.CallbackOnTargetDies,
		ExtraCondition: stung,
		Handler:        grant,
	})
}

func (hunter *Hunter) registerImprovedArcaneShot() {
	if hunter.Talents.ImprovedArcaneShot == 0 {
		return
	}

	spelldata.ParseStatic(&hunter.Character, spellData.ImprovedArcaneShot.Rank(hunter.Talents.ImprovedArcaneShot))
}

// Effect 2 is a dummy restating the bonus. The row states a stack of one, which would keep the
// parse off a pseudo-stat multiplier. With a pet out the server takes the 20% back off both units,
// so the hunter nets 1 and the pet, which never had it, is divided by 1.2: a level 30 Cat at 1.3 s
// and 152 attack power reads 36-48 against the model's 43.7-57.8, where the hidden pet row's flat
// -20 (415372) would read 35-46.
func (hunter *Hunter) registerLoneWolf() {
	if !hunter.Talents.LoneWolf {
		return
	}

	if hunter.Pet != nil {
		hunter.Pet.PseudoStats.DamageDealtMultiplier /= spellData.LoneWolf.EffectAt(1).MultiplierAt(1)
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

var rapidRecuperationBuff = spellData.RapidRecuperationTriggered.Highest()

func (hunter *Hunter) registerRapidRecuperation() {
	if hunter.Talents.RapidRecuperation == 0 {
		return
	}

	// The buff row states a flat 50; the talent's effects set the share, see gainRapidRecuperation.
	rank := spellData.RapidRecuperation.Rank(hunter.Talents.RapidRecuperation)
	hit := rank.EffectN(1).Percent()
	hunter.rapidRecuperationSpend = rank.EffectN(2).Percent()
	buff := hunter.RegisterAura(spelldata.AuraConfig(rapidRecuperationBuff))
	buff.ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
		aura.Unit.PseudoStats.SpiritRegenRateCasting += hunter.rapidRecuperationShare
		aura.Unit.UpdateManaRegenRates(sim)
	}).ApplyOnExpire(func(aura *core.Aura, sim *core.Simulation) {
		aura.Unit.PseudoStats.SpiritRegenRateCasting -= hunter.rapidRecuperationShare
		aura.Unit.UpdateManaRegenRates(sim)
	})
	hunter.rapidRecuperationAura = buff

	// The proc shape with no roll: 1223987 fires on a ranged ability and names Serpent Sting only
	// in its tooltip, so the listener is narrowed to it by hand. The sting's application deals no
	// damage of its own, which the listener has to accept.
	trigger := spelldata.ProcTrigger(&hunter.Character, rank,
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			hunter.gainRapidRecuperation(sim, hit)
		})
	trigger.Name = "Rapid Recuperation - Trigger"
	trigger.ClassFlags = serpentStingRank.ClassFlags
	trigger.RequireDamageDealt = false

	hunter.MakeProcTriggerAura(trigger)
}

// Puts Rapid Recuperation up at the share its source grants: effect 1 for a Serpent Sting hit, 25/50%
// by rank, and effect 2 for spending Rapid Killing, 50/100% by rank. A gain at another share replaces
// the one up.
func (hunter *Hunter) gainRapidRecuperation(sim *core.Simulation, share float64) {
	if hunter.rapidRecuperationAura == nil {
		return
	}

	if share != hunter.rapidRecuperationShare {
		hunter.rapidRecuperationAura.Deactivate(sim)
		hunter.rapidRecuperationShare = share
	}
	hunter.rapidRecuperationAura.Activate(sim)
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
