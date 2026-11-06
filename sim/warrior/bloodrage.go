package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var bloodrageRank = spellData.Bloodrage.Highest()
var bloodrageOverTime = bloodrageRank.EffectN(2).Trigger()
var bloodrageOverTimeTick = bloodrageOverTime.PeriodicEffect()

func (warrior *Warrior) registerBloodrage() {
	// In game the instant rage makes 5 threat per rage in Battle and Defensive Stance alike, while the
	// rage over time takes the stance's multiplier.
	// TODO: Ingame test needed -- whether Defiance or a threat buff scales the instant rage's threat.
	instantRageMetrics := warrior.NewRageMetrics(core.ActionID{SpellID: bloodrageRank.ID})
	instantRageMetrics.UnscaledThreat = true
	rageMetrics := warrior.NewRageMetrics(core.ActionID{SpellID: bloodrageOverTime.ID})
	healthCost := warrior.BaseHealth() * float64(bloodrageRank.Power(dbcenums.POWER_HEALTH).CostPct) / 100
	improvedBloodrage := spellData.ImprovedBloodrage.MultiplierAt(warrior.Talents.ImprovedBloodrage)
	instantRage := bloodrageRank.EnergizeEffect().Tenths() * improvedBloodrage
	ragePerTick := bloodrageOverTimeTick.Tenths() * improvedBloodrage

	config := spelldata.SpellConfig(&warrior.Unit, bloodrageRank)

	config.Hot = spelldata.DotConfig(bloodrageOverTime, bloodrageOverTimeTick)
	config.Hot.SelfOnly = true
	config.Hot.OnTick = func(sim *core.Simulation, _ *core.Unit, _ *core.Dot) {
		warrior.AddRage(sim, ragePerTick, rageMetrics)
	}

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		warrior.AddRage(sim, instantRage, instantRageMetrics)
		warrior.SpendHealth(sim, healthCost, spell.HealthMetrics(&warrior.Unit))
		spell.SelfHot().Apply(sim)
	}

	spell := warrior.RegisterSpell(config)

	warrior.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
			return warrior.CurrentRage() < 70
		},
	})
}
