package paladin

import (
	"fmt"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var LightsVigilRankMap = spellData.LightsVigil

// The vigil aura, party heal and enemy strike each rank triggers. They sit in LightsVigilTriggered
// in the order the game files list them, not by rank, so the pairing is by hand.
var lightsVigilTriggered = map[int32]struct{ aura, heal, strike int32 }{
	1: {aura: 1310909, heal: 1310912, strike: 1310914},
	2: {aura: 1311593, heal: 1311591, strike: 1311592},
	3: {aura: 1311597, heal: 1311596, strike: 1311598},
}

// One rank's vigils on enemies, the strike that consumes one, and the refund it gives.
type lightsVigil struct {
	auras  core.AuraArray
	strike *core.Spell
	refund float64
	cost   float64

	// Ends the vigil on the target: the strike lands and the refund comes back.
	consume func(sim *core.Simulation, target *core.Unit)
}

// Light's Vigil (talent)
// https://www.wowhead.com/forever/spell=1311595
//
// Applies Light's Vigil to the target for 30 sec. Your next Holy Shock cast on them triggers no
// cooldown and causes friendly targets to heal their party for 704, or enemy targets to suffer 395
// Holy damage and refund 75% of Light's Vigil's Mana cost. You may only have 1 Light's Vigil
// active per Paladin, per party.
//
// The sim casts it on enemies: Holy Shock on a vigiled enemy fires the strike and the refund. The
// friendly path, a party heal off a Holy Shock heal, has no target model here and is not built.
func (paladin *Paladin) registerLightsVigil() {
	LightsVigilRankMap.Each(paladin.registerLightsVigilRank)
}

func (paladin *Paladin) registerLightsVigilRank(_ int32, rank *spelldata.Spell) {
	triggered := lightsVigilTriggered[rank.RankNumber()]
	auraRank := spellData.LightsVigilTriggered.ByID(triggered.aura)
	strikeRank := spellData.LightsVigilTriggered.ByID(triggered.strike)
	strikeDamage := strikeRank.DamageEffect()

	vigil := &lightsVigil{
		// The refund share is the rank's second effect.
		refund: rank.EffectN(2).Percent(),
	}
	paladin.lightsVigils = append(paladin.lightsVigils, vigil)

	vigil.auras = paladin.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.GetOrRegisterAura(core.Aura{
			Label:    fmt.Sprintf("Light's Vigil%s Rank %d", paladin.Label, rank.RankNumber()),
			ActionID: core.ActionID{SpellID: auraRank.ID},
			Duration: auraRank.Duration(),
		})
	})

	vigil.strike = paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: strikeRank.ID},
		SpellSchool:    strikeRank.SpellSchool(),
		DefenseType:    strikeRank.DefenseTypeCore(),
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagPassiveSpell,
		ClassSpellMask: SpellMaskLightsVigilStrike,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: strikeDamage.Coeff(),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, strikeDamage.Roll(sim), spell.OutcomeMagicHitAndCrit)
		},
	})

	manaMetrics := paladin.NewManaMetrics(core.ActionID{SpellID: rank.ID})
	vigil.strike.RelatedAuraArrays = vigil.auras.ToMap()

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rank.ID},
		SpellSchool:    rank.SpellSchool(),
		DefenseType:    rank.DefenseTypeCore(),
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskLightsVigil,
		Rank:           rank.RankNumber(),
		MaxRange:       float64(rank.MaxRange),

		ManaCost: manaCost(rank),
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      rank.GCD(),
				CastTime: rank.CastTime(),
			},
			CD: core.Cooldown{
				Timer:    paladin.sharedTimer(&paladin.lightsVigilTimer),
				Duration: cooldown(rank),
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			vigil.cost = spell.CurCast.Cost
			// One vigil per paladin: a new one replaces whatever was up.
			for _, other := range paladin.lightsVigils {
				for _, aura := range other.auras {
					if aura != nil && aura.IsActive() {
						aura.Deactivate(sim)
					}
				}
			}
			vigil.auras.Get(target).Activate(sim)
		},

		RelatedAuraArrays: vigil.auras.ToMap(),
	})

	vigil.consume = func(sim *core.Simulation, target *core.Unit) {
		vigil.auras.Get(target).Deactivate(sim)
		vigil.strike.Cast(sim, target)
		paladin.AddMana(sim, vigil.cost*vigil.refund, manaMetrics)
	}
}

// The vigil this paladin has on the target, if any.
func (paladin *Paladin) activeLightsVigil(target *core.Unit) *lightsVigil {
	for _, vigil := range paladin.lightsVigils {
		if vigil.auras.Get(target).IsActive() {
			return vigil
		}
	}
	return nil
}
