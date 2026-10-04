package buffs

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The readers a Meta and the judgement ranks turn a store row into a buff's numbers with. An amount
// is the caster's at core.CharacterLevel, in the game files' units.

func amount(e *spelldata.Effect) float64 {
	return e.Average()
}

// The game files state a permanent aura as -1 and an aura with no duration of its own, a totem's,
// as 0.
func auraDuration(s *spelldata.Spell) time.Duration {
	if s.DurationMs <= 0 {
		return core.NeverExpires
	}
	return s.Duration()
}

func cooldown(s *spelldata.Spell) time.Duration {
	return max(s.Cooldown(), s.CategoryCooldown())
}

func talentScaledDuration(s *spelldata.Spell, mod *spelldata.Effect) time.Duration {
	if s.DurationMs <= 0 {
		return core.NeverExpires
	}
	return time.Duration(spelldata.Scaled(float64(s.DurationMs), mod)) * time.Millisecond
}
