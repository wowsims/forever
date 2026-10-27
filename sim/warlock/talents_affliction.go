package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (warlock *Warlock) registerAfflictionTalents() {
	// Tier 1
	// Improved Life Tap implemented in lifetap.go
	warlock.applySuppression()
	warlock.applyImprovedCorruption()

	// Tier 2
	warlock.applyGrimReach()
	warlock.applySoulHarvest()
	warlock.applyImprovedDrains()

	// Tier 3
	warlock.applyImprovedBaneOfAgony()
	warlock.applyFelConcentration()
	warlock.registerAmplifyCurse()
	warlock.applyPandemic()

	// Tier 4
	warlock.applyMalevolence()
	warlock.applyNightfall()
	warlock.registerCurseOfExhaustion()

	// Tier 5
	warlock.applySiphonLife()
	// Soul Siphon implemented via helper functions that calculate the multiplier, and are called by the affected spells.

	// Tier 6
	warlock.applyShadowMastery()

	// Tier 7
	warlock.applyWrack()
}

// --- Tier 1 ---

func (warlock *Warlock) applySuppression() {
	if warlock.Talents.Suppression == 0 {
		return
	}

	spelldata.ParseStatic(&warlock.Character, spellData.Suppression.Rank(warlock.Talents.Suppression))
}

func (warlock *Warlock) applyImprovedCorruption() {
	if warlock.Talents.ImprovedCorruption == 0 {
		return
	}
	spelldata.ParseStatic(&warlock.Character, spellData.ImprovedCorruption.Rank(warlock.Talents.ImprovedCorruption))
}

// --- Tier 2 ---

// Grim Reach raises the range of the Affliction spells by 10/20%, read from its row.
func (warlock *Warlock) applyGrimReach() {
	if warlock.Talents.GrimReach == 0 {
		return
	}

	spelldata.ParseStatic(&warlock.Character, spellData.GrimReach.Rank(warlock.Talents.GrimReach))
}

// TODO: Not implemented due to lack of a kill event: adds "leave" the encounter on a set timer rather
// than dying. Requires a Drain Soul kill to trigger.
func (warlock *Warlock) applySoulHarvest() {
	if warlock.Talents.SoulHarvest == 0 {
		return
	}
}

func (warlock *Warlock) applyImprovedDrains() {
	if warlock.Talents.ImprovedDrains == 0 {
		return
	}

	spelldata.ParseStatic(&warlock.Character, spellData.ImprovedDrains.Rank(warlock.Talents.ImprovedDrains))
}

// --- Tier 3 ---
func (warlock *Warlock) applyImprovedBaneOfAgony() {
	if warlock.Talents.ImprovedBaneOfAgony == 0 {
		return
	}

	spelldata.ParseStatic(&warlock.Character, spellData.ImprovedBaneOfAgony.Rank(warlock.Talents.ImprovedBaneOfAgony))
}

// TODO: Not implemented, as it does not affect sims for now.
// The row is SPELLMOD_NOT_LOSE_CASTING_TIME (pushback protection) on the drains and Wrack,
// which the parse skips; it matters only once the sim models pushback from damage taken.
func (warlock *Warlock) applyFelConcentration() {
	if warlock.Talents.FelConcentration == 0 {
		return
	}
}

func (warlock *Warlock) registerAmplifyCurse() {
	if !warlock.Talents.AmplifyCurse {
		return
	}

	rankedSpell := spellData.AmplifyCurse.ByID(18288)
	warlock.AmplifyCurseMultiplier = 1 + rankedSpell.EffectN(1).Percent()

	// The curses the tooltip names spend it. Effect 1 names Curse of Weakness and effect 2 Curse of Exhaustion;
	// Bane of Agony is in neither class mask in the game files, so it is matched by its own spell mask.
	ampWeakness, ampExhaustion := rankedSpell.EffectN(1).ClassFlags, rankedSpell.EffectN(2).ClassFlags

	// TODO: Confirm that the amp is consumed on cast, regardless of hit.
	auraConfig := spelldata.AuraConfig(rankedSpell)
	auraConfig.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if spell.MatchesFlags(ampWeakness) || spell.MatchesFlags(ampExhaustion) || spell.Matches(WarlockSpellBaneOfAgony) {
			aura.Deactivate(sim)
		}
	}
	warlock.AmplifyCurseAura = warlock.RegisterAura(auraConfig)

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Flags(core.SpellFlagAPL))
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		warlock.AmplifyCurseAura.Activate(sim)
	}
	warlock.AmplifyCurse = warlock.RegisterSpell(config)

	// Never cast on its own: the APL decides which curse takes the amp.
	warlock.AddMajorCooldown(core.MajorCooldown{
		Spell:          warlock.AmplifyCurse,
		Type:           core.CooldownTypeDPS,
		ShouldActivate: func(*core.Simulation, *core.Character) bool { return false },
	})
}

func (warlock *Warlock) applyPandemic() {
	if warlock.Talents.Pandemic == 0 {
		return
	}

	spelldata.ParseStatic(&warlock.Character, spellData.Pandemic.Rank(warlock.Talents.Pandemic))
}

// --- Tier 4 ---

// The class flags also name Lash of Pain, which the Succubus casts, so the row is parsed on the Succubus too.
func (warlock *Warlock) applyMalevolence() {
	if warlock.Talents.Malevolence == 0 {
		return
	}

	rank := spellData.Malevolence.Rank(warlock.Talents.Malevolence)
	spelldata.ParseStatic(&warlock.Character, rank)
	spelldata.ParseStatic(&warlock.Succubus.Character, rank)
}

func (warlock *Warlock) applyNightfall() {
	if warlock.Talents.Nightfall == 0 {
		return
	}

	shadowTrance := spellData.NightfallTriggered.ByID(17941)
	// Effect 1 names the spells it makes instant, which are the ones that spend it.
	instant := shadowTrance.EffectN(1).ClassFlags

	auraConfig := spelldata.AuraConfig(shadowTrance)
	auraConfig.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if spell.MatchesFlags(instant) && spell.CurCast.CastTime == 0 {
			aura.Deactivate(sim)
		}
	}

	warlock.NightfallProcAura = warlock.RegisterAura(auraConfig)

	spelldata.ParseEffects(&warlock.Character, warlock.NightfallProcAura, shadowTrance)

	trigger := spelldata.ProcTrigger(&warlock.Character, spellData.Nightfall.Rank(warlock.Talents.Nightfall),
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			warlock.NightfallProcAura.Activate(sim)
		})
	// The talent row names no spells, so the triggering ones are gathered from their own rows.
	trigger.ClassFlags = combineClassFlags(
		spellData.Corruption.Highest().ClassFlags,
		spellData.DrainSoul.Highest().ClassFlags,
		spellData.DrainLife.Highest().ClassFlags,
		spellData.Wrack.Highest().ClassFlags,
	)

	trigger.RequireDamageDealt = true
	warlock.MakeProcTriggerAura(trigger)
}

// TODO: Not implemented, as it does not affect DPS: the curse only slows the target's movement.
func (warlock *Warlock) registerCurseOfExhaustion() {
	if !warlock.Talents.CurseOfExhaustion {
		return
	}
}

// --- Tier 5 ---

func (warlock *Warlock) applySiphonLife() {
	if !warlock.Talents.SiphonLife {
		return
	}

	warlock.registerSiphonLifeSpell()
}

func (warlock *Warlock) soulSiphonPerEffect() float64 {
	return spellData.SoulSiphon.FractionAt(warlock.Talents.SoulSiphon)
}

// Soul Siphon's cap, from its tooltip: "4/8/12% per each of your other Affliction effects active on the
// target, up to a maximum increase of 12/24/36%", three effects at every rank. Row 17804 is a single
// A_DUMMY stating the per-effect share only: no cap and no class flags.
const soulSiphonMaxEffects = 3

// The multiplier a Soul Siphon spell's tick takes. The spells are the tooltip's, since the row names none:
// Drain Life, Drain Soul and Wrack, each of which calls this.
// TODO: Appears to count any "affliction effect", confirm this includes curses in-game.
func (warlock *Warlock) soulSiphonMultiplier(target *core.Unit, multiplierPerEffect float64) float64 {
	if multiplierPerEffect == 0 {
		return 1
	}
	return 1 + multiplierPerEffect*min(warlock.AfflictionCount(target), soulSiphonMaxEffects)
}

// Builds a Soul Siphon channel: Drain Life, Drain Soul and Wrack share the tick, hit roll and expected
// damage, and differ in the tick effect and what each adds on top. afterTick, when given, runs on each
// tick's result once it is dealt (Drain Life's heal).
func (warlock *Warlock) soulSiphonChannelConfig(rankedSpell *spelldata.Spell, tick *spelldata.Effect,
	afterTick func(sim *core.Simulation, result *core.SpellResult)) core.SpellConfig {
	soulSiphonMult := warlock.soulSiphonPerEffect()

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))

	// TODO: Confirm no snapshot from Soul Siphon.
	// TODO: Confirm in game whether spell haste shortens the ticks; the rows carry no haste attribute, so
	// neither AffectedByCastSpeed nor HasteReducesDuration is set.
	config.Dot = spelldata.DotConfig(rankedSpell, tick)
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		result := dot.Spell.CalcPeriodicDamage(sim, target, tick.Average(), rankedSpell.TickOutcome(dot))
		scaleResult(sim, dot.Spell, result, warlock.soulSiphonMultiplier(target, soulSiphonMult))
		dot.Spell.DealPeriodicDamage(sim, result)
		if afterTick != nil {
			afterTick(sim, result)
		}
	}

	// Roll to hit and apply as dot
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
		spell.DealOutcome(sim, result)
	}

	// Calculate expected damage, scaled by Soul Siphon multiplier.
	config.ExpectedTickDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		result := spell.CalcPeriodicDamage(sim, target, tick.Average(), spell.OutcomeExpectedMagicAlwaysHit)
		result.Damage *= warlock.soulSiphonMultiplier(target, soulSiphonMult)
		return result
	}

	return config
}

// --- Tier 6 ---
func (warlock *Warlock) applyShadowMastery() {
	if warlock.Talents.ShadowMastery == 0 {
		return
	}

	// Effect 1 raises direct damage and effect 2 periodic damage.
	// Effect 1's class flags include Corruption, which has no direct damage.
	spelldata.ParseStatic(&warlock.Character, spellData.ShadowMastery.Rank(warlock.Talents.ShadowMastery))
}

// --- Tier 7 ---
func (warlock *Warlock) applyWrack() {
	if !warlock.Talents.Wrack {
		return
	}

	warlock.registerWrackSpell()
}
