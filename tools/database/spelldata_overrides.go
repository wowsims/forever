package database

import (
	"fmt"
	"math"
	"slices"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/tools/database/overrides"
)

// Writes the hand-supplied values onto the rows before they are rendered, so the sim reads one
// number per field instead of a game files number a call site patches. Every failure below stops the
// whole generation: a row the game files have since caught up with, or one that names a spell the
// store does not carry, is a number that has stopped being true and has to be looked at rather than
// written.
func applyOverrides(rows []storeSpell, list []overrides.Override) error {
	byID := map[int32]*storeSpell{}
	for i := range rows {
		byID[rows[i].ID] = &rows[i]
	}

	type key struct {
		id    int32
		field overrides.Field
	}
	seen := map[key]bool{}

	for _, o := range list {
		if o.Reason == "" {
			return fmt.Errorf("the %s override of spell %d states no reason", o.Field, o.SpellID)
		}
		if seen[key{o.SpellID, o.Field}] {
			return fmt.Errorf("spell %d carries two %s overrides", o.SpellID, o.Field)
		}
		seen[key{o.SpellID, o.Field}] = true

		row, ok := byID[o.SpellID]
		if !ok {
			return fmt.Errorf("the %s override names spell %d, which the store does not carry",
				o.Field, o.SpellID)
		}
		if err := applyOverride(row, o); err != nil {
			return err
		}
		row.overrideNotes = append(row.overrideNotes,
			fmt.Sprintf("override: %s %s -- %s", o.Field, num(o.Value), o.Reason))
	}
	return nil
}

func applyOverride(s *storeSpell, o overrides.Override) error {
	switch o.Field {
	case overrides.PPM:
		if s.procsPerMinuteID != 0 {
			return fmt.Errorf("spell %d now states SpellProcsPerMinuteID %d, so its hand-supplied PPM is stale",
				s.ID, s.procsPerMinuteID)
		}
		s.RPPM = o.Value

		// The rate is the whole reading: the column that said "fires on its own condition" is the
		// sentinel this override exists to answer.
		s.ProcChanceSource, s.ProcChanceEffect = procChancePPM, 0

	case overrides.FlatThreat:
		return applyFlatThreatOverride(s, o.Value)

	case overrides.ThreatPerSpellLevel:
		return applyFlatThreatOverride(s, o.Value*float64(s.SpellLevel))

	case overrides.APCoefDirect:
		return applyAPCoefOverride(s, o, s.directEffect(), "direct")

	case overrides.APCoefPeriodic:
		return applyAPCoefOverride(s, o, s.periodicEffect(), "periodic")

	case overrides.APCoefThreat:
		return applyAPCoefOverride(s, o, s.threatEffect(), "threat")

	case overrides.ProcChancePct:
		if s.tooltipStatesChance {
			return fmt.Errorf("spell %d states its own proc chance in the tooltip, so the hand-supplied %v%% is stale",
				s.ID, o.Value)
		}
		if o.Value != math.Trunc(o.Value) || o.Value < 0 || o.Value > 100 {
			return fmt.Errorf("spell %d has a proc chance override of %v, which is no whole percentage",
				s.ID, o.Value)
		}
		s.ProcChance = uint8(o.Value)
		s.ProcChanceSource, s.ProcChanceEffect = procChanceColumn, 0

	case overrides.DurationMs:
		if s.DurationMs != 0 {
			return fmt.Errorf("spell %d now states a duration of %d ms, so its hand-supplied one is stale",
				s.ID, s.DurationMs)
		}
		s.DurationMs = int32(o.Value)

	case overrides.ICDMs:
		if s.ICDMs != 0 {
			return fmt.Errorf("spell %d now states an internal cooldown of %d ms, so its hand-supplied one is stale",
				s.ID, s.ICDMs)
		}
		s.ICDMs = int32(o.Value)

	case overrides.Hint:
		s.ProcHint = core.ProcHint(o.Value)

	case overrides.StatMask:
		i := slices.IndexFunc(s.Effects, func(e storeEffect) bool { return e.Aura == dbcenums.A_MOD_TOTAL_STAT_PERCENTAGE })
		if i < 0 {
			return fmt.Errorf("spell %d has no A_MOD_TOTAL_STAT_PERCENTAGE effect to carry a stat mask", s.ID)
		}
		e := &s.Effects[i]
		if e.Misc != 0 || e.Misc2 != 0 {
			return fmt.Errorf("spell %d now states misc %d and %d on its stat percentage, so its hand-supplied stat mask is stale",
				s.ID, e.Misc, e.Misc2)
		}
		e.Misc2 = int32(o.Value)

	default:
		return fmt.Errorf("spell %d carries an override of unknown field %d", s.ID, uint8(o.Field))
	}
	return nil
}

func applyFlatThreatOverride(s *storeSpell, threat float64) error {
	if e := s.threatEffect(); e != nil {
		return fmt.Errorf("spell %d now states threat on effect %d, so its hand-supplied flat threat is stale",
			s.ID, e.Index)
	}
	if s.FlatThreat != 0 {
		return fmt.Errorf("spell %d carries two flat threat overrides", s.ID)
	}
	s.FlatThreat = threat
	return nil
}

func applyAPCoefOverride(s *storeSpell, o overrides.Override, e *storeEffect, role string) error {
	if e == nil {
		return fmt.Errorf("spell %d has no %s effect to carry an attack-power coefficient", s.ID, role)
	}
	if e.APCoef != 0 {
		return fmt.Errorf("spell %d now states %v attack power on its %s effect, so its hand-supplied coefficient is stale",
			s.ID, e.APCoef, role)
	}
	e.APCoef = o.Value
	return nil
}

func (s *storeSpell) threatEffect() *storeEffect {
	for i := range s.Effects {
		if IsThreatEffect(s.Effects[i].Type) {
			return &s.Effects[i]
		}
	}
	return nil
}

// The effect an override means by "direct": the damage or heal the cast itself does. A damage
// effect that ticks belongs to the DoT, which is why a period disqualifies one here.
func (s *storeSpell) directEffect() *storeEffect {
	for i := range s.Effects {
		e := &s.Effects[i]
		if e.PeriodMs != 0 {
			continue
		}
		effect := e.Type
		if effect == dbcenums.E_SCHOOL_DAMAGE || effect == dbcenums.E_HEAL || IsWeaponDamageEffect(effect) {
			return e
		}
	}
	return nil
}

// Stale once the row states its own RequiredAreasID.
func applyAreaBonuses(rows []storeSpell, list []overrides.AreaBonus) error {
	byID := map[int32]*storeSpell{}
	for i := range rows {
		byID[rows[i].ID] = &rows[i]
	}

	seen := map[int32]bool{}
	for _, b := range list {
		if b.Reason == "" {
			return fmt.Errorf("the area bonus of spell %d states no reason", b.SpellID)
		}
		if seen[b.SpellID] {
			return fmt.Errorf("spell %d carries two area bonuses", b.SpellID)
		}
		seen[b.SpellID] = true
		if len(b.Groups) == 0 || b.Multiplier <= 0 || b.DurationMultiplier <= 0 {
			return fmt.Errorf("the area bonus of spell %d names no area group or a factor of zero", b.SpellID)
		}

		row, ok := byID[b.SpellID]
		if !ok {
			return fmt.Errorf("the area bonus names spell %d, which the store does not carry", b.SpellID)
		}
		if row.RequiredAreas != 0 {
			return fmt.Errorf("spell %d now states area group %d itself, so its hand-supplied area bonus is stale",
				b.SpellID, row.RequiredAreas)
		}
		row.AreaBonusGroups = b.Groups
		row.AreaMultiplier, row.AreaDurationMultiplier = b.Multiplier, b.DurationMultiplier
		row.overrideNotes = append(row.overrideNotes,
			fmt.Sprintf("override: AreaBonus x%s (duration x%s) in %v -- %s",
				num(b.Multiplier), num(b.DurationMultiplier), b.Groups, b.Reason))
	}
	return nil
}

// Negates the curve at each listed position, after checking every rank on it still disagrees in sign
// with the row's base points.
func applyCurveSigns(rows []storeSpell, curves map[int32][][]float64, list []overrides.CurveSign) error {
	byID := map[int32]*storeSpell{}
	for i := range rows {
		byID[rows[i].ID] = &rows[i]
	}

	type key struct {
		id     int32
		effect int
	}
	seen := map[key]bool{}

	for _, c := range list {
		if c.Reason == "" {
			return fmt.Errorf("the curve sign of spell %d effect %d states no reason", c.SpellID, c.Effect)
		}
		if seen[key{c.SpellID, c.Effect}] {
			return fmt.Errorf("spell %d carries two curve signs on effect %d", c.SpellID, c.Effect)
		}
		seen[key{c.SpellID, c.Effect}] = true

		row, ok := byID[c.SpellID]
		if !ok {
			return fmt.Errorf("the curve sign names spell %d, which the store does not carry", c.SpellID)
		}
		position := c.Effect - 1
		if position < 0 || position >= len(curves[c.SpellID]) || position >= len(row.Effects) {
			return fmt.Errorf("spell %d has no curve at effect %d to take the row's sign", c.SpellID, c.Effect)
		}
		base := row.Effects[position].BasePoints
		if base == 0 {
			return fmt.Errorf("spell %d states base points 0 on effect %d, so there is no sign for its curve to take",
				c.SpellID, c.Effect)
		}
		curve := curves[c.SpellID][position]
		for _, value := range curve {
			if value == 0 || (value < 0) == (base < 0) {
				return fmt.Errorf("spell %d's curve on effect %d reads %v against base points %v, so its hand-supplied sign is stale",
					c.SpellID, c.Effect, curve, base)
			}
		}
		for i := range curve {
			curve[i] = -curve[i]
		}
		row.overrideNotes = append(row.overrideNotes,
			fmt.Sprintf("override: CurveSign effect %d -- %s", c.Effect, c.Reason))
	}
	return nil
}

// The ticking effect: a periodic aura, or the damage effect the rank's dummy times, which is how
// the rank tables read a DoT too.
func (s *storeSpell) periodicEffect() *storeEffect {
	for i := range s.Effects {
		e := &s.Effects[i]
		if IsPeriodicAura(e.Aura) || (e.Type == dbcenums.E_SCHOOL_DAMAGE && e.PeriodMs > 0) {
			return e
		}
	}
	return nil
}
