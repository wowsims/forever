package core

import "time"

// The raid's debuffs on a target are the auras the Debuffs panel registers on every target, those with
// SourceExternalDebuff, which the target's reset puts on. An AI that spawns a target holds them while the target is
// down and lands them after each spawn. A player's own cast of one of those debuffs is untouched by the hold.

func (aura *Aura) isHeldRaidDebuff() bool {
	return aura.Source == SourceExternalDebuff && aura.Unit.raidDebuffsHeld
}

// Whether the unit holds the raid's debuffs off. A driver that puts a raid debuff on with a timer of its own checks it
// before each application.
func (unit *Unit) RaidDebuffsHeld() bool {
	return unit.raidDebuffsHeld
}

// Takes the raid's debuffs off the target and keeps them off until ApplyRaidDebuffsAt lands them, through the
// target's reset when the hold comes during the sim's reset. A landing still to come is cancelled. Disable leaves a
// permanent aura on, so a target AI holds the debuffs of a target it takes away.
func (target *Target) HoldRaidDebuffs(sim *Simulation) {
	target.raidDebuffsLanding.Cancel(sim)
	target.raidDebuffsHeld = true
	for _, aura := range target.auras {
		if aura.Source == SourceExternalDebuff {
			aura.Deactivate(sim)
		}
	}
}

// Holds the raid's debuffs until at, then puts them on as the target's reset does, each through its OnReset: a
// permanent debuff comes back, and Sunder Armor ramps up again. A time at or before now lands them at once; during
// the sim's reset the landing waits for 0, since the target's own aura reset may not have run yet.
func (target *Target) ApplyRaidDebuffsAt(sim *Simulation, at time.Duration) {
	target.HoldRaidDebuffs(sim)
	if at <= sim.CurrentTime && !sim.isResetting {
		target.landRaidDebuffs(sim)
		return
	}
	target.raidDebuffsLanding.NextActionAt = max(at, sim.CurrentTime)
	sim.AddPendingAction(&target.raidDebuffsLanding)
}

func (target *Target) landRaidDebuffs(sim *Simulation) {
	target.raidDebuffsHeld = false
	for _, aura := range target.auras {
		if aura.Source == SourceExternalDebuff && aura.OnReset != nil {
			aura.OnReset(aura, sim)
		}
	}
}
