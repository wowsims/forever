package paladin

// Swift Judgement (talent)
// https://www.wowhead.com/forever/spell=1310994
//
// Unleash the energy of a Seal spell upon an enemy. Does not consume the Seal. Refer to individual
// Seals for Judgement effect. Does not share a cooldown with your Judgement ability.
func (paladin *Paladin) registerSwiftJudgement() {
	config := paladin.judgementConfig(spellData.SwiftJudgement.Highest(), SpellMaskSwiftJudgement)
	// Free and off the global cooldown, so the spell leaves the cast empty.
	config.Cast.DefaultCast.NonEmpty = true

	paladin.RegisterSpell(config)
}
