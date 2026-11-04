package core

import "time"

// The raid's debuffs on a target are the auras the Debuffs panel registers on every target, those with
// SourceExternalDebuff, which the target's reset puts on. The boss, the first target, gets them at the pull. Every
// add gets them by the encounter's settings each time it enters the fight: at the pull for an add up from the start,
// and at each spawn for one an AI spawns, which holds them while the add is down. A player's own cast of one of those
// debuffs is untouched by the hold.

// Holds or lands the raid's debuffs on each add up at the pull that no AI spawns.
func (encounter *Encounter) resetAddRaidDebuffs(sim *Simulation) {
	for _, target := range encounter.AllTargets[1:] {
		if target.Schedule == nil && target.IsEnabled() {
			encounter.LandAddRaidDebuffs(sim, target)
		}
	}
}

// Lands the raid's debuffs on an add entering the fight now, the encounter's delay ± its jitter later, or holds them
// for the fight when the encounter keeps them off the adds. An AI that spawns an add calls it at each spawn.
func (encounter *Encounter) LandAddRaidDebuffs(sim *Simulation, target *Target) {
	if encounter.holdAddDebuffs {
		target.HoldRaidDebuffs(sim)
		return
	}
	target.ApplyRaidDebuffsAt(sim, sim.CurrentTime+encounter.rollAddDebuffDelay(sim))
}

// The delay ± the jitter, never below 0. Only a jitter draws a roll.
func (encounter *Encounter) rollAddDebuffDelay(sim *Simulation) time.Duration {
	if encounter.addDebuffJitter == 0 {
		return encounter.addDebuffDelay
	}
	offset := time.Duration((2*sim.RandomFloat("Add Debuff Jitter") - 1) * float64(encounter.addDebuffJitter))
	return max(encounter.addDebuffDelay+offset, 0)
}

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
	if target.raidDebuffsQueued {
		target.raidDebuffsLanding.Cancel(sim)
		target.raidDebuffsQueued = false
	}
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
		target.raidDebuffsQueued = true
		return
	}
	if sim.isResetting && !target.aurasReset {
		target.raidDebuffsHeld = false
		return
	}
	target.landRaidDebuffs(sim)
}

func (target *Target) landRaidDebuffs(sim *Simulation) {
	target.raidDebuffsQueued = false
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
