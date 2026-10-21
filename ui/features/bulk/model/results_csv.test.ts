import { ItemSlot } from '@generated/proto/common';
import type { BulkEncounterResults, TopGearResult } from '@sim/bulk/types';
import { ActionId } from '@sim/proto/action_id';
import type { EquippedItem } from '@sim/proto/equipped_item';
import type { Gear } from '@sim/proto/gear';
import { stDevToConf95 } from '@sim/utils/math';
import { describe, expect, it } from 'vitest';

import { bulkResultsCsv } from './results_csv';

const SLOTS = [ItemSlot.ItemSlotHead, ItemSlot.ItemSlotMainHand];

const equipped = (id: number, name: string, enchant?: { name: string; effectId: number }, suffix?: { id: number; name: string }) =>
	({ id, item: { id, name }, enchant: enchant ?? null, randomSuffix: suffix ?? null }) as unknown as EquippedItem;

// A gear set over two slots; sets are equal when they are the same object.
const gear = (head: EquippedItem | null, mainHand: EquippedItem | null): Gear => {
	const set = {
		getItemSlots: () => SLOTS,
		getEquippedItem: (slot: ItemSlot) => (slot === ItemSlot.ItemSlotHead ? head : mainHand),
		equals: (other: Gear) => other === set,
	};
	return set as unknown as Gear;
};

const result = (set: Gear, avg: number, stdev = 0): TopGearResult => ({ gear: set, dpsMetrics: { avg, stdev } as never });

const LIONHEART = equipped(12640, 'Lionheart Helm');
const THUNDERFURY = equipped(19019, 'Thunderfury, Blessed Blade of the Windseeker', { name: 'Crusader', effectId: 1900 });
const BEAR_HELM = equipped(10001, 'Helm of "the" Wild', undefined, { id: 23, name: 'of the Bear' });
const ASHKANDI = equipped(19364, 'Ashkandi');
const CURRENT_GEAR = gear(LIONHEART, THUNDERFURY);
const UPGRADE = gear(BEAR_HELM, ASHKANDI);

// A timed encounter of 180 ±5 seconds, or a health-based one when `timed` is false.
const entry = (name: string, isCurrent: boolean, best: number, timed = true, upgrade = UPGRADE): BulkEncounterResults => {
	const base = result(CURRENT_GEAR, 1000, 10);
	return {
		name,
		isCurrent,
		duration: timed ? 180 : null,
		durationVariation: timed ? 5 : null,
		results: { chains: [[result(upgrade, best, 10)], [base]], originalGearResults: base, iterations: 100 },
	};
};

const lines = (csv: string) => csv.trimEnd().split('\n');

describe('bulkResultsCsv', () => {
	it('heads the file with the result columns, then four columns per gear slot', () => {
		expect(lines(bulkResultsCsv([entry('', true, 1100)]))[0]).toBe(
			[
				'encounter,duration,variation,rank,current_gear,dps,dps_margin_95,diff,diff_pct',
				'slots.head,slots.head wowhead,slots.head item id,slots.head enchant id',
				'slots.main_hand,slots.main_hand wowhead,slots.main_hand item id,slots.main_hand enchant id',
			].join(','),
		);
	});

	it('writes one row per ranked set of every encounter, in ranked order, with the length in seconds', () => {
		const rows = lines(bulkResultsCsv([entry('', true, 1100), entry('Patchwerk', false, 1050, false)])).slice(1);

		expect(rows.map(row => row.split(',').slice(0, 5))).toEqual([
			['bulk_tab.results.current', '180', '5', '1', ''],
			['bulk_tab.results.current', '180', '5', '2', 'yes'],
			['Patchwerk', '', '', '1', ''],
			['Patchwerk', '', '', '2', 'yes'],
		]);
	});

	it('keeps the DPS and the difference from the current gear at full precision', () => {
		const [, upgrade, current] = lines(bulkResultsCsv([entry('', true, 1036.125)]));

		expect(upgrade.split(',').slice(5, 9)).toEqual(['1036.125', String(stDevToConf95(10, 100)), '36.125', '3.6125']);
		expect(current.split(',').slice(7, 9)).toEqual(['0', '0']);
	});

	it('writes each slot as its name, wowhead link, item id and enchant id, quoting names holding commas or quotes', () => {
		const [, upgrade, current] = lines(bulkResultsCsv([entry('', true, 1100)]));
		const link = (id: number, suffixId?: number) => ActionId.makeItemUrl(id, suffixId);

		const thunderfury = `"Thunderfury, Blessed Blade of the Windseeker (Crusader)",${link(19019)},19019,1900`;
		expect(upgrade.endsWith(`,"Helm of ""the"" Wild of the Bear",${link(10001, 23)},10001,,Ashkandi,${link(19364)},19364,`)).toBe(true);
		expect(current.endsWith(`,Lionheart Helm,${link(12640)},12640,,${thunderfury}`)).toBe(true);
	});

	it('links a suffixed item to wowhead with the suffix as its bonus', () => {
		const [, upgrade] = lines(bulkResultsCsv([entry('', true, 1100)]));

		expect(upgrade).toMatch(/,https:\/\/wowhead\.com\/forever\/item=10001\?level=\d+&bonus=23,10001,/);
	});

	it('leaves all four cells of an empty slot empty', () => {
		const [, upgrade] = lines(bulkResultsCsv([entry('', true, 1100, true, gear(null, ASHKANDI))]));

		expect(upgrade).toContain(`,,,,,Ashkandi,${ActionId.makeItemUrl(19364)},19364,`);
	});
});
