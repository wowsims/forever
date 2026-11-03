package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// Poised, the passive the November 2 development notes add at level 32: out of combat at maximum Energy, the
// next Energy ability costs 10 less. The game build the database is made from has no row for it; game
// build 1.60.1.70753 states the passive as 1328829 and its buff as 1328827, a one-charge -10 Energy cost
// mod. The sim leaves combat only before the pull, so the buff is up for the opener when the druid starts
// it at maximum Energy.
// TODO: Read 1328829 and 1328827 from the spell store once the database is made from a game build that has them.
// TODO: Ingame research needed -- whether a Clearcasting cast, which costs nothing, takes the charge.
const poisedBuffSpellID = 1328827

func (druid *Druid) registerPoised() {
	if !druid.HasEnergyBar() {
		return
	}
	druid.PoisedAura = druid.RegisterAura(core.Aura{
		Label:    "Poised",
		ActionID: core.ActionID{SpellID: poisedBuffSpellID},
		Duration: core.NeverExpires,
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.Cost == nil || spell.Cost.BaseCost <= 0 {
				return
			}
			if _, ok := spell.Cost.ResourceCostImpl.(*core.EnergyCost); ok {
				aura.Deactivate(sim)
			}
		},
	}).AttachSpellMod(core.SpellModConfig{
		Kind:         core.SpellMod_PowerCost_Flat,
		ResourceType: proto.ResourceType_ResourceTypeEnergy,
		IntValue:     -10,
	})
}

func (druid *Druid) activatePoised(sim *core.Simulation) {
	if druid.PoisedAura != nil && druid.CurrentEnergy() >= druid.MaximumEnergy() {
		druid.PoisedAura.Activate(sim)
	}
}
