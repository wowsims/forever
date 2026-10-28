package core

// Kills the target. One roll on the design rand against killShare decides whether the raid gets the
// kill; a share of 0 or 1 settles it without a draw. Every player's and pet's OnTargetDies runs first, then OnKill when
// the raid gets the kill, and the target is disabled last, so its debuffs and the channels on it are
// still there while the callbacks run. A disabled target does not die again.
func (target *Target) Die(sim *Simulation, killShare float64) {
	if !target.IsEnabled() {
		return
	}

	credited := killShare >= 1 || (killShare > 0 && sim.designRand.NextFloat64() < killShare)
	if sim.Log != nil {
		target.Log(sim, "Died (kill credited: %t)", credited)
	}

	for _, unit := range sim.Raid.AllUnits {
		unit.OnTargetDies(sim, &target.Unit)
	}
	if credited {
		for _, unit := range sim.Raid.AllUnits {
			unit.OnKill(sim, &target.Unit)
		}
	}

	target.Disable(sim, true)
}
