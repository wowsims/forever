package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerPrayerOfHealing(_ int32, rank *spelldata.Spell) {
	heal := rank.HealEffect()

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellHealing))
	config.ClassSpellMask = PriestSpellPrayerOfHealing
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		party := priest.Party
		if targetAgent := priest.Env.Raid.GetPlayerFromUnitIndex(target.UnitIndex); targetAgent != nil {
			party = targetAgent.GetCharacter().Party
		}
		if party == nil || len(party.Players) == 0 {
			return
		}

		targets := make([]*core.Unit, 0, min(5, len(party.Players)))
		for _, member := range party.Players {
			targets = append(targets, &member.GetCharacter().Unit)
		}

		if len(targets) > 5 {
			targets = targets[:5]
		}
		if priest.PrayerOfHealingMode == proto.PrayerOfHealingMode_Dynamic && len(targets) > 1 {
			count := 1 + int(sim.RandomFloat("Prayer of Healing target count")*float64(len(targets)))
			for i := len(targets) - 1; i > 0; i-- {
				j := int(sim.RandomFloat("Prayer of Healing target selection") * float64(i+1))
				targets[i], targets[j] = targets[j], targets[i]
			}
			targets = targets[:count]
		}

		for _, target := range targets {
			spell.CalcAndDealHealing(sim, target, heal.Average(), spell.OutcomeHealingCrit)
		}
	}

	priest.RegisterSpell(config)
}
