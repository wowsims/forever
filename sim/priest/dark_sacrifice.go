package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func (priest *Priest) registerDarkSacrifice(_ int32, rank *spelldata.Spell) {
	healthPerTick := rank.Effect(dbcenums.A_PERIODIC_DUMMY, 0).Average()
	manaTick := rank.FirstAura(dbcenums.A_PERIODIC_ENERGIZE)
	manaPerTick := manaTick.Average()
	manaMetrics := priest.NewManaMetrics(core.ActionID{SpellID: rank.ID})

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Flags(core.SpellFlagAPL))

	config.Hot = spelldata.DotConfig(rank, manaTick, spelldata.RankLabel(rank, &priest.Unit))
	config.Hot.SelfOnly = true
	ticks := float64(config.Hot.NumberOfTicks)
	rowMana := manaPerTick * ticks
	config.Hot.OnTick = func(sim *core.Simulation, _ *core.Unit, _ *core.Dot) {
		// The tooltip's ${$o2+$SPI}: the row's mana over the duration, and Spirit once, shared over the ticks.
		priest.RemoveHealth(sim, healthPerTick)
		priest.AddMana(sim, manaPerTick+priest.GetStat(stats.Spirit)/ticks, manaMetrics)
	}

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.SelfHot().Apply(sim)
	}

	darkSacrifice := priest.RegisterSpell(config)
	if rank != spellData.DarkSacrifice.Highest() {
		return
	}

	priest.AddMajorCooldown(core.MajorCooldown{
		Spell: darkSacrifice,
		Type:  core.CooldownTypeMana,
		ShouldActivate: func(_ *core.Simulation, character *core.Character) bool {
			return character.MaxMana()-character.CurrentMana() >= rowMana+character.GetStat(stats.Spirit)
		},
	})
}
