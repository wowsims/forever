package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerBalanceTalents() {
	// Tier 1
	druid.applyImprovedWrath()
	druid.applyGenesis()

	// Tier 2
	druid.applyMoonglow()
	druid.applyImprovedMoonfire()
	druid.applyNaturesMajesty()
	druid.applyNaturesReach()

	// Tier 3
	druid.applyImprovedEntanglingRoots()
	druid.applyNaturesSplendor()

	// Tier 4
	druid.applyInsectSwarm()
	druid.applyVengeance()
	druid.applyImprovedStarfire()

	// Tier 5
	druid.applyOvergrowth()
	druid.applyNaturesGrace()
	druid.applyEclipse()

	// Tier 6
	druid.applyMoonfury()

	// Tier 7
	// Moonkin Form implemented in forms.go
}

// Wrath's cast time and mana cost.
func (druid *Druid) applyImprovedWrath() {
	if druid.Talents.ImprovedWrath == 0 {
		return
	}
	spelldata.ParseStatic(&druid.Character, spellData.ImprovedWrath.Rank(druid.Talents.ImprovedWrath))
}

// Periodic damage and healing, and Swiftmend's.
func (druid *Druid) applyGenesis() {
	if druid.Talents.Genesis == 0 {
		return
	}
	spelldata.ParseStatic(&druid.Character, spellData.Genesis.Rank(druid.Talents.Genesis))
}

// The mana cost of the damaging spells.
func (druid *Druid) applyMoonglow() {
	if druid.Talents.Moonglow == 0 {
		return
	}
	spelldata.ParseStatic(&druid.Character, spellData.Moonglow.Rank(druid.Talents.Moonglow))
}

// Moonfire's crit chance, and the damage of its hit and of its dot.
func (druid *Druid) applyImprovedMoonfire() {
	if druid.Talents.ImprovedMoonfire == 0 {
		return
	}
	spelldata.ParseStatic(&druid.Character, spellData.ImprovedMoonfire.Rank(druid.Talents.ImprovedMoonfire))
}

// Spell and melee crit chance.
func (druid *Druid) applyNaturesMajesty() {
	if druid.Talents.NaturesMajesty == 0 {
		return
	}
	spelldata.ParseStatic(&druid.Character, spellData.NaturesMajesty.Rank(druid.Talents.NaturesMajesty))
}

// The range of the Balance spells, and spell and melee hit.
func (druid *Druid) applyNaturesReach() {
	if druid.Talents.NaturesReach == 0 {
		return
	}
	spelldata.ParseStatic(&druid.Character, spellData.NaturesReach.Rank(druid.Talents.NaturesReach))
}

// Entangling Roots' damage.
func (druid *Druid) applyImprovedEntanglingRoots() {
	if druid.Talents.ImprovedEntanglingRoots == 0 {
		return
	}
	spelldata.ParseStatic(&druid.Character, spellData.ImprovedEntanglingRoots.Rank(druid.Talents.ImprovedEntanglingRoots))
}

// The duration of Moonfire, Rejuvenation, Regrowth and Insect Swarm.
func (druid *Druid) applyNaturesSplendor() {
	if !druid.Talents.NaturesSplendor {
		return
	}
	spelldata.ParseStatic(&druid.Character, spellData.NaturesSplendor.Highest())
}

func (druid *Druid) applyInsectSwarm() {
	if !druid.Talents.InsectSwarm {
		return
	}

	druid.registerInsectSwarmSpell()
}

// The crit damage bonus of the Arcane and Nature spells.
func (druid *Druid) applyVengeance() {
	if druid.Talents.Vengeance == 0 {
		return
	}
	spelldata.ParseStatic(&druid.Character, spellData.Vengeance.Rank(druid.Talents.Vengeance))
}

// Starfire's cast time. Its second effect, the chance to stun, is reported as skipped: a stun does
// nothing to the targets the sim fights.
func (druid *Druid) applyImprovedStarfire() {
	if druid.Talents.ImprovedStarfire == 0 {
		return
	}
	spelldata.ParseStatic(&druid.Character, spellData.ImprovedStarfire.Rank(druid.Talents.ImprovedStarfire))
}

// More targets for Entangling Roots, which the sim does not cast.
func (druid *Druid) applyOvergrowth() {
}

var naturesGraceBuff = spellData.NaturesGraceTriggered.Highest()

// A non-periodic spell crit grants cast speed and a shorter global cooldown for 3 seconds.
func (druid *Druid) applyNaturesGrace() {
	if !druid.Talents.NaturesGrace {
		return
	}

	aura := druid.RegisterAura(spelldata.AuraConfig(naturesGraceBuff))
	spelldata.ParseEffects(&druid.Character, aura, naturesGraceBuff)

	trigger := spelldata.ProcTrigger(&druid.Character, spellData.NaturesGrace.Highest(),
		func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			aura.Activate(sim)
		})
	// The tooltip's "cast" hint decodes to a cast trigger, but it asks for a crit, which is known
	// only once the hit resolves. The direct hit and heal callbacks leave the periodic ones out.
	trigger.Name = "Nature's Grace - Trigger"
	trigger.Callback = core.CallbackOnSpellHitDealt | core.CallbackOnHealDealt
	trigger.Outcome = core.OutcomeCrit
	// Up before the next cast starts, which the crit that granted it can be followed by at once.
	trigger.TriggerImmediately = true

	druid.MakeProcTriggerAura(trigger)
}

var eclipseBuff = spellData.EclipseTriggered.Highest()

// "Your Wrath spell reduces the cast time of your next $s3 Starfire spells by ${$m2/-1000}.2 sec.
// Stores up to $408255u charges." The talent's effects are dummies, so the charges each Wrath grants
// (effect 3) and the cast time taken off each Starfire (effect 2, the one with a rank curve) are read
// by hand, and the buff row supplies the duration and the charge cap.
func (druid *Druid) applyEclipse() {
	if druid.Talents.Eclipse == 0 {
		return
	}

	rank := spellData.Eclipse.Rank(druid.Talents.Eclipse)
	chargesPerWrath := int32(rank.EffectN(3).BaseValue())

	aura := druid.RegisterAura(spelldata.AuraConfig(eclipseBuff))
	aura.AttachSpellMod(core.SpellModConfig{
		ClassMask: DruidSpellStarfire,
		Kind:      core.SpellMod_CastTime_Flat,
		TimeValue: rank.EffectN(2).TimeValue(),
	})
	aura.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if spell.Matches(DruidSpellStarfire) {
			aura.RemoveStack(sim)
		}
	}

	trigger := spelldata.ProcTrigger(&druid.Character, rank,
		func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			aura.Activate(sim)
			aura.AddStacks(sim, chargesPerWrath)
		})
	trigger.Name = "Eclipse - Trigger"
	// Up before the next cast starts, so the Starfire straight after a Wrath takes a charge.
	trigger.TriggerImmediately = true
	trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		return spellData.Wrath.Contains(spell)
	}

	druid.MakeProcTriggerAura(trigger)
}

// Arcane and Nature damage done.
func (druid *Druid) applyMoonfury() {
	if druid.Talents.Moonfury == 0 {
		return
	}
	spelldata.ParseStatic(&druid.Character, spellData.Moonfury.Rank(druid.Talents.Moonfury))
}
