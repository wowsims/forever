package rogue

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var shankRanks = []struct {
	spellID    int32
	rank       int32
	energyGain float64
}{
	{spellID: 15248, rank: 1, energyGain: 15},
	{spellID: 15617, rank: 2, energyGain: 25},
}

func (rogue *Rogue) registerShankSpell() {
	cdTimer := rogue.NewTimer()
	flashOfSteel := float64(rogue.Talents.FlashOfSteel)

	// The captured rows describe a behind-target strike; the beta notes define these ranks and cast values.
	for _, rank := range shankRanks {
		config := spelldata.SpellConfig(&rogue.Unit, spelldata.MustFind(rank.spellID),
			spelldata.Melee(core.ProcMaskMeleeOHSpecial))
		config.Rank = rank.rank
		config.EnergyCost = core.EnergyCostOptions{Cost: 5}
		config.Cast.DefaultCast.CastTime = 0
		config.Cast.DefaultCast.GCD = core.GCDDefault
		config.Cast.CD = core.Cooldown{Timer: cdTimer, Duration: 10*time.Second - time.Duration(flashOfSteel*0.4*float64(time.Second))}
		config.BonusHitPercent = 100
		config.DamageMultiplier = 0.5
		config.ClassSpellMask = RogueSpellShank
		config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
			return rogue.OffHand() != nil
		}
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)
			baseDamage := rogue.OHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			if result.Landed() {
				rogue.AddEnergy(sim, rank.energyGain+3*flashOfSteel, spell.EnergyMetrics())
			}
		}

		shank := rogue.RegisterSpell(config)
		if rank.rank == 2 {
			rogue.Shank = shank
		}
	}
}
