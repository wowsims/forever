import { Class, EquipmentSpec, ItemSlot, ItemSpec, Profession, Race } from '@generated/proto/common';
import i18n from '@i18n/config';
import { CHARACTER_LEVEL } from '@sim/constants/mechanics';
import { nameToClass, nameToRace } from '@sim/proto/names';
import { WOWHEAD_DOMAIN, WOWHEAD_GEAR_PLANNER_URL } from '@sim/proto/wowhead';

import { finishIndividualImport } from './finish_individual_import';
import type { ImporterDefinition } from './types';

interface WowheadGearPlannerImportJSON {
	classId: string;
	raceId: string;
	genderId: number;
	level: number;
	specIndex: number;
	talentString: string;
	items: {
		slotId: number;
		itemId: number;
		randomEnchantId?: number;
		enchantId?: number;
	}[];
}

// Taken from Wowhead
function readHash(hash: string): WowheadGearPlannerImportJSON {
	const enchantOffset = 128;
	const randomEnchantOffset = 64;

	const t: WowheadGearPlannerImportJSON = {
		classId: '',
		raceId: '',
		genderId: 0,
		level: 0,
		specIndex: 0,
		talentString: '',
		items: [],
	};
	const l = /^([a-z-]+)\/([a-z-]+)(?:\/([a-zA-Z0-9_-]+))?$/.exec(hash);
	if (!l) return t;

	t.classId = l[1];
	t.raceId = l[2];

	t.level = CHARACTER_LEVEL;

	if (!l[3]) return t;

	const gearPlannerString = l[3].replace(/-/g, '+').replace(/_/g, '/');
	const gearPlannerHash = atob(gearPlannerString);
	const gearPlannerBits: number[] = [];
	for (let e = 0; e < gearPlannerHash.length; e++) {
		gearPlannerBits.push(gearPlannerHash.charCodeAt(e));
	}

	const r = gearPlannerBits.shift()!;
	if (r < 4) {
		if (r > 0) {
			t.level = gearPlannerBits.shift()!;
		}
		t.talentString = '';
		if (r > 1) {
			const e = gearPlannerBits.shift()!;
			const a = gearPlannerBits.splice(0, e);
			const s = [];
			for (let e = 0; e < a.length; e++) {
				s.push(a[e] >> 4, a[e] & 15);
			}
			let n = 0;
			for (let e = 0; e < s.length && n < 3; e++) {
				if (s[e] === 15) {
					t.talentString += '-';
					n++;
				} else {
					t.talentString += '' + s[e];
				}
			}
			t.talentString = t.talentString.replace(/-+$/, '');
		}
		while (gearPlannerBits.length >= 3) {
			let slotIdx = gearPlannerBits.shift()!;
			let n = 0;
			let itemID = 0;
			if (r >= 3) {
				const e = gearPlannerBits.shift()!;
				n = (e & 224) >> 5;
				itemID |= (e & 31) << 16;
			}
			itemID |= gearPlannerBits.shift()! << 8;
			itemID |= gearPlannerBits.shift()!;
			const o = (slotIdx & enchantOffset) > 0;
			const c = (slotIdx & randomEnchantOffset) > 0;
			slotIdx = slotIdx & ~enchantOffset & ~randomEnchantOffset;
			const item: WowheadGearPlannerImportJSON['items'][number] = {
				slotId: slotIdx,
				itemId: itemID,
				randomEnchantId: undefined,
				enchantId: undefined,
			};

			if (o) {
				let enchantId = gearPlannerBits.shift()! << 8;
				enchantId |= gearPlannerBits.shift()!;
				item.enchantId = enchantId;
			}
			if (c) {
				let randomEnchantId = gearPlannerBits.shift()! << 8;
				randomEnchantId |= gearPlannerBits.shift()!;
				if ((randomEnchantId & 32768) > 0) {
					randomEnchantId -= 65536;
				}
				item.randomEnchantId = randomEnchantId;
			}
			// Skips the item's n gem ids, three bytes each.
			gearPlannerBits.splice(0, 3 * n);
			t.items.push(item);
		}
	}

	return t;
}

function parseWowheadGearLink(link: string): WowheadGearPlannerImportJSON {
	const match = link.match(new RegExp(`${WOWHEAD_DOMAIN}/gear-planner/(.+)`));
	if (!match) {
		throw new Error(i18n.t('import.wowhead.error_invalid_link', { url: link, example: `${WOWHEAD_GEAR_PLANNER_URL}/CLASS/RACE/XXXX` }));
	}
	return readHash(match[1]);
}

export const WOWHEAD_SLOT_IDS: Record<ItemSlot, number> = {
	[ItemSlot.ItemSlotHead]: 1,
	[ItemSlot.ItemSlotNeck]: 2,
	[ItemSlot.ItemSlotShoulder]: 3,
	[ItemSlot.ItemSlotBack]: 15,
	[ItemSlot.ItemSlotChest]: 5,
	[ItemSlot.ItemSlotWrist]: 9,
	[ItemSlot.ItemSlotHands]: 10,
	[ItemSlot.ItemSlotWaist]: 6,
	[ItemSlot.ItemSlotLegs]: 7,
	[ItemSlot.ItemSlotFeet]: 8,
	[ItemSlot.ItemSlotFinger1]: 11,
	[ItemSlot.ItemSlotFinger2]: 12,
	[ItemSlot.ItemSlotTrinket1]: 13,
	[ItemSlot.ItemSlotTrinket2]: 14,
	[ItemSlot.ItemSlotMainHand]: 16,
	[ItemSlot.ItemSlotOffHand]: 17,
	[ItemSlot.ItemSlotRanged]: 18,
};

export const WOWHEAD_GEAR_PLANNER_IMPORTER: ImporterDefinition = {
	title: i18n.t('import.wowhead.title'),
	allowFileUpload: true,
	holdsLink: true,
	onImport: async (host, url) => {
		const match = url.match(new RegExp(`www\\.wowhead\\.com/${WOWHEAD_DOMAIN}/gear-planner/([a-z\\-]+)/([a-z\\-]+)/([a-zA-Z0-9_\\-]+)`));
		if (!match) {
			throw new Error(i18n.t('import.wowhead.error_invalid_url', { url }));
		}
		const missingItems: number[] = [];
		const missingEnchants: number[] = [];
		const professions: Profession[] = [];

		const parsed = parseWowheadGearLink(url);
		const charClass = nameToClass(parsed.classId.replaceAll('-', ''));
		if (charClass == Class.ClassUnknown) {
			throw new Error(i18n.t('import.wowhead.error_cannot_parse_class', { classId: parsed.classId }));
		}

		const race = nameToRace(parsed.raceId.replaceAll('-', ''));
		if (race == Race.RaceUnknown) {
			throw new Error(i18n.t('import.wowhead.error_cannot_parse_race', { raceId: parsed.raceId }));
		}

		const equipmentSpec = EquipmentSpec.create();

		parsed.items.forEach(item => {
			const dbItem = host.sim.db.getItemById(item.itemId);
			if (!dbItem) {
				missingItems.push(item.itemId);
				return;
			}
			const itemSpec = ItemSpec.create();
			itemSpec.id = item.itemId;
			const slotId = item.slotId;
			if (item.enchantId) {
				const enchant = host.sim.db.enchantSpellIdToEnchant(item.enchantId);
				if (!enchant) {
					missingEnchants.push(item.enchantId);
					return;
				}
				itemSpec.enchant = enchant.effectId;
			}
			if (item.randomEnchantId) {
				itemSpec.randomSuffix = item.randomEnchantId;
			}
			const itemSlotEntry = Object.entries(WOWHEAD_SLOT_IDS).find(e => e[1] == slotId);
			if (itemSlotEntry != null) {
				equipmentSpec.items.push(itemSpec);
			}
		});

		await finishIndividualImport(host, {
			charClass,
			race,
			equipmentSpec,
			talentsStr: parsed.talentString ?? '',
			professions,
			missingEnchants,
			missingItems,
		});
	},
};
