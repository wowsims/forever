package rogue

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var sliceAndDiceRank = spellData.SliceAndDice.Highest()

func (rogue *Rogue) registerSliceAndDice() {
	config := spelldata.SpellConfig(&rogue.Unit, sliceAndDiceRank, spelldata.Melee(core.ProcMaskEmpty))
	config.ExtraCastCondition = rogue.FinisherExtraCastCondition
	talentModifier := spellData.ImprovedSliceAndDice.MultiplierAt(rogue.Talents.ImprovedSliceAndDice)

	sndAura := rogue.RegisterAura(spelldata.AuraConfig(sliceAndDiceRank))
	spelldata.ParseEffects(&rogue.Character, sndAura, sliceAndDiceRank)

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim) // This was true in Vanilla, but needs to be tested as future expansions may change this behavior.
		baseDuration := sliceAndDiceRank.Duration().Seconds()
		pointDuration := 3.0 * float64(rogue.ComboPoints())
		fullDuration := (baseDuration + pointDuration) * talentModifier
		sndAura.Duration = time.Duration(fullDuration*1000) * time.Millisecond
		rogue.ApplyFinisher(sim, spell)
		sndAura.Activate(sim)
	}

	config.MetricSplits = 6
	config.Cast.ModifyCast = func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
		spell.SetMetricsSplit(rogue.ComboPoints())
	}

	rogue.SliceAndDice = rogue.RegisterSpell(config)
}
