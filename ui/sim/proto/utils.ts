import { Player } from '@generated/proto/api';
import { IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { Class, ConsumesSpec, Faction, Race, Spec, TristateEffect, UnitReference, UnitReference_Type } from '@generated/proto/common';
import { ResourceType } from '@generated/proto/spell';
import { UIItem as Item } from '@generated/proto/ui';

import { PlayerClasses } from '../player/classes';
import { PlayerClass } from '../player/player_class';
import { PlayerSpec } from '../player/player_spec';
import { PlayerSpecs } from '../player/specs';
import { getEnumValues } from '../utils/collections';
import { sum } from '../utils/math';
import { isBluntWeaponType, isSharpWeaponType } from './items';
import { DENSE_SHARPENING_STONE_ID, DENSE_WEIGHTSTONE_ID } from './sim_tables_auto_gen';

export const NUM_SPECS = getEnumValues(Spec).length;

// Converts '1231321-12313123-0' to [13, 16, 0]. Talent strings are one run of per-row digits
// per tree, so the interesting number is the per-tree total and not the digits themselves.
export function getTalentTreePoints(talentsString: string): Array<number> {
	const trees = talentsString.split('-');
	if (trees.length == 2) {
		trees.push('0');
	}
	return trees.map(tree => sum([...tree].map(char => parseInt(char) || 0)));
}

export function getTalentPoints(talentsString: string): number {
	return sum(getTalentTreePoints(talentsString));
}

const CLASS_TEXT: Record<string, string> = {
	druid: 'text-class-druid',
	hunter: 'text-class-hunter',
	mage: 'text-class-mage',
	paladin: 'text-class-paladin',
	priest: 'text-class-priest',
	rogue: 'text-class-rogue',
	shaman: 'text-class-shaman',
	warlock: 'text-class-warlock',
	warrior: 'text-class-warrior',
};

/** The class-colour class. `--color-class-*` in `styles/theme/colors.css` is what makes it resolve, so the argument has to be a `getCssScheme` slug and not a display name. */
export function textClassName(className: string): string {
	return CLASS_TEXT[className];
}
export function textClassNameForClass<ClassType extends Class>(playerClass: PlayerClass<ClassType>): string {
	return textClassName(PlayerClasses.getCssScheme(playerClass));
}
export function textClassNameForSpec<SpecType extends Spec>(playerSpec: PlayerSpec<SpecType>): string {
	return textClassNameForClass(PlayerSpecs.getPlayerClass(playerSpec));
}

export const raceToFaction: Record<Race, Faction> = {
	[Race.RaceUnknown]: Faction.Unknown,

	[Race.RaceDwarf]: Faction.Alliance,
	[Race.RaceGnome]: Faction.Alliance,
	[Race.RaceHuman]: Faction.Alliance,
	[Race.RaceNightElf]: Faction.Alliance,
	[Race.RaceHighOrderSkyborne]: Faction.Alliance,

	[Race.RaceBloodElf]: Faction.Horde,
	[Race.RaceOrc]: Faction.Horde,
	[Race.RaceTauren]: Faction.Horde,
	[Race.RaceTroll]: Faction.Horde,
	[Race.RaceUndead]: Faction.Horde,
	[Race.RaceWindshaperSkyborne]: Faction.Horde,
};

// Returns a copy of playerOptions, with the class field set.
export function getPlayerSpecFromPlayer<SpecType extends Spec>(player: Player): PlayerSpec<SpecType> {
	const specValues = getEnumValues(Spec);
	for (let i = 0; i < specValues.length; i++) {
		const spec = specValues[i] as SpecType;
		let specString = Spec[spec]; // Returns 'SpecBalanceDruid' for BalanceDruid.
		specString = specString.substring('Spec'.length); // 'BalanceDruid'
		specString = specString.charAt(0).toLowerCase() + specString.slice(1); // 'balanceDruid'

		if (player.spec.oneofKind == specString) {
			return PlayerSpecs.fromProto(spec);
		}
	}

	throw new Error('Unable to parse spec from player proto: ' + JSON.stringify(Player.toJson(player), null, 2));
}

// Swaps the Dense sharpening stone and weightstone to match the weapon; any other imbue passes through.
export function adjustWeaponImbueId(imbueId: number, hasSharp: boolean, hasBlunt: boolean): number {
	if (imbueId !== DENSE_SHARPENING_STONE_ID && imbueId !== DENSE_WEIGHTSTONE_ID) return imbueId;
	if (hasSharp) return DENSE_SHARPENING_STONE_ID;
	if (hasBlunt) return DENSE_WEIGHTSTONE_ID;
	return 0;
}

const adjustHandImbue = (imbueId: number, weapon?: Item | null) =>
	adjustWeaponImbueId(imbueId, !!weapon && isSharpWeaponType(weapon.weaponType), !!weapon && isBluntWeaponType(weapon.weaponType));

// Matches each hand's imbue field to the weapon in that hand; the same spec comes back when nothing changes.
const adjustImbuePair = (
	consumes: ConsumesSpec,
	[mhField, ohField]: ['mhImbueId', 'ohImbueId'] | ['swapMhImbueId', 'swapOhImbueId'],
	mainHand?: Item | null,
	offHand?: Item | null,
): ConsumesSpec => {
	const mhImbueId = adjustHandImbue(consumes[mhField], mainHand);
	const ohImbueId = adjustHandImbue(consumes[ohField], offHand);
	if (mhImbueId === consumes[mhField] && ohImbueId === consumes[ohField]) return consumes;
	return ConsumesSpec.clone({ ...consumes, [mhField]: mhImbueId, [ohField]: ohImbueId });
};

export const adjustImbues = (consumes: ConsumesSpec, mainHand?: Item | null, offHand?: Item | null): ConsumesSpec =>
	adjustImbuePair(consumes, ['mhImbueId', 'ohImbueId'], mainHand, offHand);

export const adjustSwapImbues = (consumes: ConsumesSpec, mainHand?: Item | null, offHand?: Item | null): ConsumesSpec =>
	adjustImbuePair(consumes, ['swapMhImbueId', 'swapOhImbueId'], mainHand, offHand);

export function newUnitReference(raidIndex: number): UnitReference {
	return UnitReference.create({
		type: UnitReference_Type.Player,
		index: raidIndex,
	});
}

export function emptyUnitReference(): UnitReference {
	return UnitReference.create();
}

export const orderedResourceTypes: Array<ResourceType> = [
	ResourceType.ResourceTypeHealth,
	ResourceType.ResourceTypeMana,
	ResourceType.ResourceTypeEnergy,
	ResourceType.ResourceTypeRage,
	ResourceType.ResourceTypeComboPoints,
	ResourceType.ResourceTypeFocus,
	ResourceType.ResourceTypeGenericResource,
];

export const AL_CATEGORY_HARD_MODE = 'Hard Mode';
export const AL_CATEGORY_TITAN_RUNE = 'Titan Rune';

// The buffs every healer gear planner starts with: the caster stat buffs, the two caster totems
// and the blessings. Nothing here depends on the class.
export const defaultHealerRaidBuffs = (): RaidBuffs =>
	RaidBuffs.create({
		arcaneBrilliance: true,
		giftOfTheWild: true,
		prayerOfFortitude: true,
		prayerOfSpirit: true,
	});

export const defaultHealerPartyBuffs = (): PartyBuffs =>
	PartyBuffs.create({
		manaSpringTotem: TristateEffect.TristateEffectRegular,
	});

export const defaultHealerIndividualBuffs = (): IndividualBuffs =>
	IndividualBuffs.create({
		greaterBlessingOfKings: true,
		greaterBlessingOfWisdom: true,
	});
