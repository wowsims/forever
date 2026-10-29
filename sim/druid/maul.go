package druid

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerMaulSpell() {
	rank := spellData.Maul.Highest()
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial),
		spelldata.RemoveFlags(core.SpellFlagAPL))
	// The queued cast keeps the spell id the APL names. The swing's damage is the same id, tagged,
	// so a lookup of the untagged id finds the queue.
	config.ActionID = config.ActionID.WithTag(1)
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		if druid.maulQueueAura != nil {
			druid.maulQueueAura.Deactivate(sim)
		}
		result := spell.CalcAndDealDamage(sim, target, druid.weaponSpecialDamage(sim, spell, target, rank), spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		druid.refundOnMiss(sim, spell, result)
	}
	damage := druid.RegisterSpell(config)
	druid.Maul = druid.makeMaulQueueSpellAndAura(damage)
}

func (druid *Druid) makeMaulQueueSpellAndAura(maulSpell *DruidSpell) *DruidSpell {
	isMaulQueued := false

	druid.maulQueueAura = druid.RegisterAura(core.Aura{
		Label:    "Maul Queue Aura",
		ActionID: maulSpell.ActionID,
		Duration: core.NeverExpires,
		OnReset: func(_ *core.Aura, _ *core.Simulation) {
			isMaulQueued = false
		},
		OnGain: func(_ *core.Aura, _ *core.Simulation) {
			druid.maulQueueSpell = maulSpell.Spell
		},
		OnExpire: func(_ *core.Aura, _ *core.Simulation) {
			druid.maulQueueSpell = nil
		},
	})

	druid.maulRealismICD = &core.Cooldown{
		Timer:    druid.NewTimer(),
		Duration: time.Millisecond * 50,
	}

	queueSpell := druid.RegisterSpell(core.SpellConfig{
		ActionID:        maulSpell.ActionID.WithTag(0),
		CastRequirement: maulSpell.CastRequirement,
		SpellSchool:     core.SpellSchoolPhysical,
		DefenseType:     core.DefenseTypeMelee,
		ProcMask:        core.ProcMaskMeleeMHSpecial,
		Flags:           core.SpellFlagMeleeMetrics | core.SpellFlagAPL | core.SpellFlagNoMetrics,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
		},

		ExtraCastCondition: func(sim *core.Simulation, _ *core.Unit) bool {
			return druid.maulQueueAura != nil &&
				!druid.maulQueueAura.IsActive() &&
				!isMaulQueued &&
				druid.CurrentRage() >= maulSpell.Cost.GetCurrentCost() &&
				druid.maulRealismICD.IsReady(sim)
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
			maulSpell.CountManualCast(target)
			if druid.maulRealismICD.IsReady(sim) {
				isMaulQueued = true
				druid.maulRealismICD.Use(sim)
				sim.AddPendingAction(&core.PendingAction{
					NextActionAt: sim.CurrentTime + druid.maulRealismICD.Duration,
					OnAction: func(sim *core.Simulation) {
						if druid.InForm(Bear) {
							druid.maulQueueAura.Activate(sim)
						}
						isMaulQueued = false
					},
				})
			}
		},
	})

	return queueSpell
}

func (druid *Druid) TryMaul(sim *core.Simulation, mhSwingSpell *core.Spell) *core.Spell {
	if druid.maulQueueAura == nil || !druid.maulQueueAura.IsActive() {
		return mhSwingSpell
	}
	if druid.maulQueueSpell == nil || !druid.maulQueueSpell.CanCast(sim, druid.CurrentTarget) {
		druid.maulQueueAura.Deactivate(sim)
		return mhSwingSpell
	}
	return druid.maulQueueSpell
}
