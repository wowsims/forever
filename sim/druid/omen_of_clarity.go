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
	// ICD starts.
	trigger.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		// The 15 October development notes take Faerie Fire cast in Cat, Bear or Dire Bear Form off
		// Omen of Clarity. Not in game files.
		if spellData.FaerieFire.Contains(spell) && druid.InForm(Cat|Bear|DireBear) {
			return false
		}
		if !spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) && spell.ClassFlags.Family != omen.ClassFlags.Family {
			return false
		}
		// Hurricane's ticks never roll it, and the cast trigger below rolls its cast, so the cast's
		// landings roll nothing here.
		if spellData.HurricaneTriggered.Contains(spell) || spellData.Hurricane.Contains(spell) {
			return false
		}
		return roll(sim, spell)
	}
	druid.OmenOfClarityAura = druid.MakeProcTriggerAura(trigger)

	// Hurricane and, in Moonkin Form only, the helpful buffs Thorns and Innervate roll it at the cast, on
	// the same chance and cooldown. Mark of the Wild and Nature's Grasp roll as those do, and the sim
	// casts neither; a shapeshift rolls nothing. The buff comes a spell batch later, after the cast has
	// paid its cost or used Clearcasting, so a proc never makes its own cast free.
	//
	// Core gives every trigger a cooldown of its own, and the spell's proc mask hears no helpful cast, so
	// this trigger gates on and starts the hit trigger's cooldown itself.
	icd := druid.OmenOfClarityAura.Icd
	druid.OmenOfClarityAura.AttachProcTrigger(core.ProcTrigger{
		Name:     "Omen of Clarity - Cast Trigger",
		Callback: core.CallbackOnCastComplete,
		Handler:  trigger.Handler,
		ExtraCondition: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
			rolls := int32(0)
			switch {
			case spellData.Hurricane.Contains(spell):
				rolls = sim.Environment.ActiveTargetCount()
			case druid.InForm(Moonkin) && (spellData.Thorns.Contains(spell) || spellData.Innervate.Contains(spell)):
				rolls = 1
			}
			if rolls == 0 || !icd.IsReady(sim) {
				return false
			}
			for range rolls {
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
