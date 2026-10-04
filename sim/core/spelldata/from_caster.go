package spelldata

import (
	"github.com/wowsims/forever/sim/core"
)

// The slot a parse holds on its caster's attack table against the aura's unit: see FromCaster.
type casterSlot struct {
	caster, target *core.Unit
	index, total   int

	damage, crit     []*casterPart
	damageOn, critOn bool
}

// One effect in the slot: whether it is up, what it is worth at its current level and the spells it
// reaches.
type casterPart struct {
	up      bool
	worth   float64
	reaches func(*core.Spell) bool
}

// A damage effect that reaches only the caster's attacks: a percentage the game applies once per
// stack, so a level of 2 is worth 1 + 2*value/100.
func (p *parser) casterDamage(kind string, v float64, reaches func(*core.Spell) bool) *attachment {
	a := &attachment{kind: kind, value: core.MultiplierFromPercent(v), multiplicative: true}
	return p.casterEffect(a, reaches, func(s *casterSlot) *[]*casterPart { return &s.damage },
		func(level float64) float64 { return core.MultiplierFromPercent(v * level) })
}

// A crit effect that reaches only the caster's attacks: percentage points, once per stack.
func (p *parser) casterCrit(kind string, v float64, reaches func(*core.Spell) bool) *attachment {
	a := &attachment{kind: kind, value: v}
	return p.casterEffect(a, reaches, func(s *casterSlot) *[]*casterPart { return &s.crit },
		func(level float64) float64 { return v * level })
}

// The effect as one of the slot's parts, in the handler parts picks, priced by worth at each level.
// The static path has no aura on a target, and without FromCaster there is no caster, so both skip
// the row.
func (p *parser) casterEffect(a *attachment, reaches func(*core.Spell) bool,
	parts func(*casterSlot) *[]*casterPart, worth func(level float64) float64) *attachment {
	if p.casterSlot == nil || p.static {
		return nil
	}
	if p.dry {
		return a
	}

	slot := p.casterSlot
	part := &casterPart{reaches: reaches}
	list := parts(slot)
	*list = append(*list, part)
	a.set = func(_ *core.Simulation, level float64) {
		part.up, part.worth = level > 0, worth(level)
		slot.update()
	}
	return a
}

// Enables or disables each handler to match whether any of its effects is up. The attack tables
// exist once the environment is finalized, which is before any aura can be gained.
func (s *casterSlot) update() {
	table := s.caster.AttackTables[s.target.UnitIndex]

	if on := anyUp(s.damage); on != s.damageOn {
		s.damageOn = on
		if on {
			core.EnableDamageDoneByCaster(s.index, s.total, table, s.damageMultiplier)
		} else {
			core.DisableDamageDoneByCaster(s.index, table)
		}
	}

	if on := anyUp(s.crit); on != s.critOn {
		s.critOn = on
		if on {
			core.EnableCritChanceByCaster(s.index, s.total, table, s.critChance)
		} else {
			core.DisableCritChanceByCaster(s.index, table)
		}
	}
}

func (s *casterSlot) damageMultiplier(_ *core.Simulation, spell *core.Spell, _ *core.AttackTable) float64 {
	mult := 1.0
	for _, part := range s.damage {
		if part.up && part.reaches(spell) {
			mult *= part.worth
		}
	}
	return mult
}

func (s *casterSlot) critChance(spell *core.Spell) float64 {
	bonus := 0.0
	for _, part := range s.crit {
		if part.up && part.reaches(spell) {
			bonus += part.worth
		}
	}
	return bonus
}

func anyUp(parts []*casterPart) bool {
	for _, part := range parts {
		if part.up {
			return true
		}
	}
	return false
}
