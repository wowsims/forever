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
	druid.parseTalent(druid.Talents.ImprovedWrath, spellData.ImprovedWrath)
}

// Periodic damage and healing on the spells its class flags name: Moonfire, Insect Swarm, Rip, Rake,
// Pounce, Lacerate, Entangling Roots, Rejuvenation, Regrowth and Wild Growth, and Swiftmend's heal.
func (druid *Druid) applyGenesis() {
	druid.parseTalent(druid.Talents.Genesis, spellData.Genesis)
}

// The mana cost of the damaging spells.
func (druid *Druid) applyMoonglow() {
	druid.parseTalent(druid.Talents.Moonglow, spellData.Moonglow)
}

// Moonfire's crit chance, and the damage of its hit and of its dot.
func (druid *Druid) applyImprovedMoonfire() {
	druid.parseTalent(druid.Talents.ImprovedMoonfire, spellData.ImprovedMoonfire)
}

// Nature's Majesty is spell and melee critical strike chance, in every form.
func (druid *Druid) applyNaturesMajesty() {
	druid.parseTalent(druid.Talents.NaturesMajesty, spellData.NaturesMajesty)
}

// The range of the Balance spells, and spell and melee hit.
func (druid *Druid) applyNaturesReach() {
	druid.parseTalent(druid.Talents.NaturesReach, spellData.NaturesReach)
}

// Entangling Roots' damage: the sim does not cast Entangling Roots, as the bosses it fights cannot be rooted.
func (druid *Druid) applyImprovedEntanglingRoots() {
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
	druid.parseTalent(druid.Talents.Vengeance, spellData.Vengeance)
}

// Starfire's cast time. Its second effect, the chance to stun, is reported as skipped: a stun does
// nothing to the targets the sim fights.
func (druid *Druid) applyImprovedStarfire() {
	druid.parseTalent(druid.Talents.ImprovedStarfire, spellData.ImprovedStarfire)
}

// More targets for Entangling Roots, which the sim does not cast: the bosses it fights cannot be rooted.
func (druid *Druid) applyOvergrowth() {
}

var naturesGraceBuff = spellData.NaturesGraceTriggered.Highest()

// A non-periodic spell crit grants cast speed and a shorter global cooldown for 3 seconds. Nothing
// spends the buff, and a global cooldown under it takes both the cast speed and the 10% cut (1.227 s).
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
	// only once the hit resolves. The direct hit and heal callbacks leave the periodic ones out. The
	// buff comes a spell batch after the crit, so a cast that starts as the crit lands, such as the
	// one queued behind a Starfire, is not hasted.
	// TODO: Ingame research needed once characters reach level 60 -- whether a Hurricane tick crit
	// grants it. Each tick is its own direct hit here and procs it, the way Blizzard's ticks roll
	// Fingers of Frost in the logs.
	trigger.Name = "Nature's Grace - Trigger"
	trigger.Callback = core.CallbackOnSpellHitDealt | core.CallbackOnHealDealt
	trigger.Outcome = core.OutcomeCrit

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
		ClassFlags: spellData.Starfire.Highest().ClassFlags,
		Kind:       core.SpellMod_CastTime_Flat,
		TimeValue:  rank.EffectN(2).TimeValue(),
	})
	// Every Starfire cast spends a charge, a missed one included.
	aura.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if spellData.Starfire.Contains(spell) {
			aura.RemoveStack(sim)
		}
	}

	// Each Wrath whose hit roll lands grants the charges with its cast result, at its cast, so a
	// Starfire queued behind it takes one. A miss or a full resist grants none.
	trigger := spelldata.ProcTrigger(&druid.Character, rank,
		func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			aura.Activate(sim)
			aura.AddStacks(sim, chargesPerWrath)
		})
	trigger.Name = "Eclipse - Trigger"
	trigger.Callback = core.CallbackOnCastResult
	trigger.Outcome = core.OutcomeLanded
	trigger.TriggerImmediately = true
	trigger.ClassFlags = spellData.Wrath.Highest().ClassFlags
	trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		return spellData.Wrath.Contains(spell)
	}
	druid.MakeProcTriggerAura(trigger)
}

// Arcane and Nature damage done.
func (druid *Druid) applyMoonfury() {
	druid.parseTalent(druid.Talents.Moonfury, spellData.Moonfury)
}
