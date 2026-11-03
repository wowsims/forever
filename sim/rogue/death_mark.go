package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

const deathMarkStorePercentPerComboPoint = 0.12

func (rogue *Rogue) registerDeathMark() {
	if !rogue.Talents.DeathMark {
		return
	}

	deathMarkRank := spelldata.MustFind(1325758)

	type markState struct {
		damage      float64
		storeFactor float64
		comboPoints int32
	}

	maxStoredDamage := func(state *markState) float64 {
		// The rank 4 tooltip cap is 156 + 245 per combo point + 4% AP per point. Its
		// attackP variable resolves to AP; the tooltip also names Critical Strike chance,
		// but its displayed cap formula has no critical strike term.
		points := float64(state.comboPoints)
		return 156 + 245*points + 0.04*rogue.GetStat(stats.AttackPower)*points
	}

	states := make(map[*core.Unit]*markState)
	var detonation *core.Spell
	marks := rogue.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		state := &markState{}
		states[target] = state
		return target.RegisterAura(core.Aura{
			Label:    "Death Mark - " + rogue.Label,
			ActionID: core.ActionID{SpellID: deathMarkRank.ID},
			Duration: deathMarkRank.Duration(),
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				damage := min(state.damage, maxStoredDamage(state))
				state.damage = 0
				if damage > 0 {
					detonation.CalcAndDealDamage(sim, target, damage, detonation.OutcomeAlwaysHit)
				}
			},
		})
	})

	recordDamage := func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		if result.Damage <= 0 || spell.Matches(RogueSpellDeathMark) {
			return
		}
		aura := marks.Get(result.Target)
		if !aura.IsActive() {
			return
		}
		state := states[result.Target]
		factor := state.storeFactor
		state.damage = min(state.damage+result.Damage*factor, maxStoredDamage(state))
	}

	rogue.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Death Mark - Store Damage",
		ActionID:           core.ActionID{SpellID: deathMarkRank.ID},
		Callback:           core.CallbackOnSpellHitDealt | core.CallbackOnPeriodicDamageDealt,
		TriggerImmediately: true,
		ProcChance:         1,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			recordDamage(sim, spell, result)
		},
	})

	// The game files state 25 yd; the tooltip text states 20 yd.
	config := spelldata.SpellConfig(&rogue.Unit, deathMarkRank,
		spelldata.Flags(core.SpellFlagAPL|SpellFlagFinisher))
	config.ClassSpellMask = RogueSpellDeathMark
	config.ProcMask = core.ProcMaskEmpty
	config.ExtraCastCondition = rogue.FinisherExtraCastCondition
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
		if result.Landed() {
			points := rogue.ComboPoints()
			state := states[target]
			state.damage = 0
			state.comboPoints = points
			state.storeFactor = deathMarkStorePercentPerComboPoint * float64(points) *
				(1 + rogue.deathMarkDamageBonus)
			marks.Get(target).Activate(sim)
			rogue.ApplyFinisher(sim, spell)
		} else {
			spell.IssueRefund(sim)
		}
		spell.DealOutcome(sim, result)
	}
	rogue.RegisterSpell(config)

	detonationConfig := spelldata.SpellConfig(&rogue.Unit, deathMarkRank,
		spelldata.Tag(1), spelldata.Proc())
	detonationConfig.ClassSpellMask = RogueSpellDeathMark
	detonationConfig.ProcMask = core.ProcMaskEmpty
	detonationConfig.DamageMultiplier = 1
	detonationConfig.ThreatMultiplier = 1
	detonationConfig.ApplyEffects = func(_ *core.Simulation, _ *core.Unit, _ *core.Spell) {}
	detonation = rogue.RegisterSpell(detonationConfig)
}
