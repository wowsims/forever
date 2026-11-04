package core

import "time"

// Shoot's scheduler. Shoot applies a channel that carries the channeling state (Is Channeling, the
// rotation's hold, the timeline bar) but never ticks; the wand fires the shots at its own slots.
type Wand struct {
	unit  *Unit
	shoot *Spell
	speed func() time.Duration // The equipped wand's speed; zero with no wand.
	shot  func(sim *Simulation, target *Unit)

	target      *Unit
	interruptIf APLValue      // The channel action's condition; nil for a cast, which fires one shot.
	interval    time.Duration // Fixed when Shoot starts.
	slot        *PendingAction
	lastShotAt  time.Duration
	timerAt     time.Duration // When the wand's timer last restarted: a rotation cast, another channel's end, or a shot.
}

const wandWindup = 500 * time.Millisecond

// Shoot's channel. The wand sets its length to the rest of the fight when Shoot starts; ChanneledTicks(Shoot)
// reads 0, since the shots are not its ticks. NumberOfTicks stays 1: the rotation and the dot end a channel
// with no ticks left.
func WandChannel() DotConfig {
	return DotConfig{
		Aura:          Aura{Label: "Shoot"},
		NumberOfTicks: 1,
		TickLength:    wandWindup,
	}
}

// Makes shoot the unit's wand Shoot, which starts wanding once its own ApplyEffects has run and fires
// shot at each of its slots; Shoot cannot be cast while it runs.
func (unit *Unit) EnableWand(shoot *Spell, swingSpeed func() time.Duration, shot func(sim *Simulation, target *Unit)) {
	wand := &Wand{unit: unit, shoot: shoot, speed: swingSpeed, shot: shot}
	unit.Wand = wand
	unit.RegisterResetEffect(func(_ *Simulation) {
		wand.lastShotAt = 0
		wand.timerAt = -NeverExpires
	})

	castCondition := shoot.ExtraCastCondition
	shoot.ExtraCastCondition = func(sim *Simulation, target *Unit) bool {
		return !wand.wanding() && (castCondition == nil || castCondition(sim, target))
	}
	applyEffects := shoot.ApplyEffects
	shoot.ApplyEffects = func(sim *Simulation, target *Unit, spell *Spell) {
		applyEffects(sim, target, spell)
		wand.start(sim, target)
	}

	unit.OnChannelEnd(func(sim *Simulation, dot *Dot) {
		if dot.Spell != shoot {
			wand.timerAt = sim.CurrentTime
			return
		}
		wand.slot.Cancel(sim)
		// The rotation acts at the slot with no clip delay, or when the GCD ends: a wake-up that came
		// while it was held is spent.
		if unit.GCD.IsReady(sim) {
			unit.WaitUntil(sim, sim.CurrentTime)
		} else if unit.rotationAction.IsConsumed() || unit.NextRotationActionAt() > unit.GCD.ReadyAt() {
			unit.WaitUntil(sim, unit.GCD.ReadyAt())
		}
	})

	// Starting to move ends wanding, with no shot.
	unit.RegisterMovementCallback(func(sim *Simulation, _ float64, kind MovementUpdateType) {
		if kind == MovementStart && wand.wanding() {
			unit.ChanneledDot.Deactivate(sim)
		}
	})
}

// Starts wanding at target. The first shot comes after the wind-up, and one wand speed after the timer
// last restarted. A channel action's interruptIf is taken before the channel starts, so the rotation
// never holds it and never cuts Shoot between slots.
func (wand *Wand) start(sim *Simulation, target *Unit) {
	wand.interruptIf = nil
	if action := wand.unit.takeChannelAction(wand.shoot); action != nil {
		wand.interruptIf = action.interruptIf
	}

	wand.target = target
	wand.interval = wand.swingSpeed()
	firstShotAt := max(sim.CurrentTime+wandWindup, wand.timerAt+wand.interval)

	dot := wand.shoot.Dot(target)
	dot.BaseTickLength = max(sim.GetRemainingDuration(), firstShotAt-sim.CurrentTime+wand.interval)
	dot.Apply(sim)
	dot.stopTicking(sim)

	if wand.slot == nil {
		wand.slot = &PendingAction{OnAction: wand.fire}
	}
	wand.slot.NextActionAt = firstShotAt
	sim.AddPendingAction(wand.slot)
}

// A slot that an interrupt takes fires no shot. A cast rather than a channel action fires one shot and stops.
func (wand *Wand) fire(sim *Simulation) {
	dot := wand.unit.ChanneledDot
	if wand.interruptIf != nil && wand.slotTaken(sim) {
		dot.Deactivate(sim)
		return
	}

	wand.shot(sim, wand.target)
	wand.lastShotAt = sim.CurrentTime
	wand.timerAt = sim.CurrentTime
	if !wand.wanding() {
		return
	}
	if wand.interruptIf == nil {
		dot.Deactivate(sim)
		return
	}

	wand.slot.NextActionAt = sim.CurrentTime + wand.interval
	// A fight that runs past its expected length keeps the channel going.
	if dot.ExpiresAt() <= wand.slot.NextActionAt {
		dot.UpdateExpires(wand.slot.NextActionAt + wand.interval)
	}
	sim.AddPendingAction(wand.slot)
}

// Whether the rotation takes the slot: interruptIf holds and the rotation, with Shoot ended, would do
// something other than Shoot. With nothing else ready Shoot keeps shooting. Inside a strict sequence
// (which Shoot started) the slot goes to its current subaction once that is ready, read without asking
// the sequence for its next action, which moves it on; any other controlling action takes the slot.
func (wand *Wand) slotTaken(sim *Simulation) bool {
	rot := wand.unit.Rotation
	rot.evalGeneration++
	if !wand.interruptIf.GetBool(sim) {
		return false
	}
	if step, controlled := rot.controlledStep(); controlled {
		return step == nil || (step.IsReady(sim) && !step.casts(wand.shoot))
	}

	var next *APLAction
	rot.withChannelEnded(func() {
		next = rot.getNextAction(sim)
	})
	return next != nil && !next.casts(wand.shoot)
}

// A completed cast by the rotation or a major cooldown of any spell but Shoot.
func (wand *Wand) castCompleted(sim *Simulation, spell *Spell) {
	if wand != nil && spell != wand.shoot {
		wand.timerAt = sim.CurrentTime
	}
}

func (wand *Wand) wanding() bool {
	return wand != nil && wand.unit.ChanneledDot != nil && wand.unit.ChanneledDot.Spell == wand.shoot
}

// Shoot's channel or one of its shots, which share its spell id.
func (wand *Wand) IsShoot(spell *Spell) bool {
	return wand != nil && spell.ActionID.SpellID == wand.shoot.ActionID.SpellID
}

func (wand *Wand) nextShotAt() (time.Duration, bool) {
	if wand.wanding() {
		return wand.slot.NextActionAt, true
	}
	return 0, false
}

// Zero while idle; otherwise the time to the next slot.
func (wand *Wand) timeToNext(sim *Simulation) time.Duration {
	if at, ok := wand.nextShotAt(); ok {
		return max(0, at-sim.CurrentTime)
	}
	return 0
}

func (wand *Wand) timeSinceLast(sim *Simulation) time.Duration {
	if wand == nil {
		return 0
	}
	return sim.CurrentTime - wand.lastShotAt
}

// The running Shoot's shot interval, fixed when it started; otherwise the wand speed at the current cast
// speed, the interval of a Shoot started now. Zero with no wand.
func (wand *Wand) swingSpeed() time.Duration {
	if wand == nil {
		return 0
	}
	if wand.wanding() {
		return wand.interval
	}
	speed := wand.speed()
	if speed == 0 {
		return 0
	}
	return wand.unit.ApplyCastSpeedForSpell(speed, wand.shoot).Round(time.Millisecond)
}
