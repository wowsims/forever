package core

import "time"

// The raid's debuffs on a target are the auras the Debuffs panel registers on every target, those with
// SourceExternalDebuff, which the target's reset puts on. An AI that spawns a target holds them while the target is
// down and lands them after each spawn. A player's own cast of one of those debuffs is untouched by the hold.

func (aura *Aura) isHeldRaidDebuff() bool {
	return aura.Source == SourceExternalDebuff && aura.Unit.raidDebuffsHeld
}

// Whether the unit holds the raid's debuffs off. A raid debuff on a timer of its own checks it at each tick.
func (unit *Unit) RaidDebuffsHeld() bool {
	return unit.raidDebuffsHeld
}

// Takes the raid's debuffs off the target and keeps them off until ApplyRaidDebuffsAt lands them, through the
// target's reset when the hold comes during the sim's reset. A landing still to come is cancelled. Disable leaves a
// permanent aura on, so a target AI holds the debuffs of a target it takes away.
func (target *Target) HoldRaidDebuffs(sim *Simulation) {
	target.raidDebuffsLanding.Cancel(sim)
	target.raidDebuffsHeld = true
	for _, aura := range target.raidDebuffAuras() {
		aura.Deactivate(sim)
	}
}

// Holds the raid's debuffs until at, then puts them on as the target's reset does, each through its OnReset: a
// permanent debuff comes back, and Sunder Armor ramps up again. A time at or before now lands them at once. During
// the sim's reset, before the target's own reset has run, it only releases the hold, and that reset puts them on, so
// a player's Reset handler and the prepull see them.
func (target *Target) ApplyRaidDebuffsAt(sim *Simulation, at time.Duration) {
	target.HoldRaidDebuffs(sim)
	if at > sim.CurrentTime {
		target.raidDebuffsLanding.NextActionAt = at
		sim.AddPendingAction(&target.raidDebuffsLanding)
		return
	}
	if sim.isResetting && !target.aurasReset {
		target.raidDebuffsHeld = false
		return
	}
	target.landRaidDebuffs(sim)
}

func (target *Target) landRaidDebuffs(sim *Simulation) {
	target.raidDebuffsHeld = false
	for _, aura := range target.raidDebuffAuras() {
		if aura.OnReset != nil {
			aura.OnReset(aura, sim)
		}
	}
}

// Every raid debuff registers while the environment is built, so the list gathered at the first call holds them all.
func (target *Target) raidDebuffAuras() []*Aura {
	if target.raidDebuffs == nil {
		target.raidDebuffs = []*Aura{}
		for _, aura := range target.auras {
			if aura.Source == SourceExternalDebuff {
				target.raidDebuffs = append(target.raidDebuffs, aura)
			}
		}
	}
	return target.raidDebuffs
}
