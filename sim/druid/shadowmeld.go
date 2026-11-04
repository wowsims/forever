package druid

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// Shadowmeld 20580, the Night Elf racial. The spell store does not carry racial rows, so the
// values are read off 20580 by hand: no GCD, a 10s category cooldown that the tooltip raises to
// 2 min in combat ("Using this ability in combat ... increases the cooldown to 2 min"), A_MOD_STEALTH,
// and the auto attack cancel bit (Attributes[0] 0x100000). It does not enable Prowl attacks.
// The next cast that breaks Prowl breaks it too. The threat drop (A_MOD_TOTAL_THREAT) is not modelled.
const shadowmeldSpellID = 20580

func (druid *Druid) registerShadowmeld() {
	if druid.Race != proto.Race_RaceNightElf {
		return
	}
	druid.ShadowmeldAura = druid.RegisterAura(core.Aura{
		Label:       "Shadowmeld",
		ActionID:    core.ActionID{SpellID: shadowmeldSpellID},
		Duration:    core.NeverExpires,
		HideFromAPL: true, // Out of the APL with its spell.
		OnGain: func(_ *core.Aura, sim *core.Simulation) {
			druid.AutoAttacks.PauseMeleeBy(sim, core.NeverExpires)
		},
		OnExpire: func(_ *core.Aura, sim *core.Simulation) {
			druid.AutoAttacks.ResumeMeleeAt(sim, 0, 0)
		},
	})
	druid.Shadowmeld = druid.RegisterSpell(core.SpellConfig{
		ActionID: core.ActionID{SpellID: shadowmeldSpellID},
		Flags:    core.SpellFlagNoOnCastComplete,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: 2 * time.Minute,
			},
		},
		ExtraCastCondition: func(_ *core.Simulation, _ *core.Unit) bool {
			return !druid.ShadowmeldAura.IsActive()
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			druid.ShadowmeldAura.Activate(sim)
		},
	})
}
