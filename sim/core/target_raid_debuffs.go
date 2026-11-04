package core

import "time"

// The raid's debuffs on a target are the auras the Debuffs panel registers on every target, those with
// SourceExternalDebuff, which the target's reset puts on. The boss, the first target, gets them at the pull. Every
// add gets them by the encounter's settings each time it enters the fight: at the pull for an add up from the start,
// and at each spawn for one an AI spawns, which holds them while the add is down. A player's own cast of one of those
// debuffs is untouched by the hold.

// Holds or lands the raid's debuffs on each add with no Schedule. One up at the pull gets them by the encounter's
// settings. One disabled at the pull holds them for the fight, since only an AI brings a target in later. An AI that
// spawns adds sets their Schedule and schedules their debuffs itself through ScheduleAddRaidDebuffs.
func (encounter *Encounter) resetAddRaidDebuffs(sim *Simulation) {
	for _, target := range encounter.AllTargets[1:] {
		if target.Schedule != nil {
			continue
		}
		if target.IsEnabled() {
			encounter.ScheduleAddRaidDebuffs(sim, target)
		} else {
			target.DisableRaidDebuffs(sim)
		}
	}
}

// Schedules the raid's debuffs on an add entering the fight now, the encounter's delay ± its jitter later, or disables
// them for the fight when the encounter keeps them off the adds. An AI that spawns an add calls it at each spawn.
func (encounter *Encounter) ScheduleAddRaidDebuffs(sim *Simulation, target *Target) {
	if encounter.holdAddDebuffs {
		target.DisableRaidDebuffs(sim)
		return
	}
	target.ApplyRaidDebuffsAt(sim, sim.CurrentTime+encounter.rollAddDebuffDelay(sim))
}

// The delay ± the jitter, never below 0. Only a jitter draws a roll, on a stream apart from the combat rolls.
func (encounter *Encounter) rollAddDebuffDelay(sim *Simulation) time.Duration {
	if encounter.addDebuffJitter == 0 {
		return encounter.addDebuffDelay
	}
	offset := time.Duration((2*sim.addDebuffRand.NextFloat64() - 1) * float64(encounter.addDebuffJitter))
	return max(encounter.addDebuffDelay+offset, 0)
}

func (aura *Aura) isHeldRaidDebuff() bool {
	return aura.Source == SourceExternalDebuff && aura.Unit.raidDebuffsHeld
}

// Whether the raid's debuffs are disabled on the unit. A raid debuff on a timer of its own checks it at each tick.
func (unit *Unit) RaidDebuffsDisabled() bool {
	return unit.raidDebuffsHeld
}

// Takes the raid's debuffs off the target and keeps them off until ApplyRaidDebuffsAt lands them, through the
// target's reset when the disable comes during the sim's reset. A landing still to come is cancelled. Disable leaves a
// permanent aura on, so a target AI disables the debuffs of a target it takes away.
func (target *Target) DisableRaidDebuffs(sim *Simulation) {
	if !target.raidDebuffsLanding.consumed {
		target.raidDebuffsLanding.Cancel(sim)
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
	// Nothing is on or queued yet, so there is nothing to hold or release.
	if sim.isResetting && !target.aurasReset && at <= sim.CurrentTime {
		target.raidDebuffsHeld = false
		return
	}
	target.DisableRaidDebuffs(sim)
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
