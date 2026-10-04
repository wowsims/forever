package spelldata

import (
	"fmt"
	"math"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/stats"
)

// The rows a raid buff reads and nothing else does, which BuffAuras turns on. A class's own spell
// states these auras for a resource or a value the class wires itself: Bloodrage's rage over time is
// an A_PERIODIC_ENERGIZE, and Berserker Stance states an attack power percentage of 0.
var buffAuraTable = map[dbcenums.EffectAuraType]row{
	// Mana over time, which the sim keeps as mana per five seconds. The game files' tick is whole;
	// what the conversion makes of it is not, and it is not truncated.
	dbcenums.A_PERIODIC_ENERGIZE: func(p *parser, e *Effect, v float64) *attachment {
		if e.Misc != int32(dbcenums.POWER_MANA) || e.PeriodMs <= 0 {
			return nil
		}
		return p.statBuff(stats.MP5, v*5000/float64(e.PeriodMs))
	},

	dbcenums.A_MOD_ATTACK_POWER_PCT: func(p *parser, e *Effect, v float64) *attachment {
		return p.statMultiplier([]stats.Stat{stats.AttackPower}, core.MultiplierFromPercent(v))
	},
}

// The options every raid buff and debuff is parsed with: priced at core.CharacterLevel, with the
// buff table's rows, less the auras skip names, and a finisher at full combo points where
// fullComboPoints says so.
func RaidBuffOptions(skip []dbcenums.EffectAuraType, fullComboPoints bool) []ParseOpt {
	opts := []ParseOpt{Level(core.CharacterLevel), BuffAuras()}
	if len(skip) > 0 {
		opts = append(opts, SkipAuras(skip...))
	}
	if fullComboPoints {
		opts = append(opts, FullComboPoints())
	}
	return opts
}

// What a buff built from row s leaves out, one note per aura effect the parse skipped. An empty answer
// means every aura effect the row states is attached. A generator reads the notes off DryRun, so a
// buff it writes is one the sim builds the same way.
func (p *Parsed) SkippedNotes(s *Spell) []string {
	if s == nil || s == Nil || s.ID == 0 {
		return []string{"no row in the store"}
	}

	var notes []string
	for _, e := range p.Skipped {
		if e.PointsPerResource != 0 && e.BasePoints == 0 {
			notes = append(notes, fmt.Sprintf("effect %d is worth %v per combo point", e.Position, e.PointsPerResource))
			continue
		}
		notes = append(notes, effectNote(e.Position, e.Effect)+reasonNote(e.Reason))
	}
	return notes
}

// An improving talent's modifier on an amount, truncated the way the game resolves one: a
// percentage of it, or a flat addition. An untaken talent is NilEffect, which adds nothing.
func Scaled(amount float64, mod *Effect) float64 {
	if mod.Aura == dbcenums.A_ADD_PCT_MODIFIER {
		return math.Trunc(amount * core.MultiplierFromPercent(mod.BaseValue()))
	}
	return math.Trunc(amount + mod.BaseValue())
}
