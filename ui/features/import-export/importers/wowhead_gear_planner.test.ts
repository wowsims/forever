// The gear-planner hash is Wowhead's own format. Nothing about it is readable, so the test that
// means anything is a round trip: the *exporter* beside it carries an independent writer, and every
// field the importer hands on has to come back out of the link the exporter built.
import { Class, ItemSlot, ItemSpec, Race } from '@generated/proto/common';
import { WOWHEAD_DOMAIN } from '@sim/proto/wowhead';
import { fakeHost } from '@testing/sim/fake_host';
import { describe, expect, it, vi } from 'vitest';

import { WOWHEAD_GEAR_PLANNER_EXPORTER } from '../exporters/wowhead_gear_planner';
import type { IndividualImport } from './finish_individual_import';
import { WOWHEAD_GEAR_PLANNER_IMPORTER, WOWHEAD_SLOT_IDS } from './wowhead_gear_planner';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

const imported = vi.hoisted(() => [] as Array<IndividualImport>);
vi.mock('./finish_individual_import', () => ({
	finishIndividualImport: async (_host: unknown, parsed: IndividualImport) => {
		imported.push(parsed);
	},
}));

interface Equipped {
	id: number;
	enchantSpellId?: number;
	randomSuffix?: number;
}

interface Exported {
	className?: string;
	race?: Race;
	talents?: string;
	gear?: Partial<Record<ItemSlot, Equipped>>;
}

const exportLink = ({ className = 'Warrior', race = Race.RaceHuman, talents = '', gear = {} }: Exported) =>
	WOWHEAD_GEAR_PLANNER_EXPORTER.getData(
		fakeHost({
			player: {
				getPlayerClass: () => ({ friendlyName: className }),
				getRace: () => race,
				getTalentsString: () => talents,
				getGear: () => ({
					getItemSlots: () => Object.keys(gear).map(Number) as Array<ItemSlot>,
					getEquippedItem: (slot: ItemSlot) => {
						const item = gear[slot];
						if (!item) return null;
						return {
							id: item.id,
							_enchant: item.enchantSpellId ? { spellId: item.enchantSpellId } : null,
							_randomSuffix: item.randomSuffix ? { id: item.randomSuffix } : null,
						};
					},
				}),
			},
		}),
		{} as never,
	);

// Every item is known, and an enchant spell maps to the effect id one above it.
const importHost = fakeHost({
	sim: { db: { getItemById: () => ({}), enchantSpellIdToEnchant: (spellId: number) => ({ effectId: spellId + 1 }) } },
});

const importLink = async (url: string): Promise<IndividualImport> => {
	imported.length = 0;
	await WOWHEAD_GEAR_PLANNER_IMPORTER.onImport(importHost, url);
	return imported[0];
};

const roundTrip = async (exported: Exported) => importLink(await exportLink(exported));

describe('the Wowhead gear planner link', () => {
	it('carries class and race in the path', async () => {
		const parsed = await roundTrip({ className: 'Hunter', race: Race.RaceNightElf });
		expect(parsed.charClass).toBe(Class.ClassHunter);
		expect(parsed.race).toBe(Race.RaceNightElf);
	});

	it.each([Race.RaceHighOrderSkyborne, Race.RaceWindshaperSkyborne])('round-trips the multi-word race %s', async race => {
		expect((await roundTrip({ race })).race).toBe(race);
	});

	it('round-trips the three-tree talent string', async () => {
		const parsed = await roundTrip({ talents: '05002001-0550000502-05032' });
		expect(parsed.talentsStr).toBe('05002001-0550000502-05032');
	});

	// An empty talent string writes three separators and nothing else, and must not come back as '---'.
	it('reads an empty talent string back as empty', async () => {
		const parsed = await roundTrip({ talents: '' });
		expect(parsed.talentsStr).toBe('');
	});

	it('round-trips an item with an enchant and a random suffix', async () => {
		const parsed = await roundTrip({
			talents: '05002001',
			gear: { [ItemSlot.ItemSlotHead]: { id: 29011, enchantSpellId: 28909, randomSuffix: -5 } },
		});
		expect(parsed.equipmentSpec?.items).toEqual([ItemSpec.create({ id: 29011, enchant: 28910, randomSuffix: -5 })]);
	});

	// A Wowhead link can carry gem ids, three bytes each, after an item: they are skipped, and the
	// item after them still parses.
	it('skips the gem ids an item carries', async () => {
		const id3 = (id: number, high = 0) => [high | (id >> 16), (id >> 8) & 255, id & 255];
		const enchantOffset = 128;
		const bytes = [
			3, // format version
			60, // level
			0, // no talent bytes
			WOWHEAD_SLOT_IDS[ItemSlot.ItemSlotHead] | enchantOffset,
			...id3(29011, 2 << 5), // two gems
			28909 >> 8,
			28909 & 255,
			...id3(24027, 0 << 5),
			...id3(24061, 1 << 5),
			WOWHEAD_SLOT_IDS[ItemSlot.ItemSlotMainHand],
			...id3(28773),
		];
		const hash = btoa(String.fromCharCode(...bytes))
			.replace(/\+/g, '-')
			.replace(/\//g, '_')
			.replace(/=+$/, '');

		const parsed = await importLink(`https://www.wowhead.com/${WOWHEAD_DOMAIN}/gear-planner/warrior/human/${hash}`);

		expect(parsed.equipmentSpec?.items).toEqual([ItemSpec.create({ id: 29011, enchant: 28910 }), ItemSpec.create({ id: 28773 })]);
	});

	it('round-trips a bare item, leaving the optional fields off', async () => {
		const parsed = await roundTrip({ talents: '05002001', gear: { [ItemSlot.ItemSlotMainHand]: { id: 28773 } } });
		expect(parsed.equipmentSpec?.items).toEqual([ItemSpec.create({ id: 28773 })]);
	});

	// The ranged slot is slot id 18, and the last one the picker offers.
	it('round-trips a full seventeen-slot set, ranged included, in slot id order', async () => {
		const slots = (Object.keys(WOWHEAD_SLOT_IDS).map(Number) as Array<ItemSlot>).sort((a, b) => WOWHEAD_SLOT_IDS[a] - WOWHEAD_SLOT_IDS[b]);
		const gear = Object.fromEntries(slots.map((slot, index) => [slot, { id: 28000 + index * 37 }]));
		const parsed = await roundTrip({ talents: '05002001', gear });
		expect(slots).toHaveLength(17);
		expect(parsed.equipmentSpec?.items.map(item => item.id)).toEqual(slots.map((_, index) => 28000 + index * 37));
	});

	it('rejects a link that is not a gear planner link', async () => {
		await expect(importLink('https://www.wowhead.com/item=29011')).rejects.toThrow('import.wowhead.error_invalid_url');
	});

	it('rejects a gear planner link with no hash', async () => {
		await expect(importLink(`https://www.wowhead.com/${WOWHEAD_DOMAIN}/gear-planner/warrior/human`)).rejects.toThrow('import.wowhead.error_invalid_url');
	});
});
