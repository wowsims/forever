import { ArmorType, RangedWeaponType, Stat, WeaponType } from '@generated/proto/common';
import { DatabaseFilters } from '@generated/proto/ui';
import { PlayerSpecs } from '@sim/player/specs';
import { dropsItemListFixture, dropsNamesFixture, UNTICKED_ITEM_ID, WORN_ITEM_ID } from '@testing/storybook/fixtures/bulk-drops';
import { describe, expect, it } from 'vitest';

import { dropsActiveFilterCount, dropsDefaultFilters, dropsHiddenTicked, dropsMatchesSearch, dropsPassesFilters } from './filters';
import type { DropsItemList } from './items';

const WARRIOR = PlayerSpecs.getPlayerClass(PlayerSpecs.DpsWarrior);
const FELSTRIKER = 2001;
const BLACKBLADE = 2002;
const DRACORIAN = 2004;
const REIVER = 2005;
const MARK = 2006;
const LIONHEART = 2011;

const list = dropsItemListFixture();
const filtersWith = (fields: Partial<DatabaseFilters>) => ({ ...dropsDefaultFilters(WARRIOR), ...fields });
const passing = (filters: DatabaseFilters, from: DropsItemList = list) =>
	from.items
		.filter(entry => dropsPassesFilters(entry.item, filters))
		.map(entry => entry.item.id)
		.sort((a, b) => a - b);
const matching = (search: string) =>
	list.items
		.filter(entry => dropsMatchesSearch(entry, search, dropsNamesFixture()))
		.map(entry => entry.item.id)
		.sort((a, b) => a - b);
const allIds = list.items.map(entry => entry.item.id).sort((a, b) => a - b);

describe('dropsDefaultFilters', () => {
	it('lets every listed item through', () => {
		expect(passing(dropsDefaultFilters(WARRIOR))).toEqual(allIds);
	});
});

describe('dropsPassesFilters', () => {
	it('checks armor on the armor slots only, by the item’s own slot', () => {
		expect(passing(filtersWith({ armorTypes: [ArmorType.ArmorTypeMail] }))).toEqual([FELSTRIKER, BLACKBLADE, WORN_ITEM_ID, MARK]);
	});

	it('checks weapon types on the weapons only', () => {
		expect(passing(filtersWith({ weaponTypes: [WeaponType.WeaponTypeMace] }))).toEqual([DRACORIAN, REIVER, MARK, UNTICKED_ITEM_ID, LIONHEART]);
	});

	it('takes the item level bounds and the picked stats', () => {
		expect(passing(filtersWith({ minIlvl: 62 }))).toEqual(allIds.filter(id => id !== REIVER && id !== UNTICKED_ITEM_ID));
		expect(passing(filtersWith({ itemStats: [Stat.StatStrength] }))).toEqual([]);
	});
});

describe('dropsMatchesSearch', () => {
	it('matches every word against the name or a source’s name', () => {
		expect(matching('')).toEqual(allIds);
		expect(matching('claws')).toEqual([REIVER]);
		expect(matching('rivendare')).toEqual([FELSTRIKER, DRACORIAN]);
		expect(matching('rivendare felstriker')).toEqual([FELSTRIKER]);
		expect(matching('Rivendare gauntlets')).toEqual([DRACORIAN]);
		expect(matching('stratholme')).toEqual([]);
	});
});

describe('dropsActiveFilterCount', () => {
	it('counts nothing at the defaults, and each narrowed row once', () => {
		expect(dropsActiveFilterCount(dropsDefaultFilters(WARRIOR), WARRIOR)).toBe(0);
		expect(
			dropsActiveFilterCount(
				filtersWith({
					minIlvl: 60,
					itemStats: [Stat.StatStrength, Stat.StatAgility],
					armorTypes: [ArmorType.ArmorTypePlate],
					rangedWeaponTypes: [RangedWeaponType.RangedWeaponTypeBow],
					maxMhWeaponSpeed: 2.6,
				}),
				WARRIOR,
			),
		).toBe(5);
	});
});

describe('dropsHiddenTicked', () => {
	it('counts the ticked, non-worn items the filters hide', () => {
		const filters = filtersWith({ armorTypes: [ArmorType.ArmorTypeMail] });
		// Plate hands: Dracorian, Reiver and Lionheart are ticked, Death Grips is not. The worn Doomsaw passes anyway.
		expect(dropsHiddenTicked(list, filters)).toBe(3);
		expect(dropsHiddenTicked(list, filtersWith({ weaponTypes: [] }))).toBe(2);
	});
});
