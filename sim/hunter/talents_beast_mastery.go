package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Endurance Training, Improved Aspect of the Monkey, Pathfinding, Improved Revive Pet, Bestial
// Swiftness, Improved Mend Pet, Spirit Bond and Intimidation change nothing the sim measures.
func (hunter *Hunter) registerBeastMasteryTalents() {
	// Tier 1
	// Deadly Aspects: aspects.go
	hunter.registerFocusedFire()

	// Tier 3
	hunter.registerUnleashedFury()

	// Tier 4
	hunter.registerFerocity()
	hunter.registerSummonHawk()

	// Tier 5
	hunter.registerBestialDiscipline()
	hunter.registerFrenzy()

	// Tier 6
	hunter.registerBestialWrath()
}

// A dummy the tooltip reads as damage done by the hunter and the pet, the hunter's share only
// while the pet is up.
func (hunter *Hunter) registerFocusedFire() {
	if hunter.Talents.FocusedFire == 0 || hunter.Pet == nil {
		return
	}

	multiplier := spellData.FocusedFire.MultiplierAt(hunter.Talents.FocusedFire)
	hunter.Pet.PseudoStats.DamageDealtMultiplier *= multiplier

	aura := hunter.RegisterAura(core.Aura{
		Label:    "Focused Fire",
		ActionID: core.ActionID{SpellID: spellData.FocusedFire.Rank(hunter.Talents.FocusedFire).ID},
		Duration: core.NeverExpires,
	}).AttachMultiplicativePseudoStatBuff(&hunter.PseudoStats.DamageDealtMultiplier, multiplier)

	hunter.Pet.OnPetEnable = func(sim *core.Simulation) { aura.Activate(sim) }
	hunter.Pet.OnPetDisable = func(sim *core.Simulation) { aura.Deactivate(sim) }
}

// Effect 1 scales the damage done of Tamed Pet Passive (8875), which states 0: the talent's
// amount is the whole bonus. It names Summon Hawk too, which the store carries.
func (hunter *Hunter) registerUnleashedFury() {
	if hunter.Talents.UnleashedFury == 0 {
		return
	}

	spelldata.ParseStatic(&hunter.Character, spellData.UnleashedFury.Rank(hunter.Talents.UnleashedFury))
	for _, pet := range hunter.pets() {
		pet.PseudoStats.DamageDealtMultiplier *= spellData.UnleashedFury.MultiplierAt(hunter.Talents.UnleashedFury)
	}
}

// Effect 1 adds to the crit chance of Tamed Pet Passive (19591), which states 0. It names Summon
// Hawk's damage too, which the store carries.
func (hunter *Hunter) registerFerocity() {
	if hunter.Talents.Ferocity == 0 {
		return
	}

	spelldata.ParseStatic(&hunter.Character, spellData.Ferocity.Rank(hunter.Talents.Ferocity))
	crit := spellData.Ferocity.ValueAt(hunter.Talents.Ferocity)
	for _, pet := range hunter.pets() {
		pet.AddStats(stats.Stats{stats.PhysicalCritPercent: crit, stats.SpellCritPercent: crit})
	}
}

// Effect 1 is the chance of Tamed Pet Passive (20784), which states 0, to cast Frenzy (19615) off
// a crit. 20784 is not in the store; its ProcTypeMask is every direct hit dealt.
func (hunter *Hunter) registerFrenzy() {
	if hunter.Talents.Frenzy == 0 || hunter.Pet == nil {
		return
	}

	rank := spellData.FrenzyTriggered.Highest()
	buff := hunter.Pet.RegisterAura(spelldata.AuraConfig(rank))
	spelldata.ParseEffects(&hunter.Pet.Character, buff, rank)

	decoded := core.DecodeProcTypeMask([2]uint32{dbcenums.PROC_FLAG_ANY_DIRECT_DEALT}, core.ProcHintCrit)
	hunter.Pet.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Frenzy - Trigger",
		ActionID:           core.ActionID{SpellID: spellData.Frenzy.Rank(hunter.Talents.Frenzy).ID},
		Callback:           decoded.Callback,
		ProcMask:           decoded.ProcMask,
		Outcome:            decoded.Outcome,
		RequireDamageDealt: decoded.RequireDamageDealt,
		ProcChance:         spellData.Frenzy.FractionAt(hunter.Talents.Frenzy),
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			buff.Activate(sim)
		},
	})
}

var bestialWrathRank = spellData.BestialWrath.Highest()

// Effect 2, the pet's damage, is the one the sim measures: effect 1 is its size and the rest its
// immunities.
func (hunter *Hunter) registerBestialWrath() {
	if !hunter.Talents.BestialWrath || hunter.Pet == nil {
		return
	}

	pet := hunter.Pet
	aura := pet.RegisterAura(spelldata.AuraConfig(bestialWrathRank))
	spelldata.ParseEffects(&pet.Character, aura, bestialWrathRank, spelldata.Effects(2))

	config := spelldata.SpellConfig(&hunter.Unit, bestialWrathRank, spelldata.Flags(core.SpellFlagAPL))
	config.ExtraCastCondition = func(sim *core.Simulation, _ *core.Unit) bool {
		return pet.IsEnabled()
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		aura.Activate(sim)
	}

	hunter.AddMajorCooldown(core.MajorCooldown{
		Spell: hunter.RegisterSpell(config),
		Type:  core.CooldownTypeDPS,
	})
}

// Effect 1 is the pet's focus regeneration. Effect 2 is the hunter's own mana regeneration while
// casting, A_MOD_MANA_REGEN_INTERRUPT, which the parse table has no row for.
func (hunter *Hunter) registerBestialDiscipline() {
	if hunter.Talents.BestialDiscipline == 0 {
		return
	}

	hunter.PseudoStats.SpiritRegenRateCasting += spellData.BestialDiscipline.EffectAt(2).FractionAt(hunter.Talents.BestialDiscipline)
}
