import { type VersusEffect, VersusEffectKind } from '@generated/proto/api';
import { type AreaType, CreatureFamily, MobType, type Target } from '@generated/proto/common';
import type { UIItem } from '@generated/proto/ui';
import i18n from '@i18n/config';

/** `partial` when some targets match and others do not; an area is only ever `on` or `off`. */
export type VersusState = 'on' | 'partial' | 'off';

export interface VersusMatch {
	state: VersusState;
	matched: number;
	total: number;
}

export interface VersusChip extends VersusMatch {
	kind: 'mobType' | 'family' | 'area';
	/** The `MobType`, `CreatureFamily` or `AreaType`, by `kind`. */
	value: number;
}

// Zero matched is `off` whatever the total, so an encounter with no targets reads as `off`.
const countMatch = (matched: number, total: number): VersusMatch => ({
	state: matched === 0 ? 'off' : matched === total ? 'on' : 'partial',
	matched,
	total,
});

// A copy of the sim's table of the mob type each creature family belongs to (`creatureFamilyMobTypes` in
// sim/core/versus.go), which this one follows: Aquatic and Serpentine span several mob types and give none.
const FAMILY_MOB_TYPES: Record<CreatureFamily, MobType> = {
	[CreatureFamily.CreatureFamilyUnknown]: MobType.MobTypeUnknown,
	[CreatureFamily.CreatureFamilyNaga]: MobType.MobTypeHumanoid,
	[CreatureFamily.CreatureFamilySatyr]: MobType.MobTypeDemon,
	[CreatureFamily.CreatureFamilyAquatic]: MobType.MobTypeUnknown,
	[CreatureFamily.CreatureFamilyEarthElemental]: MobType.MobTypeElemental,
	[CreatureFamily.CreatureFamilyFireElemental]: MobType.MobTypeElemental,
	[CreatureFamily.CreatureFamilyMountainGiant]: MobType.MobTypeGiant,
	[CreatureFamily.CreatureFamilySwine]: MobType.MobTypeBeast,
	[CreatureFamily.CreatureFamilyPlant]: MobType.MobTypeElemental,
	[CreatureFamily.CreatureFamilyTroll]: MobType.MobTypeHumanoid,
	[CreatureFamily.CreatureFamilyOrc]: MobType.MobTypeHumanoid,
	[CreatureFamily.CreatureFamilyOgre]: MobType.MobTypeHumanoid,
	[CreatureFamily.CreatureFamilyWolf]: MobType.MobTypeBeast,
	[CreatureFamily.CreatureFamilyWorgen]: MobType.MobTypeHumanoid,
	[CreatureFamily.CreatureFamilyMurloc]: MobType.MobTypeHumanoid,
	[CreatureFamily.CreatureFamilyFurbolg]: MobType.MobTypeHumanoid,
	[CreatureFamily.CreatureFamilyDwarf]: MobType.MobTypeHumanoid,
	[CreatureFamily.CreatureFamilyDinosaur]: MobType.MobTypeBeast,
	[CreatureFamily.CreatureFamilySerpentine]: MobType.MobTypeUnknown,
};

/** The target's own mob type, or its family's where it has none, as the sim counts it. */
export const targetMobType = (target: Target): MobType =>
	target.mobType === MobType.MobTypeUnknown ? (FAMILY_MOB_TYPES[target.creatureFamily] ?? MobType.MobTypeUnknown) : target.mobType;

/** A target of `MobTypeUnknown` with no family of a mob type never matches. */
export const mobTypeState = (mobType: MobType, targets: ReadonlyArray<Target>): VersusMatch =>
	countMatch(mobType === MobType.MobTypeUnknown ? 0 : targets.filter(target => targetMobType(target) === mobType).length, targets.length);

/** Matches on the target's family alone, as the sim does; a target of `CreatureFamilyUnknown` never matches. */
export const familyState = (family: CreatureFamily, targets: ReadonlyArray<Target>): VersusMatch =>
	countMatch(family === CreatureFamily.CreatureFamilyUnknown ? 0 : targets.filter(target => target.creatureFamily === family).length, targets.length);

export const areaState = (areaType: AreaType, areaTypes: ReadonlyArray<AreaType>): VersusMatch => countMatch(areaTypes.includes(areaType) ? 1 : 0, 1);

export const namesVersus = (item: UIItem): boolean => item.mobTypes.length > 0 || item.creatureFamilies.length > 0 || item.areaTypes.length > 0;

export const itemVersusChips = (item: UIItem, targets: ReadonlyArray<Target>, areaTypes: ReadonlyArray<AreaType>): Array<VersusChip> => [
	...item.mobTypes.map(value => ({ kind: 'mobType' as const, value, ...mobTypeState(value, targets) })),
	...item.creatureFamilies.map(value => ({ kind: 'family' as const, value, ...familyState(value, targets) })),
	...item.areaTypes.map(value => ({ kind: 'area' as const, value, ...areaState(value, areaTypes) })),
];

const EFFECT_KEYS: Record<VersusEffectKind, string> = {
	[VersusEffectKind.VersusEffectKindDamage]: 'effect_damage',
	[VersusEffectKind.VersusEffectKindRestore]: 'effect_restore',
	[VersusEffectKind.VersusEffectKindProcChance]: 'effect_proc_chance',
	[VersusEffectKind.VersusEffectKindCritChance]: 'effect_crit_chance',
	[VersusEffectKind.VersusEffectKindProcOnly]: 'effect_proc_only',
	[VersusEffectKind.VersusEffectKindDuration]: 'effect_duration',
};

export const versusEffectFactor = (effect: VersusEffect): string =>
	i18n.t(`sidebar.character_stats.versus.${EFFECT_KEYS[effect.kind]}`, { factor: Number(effect.multiplier.toFixed(2)) });
