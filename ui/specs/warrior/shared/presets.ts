import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec } from '@generated/proto/common';

export const DefaultIndividualBuffs = IndividualBuffs.create({
	greaterBlessingOfKings: true,
	greaterBlessingOfMight: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	strengthOfEarthTotem: true,
	windfuryTotem: true,
	leaderOfThePack: true,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	prayerOfFortitude: true,
	giftOfTheWild: true,
});

export const DefaultDebuffs = Debuffs.create({
	giftOfArthas: true,
	exposeArmor: true,
	faerieFire: true,
	sunderArmor: true,
	curseOfRecklessness: true,
});

export const DefaultConsumables = ConsumesSpec.create({
	potId: Consumables.MajorFrenzyPotion.id,
	foodId: Consumables.SmokedDesertDumplings.id,
	explosiveId: Consumables.EZThroDarkBomb.id,
	elixirIds: [Consumables.ElixirOfFerocity.id, Consumables.ElixirOfTheMongoose.id],
	extraIds: [Consumables.GoblinSapperCharge.id],
});
