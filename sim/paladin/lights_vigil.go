package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var LightsVigilRankMap = spellData.LightsVigil

// The ally's vigil aura and the party heal each rank triggers. They sit in LightsVigilTriggered in
// the order the game files list them, not by rank, so the pairing is by hand.
var lightsVigilTriggered = map[int32]struct{ aura, heal int32 }{
	1: {aura: 1310909, heal: 1310912},
	2: {aura: 1311593, heal: 1311591},
	3: {aura: 1311597, heal: 1311596},
}

// One rank's vigils on allies, and the party heal that consuming one casts.
type lightsVigil struct {
	auras core.AuraArray
	heal  *core.Spell
}

// Light's Vigil (talent)
// https://www.wowhead.com/forever/spell=1311595
//
// Applies Light's Vigil to a friendly target for 30 sec. The caster's next Holy Shock on the target
// incurs no cooldown and heals the target and their party for 700 to 740. Only one Vigil per Paladin
// may be active in each party.
func (paladin *Paladin) registerLightsVigil() {
	LightsVigilRankMap.Each(func(_ int32, rank *spelldata.Spell) {
		paladin.registerLightsVigilRank(rank)
	})
}

func (paladin *Paladin) registerLightsVigilRank(rank *spelldata.Spell) {
	triggered := lightsVigilTriggered[rank.RankNumber()]
	auraRank := spellData.LightsVigilTriggered.ByID(triggered.aura)
	healRank := spellData.LightsVigilTriggered.ByID(triggered.heal)
	heal := healRank.HealEffect()

	vigil := &lightsVigil{}
	paladin.lightsVigils = append(paladin.lightsVigils, vigil)

	vigil.auras = paladin.NewAllyAuraArray(func(target *core.Unit) *core.Aura {
		return target.GetOrRegisterAura(core.Aura{
			Label:    rank.RankedLabel(&paladin.Unit),
			ActionID: core.ActionID{SpellID: auraRank.ID},
			Duration: auraRank.Duration(),
		})
	})

	// TODO: Verify in game the threat the party heal makes; the paladin's other heals make half.
	healConfig := spelldata.SpellConfig(&paladin.Unit, healRank, spelldata.Magic(core.ProcMaskSpellHealing), spelldata.Proc())
	healConfig.Flags |= core.SpellFlagHelpful
	healConfig.ClassSpellMask = SpellMaskLightsVigilHeal
	healConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealHealing(sim, target, heal.Average(), spell.OutcomeHealingCrit)
	}
	vigil.heal = paladin.RegisterSpell(healConfig)

	config := spelldata.SpellConfig(&paladin.Unit, rank, spelldata.Flags(core.SpellFlagAPL|core.SpellFlagHelpful))
	config.ProcMask = core.ProcMaskEmpty
	config.ClassSpellMask = SpellMaskLightsVigil
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
		// One vigil per paladin: a new one replaces whatever was up.
		for _, other := range paladin.lightsVigils {
			for _, aura := range other.auras {
				if aura != nil && aura.IsActive() {
					aura.Deactivate(sim)
				}
			}
		}
		vigil.auras.Get(target).Activate(sim)
	}
	config.RelatedAuraArrays = vigil.auras.ToMap()
	paladin.RegisterSpell(config)
}

// Ends the vigil on the target and heals the target and their party, or the target alone where it is
// in no party.
func (paladin *Paladin) consumeLightsVigil(sim *core.Simulation, target *core.Unit) bool {
	for _, vigil := range paladin.lightsVigils {
		if !vigil.auras.Get(target).IsActive() {
			continue
		}
		vigil.auras.Get(target).Deactivate(sim)
		var party *core.Party
		if agent := paladin.Env.Raid.GetPlayerFromUnit(target); agent != nil {
			party = agent.GetCharacter().Party
		}
		if party == nil {
			vigil.heal.Cast(sim, target)
			return true
		}
		for _, member := range party.Players {
			vigil.heal.Cast(sim, &member.GetCharacter().Unit)
		}
		return true
	}
	return false
}
