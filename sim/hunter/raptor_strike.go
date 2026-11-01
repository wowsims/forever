package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var raptorStrikeRank = spellData.RaptorStrike.Highest()

func (hunter *Hunter) registerRaptorStrike() {
	config := spelldata.SpellConfig(&hunter.Unit, raptorStrikeRank,
		spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.Flags(core.SpellFlagNoOnCastComplete))
	// The APL queues it through the tag 1 spell below; the next swing casts it.
	config.Flags &^= core.SpellFlagAPL

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := raptorStrikeRank.WeaponStrike(sim, spell, hunter.MHWeaponDamage(sim, spell.MeleeAttackPower(target)), 0)
		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
	}

	hunter.RaptorStrike = hunter.RegisterSpell(config)

	// Raptor Strike goes off on the next main-hand swing, as Heroic Strike does: the APL casts the
	// tag 1 spell, which queues it, and the next swing casts it in the swing's place.
	hunter.raptorStrikeQueue = hunter.RegisterAura(core.Aura{
		Label:    "Raptor Strike Queue",
		ActionID: hunter.RaptorStrike.ActionID.WithTag(1),
		Duration: core.NeverExpires,
	})
	hunter.AutoAttacks.SetQueuedSwingCanceler(hunter.raptorStrikeQueue.Deactivate)

	hunter.RegisterSpell(core.SpellConfig{
		ActionID:    hunter.RaptorStrike.ActionID.WithTag(1),
		SpellSchool: hunter.RaptorStrike.SpellSchool,
		DefenseType: hunter.RaptorStrike.DefenseType,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagAPL | core.SpellFlagNoMetrics,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
		},

		// Queuing has no range: it can be set walking in, so the first swing on arrival takes it.
		ExtraCastCondition: func(sim *core.Simulation, _ *core.Unit) bool {
			return !hunter.raptorStrikeQueue.IsActive() && hunter.RaptorStrike.IsReady(sim) &&
				hunter.CurrentMana() >= hunter.RaptorStrike.Cost.GetCurrentCost()
		},
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
			hunter.RaptorStrike.CountManualCast(target)
			hunter.raptorStrikeQueue.Activate(sim)
		},
	})
}

// The swing a main-hand auto becomes: the queued Raptor Strike when it can still be cast, the plain
// swing otherwise. A swing that cannot cast it drops the queue.
func (hunter *Hunter) TryRaptorStrike(sim *core.Simulation, mhSwingSpell *core.Spell) *core.Spell {
	if mhSwingSpell.ActionID.Tag != int32(proto.ItemSlot_ItemSlotMainHand) || !hunter.raptorStrikeQueue.IsActive() {
		return mhSwingSpell
	}

	hunter.raptorStrikeQueue.Deactivate(sim)
	if !hunter.RaptorStrike.CanCast(sim, hunter.CurrentTarget) {
		return mhSwingSpell
	}

	return hunter.RaptorStrike
}
