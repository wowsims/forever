package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerOmenOfClarity() {
	clearRow := spellData.OmenOfClarityTriggered.Highest()
	druid.ClearcastingAura = druid.RegisterAura(spelldata.AuraConfig(clearRow))
	spelldata.ParseEffects(&druid.Character, druid.ClearcastingAura, clearRow)
	mask := clearRow.EffectN(1).ClassFlags
	druid.ClearcastingAura.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if mask.Matches(spell.ClassFlags) {
			aura.Deactivate(sim)
		}
	}

	omen := spellData.OmenOfClarity.Highest()
	trigger := spelldata.ProcTrigger(&druid.Character, omen, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		druid.ClearcastingAura.Activate(sim)
	}, spelldata.Chance(1))
	// A landed caster-form debuff can roll it without dealing damage.
	trigger.RequireDamageComponent = false
	roll := func(sim *core.Simulation, spell *core.Spell) bool {
		chance := float64(omen.RPPM) * druid.omenSpeed(spell) / 60
		return sim.Proc(druid.ModifiedProcChance(omen.ClassFlags, chance), trigger.Name)
	}
	// Chance(1) keeps the resolver from building a procs-per-minute manager and leaves core no roll of
	// its own. The chance depends on what hit, so it is drawn here, after the ICD check and before the
	// ICD starts. Only auto attacks and druid spells roll it.
	trigger.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		// The 15 October development notes take Faerie Fire cast in Cat, Bear or Dire Bear Form off
		// Omen of Clarity. No row states it.
		if spellData.FaerieFire.Contains(spell) && druid.InForm(Cat|Bear|DireBear) {
			return false
		}
		if !spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) && spell.ClassFlags.Family != omen.ClassFlags.Family {
			return false
		}
		// Hurricane's ticks never roll it.
		if spellData.HurricaneTriggered.Contains(spell) {
			return false
		}
		return roll(sim, spell)
	}
	druid.OmenOfClarityAura = druid.MakeProcTriggerAura(trigger)

	// The Hurricane cast rolls it at the cast, once for each target its area reaches, on the same chance
	// and cooldown; a proc ends the rolls. The buff comes a spell batch later, after the cast has paid
	// its cost or spent Clearcasting, so a proc never makes its own cast free.
	icd := druid.OmenOfClarityAura.Icd
	druid.MakeProcTriggerAura(core.ProcTrigger{
		Name:     "Omen of Clarity - Cast Trigger",
		Callback: core.CallbackOnCastComplete,
		Handler:  trigger.Handler,
		ExtraCondition: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
			if !spellData.Hurricane.Contains(spell) || !icd.IsReady(sim) {
				return false
			}
			for range sim.Encounter.ActiveTargetUnits {
				if roll(sim, spell) {
					icd.Set(sim.CurrentTime + druid.ModifiedProcCooldown(trigger.ProcSpellClassFlags, trigger.ICD))
					return true
				}
			}
			return false
		},
	})
}

// The speed, in seconds, that turns Omen of Clarity's procs per minute into a chance for this hit.
// Auto attacks and specials use the swing speed: the paw in Cat and Bear Form, the weapon otherwise.
// A spell uses its cast time after talent and set reductions, and an instant 1.5 s; haste and cast
// speed such as Nature's Grace do not change it.
func (druid *Druid) omenSpeed(spell *core.Spell) float64 {
	switch {
	case spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit | core.ProcMaskMeleeSpecial):
		return druid.AutoAttacks.MH().SwingSpeed
	case spell.DefaultCast.CastTime > 0:
		return spell.DefaultCast.CastTime.Seconds()
	default:
		return core.GCDDefault.Seconds()
	}
}
