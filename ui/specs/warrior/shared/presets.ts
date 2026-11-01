import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec } from '@generated/proto/common';

export const DefaultIndividualBuffs = IndividualBuffs.create({
	greaterBlessingOfKings: true,
	greaterBlessingOfMight: 1,
});

export const DefaultPartyBuffs = PartyBuffs.create({});

export const DefaultRaidBuffs = RaidBuffs.create({
	prayerOfFortitude: 1,
	giftOfTheWild: 1,
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

export const DefaultConsumables = ConsumesSpec.create({
	foodId: Consumables.SmokedDesertDumplings.id,
	explosiveId: Consumables.EZThroDarkBomb.id,
	elixirIds: [Consumables.ElixirOfFerocity.id, Consumables.ElixirOfTheMongoose.id],
	extraIds: [Consumables.GoblinSapperCharge.id],
});
