package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerShadowTalents() {
	// Tier 1
	priest.applyShadowFocus()
	priest.applyBlackout()
	priest.applySpiritTap()

	// Tier 2
	priest.applyImprovedShadowWordPain()
	priest.applyShadowReach()

	// Tier 3
	priest.applyImprovedMindBlast()
	priest.applyImprovedPsychicScream()
	priest.applyMindFlay()
	priest.applyImprovedMindFlay()

	// Tier 4
	priest.applyDarkInspiration()
	priest.applyVampiricEmbrace()
	priest.applyShadowWeaving()

	// Tier 5
	priest.applySilence()
	priest.applyDevouringContagion()

	// Tier 6
	priest.applyEarlyDemise()
	priest.applyDarkness()

	// Tier 7
	priest.applyShadowform()
}

// Improves your chance to hit with Shadow spells by 5%.
func (priest *Priest) applyShadowFocus() {
	if priest.Talents.ShadowFocus == 0 {
		return
	}

	spelldata.ParseStatic(&priest.Character, spellData.ShadowFocus.Rank(priest.Talents.ShadowFocus), spelldata.ShowAsSchoolHit(core.SpellSchoolShadow))
}

// Gives your Shadow damage spells a 10% chance to stun the target for 3 sec.
// The damage sim does not model the stun.
func (priest *Priest) applyBlackout() {}

// Gives you a X% chance to increase your Spirit by Y% for 15 sec when an enemy afflicted by your Vampiric Embrace dies.
// For the duration, 50% of your Mana regeneration will continue while casting.
// TODO: Procs on a target's death, which the sim does not model.
func (priest *Priest) applySpiritTap() {
	if priest.Talents.SpiritTap == 0 {
		return
	}
}

// Increases the duration of your Shadow Word: Pain spell by 6 sec.
func (priest *Priest) applyImprovedShadowWordPain() {
	if priest.Talents.ImprovedShadowWordPain == 0 {
		return
	}

	spelldata.ParseStatic(&priest.Character, spellData.ImprovedShadowWordPain.Rank(priest.Talents.ImprovedShadowWordPain))
}

// Increases the range of your offensive Shadow spells by X%.
func (priest *Priest) applyShadowReach() {
	if priest.Talents.ShadowReach == 0 {
		return
	}

	spelldata.ParseStatic(&priest.Character, spellData.ShadowReach.Rank(priest.Talents.ShadowReach))
}

// Reduces the cooldown of your Mind Blast spell by X sec.
func (priest *Priest) applyImprovedMindBlast() {
	if priest.Talents.ImprovedMindBlast == 0 {
		return
	}

	spelldata.ParseStatic(&priest.Character, spellData.ImprovedMindBlast.Rank(priest.Talents.ImprovedMindBlast))
}

// Reduces the cooldown of your Psychic Scream spell by 4 sec.
// The damage sim does not model crowd control.
func (priest *Priest) applyImprovedPsychicScream() {}

// Grants the Mind Flay ability
func (priest *Priest) applyMindFlay() {
	if !priest.Talents.MindFlay {
		return
	}
	spellData.MindFlay.Each(priest.registerMindFlaySpell)
}

// Each instance of damage of Mind Flay deals X% more damage than its previous instance of damage.
// The ramp is on Mind Flay's ticks, in registerMindFlaySpell.
func (priest *Priest) applyImprovedMindFlay() {}

// Your Mind Blast and fully channeled Mind Flay causes Dark Inspiration, increasing yours and your party's Spirit by X% for 10 sec.
// The sim models only the priest's own Spirit.
func (priest *Priest) applyDarkInspiration() {
	if priest.Talents.DarkInspiration == 0 {
		return
	}

	buff := spellData.DarkInspirationTriggered.Highest()
	aura := priest.RegisterAura(spelldata.AuraConfig(buff))
	// The buff's row states the full five-rank amount, which the talent's ranks scale.
	spelldata.ParseEffects(&priest.Character, aura, buff, spelldata.ScaledByRanks(priest.Talents.DarkInspiration, spellData.DarkInspiration.Len()))

	priest.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Dark Inspiration Trigger",
		ClassSpellMask: PriestSpellMindBlast,
		Callback:       core.CallbackOnSpellHitDealt,
		Outcome:        core.OutcomeLanded,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			aura.Activate(sim)
		},
	})
	priest.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Dark Inspiration Channel Trigger",
		ClassSpellMask: PriestSpellMindFlay,
		Callback:       core.CallbackOnPeriodicDamageDealt,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.Dot(result.Target).RemainingTicks() == 0 {
				aura.Activate(sim)
			}
		},
	})
}

// Grants the Vampiric Embrace Spell
func (priest *Priest) applyVampiricEmbrace() {
	if !priest.Talents.VampiricEmbrace {
		return
	}

	spellData.VampiricEmbrace.Each(priest.registerVampiricEmbrace)
}

// Your Shadow damage spells have a X% chance to increase the Shadow damage you deal by 2% for 15 sec, stacking up to 5 times.
func (priest *Priest) applyShadowWeaving() {
	if priest.Talents.ShadowWeaving == 0 {
		return
	}

	debuff := spellData.ShadowWeavingTriggered.Highest()
	swAuras := priest.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		aura := target.GetOrRegisterAura(spelldata.AuraConfig(debuff, spelldata.RankLabel(debuff, &priest.Unit)))
		spelldata.ParseEffects(nil, aura, debuff, spelldata.FromCaster(&priest.Unit, 0, 2))
		return aura
	})

	priest.MakeProcTriggerAura(core.ProcTrigger{
		Name:             "Shadow Weaving Trigger",
		CanProcFromProcs: true, // 15257 carries the bit.
		ClassSpellMask:   PriestShadowSpells,
		Callback:         core.CallbackOnSpellHitDealt,
		Outcome:          core.OutcomeLanded,
		ProcChance:       spellData.ShadowWeaving.FractionAt(priest.Talents.ShadowWeaving),
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			aura := swAuras.Get(result.Target)
			aura.Activate(sim)
			aura.AddStack(sim)
		},
	})
}

// Grants the Silence ability
// The damage sim does not model interrupts.
func (priest *Priest) applySilence() {}

// Reduces the mana cost of your Devouring Plague by (25/50)%.
// Targets that die while Devouring Plague is active spreads it, jumping to a nearby enemy within (5/10) yds for the remaining duration.
func (priest *Priest) applyDevouringContagion() {
	if priest.Talents.DevouringContagion == 0 {
		return
	}

	spelldata.ParseStatic(&priest.Character, spellData.DevouringContagion.Rank(priest.Talents.DevouringContagion))
}

// Increases Shadow Word: Death's critical strike chance on targets at or below X% health by Y%.
func (priest *Priest) applyEarlyDemise() {
	if priest.Talents.EarlyDemise == 0 {
		return
	}

	// The sim tracks the target's health as the execute phases 90, 45, 35, 25 and 20.
	threshold := int32(spellData.EarlyDemise.EffectAt(2).ValueAt(priest.Talents.EarlyDemise))
	critMod := priest.AddDynamicMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		FloatValue: spellData.EarlyDemise.EffectAt(1).ValueAt(priest.Talents.EarlyDemise),
		ClassMask:  PriestSpellShadowWordDeath,
	})

	onExecutePhase := func(_ *core.Simulation, isExecute int32) {
		if isExecute <= threshold {
			critMod.Activate()
		}
	}
	priest.RegisterResetEffect(func(sim *core.Simulation) {
		critMod.Deactivate()
		sim.RegisterExecutePhaseCallback(onExecutePhase)
	})
}

// Increases your Shadow damage done by X%
func (priest *Priest) applyDarkness() {
	if priest.Talents.Darkness == 0 {
		return
	}

	spelldata.ParseStatic(&priest.Character, spellData.Darkness.Rank(priest.Talents.Darkness))
}

// Grants the Shadowform Ability
func (priest *Priest) applyShadowform() {
	if !priest.Talents.Shadowform {
		return
	}

	rank := spellData.Shadowform.Highest()
	form := rank.ShapeshiftForm()

	auraConfig := spelldata.AuraConfig(rank)
	auraConfig.OnReset = func(aura *core.Aura, sim *core.Simulation) {
		if priest.SelfBuffs.PreShadowform {
			aura.Activate(sim)
		}
	}
	auraConfig.OnGain = func(_ *core.Aura, _ *core.Simulation) {
		priest.ShapeshiftForm = form
	}
	auraConfig.OnExpire = func(_ *core.Aura, _ *core.Simulation) {
		priest.ShapeshiftForm = 0
	}
	shadowformAura := priest.RegisterAura(auraConfig)
	spelldata.ParseEffects(&priest.Character, shadowformAura, rank)

	priest.registerSelfBuffSpell(rank, shadowformAura)
}
