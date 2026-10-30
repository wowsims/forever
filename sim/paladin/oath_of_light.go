package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The Oath on one ally, and what it adds to each heal the paladin casts on that ally.
type oathOfLight struct {
	auras        core.AuraArray
	holyLight    float64
	flashOfLight float64
	holyShock    float64
}

// Oath of Light (talent)
// https://www.wowhead.com/forever/spell=1328548
//
// Swear an Oath to protect target ally for 5 min, increasing the healing of your Holy Light on them
// by up to 400, your Flash of Light by up to 115, and your Holy Shock by up to 229. Only one Oath of
// Light may be sworn by the Paladin at any one time.
//
// The three effects state the bonuses, which grow with the paladin's level; the heals read them when
// the target carries the Oath.
// TODO: Verify in game whether a lower rank of Holy Light, Flash of Light or Holy Shock gets less than
// the full bonus, as the tooltip's "up to" says.
func (paladin *Paladin) registerOathOfLight() {
	rank := spellData.OathOfLight.Highest()

	oath := &oathOfLight{
		holyLight:    rank.EffectN(1).Average(),
		flashOfLight: rank.EffectN(2).Average(),
		holyShock:    rank.EffectN(3).Average(),
	}
	oath.auras = paladin.NewAllyAuraArray(func(target *core.Unit) *core.Aura {
		return target.GetOrRegisterAura(core.Aura{
			Label:    "Oath of Light" + paladin.Label,
			ActionID: core.ActionID{SpellID: rank.ID},
			Duration: rank.Duration(),
		})
	})
	paladin.oathOfLight = oath

	config := spelldata.SpellConfig(&paladin.Unit, rank, spelldata.Flags(core.SpellFlagAPL|core.SpellFlagHelpful))
	config.ProcMask = core.ProcMaskEmpty
	config.ClassSpellMask = SpellMaskOathOfLight
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
		// One Oath per paladin: a new one replaces whatever was up.
		for _, aura := range oath.auras {
			if aura != nil && aura.IsActive() {
				aura.Deactivate(sim)
			}
		}
		oath.auras.Get(target).Activate(sim)
	}
	config.RelatedAuraArrays = oath.auras.ToMap()
	paladin.RegisterSpell(config)
}

// Whether the paladin's Oath of Light is on the target.
func (paladin *Paladin) hasOathOfLight(target *core.Unit) bool {
	return paladin.oathOfLight != nil && paladin.oathOfLight.auras.Get(target).IsActive()
}
