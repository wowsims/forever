package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Rank 6 is the last rank the trainer teaches; rank 7 comes from the Guide: Aspect of the Hawk VII book.
var aspectOfTheHawkTrainerRank = spellData.AspectOfTheHawk.ByID(14322)
var aspectOfTheHawkBookRank = spellData.AspectOfTheHawk.ByID(25296)
var aspectOfTheBeastRank = spellData.AspectOfTheBeast.Highest()
var aspectOfTheCheetahRank = spellData.AspectOfTheCheetah.Highest()
var dazedRank = spellData.AspectOfTheCheetahTriggered.Highest()

// The hastes the aspects proc: Quick Shots off Hawk, Quick Strikes off Beast.
var quickShots = spellData.AspectOfTheHawkTriggered.Highest()
var quickStrikes = spellData.AspectOfTheBeastTriggered.Highest()

func (hunter *Hunter) registerAspects() {
	hunter.AspectOfTheHawkAuras = []*core.Aura{hunter.registerAspect(aspectOfTheHawkTrainerRank), hunter.registerAspect(aspectOfTheHawkBookRank)}
	hunter.AspectOfTheBeastAura = hunter.registerAspect(aspectOfTheBeastRank)
	hunter.registerAspectOfTheCheetah()

	hunter.registerDeadlyAspects()
}

// One aspect at a time: a permanent self-buff the row's effects parse onto, and the spell that swaps
// to it.
func (hunter *Hunter) registerAspect(row *spelldata.Spell) *core.Aura {
	aura := hunter.RegisterAura(spelldata.AuraConfig(row, spelldata.Label(row.RankedName())))
	spelldata.ParseEffects(&hunter.Character, aura, row)
	aura.NewExclusiveEffect("Aspect", true, core.ExclusiveEffect{})

	config := spelldata.SpellConfig(&hunter.Unit, row, spelldata.Flags(core.SpellFlagAPL))

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		aura.Activate(sim)
	}

	config.RelatedSelfBuff = aura

	hunter.RegisterSpell(config)

	return aura
}

// The parse has no row for movement speed, so effect 1 is set by hand, with Pathfinding's
// SPELLMOD_EFFECT1 added to it. Effect 2 dazes the hunter when struck: a slow of its own, which
// MultiplyMovementSpeed stacks on the aspect's speed.
func (hunter *Hunter) registerAspectOfTheCheetah() {
	speed := aspectOfTheCheetahRank.EffectN(1).Percent() + spellData.Pathfinding.EffectAt(1).FractionAt(hunter.Talents.Pathfinding)
	aura := hunter.registerAspect(aspectOfTheCheetahRank)
	aura.NewActiveMovementSpeedEffect(speed)

	slow := 1 + dazedRank.EffectN(1).Percent()
	dazed := hunter.RegisterAura(spelldata.AuraConfig(dazedRank))
	dazed.OnGain = func(_ *core.Aura, sim *core.Simulation) {
		hunter.MultiplyMovementSpeed(sim, slow)
	}
	dazed.OnExpire = func(_ *core.Aura, sim *core.Simulation) {
		hunter.MultiplyMovementSpeed(sim, 1/slow)
	}

	daze := spelldata.ProcTrigger(&hunter.Character, aspectOfTheCheetahRank,
		func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			dazed.Activate(sim)
		})
	daze.Name = "Aspect of the Cheetah - Dazed"
	daze.TriggerImmediately = true
	aura.AttachProcTrigger(daze)
}

// Deadly Aspects is the chance the aspects' own procs fire with: the Hawk and Beast rows state the
// trigger and no rate, and the talent adds SPELLMOD_CHANCE_OF_SUCCESS to each - 2% a rank - which
// the parse has no row for. The tooltip names Auto Shot and melee auto attacks where the rows hear
// every ranged and melee hit, so each mask is narrowed to the white hits by hand.
func (hunter *Hunter) registerDeadlyAspects() {
	rank := hunter.Talents.DeadlyAspects
	if rank == 0 {
		return
	}

	quickShotsAura := hunter.RegisterAura(spelldata.AuraConfig(quickShots))
	spelldata.ParseEffects(&hunter.Character, quickShotsAura, quickShots)
	quickStrikesAura := hunter.RegisterAura(spelldata.AuraConfig(quickStrikes))
	spelldata.ParseEffects(&hunter.Character, quickStrikesAura, quickStrikes)

	hawk := spelldata.ProcTrigger(&hunter.Character, aspectOfTheHawkBookRank,
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			quickShotsAura.Activate(sim)
		},
		spelldata.Chance(spellData.DeadlyAspects.EffectAt(1).FractionAt(rank)))
	hawk.Name = "Deadly Aspects - Quick Shots"
	hawk.ProcMask = core.ProcMaskRangedAuto
	// Quick Shots fires at the Auto Shot's release, a missed shot included.
	hawk.Callback = core.CallbackOnCastResult
	hawk.Outcome = core.OutcomeEmpty
	hawk.RequireDamageDealt = false
	hawk.TriggerImmediately = true
	for _, aura := range hunter.AspectOfTheHawkAuras {
		aura.AttachProcTrigger(hawk)
	}

	beast := spelldata.ProcTrigger(&hunter.Character, aspectOfTheBeastRank,
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			quickStrikesAura.Activate(sim)
		},
		spelldata.Chance(spellData.DeadlyAspects.EffectAt(2).FractionAt(rank)))
	beast.Name = "Deadly Aspects - Quick Strikes"
	beast.ProcMask = core.ProcMaskMeleeWhiteHit
	hunter.AspectOfTheBeastAura.AttachProcTrigger(beast)
}
