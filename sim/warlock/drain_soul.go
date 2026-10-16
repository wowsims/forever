package warlock

func (warlock *Warlock) registerDrainSoul() {
	rankedSpell := spellData.DrainSoul.ByID(11675)
	drainSoulTick := rankedSpell.PeriodicEffect()

	config := warlock.soulSiphonChannelConfig(rankedSpell, drainSoulTick, nil)
	config.ClassSpellMask = WarlockSpellDrainSoul

	warlock.DrainSoul = warlock.RegisterSpell(config)
}
