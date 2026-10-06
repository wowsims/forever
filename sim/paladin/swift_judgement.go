package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Swift Judgement (talent)
// https://www.wowhead.com/forever/spell=1310994
//
// Finishes the remaining cooldown on your Judgement ability and reduces the Mana cost of your next
// Judgement by 100%.
func (paladin *Paladin) registerSwiftJudgement() {
	rank := spellData.SwiftJudgement.Highest()

	// The free cost is the row's. Its one charge goes to the next Judgement, which spends it itself:
	// Judgement fires no on-cast callbacks for a listener to hear.
	freeJudgement := paladin.RegisterAura(spelldata.AuraConfig(rank, spelldata.Label("Swift Judgement"+paladin.Label)))
	spelldata.ParseEffects(&paladin.Character, freeJudgement, rank)
	paladin.swiftJudgementAura = freeJudgement

	config := spelldata.SpellConfig(&paladin.Unit, rank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ClassSpellMask = SpellMaskSwiftJudgement
	// Free and off the global cooldown, so the row leaves the cast empty.
	config.Cast.DefaultCast.NonEmpty = true
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		paladin.Judgement.CD.Reset()
		spell.RelatedSelfBuff.Activate(sim)
	}
	config.RelatedSelfBuff = freeJudgement
	paladin.RegisterSpell(config)
}
