package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerPrayerOfMending(_ int32, rank *spelldata.Spell) {
	heal := rank.EffectN(1)
	triggerAura := spellData.PrayerOfMendingTriggered.ByID(401877)

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellHealing))
	config.ClassSpellMask = PriestSpellPrayerOfMending
	config.BonusCoefficient = heal.Coeff()
	var spell *core.Spell

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
		isPartyMember := false
		if priest.Party != nil {
			for _, member := range priest.Party.Players {
				if target == &member.GetCharacter().Unit {
					isPartyMember = true
					break
				}
			}
		}
		if !isPartyMember {
			target = &priest.Unit
		}

		charges := priest.PrayerOfMendingJumps
		if charges <= 0 {
			charges = int32(triggerAura.MaxStack)
		}
		charges = min(charges, int32(triggerAura.MaxStack))
		delay := min(5.0, priest.PrayerOfMendingDelay)
		if delay < 0 {
			delay = 0
		}

		var scheduleHeal func(*core.Simulation, *core.Unit, int32)
		scheduleHeal = func(sim *core.Simulation, healTarget *core.Unit, remainingJumps int32) {
			pa := sim.GetConsumedPendingActionFromPool()
			pa.NextActionAt = sim.CurrentTime + core.DurationFromSeconds(delay)
			pa.OnAction = func(sim *core.Simulation) {
				spell.CalcAndDealHealing(sim, healTarget, heal.Average(), spell.OutcomeHealingCrit)
				if remainingJumps <= 0 || priest.Party == nil {
					return
				}

				candidates := make([]*core.Unit, 0, len(priest.Party.Players))
				for _, member := range priest.Party.Players {
					candidate := &member.GetCharacter().Unit
					if candidate != healTarget {
						candidates = append(candidates, candidate)
					}
				}
				if len(candidates) == 0 {
					return
				}

				nextTarget := candidates[int(sim.RandomFloat("Prayer of Mending jump target")*float64(len(candidates)))]
				scheduleHeal(sim, nextTarget, remainingJumps-1)
			}
			sim.AddPendingAction(pa)
		}

		scheduleHeal(sim, target, charges-1)
	}

	spell = priest.RegisterSpell(config)
}
