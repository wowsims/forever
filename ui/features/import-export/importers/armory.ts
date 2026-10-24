// Character import from the Blizzard Profile API, through the Armory Worker (workers/armory), which
// holds the Battle.net credentials and returns a trimmed character. Dev-only: the Worker runs locally
// until the WoW Forever Profile API namespace is published.
import { Class, EquipmentSpec, ItemSlot, ItemSpec, Profession, Race } from '@generated/proto/common';
import i18n from '@i18n/config';
import { CHARACTER_LEVEL } from '@sim/constants/mechanics';
import { ARMORY_API_URL } from '@sim/constants/other';
import type { Database } from '@sim/proto/database';
import { nameToProfession, nameToRace } from '@sim/proto/names';
import type { IndividualSimHost } from '@sim/sim_host';
import { classTalentsConfig } from '@sim/talents/factory';
import { serializeTalentsString } from '@sim/talents/talents_string';
import { toastManager } from '@ui-kit/Toast';

import type { IndividualImport } from './finish_individual_import';
import { finishIndividualImport } from './finish_individual_import';

export const ARMORY_REGIONS = ['us', 'eu', 'kr', 'tw'] as const;
export type ArmoryRegion = (typeof ARMORY_REGIONS)[number];

export interface ArmoryLookup {
	region: ArmoryRegion;
	realm: string;
	// `First-Last` for a Forever character, or a single name.
	name: string;
}

// The Worker's response (`ArmoryCharacter` in workers/armory/src/character.ts).
export interface ArmoryCharacter {
	name: string;
	realm: string;
	level: number;
	classId: number;
	raceName: string;
	items: Array<{ slot: string; id: number; enchant?: number }>;
	talents: Array<{ spellId: number; talentId: number; rank: number }>;
	professions: string[];
}

// Reads a character name as typed or as the game writes it in logs: `Erlyn Farthing`, `Erlyn-Farthing`, or
// `Erlyn-Farthing-ClassicBetaPvE2` with the realm last. Name parts hold no hyphens or spaces in game, so a
// third part can only be the realm.
export const parseCharacterName = (input: string): { name: string; realm?: string } | null => {
	const parts = input
		.trim()
		.split(/[\s-]+/)
		.filter(Boolean);
	if (parts.length < 1 || parts.length > 3) return null;
	return { name: parts.slice(0, 2).join('-'), ...(parts.length == 3 ? { realm: parts[2] } : {}) };
};

// The query param a link uses to import a character on page load: `?armory=us/ClassicBetaPvE2/Erlyn-Farthing`.
export const ARMORY_LINK_PARAM = 'armory';

export const parseArmoryParam = (search: string): ArmoryLookup | null => {
	const value = new URLSearchParams(search).get(ARMORY_LINK_PARAM);
	const segments = value?.split('/').map(segment => segment.trim());
	if (!segments || segments.length != 3) return null;

	const [region, realm, name] = segments;
	const parsed = parseCharacterName(name);
	const lowerRegion = region.toLowerCase() as ArmoryRegion;
	if (!ARMORY_REGIONS.includes(lowerRegion) || !realm || !parsed) return null;
	return { region: lowerRegion, realm, name: parsed.name };
};

export const fetchArmoryCharacter = async ({ region, realm, name }: ArmoryLookup): Promise<ArmoryCharacter> => {
	const path = [region, realm.trim(), name.trim()].map(encodeURIComponent).join('/');
	let response: Response;
	try {
		response = await fetch(`${ARMORY_API_URL}/character/${path}`);
	} catch {
		throw new Error(i18n.t('import.armory.errors.unreachable', { url: ARMORY_API_URL }));
	}
	if (response.status == 404) throw new Error(i18n.t('import.armory.errors.not_found', { name, realm }));
	if (!response.ok) {
		const body = (await response.json().catch(() => null)) as { error?: string } | null;
		throw new Error(i18n.t('import.armory.errors.lookup_failed', { status: response.status, error: body?.error ?? '' }));
	}
	return response.json();
};

const ARMORY_SLOTS: Record<string, ItemSlot> = {
	HEAD: ItemSlot.ItemSlotHead,
	NECK: ItemSlot.ItemSlotNeck,
	SHOULDER: ItemSlot.ItemSlotShoulder,
	BACK: ItemSlot.ItemSlotBack,
	CHEST: ItemSlot.ItemSlotChest,
	WRIST: ItemSlot.ItemSlotWrist,
	HANDS: ItemSlot.ItemSlotHands,
	WAIST: ItemSlot.ItemSlotWaist,
	LEGS: ItemSlot.ItemSlotLegs,
	FEET: ItemSlot.ItemSlotFeet,
	FINGER_1: ItemSlot.ItemSlotFinger1,
	FINGER_2: ItemSlot.ItemSlotFinger2,
	TRINKET_1: ItemSlot.ItemSlotTrinket1,
	TRINKET_2: ItemSlot.ItemSlotTrinket2,
	MAIN_HAND: ItemSlot.ItemSlotMainHand,
	OFF_HAND: ItemSlot.ItemSlotOffHand,
	RANGED: ItemSlot.ItemSlotRanged,
};

export interface ArmoryImport extends IndividualImport {
	// Armory talents that match no talent in the class's trees, by spell id.
	unmatchedTalents: number[];
}

export const armoryToImport = (db: Database, character: ArmoryCharacter): ArmoryImport => {
	// The proto Class values are the game's class ids. Race values are the sim's own, so race goes by name.
	const charClass = Class[character.classId] ? (character.classId as Class) : Class.ClassUnknown;
	if (charClass == Class.ClassUnknown) throw new Error(i18n.t('import.errors.cannot_parse_class'));

	const race = nameToRace(character.raceName);
	if (race == Race.RaceUnknown) throw new Error(i18n.t('import.errors.cannot_parse_race'));

	const missingItems: number[] = [];
	const missingEnchants: number[] = [];
	const equipmentSpec = EquipmentSpec.create();
	// Slot order, since the gear lookup fills a slot pair (rings, trinkets) in the order it is given.
	const slotted = character.items.filter(item => item.slot in ARMORY_SLOTS).sort((a, b) => ARMORY_SLOTS[a.slot] - ARMORY_SLOTS[b.slot]);
	for (const item of slotted) {
		if (!db.getItemById(item.id)) {
			missingItems.push(item.id);
			continue;
		}
		const itemSpec = ItemSpec.create({ id: item.id });
		if (item.enchant) {
			if (db.getEnchantsByEffectId(item.enchant).length) {
				itemSpec.enchant = item.enchant;
			} else {
				missingEnchants.push(item.enchant);
			}
		}
		equipmentSpec.items.push(itemSpec);
	}

	// Each Forever talent reports one spell for all its ranks, and the Armory gives the rank with it.
	const config = classTalentsConfig[charClass];
	const points = config.map(tree => tree.talents.map(() => 0));
	const talentConfigs = config.flatMap((tree, treeIdx) => tree.talents.map((talentConfig, talentIdx) => ({ talentConfig, treeIdx, talentIdx })));
	const unmatchedTalents: number[] = [];
	for (const talent of character.talents) {
		const found = talentConfigs.find(
			({ talentConfig }) => talentConfig.spellId == talent.spellId || (!!talentConfig.definitionId && talentConfig.definitionId == talent.talentId),
		);
		if (!found) {
			unmatchedTalents.push(talent.spellId || talent.talentId);
			continue;
		}
		points[found.treeIdx][found.talentIdx] = Math.min(Math.max(talent.rank, 0), found.talentConfig.maxPoints);
	}

	return {
		charClass,
		race,
		equipmentSpec,
		talentsStr: serializeTalentsString(points),
		talentsForTheseTrees: true,
		professions: character.professions.map(nameToProfession).filter(profession => profession != Profession.ProfessionUnknown),
		missingItems,
		missingEnchants,
		unmatchedTalents,
	};
};

export const importArmoryCharacter = async (host: IndividualSimHost<any>, lookup: ArmoryLookup): Promise<void> => {
	const character = await fetchArmoryCharacter(lookup);

	// The sim only models max level, so anything below it is missing gear it cannot show.
	if (character.level < CHARACTER_LEVEL) {
		toastManager.add({ variant: 'warning', body: i18n.t('import.addon.below_max_level', { level: character.level }), delay: 5000 });
	}

	const { unmatchedTalents, ...parsed } = armoryToImport(host.sim.db, character);
	await finishIndividualImport(host, parsed);

	if (unmatchedTalents.length) {
		toastManager.add({
			variant: 'warning',
			body: i18n.t('import.armory.unmatched_talents', { ids: unmatchedTalents.join(', ') }),
			delay: 8000,
		});
	}
};
