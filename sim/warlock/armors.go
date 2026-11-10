package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Demon Armor (11735), parsed from its row: armor, Shadow resistance and in-combat health regen. Demonic
// Aegis raises every amount it states.
func (warlock *Warlock) registerArmors() {
	rankedSpell := spellData.DemonArmor.ByID(11735)

	opts := []spelldata.ParseOpt{}
	if warlock.Talents.DemonicAegis > 0 {
		opts = append(opts, spelldata.Count(1+spellData.DemonicAegis.Rank(warlock.Talents.DemonicAegis).EffectN(1).Percent()))
	}

	warlock.DemonArmor = warlock.RegisterAura(spelldata.AuraConfig(rankedSpell))
	spelldata.ParseEffects(&warlock.Character, warlock.DemonArmor, rankedSpell, opts...)

	if warlock.Options.Armor == proto.WarlockOptions_DemonArmor {
		core.MakePermanent(warlock.DemonArmor)
	}
}
