package shaman

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func (shaman *Shaman) registerElementalTalents() {
	// Tier 1
	shaman.applyConvection()
	shaman.applyConcussion()

	// Tier 2
	shaman.applyElementalDevastation()
	shaman.applyReverberation()
	shaman.applyElementalAlacrity()
	shaman.applyCallOfFlame()

	// Tier 3
	shaman.applyElementalFocus()
	shaman.applyElementalWarding()

	// Tier 4
	shaman.applyTrialByFire()
	// Eye of the Storm (29062) reduces pushback only, which the sim does not model.
	// Earthbound (1222988) makes Earthbind Totem root, which the sim does not model.

	// Tier 5
	shaman.applyElementalReach()
	shaman.applyCallOfThunder()
	// Lightning Overload: shaman.go, GetOverloadChance

	// Tier 6
	shaman.applyElementalFury()

	// Tier 7
	shaman.applyLightningMastery()
}

func (shaman *Shaman) applyConvection() {
	if shaman.Talents.Convection == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.Convection.Rank(shaman.Talents.Convection))
}

func (shaman *Shaman) applyConcussion() {
	if shaman.Talents.Concussion == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.Concussion.Rank(shaman.Talents.Concussion))
}

func (shaman *Shaman) applyElementalWarding() {
	if shaman.Talents.ElementalWarding == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.ElementalWarding.Rank(shaman.Talents.ElementalWarding))
}

func (shaman *Shaman) applyReverberation() {
	if shaman.Talents.Reverberation == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.Reverberation.Rank(shaman.Talents.Reverberation))
}

func (shaman *Shaman) applyCallOfFlame() {
	if shaman.Talents.CallOfFlame == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.CallOfFlame.Rank(shaman.Talents.CallOfFlame))
}

var elementalDevastationBuff = spellData.ElementalDevastationTriggered.Highest()

func (shaman *Shaman) applyElementalDevastation() {
	if shaman.Talents.ElementalDevastation == 0 {
		return
	}

	// 30165 supplies the duration. The crit is the talent's own ladder, 3/6/9 by rank, which is
	// what the tooltip of 30160 reads; the buff row states a flat 3.
	critAura := shaman.RegisterAura(spelldata.AuraConfig(elementalDevastationBuff)).
		AttachStatBuff(stats.PhysicalCritPercent, shaman.ElementalDevastationCritPercent())

	// The row hears harmful spells with no roll, and the tooltip's "Critical Strikes with direct
	// damage spells" makes the crit the outcome.
	trigger := spelldata.ProcTrigger(&shaman.Character, spellData.ElementalDevastation.Rank(shaman.Talents.ElementalDevastation),
		func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			critAura.Activate(sim)
		})
	trigger.Name = "Elemental Devastation - Trigger"

	shaman.MakeProcTriggerAura(trigger)
}

// The physical crit Elemental Devastation's buff grants at the talent's rank, 0 without the talent.
func (shaman *Shaman) ElementalDevastationCritPercent() float64 {
	return spellData.ElementalDevastation.ValueAt(shaman.Talents.ElementalDevastation)
}

var clearcastingBuff = spellData.ElementalFocusTriggered.Highest()

func (shaman *Shaman) applyElementalFocus() {
	if !shaman.Talents.ElementalFocus {
		return
	}

	// The charge goes to the next spell the cost mod reaches, not to the cast that procced it.
	consumes := clearcastingBuff.EffectN(1).ClassFlags
	var procSpell *core.Spell
	var procTime time.Duration

	config := spelldata.AuraConfig(clearcastingBuff)
	config.OnGain = func(aura *core.Aura, sim *core.Simulation) {
		aura.SetStacks(sim, aura.MaxStacks)
	}
	config.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if !consumes.Reaches(spell) || (spell == procSpell && sim.CurrentTime == procTime) {
			return
		}
		aura.RemoveStack(sim)
	}
	clearcastingAura := shaman.RegisterAura(config)
	spelldata.ParseEffects(&shaman.Character, clearcastingAura, clearcastingBuff)

	// The row rolls its own 10%. The tooltip limits it to Fire, Frost and Nature damage spells.
	trigger := spelldata.ProcTrigger(&shaman.Character, spellData.ElementalFocus.Highest(),
		func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			procSpell, procTime = spell, sim.CurrentTime
			clearcastingAura.Activate(sim)
		})
	trigger.Name = "Elemental Focus - Trigger"
	// The tooltip's "after casting" is one roll per cast, a miss included, and none for what the
	// shaman does not cast, such as a Lightning Overload. The row's helpful-spell bit adds heals,
	// which the tooltip's damage spells leave out.
	trigger.Callback = core.CallbackOnCastComplete
	trigger.ProcMask &^= core.ProcMaskSpellHealing
	trigger.TriggerImmediately = true
	trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		return spell.SpellSchool.Matches(core.SpellSchoolFire | core.SpellSchoolFrost | core.SpellSchoolNature)
	}

	shaman.MakeProcTriggerAura(trigger)
}

func (shaman *Shaman) applyElementalAlacrity() {
	if shaman.Talents.ElementalAlacrity == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.ElementalAlacrity.Rank(shaman.Talents.ElementalAlacrity))
}

func (shaman *Shaman) applyTrialByFire() {
	if shaman.Talents.TrialByFire == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.TrialByFire.Rank(shaman.Talents.TrialByFire))
}

// The range the row states reaches the spells' range checks only. Effect 3 is a dummy.
func (shaman *Shaman) applyElementalReach() {
	if shaman.Talents.ElementalReach == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.ElementalReach.Rank(shaman.Talents.ElementalReach))
}

func (shaman *Shaman) applyElementalFury() {
	if shaman.Talents.ElementalFury == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.ElementalFury.Rank(shaman.Talents.ElementalFury))
}

// The crit reaches the overloads too, whose rows carry Lightning Bolt's and Chain Lightning's class flags.
func (shaman *Shaman) applyCallOfThunder() {
	if !shaman.Talents.CallOfThunder {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.CallOfThunder.Highest())
}

var lightningMasteryRank = spellData.LightningMastery.Highest()

// The row states one charge and casts that would spend it, but the tooltip reads "Lasts $d", so the
// buff lasts its full 15 s and no cast spends it.
//
// TODO: Ingame test needed: whether the first spell the buff reaches ends it.
func (shaman *Shaman) applyLightningMastery() {
	if !shaman.Talents.LightningMastery {
		return
	}

	lmAura := shaman.RegisterAura(spelldata.AuraConfig(lightningMasteryRank))
	spelldata.ParseEffects(&shaman.Character, lmAura, lightningMasteryRank)

	config := spelldata.SpellConfig(&shaman.Unit, lightningMasteryRank)
	config.ClassSpellMask = SpellMaskLightningMastery
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		lmAura.Activate(sim)
	}
	config.RelatedSelfBuff = lmAura

	shaman.AddMajorCooldown(core.MajorCooldown{
		Spell: shaman.RegisterSpell(config),
		Type:  core.CooldownTypeDPS,
	})
}
