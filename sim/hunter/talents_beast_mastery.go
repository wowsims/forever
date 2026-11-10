package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Endurance Training (the pet's health and armor), Improved Aspect of the Monkey, Improved Revive
// Pet, Improved Mend Pet and Spirit Bond change nothing the sim measures.
func (hunter *Hunter) registerBeastMasteryTalents() {
	// Tier 1
	// Deadly Aspects: aspects.go
	hunter.registerFocusedFire()

	// Tier 2
	// Pathfinding: aspects.go

	// Tier 3
	hunter.registerBestialSwiftness()
	hunter.registerUnleashedFury()

	// Tier 4
	hunter.registerFerocity()
	hunter.registerSummonHawk()

	// Tier 5
	hunter.registerIntimidation()
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
// amount is the whole bonus. Its class flags match the Summon Hawk dive too, which the store carries;
// the tooltip names the hawks, which take it as the pet does.
func (hunter *Hunter) registerUnleashedFury() {
	if hunter.Talents.UnleashedFury == 0 {
		return
	}

	spelldata.ParseStatic(&hunter.Character, spellData.UnleashedFury.Rank(hunter.Talents.UnleashedFury))
	for _, pet := range hunter.pets() {
		pet.PseudoStats.DamageDealtMultiplier *= spellData.UnleashedFury.MultiplierAt(hunter.Talents.UnleashedFury)
	}
}

// Effect 1 adds to the crit chance of Tamed Pet Passive (19591), which states 0. Its class flags
// match the Summon Hawk dive too, which the store carries; the tooltip names the hawks, which take
// it whole: a hawk with Ferocity 5/5 read 10 plus half its owner's crit.
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

	_, spell := hunter.registerPetBuff(bestialWrathRank)
	hunter.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeDPS,
	})
}

// A buff on the pet with the row's effect 2 parsed onto it, and the hunter's spell that puts it up
// while the pet is.
func (hunter *Hunter) registerPetBuff(rank *spelldata.Spell) (*core.Aura, *core.Spell) {
	pet := hunter.Pet
	aura := pet.RegisterAura(spelldata.AuraConfig(rank))
	spelldata.ParseEffects(&pet.Character, aura, rank, spelldata.Effects(2))

	config := spelldata.SpellConfig(&hunter.Unit, rank, spelldata.Flags(core.SpellFlagAPL))
	config.ExtraCastCondition = func(sim *core.Simulation, _ *core.Unit) bool {
		return pet.IsEnabled()
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		aura.Activate(sim)
	}

	return aura, hunter.RegisterSpell(config)
}

// The talent raises the speed Tamed Pet Passive (19582) gives the pet, which the store lacks: a
// speed of the kind Dash and Dive give, so the higher of them holds.
func (hunter *Hunter) registerBestialSwiftness() {
	if !hunter.Talents.BestialSwiftness || hunter.Pet == nil {
		return
	}

	rank := spellData.BestialSwiftness.Highest()
	aura := hunter.Pet.RegisterAura(spelldata.AuraConfig(rank, spelldata.Permanent()))
	aura.NewActiveMovementSpeedEffect(rank.EffectN(1).Percent())
}

var intimidationRank = spellData.Intimidation.Highest()

// The pet's next attack crits, effect 2, and spends the row's one charge. Effect 1 is the stun that
// attack carries.
func (hunter *Hunter) registerIntimidation() {
	if !hunter.Talents.Intimidation || hunter.Pet == nil {
		return
	}

	aura, _ := hunter.registerPetBuff(intimidationRank)
	aura.ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
		aura.SetStacks(sim, aura.MaxStacks)
	})
	spelldata.AttachChargeSpender(&hunter.Pet.Character, aura, intimidationRank, "Intimidation - Spend")
}

// Effect 1 is the pet's focus regeneration. Effect 2 is the hunter's own mana regeneration while
// casting, A_MOD_MANA_REGEN_INTERRUPT, which the parse table has no row for.
func (hunter *Hunter) registerBestialDiscipline() {
	if hunter.Talents.BestialDiscipline == 0 {
		return
	}

	hunter.PseudoStats.SpiritRegenRateCasting += spellData.BestialDiscipline.EffectAt(2).FractionAt(hunter.Talents.BestialDiscipline)
}
