import i18n from '@i18n/config';
import { CHARACTER_LEVEL } from '@sim/constants/mechanics';
import { raceNames } from '@sim/proto/names';
import { WOWHEAD_GEAR_PLANNER_URL } from '@sim/proto/wowhead';

import { WOWHEAD_SLOT_IDS } from '../importers';
import type { ExporterDefinition } from './types';

function writeTalents(talentStr: string): number[] {
	const bits: number[] = [];
	const t = (talentStr + '--').split('-', 3);
	for (let e = 0; e < t.length; e++) {
		for (let a = 0; a < t[e].length; a++) bits.push(parseInt(t[e].charAt(a)));
		bits.push(15);
	}
	return bits;
}

// Function to write the hash (reverse of readHash)
function writeHash(data: WowheadGearPlannerData): string {
	let hash = '';
	const enchantOffset = 128;
	const randomEnchantOffset = 64;

	// Initialize bits array
	const t = 3;
	const hashArray: number[] = [t];

	// Level
	hashArray.push(data.level ?? 0);

	// Talents
	const talentBits = writeTalents(data.talents);
	hashArray.push(Math.ceil(talentBits.length / 2));
	for (let e = 0; e < talentBits.length; e += 2) {
		hashArray.push((talentBits[e] << 4) | (talentBits[e + 1] || 0));
	}

	// Items
	const items = data.items ?? [];
	const itemArray: number[] = [];
	items.forEach(item => {
		itemArray.push(item.slotId);
	});

	for (let slotIndex = 0; slotIndex < itemArray.length; slotIndex++) {
		let bits = itemArray[slotIndex];
		const item = items[slotIndex];
		const id = item.itemId;
		const enchant = item.enchantId || 0;
		const randomEnchant = item.randomEnchantId || 0;
		if (enchant) bits |= enchantOffset;
		if (randomEnchant) bits |= randomEnchantOffset;

		hashArray.push(bits);
		hashArray.push((id >> 16) & 255, (id >> 8) & 255, id & 255);
		if (enchant) hashArray.push((enchant >> 8) & 255, enchant & 255);
		if (randomEnchant) hashArray.push((randomEnchant >> 8) & 255, randomEnchant & 255);
	}

	// Encode bits into characters
	if (hashArray.length <= 3) {
		return '';
	}

	const hashData = btoa(String.fromCharCode.apply(null, hashArray)).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');

	// Append the hash data to the URL
	if (hashData) {
		hash += hashData;
	}

	return hash;
}

interface WowheadGearPlannerData {
	class?: string;
	race?: string;
	level: number;
	talents: string;
	items: WowheadItemData[];
}

export interface WowheadItemData {
	slotId: number;
	itemId: number;
	randomEnchantId?: number;
	enchantId?: number;
}

function createWowheadGearPlannerLink(data: WowheadGearPlannerData): string {
	const baseUrl = '';
	const hash = writeHash(data);
	return baseUrl + hash;
}

export const WOWHEAD_GEAR_PLANNER_EXPORTER: ExporterDefinition = {
	title: i18n.t('export.wowhead.title'),
	allowDownload: true,
	holdsLink: true,
	getData: host => {
		const player = host.player;

		const classStr = player.getPlayerClass().friendlyName.replaceAll(/\s/g, '-').toLowerCase();
		const raceStr = raceNames.get(player.getRace())!.replaceAll(/\s/g, '-').toLowerCase();
		const url = `${WOWHEAD_GEAR_PLANNER_URL}/${classStr}/${raceStr}/`;

		const data: WowheadGearPlannerData = {
			level: CHARACTER_LEVEL,
			talents: player.getTalentsString(),
			items: [],
		};

		const gear = player.getGear();

		gear.getItemSlots()
			.sort((slot1, slot2) => WOWHEAD_SLOT_IDS[slot1] - WOWHEAD_SLOT_IDS[slot2])
			.forEach(itemSlot => {
				const item = gear.getEquippedItem(itemSlot);
				if (!item) {
					return;
				}

				const slotId = WOWHEAD_SLOT_IDS[itemSlot];
				const itemData = {
					slotId: slotId,
					itemId: item.id,
				} as WowheadItemData;
				if (item._randomSuffix?.id) {
					itemData.randomEnchantId = item._randomSuffix.id;
				}
				itemData.enchantId = item._enchant?.spellId;
				data.items.push(itemData);
			});

		const hash = createWowheadGearPlannerLink(data);

		return url + hash;
	},
};
