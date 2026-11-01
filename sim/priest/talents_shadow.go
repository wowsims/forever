package priest

import (
	"fmt"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
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

// Gives you a X% chance to increase your Spirit by Y% for 15 sec after killing a non-trivial target or when an enemy afflicted by your Vampiric Embrace dies.
// For the duration, 50% of your Mana regeneration will continue while casting.
// The row hears the killing blow, and every target in the sim counts as non-trivial. The tooltip also grants the
// buff when the enemy dies with the priest's Vampiric Embrace on it, which the row's proc flags do not state. An
// afflicted enemy always grants it, once, whether or not the raid gets the kill; any other killing blow rolls the
// row's chance.
func (priest *Priest) applySpiritTap() {
	if priest.Talents.SpiritTap == 0 {
		return
	}

	buff := spellData.SpiritTapTriggered.Highest()
	aura := priest.RegisterAura(spelldata.AuraConfig(buff))
	spelldata.ParseEffects(&priest.Character, aura, buff)

	grant := func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		aura.Activate(sim)
	}
	// Vampiric Embrace registers after this talent. Both checks run while the dying target still carries its auras.
	afflicted := func(_ *core.Simulation, _ *core.Spell, result *core.SpellResult) bool {
		return priest.VampiricEmbrace != nil && priest.VampiricEmbrace.RelatedAuraArrays.AnyActive(result.Target)
	}

	trigger := spelldata.ProcTrigger(&priest.Character, spellData.SpiritTap.Rank(priest.Talents.SpiritTap), grant)
	trigger.Name = "Spirit Tap Trigger"
	trigger.TriggerImmediately = true
	trigger.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
		return !afflicted(sim, spell, result)
	}
	priest.MakeProcTriggerAura(trigger)

	priest.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Spirit Tap - Vampiric Embrace",
		Callback:           core.CallbackOnTargetDies,
		TriggerImmediately: true,
		ExtraCondition:     afflicted,
		Handler:            grant,
	})
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
// The buff's row targets the caster's party (target 20). The sim buffs the party's players, the priest
// among them, and not their pets. Each player carries one copy per priest, and only the strongest applies,
// the party buff from the settings included.
func (priest *Priest) applyDarkInspiration() {
	if priest.Talents.DarkInspiration == 0 {
		return
	}

	auras := priest.NewPartyAuraArray(func(member *core.Character) *core.Aura {
		return buffs.NewSpellDataDarkInspirationAuraFor(&member.Unit, &priest.Unit, priest.Talents.DarkInspiration)
	})

	priest.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Dark Inspiration Trigger",
		ClassSpellMask: PriestSpellMindBlast,
		Callback:       core.CallbackOnSpellHitDealt,
		Outcome:        core.OutcomeLanded,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			auras.ActivateAll(sim)
		},
	})
	priest.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Dark Inspiration Channel Trigger",
		ClassSpellMask: PriestSpellMindFlay,
		Callback:       core.CallbackOnPeriodicDamageDealt,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.Dot(result.Target).RemainingTicks() == 0 {
				auras.ActivateAll(sim)
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

	critMod := priest.AddDynamicMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		FloatValue: spellData.EarlyDemise.EffectAt(1).ValueAt(priest.Talents.EarlyDemise),
		ClassMask:  PriestSpellShadowWordDeath,
	})
	priest.RegisterResetEffect(func(_ *core.Simulation) {
		critMod.Deactivate()
	})

	// The threshold is the target's health, and the game files state no buff that would carry the bonus to another
	// target, so the bonus follows the target of each Shadow Word: Death. The sim tracks a target's health as its
	// execute phase, so only a 20% threshold can be read.
	if threshold := spellData.EarlyDemise.EffectAt(2).ValueAt(priest.Talents.EarlyDemise); threshold != 20 {
		panic(fmt.Sprintf("priest: Early Demise %d states a %.0f%% health threshold, and the sim tracks the 20%% phase only",
			spellData.EarlyDemise.Rank(priest.Talents.EarlyDemise).ID, threshold))
	}
	core.MakePermanent(priest.RegisterAura(core.Aura{
		Label: "Early Demise",
		OnApplyEffects: func(_ *core.Aura, _ *core.Simulation, target *core.Unit, spell *core.Spell) {
			if !spell.Matches(PriestSpellShadowWordDeath) {
				return
			}
			if target.IsExecutePhase20() {
				critMod.Activate()
			} else {
				critMod.Deactivate()
			}
		},
	}))
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
