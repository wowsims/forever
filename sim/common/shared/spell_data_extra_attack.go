package shared

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// An item proc whose spell grants the wearer extra attacks: Flurry Axe's 18797 and Thrash Blade's
// 21919 each grant one. Each is a main-hand swing tagged with the granting spell, so the metrics tell
// it from the swing that procced it.
func NewSpellDataExtraAttackProc(cfg SpellDataProc, variants []ItemVariant) {
	forEachSpellDataVariant(cfg, variants, registerSpellDataExtraAttackProc)
}

func registerSpellDataExtraAttackProc(cfg SpellDataProc) {
	registerSpellDataRowProc(cfg, nil, func(agent core.Agent, source effectSource, trigger *spelldata.Spell, row *spelldata.Spell) {
		applySpellDataExtraAttackProc(agent, cfg, source, trigger, row)
	})
}

func applySpellDataExtraAttackProc(agent core.Agent, cfg SpellDataProc, source effectSource, trigger *spelldata.Spell, row *spelldata.Spell) {
	character := agent.GetCharacter()
	config := spellDataProcListener(character, cfg, source, trigger)
	onInit := character.ExtraAttackProc(&config, row.ID, int(row.ExtraAttacksEffect().BasePoints))
	source.registerTrigger(character, config).ApplyOnInit(onInit)
}
