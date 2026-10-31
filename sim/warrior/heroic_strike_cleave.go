package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Rank 8 is the top rank the trainer teaches; Rank 9 comes from Manual of Heroic Strike IX.
var heroicStrikeTrainerRank = spellData.HeroicStrike.ByID(11567)
var heroicStrikeBookRank = spellData.HeroicStrike.ByID(25286)

var cleaveRank = spellData.Cleave.Highest()
var cleaveTargets = int32(cleaveRank.DamageEffect().ChainTargets)

func (warrior *Warrior) registerHeroicStrike() {
	warrior.registerHeroicStrikeRank(heroicStrikeTrainerRank)
	warrior.registerHeroicStrikeRank(heroicStrikeBookRank)
}

func (warrior *Warrior) registerHeroicStrikeRank(rank *spelldata.Spell) {
	config := spelldata.SpellConfig(&warrior.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.Flags &^= core.SpellFlagAPL

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := rank.WeaponStrike(sim, spell, warrior.MHWeaponDamage(sim, spell.MeleeAttackPower(target)), 0)
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	warrior.makeQueueSpellsAndAura(warrior.RegisterSpell(config))
}

func (warrior *Warrior) registerCleave() {
	config := spelldata.SpellConfig(&warrior.Unit, cleaveRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.Flags &^= core.SpellFlagAPL

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := cleaveRank.WeaponStrike(sim, spell, warrior.MHWeaponDamage(sim, spell.MeleeAttackPower(target)), 0)
		results := spell.CalcCleaveDamage(sim, target, cleaveTargets, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

		// The flat threat is the cleave's total, split between the targets it lands on.
		if landed := results.NumLandedHits(); landed > 1 {
			flatThreat := spell.FlatThreatBonus
			spell.FlatThreatBonus /= float64(landed)
			for _, result := range results {
				result.Threat = spell.ThreatFromDamage(sim, result.Outcome, result.Damage, warrior.AttackTables[result.Target.UnitIndex])
			}
			spell.FlatThreatBonus = flatThreat
		}
		spell.DealBatchedAoeDamage(sim)
		if !results[0].Landed() {
			spell.IssueRefund(sim)
		}
	}

	warrior.Cleave = warrior.RegisterSpell(config)
	warrior.makeQueueSpellsAndAura(warrior.Cleave)
}

// Heroic Strike and Cleave go off on the next main-hand swing (ON_NEXT_SWING in the game files): the
// APL casts the tag 1 spell, which queues the attack after the Queue Delay, and the next swing, extra
// attacks included, casts it in the swing's place. One attack is queued at a time, and a swing the
// rage no longer covers drops the queue.
func (warrior *Warrior) makeQueueSpellsAndAura(srcSpell *core.Spell) {
	queueAura := warrior.RegisterAura(core.Aura{
		Label:    "HS/Cleave Queue Aura-" + srcSpell.ActionID.String(),
		ActionID: srcSpell.ActionID.WithTag(1),
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, _ *core.Simulation) {
			warrior.curQueueAura = aura
			warrior.curQueuedAutoSpell = srcSpell
		},
		OnExpire: func(_ *core.Aura, _ *core.Simulation) {
			warrior.curQueueAura = nil
			warrior.curQueuedAutoSpell = nil
		},
	})

	warrior.RegisterSpell(core.SpellConfig{
		ActionID:    srcSpell.ActionID.WithTag(1),
		SpellSchool: srcSpell.SpellSchool,
		DefenseType: srcSpell.DefenseType,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagAPL | core.SpellFlagNoMetrics,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
		},

		ExtraCastCondition: func(sim *core.Simulation, _ *core.Unit) bool {
			return warrior.curQueueAura == nil &&
				warrior.pendingQueue.IsConsumed() &&
				warrior.CurrentRage() >= srcSpell.Cost.GetCurrentCost() &&
				warrior.queuedRealismICD.IsReady(sim)
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
			srcSpell.CountManualCast(target)
			if warrior.queuedRealismICD.Duration == 0 {
				queueAura.Activate(sim)
				return
			}

			warrior.queuedRealismICD.Use(sim)
			warrior.pendingQueue = &core.PendingAction{
				NextActionAt: sim.CurrentTime + warrior.queuedRealismICD.Duration,
				OnAction:     queueAura.Activate,
			}
			sim.AddPendingAction(warrior.pendingQueue)
		},
	})
}

func (warrior *Warrior) TryHSOrCleave(sim *core.Simulation, mhSwingSpell *core.Spell) *core.Spell {
	if warrior.curQueueAura == nil {
		return mhSwingSpell
	}

	queued := warrior.curQueuedAutoSpell
	warrior.curQueueAura.Deactivate(sim)
	if !queued.CanCast(sim, warrior.CurrentTarget) {
		return mhSwingSpell
	}

	return queued
}
