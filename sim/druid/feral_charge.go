package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"time"
)

func (druid *Druid) registerFeralChargeSpell() {
	rank := spellData.FeralCharge.Highest()
	root := spellData.FeralChargeTriggered.Highest()
	roots := druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.RegisterAura(spelldata.AuraConfig(root))
	})
	config := spelldata.SpellConfig(&druid.Unit, rank)
	config.Flags |= core.SpellFlagAPL | core.SpellFlagMovesToTarget
	config.Cast.IgnoreHaste = true
	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return druid.Talents.FeralCharge
	}
	// TODO: Verify the ground-charge speed in game; the movement model uses a temporary run-speed multiplier.
	chargeMovement := druid.RegisterAura(core.Aura{
		Label:    "Feral Charge Movement",
		ActionID: core.ActionID{SpellID: rank.ID, Tag: 1},
		Duration: 5 * time.Second,
		OnGain: func(_ *core.Aura, sim *core.Simulation) {
			druid.MultiplyMovementSpeed(sim, 3)
		},
		OnExpire: func(_ *core.Aura, sim *core.Simulation) {
			druid.MultiplyMovementSpeed(sim, 1.0/3.0)
		},
	})
	druid.RegisterMovementCallback(func(sim *core.Simulation, _ float64, kind core.MovementUpdateType) {
		if kind == core.MovementEnd && chargeMovement.IsActive() {
			chargeMovement.Deactivate(sim)
		}
	})
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
		if druid.InForm(Cat) {
			// TODO: Verify the modeled travel speed and landing delay in game.
			travel := core.DurationFromSeconds(druid.DistanceFromTarget / 80)
			druid.ExtendGCDUntil(sim, max(druid.NextGCDAt(), sim.CurrentTime+travel))
			druid.DistanceFromTarget = 0
			druid.MoveDuration(travel, sim)
			landingDelay := core.DurationFromSeconds(0.150 + sim.RandomFloat("Cat Charge")*0.6)
			druid.AutoAttacks.CancelMeleeSwing(sim)
			sim.AddPendingAction(&core.PendingAction{
				NextActionAt: sim.CurrentTime + travel + landingDelay,
				Priority:     core.ActionPriorityDOT,
				OnAction: func(sim *core.Simulation) {
					druid.AutoAttacks.EnableMeleeSwing(sim)
				},
			})
			return
		}
		if !druid.InForm(Bear) {
			return
		}
		chargeMovement.Activate(sim)
		druid.MoveTo(core.MaxMeleeRange-1, sim) // Overshoot the discrete range updates to reach melee.
		roots.Get(target).Activate(sim)
	}
	config.RelatedAuraArrays = roots.ToMap()
	druid.FeralCharge = druid.RegisterSpell(config)
	if druid.CatFormAura != nil {
		gcd := druid.FeralCharge.DefaultCast.GCD
		druid.CatFormAura.ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) {
			druid.FeralCharge.Cost.PercentModifier -= 1
			druid.FeralCharge.DefaultCast.GCD = time.Second
		}).ApplyOnExpire(func(_ *core.Aura, _ *core.Simulation) {
			druid.FeralCharge.Cost.PercentModifier += 1
			druid.FeralCharge.DefaultCast.GCD = gcd
		})
	}
}
