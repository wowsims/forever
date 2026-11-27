package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var rapidFireRank = spellData.RapidFire.Highest()

func (hunter *Hunter) registerRapidFire() {
	aura := hunter.RegisterAura(spelldata.AuraConfig(rapidFireRank))

	spelldata.ParseEffects(&hunter.Character, aura, rapidFireRank)
	hunter.speedPetFocusWith(aura, rangedHasteOf(rapidFireRank))

	config := spelldata.SpellConfig(&hunter.Unit, rapidFireRank, spelldata.Flags(core.SpellFlagAPL))

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		aura.Activate(sim)
	}

	config.RelatedSelfBuff = aura

	hunter.RapidFire = hunter.RegisterSpell(config)

	hunter.AddMajorCooldown(core.MajorCooldown{
		Spell: hunter.RapidFire,
		Type:  core.CooldownTypeDPS,
	})
}
