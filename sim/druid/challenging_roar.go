package druid

func (druid *Druid) registerChallengingRoarSpell() {
	druid.ChallengingRoar = druid.registerBearUtility(spellData.ChallengingRoar.Highest())
}
