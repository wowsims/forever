import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';

export const DefaultIndividualBuffs = IndividualBuffs.create({
	greaterBlessingOfKings: true,
	greaterBlessingOfMight: 1,
});

export const DefaultPartyBuffs = PartyBuffs.create({});

export const DefaultRaidBuffs = RaidBuffs.create({
	powerWordFortitude: true,
	markOfTheWild: true,
	strengthOfEarthTotem: 1,
	windfuryTotem: true,
	graceOfAirTotem: 1,
	leaderOfThePack: true,
});

export const DefaultDebuffs = Debuffs.create({
	giftOfArthas: true,
	exposeArmor: true,
	faerieFire: true,
	sunderArmor: true,
	curseOfRecklessness: true,
});
