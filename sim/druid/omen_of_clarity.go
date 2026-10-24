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
		if druid.InForm(Cat) && spellData.FeralCharge.Contains(spell) {
			return
		}
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
	// Chance(1) keeps the resolver from building a procs-per-minute manager and leaves core no roll of
	// its own. The chance depends on what hit, so it is drawn here, after the ICD check and before the
	// ICD starts. Only auto attacks and druid spells roll it.
	// TODO: Ingame research needed -- whether Hurricane's ticks roll it. Each tick is its own instant hit
	// here and rolls once per target, while the logs show no periodic tick ever proccing it.
	trigger.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		// The 15 October development notes take Faerie Fire cast in Cat, Bear or Dire Bear Form off
		// Omen of Clarity. No row states it.
		if spellData.FaerieFire.Contains(spell) && druid.InForm(Cat|Bear|DireBear) {
			return false
		}
		if !spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) && spell.ClassFlags.Family != omen.ClassFlags.Family {
			return false
		}
		chance := float64(omen.RPPM) * druid.omenSpeed(spell) / 60
		return sim.Proc(druid.ModifiedProcChance(omen.ClassFlags, chance), trigger.Name)
	}
	druid.OmenOfClarityAura = druid.MakeProcTriggerAura(trigger)
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
