package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var shieldBlockRank = spellData.ShieldBlock.Highest()

func (warrior *Warrior) registerShieldBlock() {
	aura := warrior.RegisterAura(spelldata.AuraConfig(shieldBlockRank))
	spelldata.ParseEffects(&warrior.Character, aura, shieldBlockRank)

	// The block that spends the last charge leaves the aura up to the end of its spell batch, so the
	// other attacks landing in that batch are blocked too and spend nothing.
	lastChargeSpent := false
	batchEnd := &core.PendingAction{
		Priority: core.ActionPriorityDOT,
		OnAction: func(sim *core.Simulation) {
			lastChargeSpent = false
			aura.Deactivate(sim)
		},
	}
	// A recast restores both charges and an expiry ends the aura, so either drops a pending batch end.
	dropBatchEnd := func(sim *core.Simulation) {
		if lastChargeSpent {
			batchEnd.Cancel(sim)
			lastChargeSpent = false
		}
	}
	aura.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
		dropBatchEnd(sim)
	})

	// The row states the melee and ranged hits the listener hears and the two charges the aura
	// starts with. The block that spends a charge is an outcome no proc mask states, and a full
	// block deals no damage, so both of those are the caller's.
	trigger := spelldata.ProcTrigger(&warrior.Character, shieldBlockRank,
		func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			if aura.GetStacks() > 1 {
				aura.RemoveStack(sim)
				return
			}
			if lastChargeSpent {
				return
			}
			lastChargeSpent = true
			batchEnd.NextActionAt = sim.CurrentTime + core.SpellBatchWindow
			sim.AddPendingAction(batchEnd)
		})
	trigger.Outcome = core.OutcomeBlock
	trigger.RequireDamageDealt = false
	trigger.TriggerImmediately = true
	aura.AttachProcTrigger(trigger)

	config := spelldata.SpellConfig(&warrior.Unit, shieldBlockRank, spelldata.Flags(core.SpellFlagAPL))

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		dropBatchEnd(sim)
		aura.Activate(sim)
		aura.SetStacks(sim, aura.MaxStacks)
	}

	config.RelatedSelfBuff = aura

	warrior.RegisterSpell(config)

	warrior.deactivateWithoutShield(aura)
}

func (warrior *Warrior) deactivateWithoutShield(aura *core.Aura) {
	warrior.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotOffHand}, func(sim *core.Simulation, _ proto.ItemSlot) {
		if !warrior.PseudoStats.CanBlock {
			aura.Deactivate(sim)
		}
	})
}
