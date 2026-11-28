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
	shaman.applyEyeOfTheStorm()
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

	// The crit is the talent's own ladder, which is what the tooltip of 30160 reads; the buff row
	// states a flat 3.
	critAura := shaman.RegisterAura(spelldata.AuraConfig(elementalDevastationBuff)).
		AttachStatBuff(stats.PhysicalCritPercent, shaman.ElementalDevastationCritPercent())

	// The row hears harmful spells with no roll, and the tooltip makes the crit the outcome.
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

	// The charge goes to the next spell the cost mod reaches, not to the cast that procced it, nor to
	// the Chain Lightning whose overload procced it as that cast completes. An overload carries the
	// spell id of the cast that launched it.
	consumes := clearcastingBuff.EffectN(1).ClassFlags
	var procSpellID int32
	var procTime time.Duration

	// A Lightning Bolt that consumes five Maelstrom Weapon stacks keeps the charge, and so does a
	// Chain Lightning cast while Maelstrom Ready! is up (EV-0180). Both auras come off as the cast
	// completes, in no fixed order with this one, so the cast is checked before its effects apply.
	var kept bool
	config := spelldata.AuraConfig(clearcastingBuff)
	config.OnGain = func(aura *core.Aura, sim *core.Simulation) {
		aura.SetStacks(sim, aura.MaxStacks)
		kept = false
	}
	config.OnApplyEffects = func(_ *core.Aura, _ *core.Simulation, _ *core.Unit, spell *core.Spell) {
		if !consumes.Reaches(spell) {
			return
		}
		maelstrom := shaman.maelstromWeaponAura
		kept = (spell.Matches(SpellMaskLightningBolt) && maelstrom.IsActive() && maelstrom.GetStacks() == maelstrom.MaxStacks) ||
			(spell.Matches(SpellMaskChainLightning) && shaman.maelstromReadyAura.IsActive())
	}
	config.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if !consumes.Reaches(spell) || kept || (spell.SpellID == procSpellID && sim.CurrentTime == procTime) {
			return
		}
		aura.RemoveStack(sim)
	}
	clearcastingAura := shaman.RegisterAura(config)
	spelldata.ParseEffects(&shaman.Character, clearcastingAura, clearcastingBuff)

	// The tooltip limits Elemental Focus to Fire, Frost and Nature damage spells.
	trigger := spelldata.ProcTrigger(&shaman.Character, spellData.ElementalFocus.Highest(),
		func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			procSpellID, procTime = spell.SpellID, sim.CurrentTime
			// A Lightning Bolt overload launches at the bolt's impact, after the bolt's cast has
			// completed, so no cast of that instant keeps the charge.
			if spell.Matches(SpellMaskLightningBoltOverload) {
				procSpellID = 0
			}
			clearcastingAura.Activate(sim)
		})
	trigger.Name = "Elemental Focus - Trigger"
	// The tooltip's "after casting" is one roll per cast, a miss included. The row's helpful-spell
	// bit adds heals, which the tooltip's damage spells leave out.
	trigger.Callback = core.CallbackOnCastComplete
	trigger.ProcMask &^= core.ProcMaskSpellHealing
	trigger.TriggerImmediately = true
	trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		return spell.SpellSchool.Matches(core.SpellSchoolFire | core.SpellSchoolFrost | core.SpellSchoolNature)
	}

	// Overload rolls in EV-0137. No cast listener hears an overload, so its launch calls the trigger's
	// cast listener.
	triggerAura := shaman.MakeProcTriggerAura(trigger)
	shaman.OnSpellRegistered(func(spell *core.Spell) {
		if !spell.Matches(SpellMaskOverload) {
			return
		}
		applyEffects := spell.ApplyEffects
		spell.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			applyEffects(sim, target, spell)
			triggerAura.OnCastComplete(triggerAura, sim, spell)
		}
	})
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

// The chance a hit does not push back a hardcast of the spells its class flags name.
func (shaman *Shaman) applyEyeOfTheStorm() {
	if shaman.Talents.EyeOfTheStorm == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.EyeOfTheStorm.Rank(shaman.Talents.EyeOfTheStorm))
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

// The row states one charge and casts that would use it. No cast uses it (EV-0137).
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
