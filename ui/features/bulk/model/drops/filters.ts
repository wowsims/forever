import { formatSearchQuery } from '@features/gear/model/item_search';
import { DatabaseFilters, type UIItem } from '@generated/proto/ui';
import { Player } from '@sim/player/player';
import type { PlayerClass } from '@sim/player/player_class';
import { getEligibleItemSlots } from '@sim/proto/items';
import { Sim } from '@sim/sim';

import type { DropsItem, DropsItemList } from './items';
import { type DropsNames, dropsSourceLabel } from './sources';

// Search and these filters only change what the items dialog shows: a ticked item they hide still runs.

type DropsFilterTypes = Pick<PlayerClass<any>, 'armorTypes' | 'weaponTypes' | 'rangedWeaponTypes'>;

/** Every armor type the class wears and every weapon and ranged type, with no bound set. */
export const dropsDefaultFilters = (playerClass: DropsFilterTypes): DatabaseFilters =>
	DatabaseFilters.create({
		armorTypes: playerClass.armorTypes.slice(),
		weaponTypes: Sim.ALL_WEAPON_TYPES.slice(),
		rangedWeaponTypes: Sim.ALL_RANGED_WEAPON_TYPES.slice(),
	});

// The checks run as for the gear picker's list of the item's first slot. Drops lists no random-suffix item.
export const dropsPassesFilters = (item: UIItem, filters: DatabaseFilters): boolean =>
	Player.passesItemFilters(item, filters, getEligibleItemSlots(item)[0], () => []);

/** Every word of `search` is in the item's name or in the name of one of its ticked sources, as in the gear picker. */
export const dropsMatchesSearch = (entry: DropsItem, search: string, names: DropsNames): boolean => {
	const words = formatSearchQuery(search).split(' ').filter(Boolean);
	if (words.length === 0) return true;
	const texts = [entry.item.name, ...entry.sourceKeys.map(key => dropsSourceLabel(key, names).label)].map(formatSearchQuery);
	return words.every(word => texts.some(text => text.includes(word)));
};

/**
 * How many of the dialog's filter rows narrow the list: each level bound set, the stat, area, mob type and creature group
 * picks, an armor, weapon or ranged type of the class left unticked, and any weapon speed bound, each once.
 */
export const dropsActiveFilterCount = (filters: DatabaseFilters, playerClass: DropsFilterTypes): number =>
	[
		filters.minIlvl,
		filters.maxIlvl,
		filters.itemStats.length,
		filters.areaTypes.length,
		filters.mobTypes.length,
		filters.creatureFamilies.length,
		playerClass.armorTypes.some(type => !filters.armorTypes.includes(type)),
		playerClass.weaponTypes.some(({ weaponType }) => !filters.weaponTypes.includes(weaponType)),
		playerClass.rangedWeaponTypes.some(type => !filters.rangedWeaponTypes.includes(type)),
		filters.minMhWeaponSpeed || filters.maxMhWeaponSpeed || filters.minOhWeaponSpeed || filters.maxOhWeaponSpeed,
		filters.minRangedWeaponSpeed || filters.maxRangedWeaponSpeed,
	].filter(Boolean).length;

export const dropsHiddenTicked = (list: DropsItemList, filters: DatabaseFilters): number =>
	list.items.filter(entry => entry.included && !entry.worn && !dropsPassesFilters(entry.item, filters)).length;
