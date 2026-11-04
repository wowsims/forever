package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var poisedPassive = spelldata.MustFind(1328825)
var poisedBuff = poisedPassive.Refs()[0]

// At maximum energy and out of combat the next ability costs less. Every fight starts out of combat
// at full energy, so the buff is up from the start until the first ability spends it.
func (rogue *Rogue) registerPoised() {
	aura := rogue.RegisterAura(spelldata.AuraConfig(poisedBuff))
	spelldata.ParseEffects(&rogue.Character, aura, poisedBuff)
	// The discount reaches finishers that hit nothing, Slice and Dice and Blade Flurry among them, so the
	// charge goes when a cast of the abilities the buff names completes, not when one lands.
	aura.AttachProcTriggerCallback(&rogue.Unit, core.ProcTrigger{
		Name:               "Poised - Spend",
		Callback:           core.CallbackOnCastComplete,
		ClassFlags:         poisedBuff.EffectN(1).ClassFlags,
		TriggerImmediately: true,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			aura.Deactivate(sim)
		},
	})

	// Composure raises the buff's own cost reduction, which no spell the sim registers carries, so its
	// value joins the buff's modifier on the abilities the buff names.
	if rogue.Talents.Composure > 0 {
		aura.AttachSpellMod(core.SpellModConfig{
			Kind:       core.SpellMod_PowerCost_Flat,
			ClassFlags: poisedBuff.EffectN(1).ClassFlags,
			IntValue:   int32(spellData.Composure.ValueAt(rogue.Talents.Composure)),
		})
	}

	aura.ApplyOnReset(func(aura *core.Aura, sim *core.Simulation) {
		aura.Activate(sim)
		aura.SetStacks(sim, aura.MaxStacks)
	})
}
