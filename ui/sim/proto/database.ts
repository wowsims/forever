import { EquipmentSpec, ItemRandomSuffix, ItemSlot, ItemSpec, ItemSwap, PresetEncounter, PresetTarget } from '@generated/proto/common';
import { Consumable, ItemEffectRandPropPoints, SimDatabase } from '@generated/proto/db';
import { SpellEffect } from '@generated/proto/spell';
import { IconData, UIDatabase, UIEnchant as Enchant, UIFaction, UIItem as Item, UINPC as Npc, UIZone as Zone } from '@generated/proto/ui';

import { CacheHandler } from '../cache/cache_handler';
import { CHARACTER_LEVEL } from '../constants/mechanics';
import { arrayEquals, distinct } from '../utils/collections';
import { EquippedItem } from './equipped_item';
import { Gear, ItemSwapGear } from './gear';
import { enchantAppliesToItem, getEligibleEnchantSlots, getEligibleItemSlots } from './items';
import { Stats } from './stats';
import { WOWHEAD_DOMAIN, WOWHEAD_EXPANSION_ENV } from './wowhead';

const dbUrlJson = '/forever/assets/database/db.json';
const dbUrlBin = '/forever/assets/database/db.bin';
const leftoversUrlJson = '/forever/assets/database/leftover_db.json';
const leftoversUrlBin = '/forever/assets/database/leftover_db.bin';
// When changing this value, don't forget to change the html <link> for preloading!
const READ_JSON = true;
const RANK_REGEX = /Rank ([0-9]+)/g;

const iconRequestCache = new CacheHandler<Promise<IconData>>();

export class Database {
	private static loadPromise: Promise<Database> | null = null;
	private static instance: Database | null = null;

	/** Shared by the whole page, so it takes no AbortSignal; retries, and clears the memo on failure. */
	static async get(): Promise<Database> {
		if (!Database.loadPromise) {
			Database.loadPromise = (async () => {
				const dbData = await Database.loadWithRetry();
				const db = new Database(dbData);
				Database.instance = db;
				return db;
			})().catch(error => {
				// Never memoize a failure (cf. getSharedWasmModule): one transient error must not brick every later caller.
				Database.loadPromise = null;
				throw error;
			});
		}
		return Database.loadPromise;
	}

	private static async loadWithRetry(attempts = 3, backoffMs = 250): Promise<UIDatabase> {
		for (let attempt = 1; ; attempt++) {
			try {
				if (READ_JSON) {
					const resp = await fetch(dbUrlJson);
					return UIDatabase.fromJson(await resp.json());
				}
				const buf = await fetch(dbUrlBin).then(r => r.arrayBuffer());
				return UIDatabase.fromBinary(new Uint8Array(buf));
			} catch (error) {
				if (attempt >= attempts) throw error;
				console.warn(`Database load attempt ${attempt} failed, retrying:`, error);
				await new Promise(resolve => setTimeout(resolve, backoffMs * attempt));
			}
		}
	}

	static getSync(): Database {
		if (!Database.instance) {
			throw new Error('Database not yet loaded; call `await Database.get()` before using getSync()');
		}
		return Database.instance;
	}

	static async getLeftovers(): Promise<UIDatabase> {
		if (READ_JSON) {
			return fetch(leftoversUrlJson)
				.then(response => response.json())
				.then(json => UIDatabase.fromJson(json));
		} else {
			return fetch(leftoversUrlBin)
				.then(response => response.arrayBuffer())
				.then(buffer => UIDatabase.fromBinary(new Uint8Array(buffer)));
		}
	}

	// Checks if any items in the equipment are missing from the current DB. If so, loads the leftover DB.
	static async loadLeftoversIfNecessary(equipment: EquipmentSpec): Promise<Database> {
		const db = await Database.get();
		if (db.loadedLeftovers) {
			return db;
		}

		const shouldLoadLeftovers = equipment.items.some(item => item.id != 0 && !db.items.has(item.id));
		if (shouldLoadLeftovers) {
			const leftoverDb = await Database.getLeftovers();
			db.loadProto(leftoverDb);
			db.loadedLeftovers = true;
		}
		return db;
	}

	private readonly items = new Map<number, Item>();
	private readonly randomSuffixes = new Map<number, ItemRandomSuffix>();
	private readonly itemEffectRandPropPoints = new Map<number, ItemEffectRandPropPoints>();
	private readonly enchantsBySlot: Partial<Record<ItemSlot, Enchant[]>> = {};
	private readonly enchantsByEffectId = new Map<number, Enchant[]>();
	private readonly npcs = new Map<number, Npc>();
	private readonly zones = new Map<number, Zone>();
	private readonly factions = new Map<number, UIFaction>();
	private readonly presetEncounters = new Map<string, PresetEncounter>();
	private readonly presetTargets = new Map<string, PresetTarget>();
	private readonly itemIcons: Record<number, IconData> = {};
	private readonly spellIcons: Record<number, IconData> = {};
	private readonly consumables = new Map<number, Consumable>();
	private readonly spellEffects = new Map<number, SpellEffect>();

	private loadedLeftovers = false;

	private constructor(db: UIDatabase) {
		this.loadProto(db);
	}

	/** The spells the database has an icon for, with their names and ranks (the ranks of a spell are what Rotation Master's spell ladders are built from). */
	getSpellIcons(): IconData[] {
		return Object.values(this.spellIcons);
	}

	// Add all data from the db proto into this database.
	private loadProto(db: UIDatabase) {
		db.items.forEach(item => {
			const itemCopy = { ...item };
			// Pre populate the item with stats from the highest base state the item can have.
			// We use this in EP calculations
			const maxScaling = item.scalingOptions[Math.max(...Object.keys(item.scalingOptions).map(Number))];
			itemCopy.weaponDamageMax = maxScaling.weaponDamageMax;
			itemCopy.weaponDamageMin = maxScaling.weaponDamageMin;
			itemCopy.randPropPoints = maxScaling.randPropPoints;
			itemCopy.ilvl = maxScaling.ilvl;
			itemCopy.stats = Stats.fromMap(maxScaling.stats).asProtoArray();

			this.items.set(itemCopy.id, itemCopy);
		});
		db.randomSuffixes.forEach(randomSuffix => this.randomSuffixes.set(randomSuffix.id, randomSuffix));
		db.itemEffectRandPropPoints.forEach(ieRpp => this.itemEffectRandPropPoints.set(ieRpp.ilvl, ieRpp));
		db.enchants.forEach(enchant => {
			const rows = this.enchantsByEffectId.get(enchant.effectId);
			if (rows) rows.push(enchant);
			else this.enchantsByEffectId.set(enchant.effectId, [enchant]);
			const slots = getEligibleEnchantSlots(enchant);
			slots.forEach(slot => {
				if (!this.enchantsBySlot[slot]) {
					this.enchantsBySlot[slot] = [];
				}
				this.enchantsBySlot[slot]!.push(enchant);
			});
		});

		db.npcs.forEach(npc => this.npcs.set(npc.id, npc));
		db.zones.forEach(zone => this.zones.set(zone.id, zone));
		db.factions.forEach(faction => this.factions.set(faction.id, faction));
		db.encounters.forEach(encounter => this.presetEncounters.set(encounter.path, encounter));
		db.encounters
			.map(e => e.targets)
			.flat()
			.forEach(target => this.presetTargets.set(target.path, target));

		db.items.forEach(
			item =>
				(this.itemIcons[item.id] = IconData.create({
					id: item.id,
					name: item.name,
					icon: item.icon,
				})),
		);
		db.itemIcons.forEach(data => (this.itemIcons[data.id] = data));
		db.spellIcons.forEach(data => (this.spellIcons[data.id] = data));
		db.consumables.forEach(consumable => this.consumables.set(consumable.id, consumable));
		db.spellEffects.forEach(spellEffect => this.spellEffects.set(spellEffect.id, spellEffect));
	}

	getAllItems(): Array<Item> {
		return Array.from(this.items.values());
	}

	getAllItemIds(): Array<number> {
		return Array.from(this.items.keys());
	}

	getItems(slot: ItemSlot): Array<Item> {
		return this.getAllItems().filter(item => getEligibleItemSlots(item).includes(slot));
	}

	getItemById(id: number): Item | undefined {
		return this.items.get(id);
	}

	getItemIdsForSet(setId: number): Array<number> {
		return this.getAllItemIds().filter(itemId => this.getItemById(itemId)!.setId === setId);
	}
	getSpellEffect(effectId: number): SpellEffect | undefined {
		return this.spellEffects.get(effectId);
	}
	getConsumables(): Array<Consumable> {
		return Array.from(this.consumables.values());
	}
	getConsumable(itemId: number): Consumable | undefined {
		return this.consumables.get(itemId);
	}
	getRandomSuffixById(id: number): ItemRandomSuffix | undefined {
		return this.randomSuffixes.get(id);
	}

	getItemEffectRandPropPoints(ilvl: number) {
		return this.itemEffectRandPropPoints.get(ilvl);
	}

	getEnchants(slot: ItemSlot): Array<Enchant> {
		return this.enchantsBySlot[slot] || [];
	}

	getEnchantsByEffectId(effectId: number): Array<Enchant> {
		return this.enchantsByEffectId.get(effectId) ?? [];
	}

	getNpc(npcId: number): Npc | null {
		return this.npcs.get(npcId) || null;
	}
	getZone(zoneId: number): Zone | null {
		return this.zones.get(zoneId) || null;
	}
	getRepFaction(factionId: number): UIFaction | null {
		return this.factions.get(factionId) || null;
	}
	getAllZones(): Array<Zone> {
		return Array.from(this.zones.values());
	}

	lookupItemSpec(itemSpec: ItemSpec): EquippedItem | null {
		const item = this.items.get(itemSpec.id);
		if (!item) return null;

		let enchant: Enchant | null = null;
		if (itemSpec.enchant) {
			const slots = getEligibleItemSlots(item);
			for (let i = 0; i < slots.length; i++) {
				const matches = (this.enchantsBySlot[slots[i]] || []).filter(enchant =>
					[enchant.effectId, enchant.itemId, enchant.spellId].includes(itemSpec.enchant),
				);
				// Rows sharing an effect id can differ in the items they fit (1897 is 2H Lesser Impact and Superior Striking).
				enchant = matches.find(enchant => enchantAppliesToItem(enchant, item)) || matches[0] || null;
				if (enchant) {
					break;
				}
			}
		}

		let randomSuffix: ItemRandomSuffix | null = null;
		if (itemSpec.randomSuffix && !!this.getRandomSuffixById(itemSpec.randomSuffix)) {
			randomSuffix = this.getRandomSuffixById(itemSpec.randomSuffix)!;
		}

		return new EquippedItem({
			item,
			enchant,
			randomSuffix,
		});
	}

	lookupEquipmentSpec(equipSpec: EquipmentSpec): Gear {
		// EquipmentSpec is supposed to be indexed by slot, but here we assume
		// it isn't just in case.
		const gearMap: Partial<Record<ItemSlot, EquippedItem | null>> = {};
		equipSpec.items.forEach(itemSpec => {
			const item = this.lookupItemSpec(itemSpec);
			if (!item) return;

			const itemSlots = getEligibleItemSlots(item.item);

			const assignedSlot = itemSlots.find(slot => !gearMap[slot]);
			if (assignedSlot == null) throw new Error('No slots left to equip ' + Item.toJsonString(item.item));

			gearMap[assignedSlot] = item;
		});

		return new Gear(gearMap);
	}

	lookupItemSwap(itemSwap: ItemSwap): ItemSwapGear {
		const gearMap = itemSwap.items.reduce<Partial<Record<ItemSlot, EquippedItem | null>>>((gearMap, itemSpec, slot) => {
			const item = this.lookupItemSpec(itemSpec);
			if (item) {
				const eligibleItemSlots = getEligibleItemSlots(item.item);
				const isSwapSlotMatch = eligibleItemSlots.some(eligibleItemSlot => eligibleItemSlot === slot);
				const assignedSlot = isSwapSlotMatch ? (slot as ItemSlot) : eligibleItemSlots[0];
				if (typeof assignedSlot === 'number') gearMap[assignedSlot] = item;
			}
			return gearMap;
		}, {});
		return new ItemSwapGear(gearMap);
	}

	enchantSpellIdToEnchant(enchantSpellId: number): Enchant | undefined {
		const enchant = Object.values(this.enchantsBySlot)
			.flat()
			.find(enchant => enchant.spellId == enchantSpellId);
		return enchant;
	}

	getPresetEncounter(path: string): PresetEncounter | null {
		return this.presetEncounters.get(path) || null;
	}
	getPresetTarget(path: string): PresetTarget | null {
		return this.presetTargets.get(path) || null;
	}
	getAllPresetEncounters(): Array<PresetEncounter> {
		return Array.from(this.presetEncounters.values());
	}
	getAllPresetTargets(): Array<PresetTarget> {
		return Array.from(this.presetTargets.values());
	}

	static async getItemIconData(itemId: number): Promise<IconData> {
		const db = await Database.get();
		const data = db.itemIcons[itemId];
		if (!data?.icon) {
			// The request is shared between callers, so it must not inherit any one caller's signal.
			db.itemIcons[itemId] = await Database.sharedIconRequest(`item-${itemId}`, () => Database.getWowheadItemTooltipData(itemId));
		}
		return db.itemIcons[itemId];
	}

	// A talent's rank and trait definition steer the wowhead LINK, not the icon: every rank of a
	// talent is one spell with one icon, so they are no part of the lookup.
	static async getSpellIconData(spellId: number): Promise<IconData> {
		const db = await Database.get();
		const data = db.spellIcons[spellId];
		if (!data?.icon) {
			db.spellIcons[spellId] = await Database.sharedIconRequest(`spell-${spellId}`, () => Database.getWowheadSpellTooltipData(spellId));
		}

		return db.spellIcons[spellId];
	}

	// Dedupes concurrent lookups for the same id. A request that comes back without an icon is
	// dropped from the cache so a transient failure isn't remembered as "this id has no icon".
	private static sharedIconRequest(cacheKey: string, makeRequest: () => Promise<IconData>): Promise<IconData> {
		let request = iconRequestCache.get(cacheKey);
		if (!request) {
			request = makeRequest().then(data => {
				if (!data.icon) iconRequestCache.delete(cacheKey);
				return data;
			});
			iconRequestCache.set(cacheKey, request);
		}
		return request;
	}

	private static async getWowheadItemTooltipData(id: number): Promise<IconData> {
		return Database.getWowheadTooltipData(id, 'item');
	}
	private static async getWowheadSpellTooltipData(id: number): Promise<IconData> {
		return Database.getWowheadTooltipData(id, 'spell');
	}
	private static async getWowheadTooltipData(id: number, tooltipPostfix: string): Promise<IconData> {
		const params = new URLSearchParams({ lvl: String(CHARACTER_LEVEL), dataEnv: String(WOWHEAD_EXPANSION_ENV) });
		const url = `https://nether.wowhead.com/${WOWHEAD_DOMAIN}/tooltip/${tooltipPostfix}/${id}?${params}`;
		try {
			const response = await fetch(url);
			const json = await response.json();
			let reportedRank = 0;
			if (tooltipPostfix === 'spell') {
				const rankMatches = [...(json['tooltip'] as string).matchAll(RANK_REGEX)];
				reportedRank = rankMatches.length ? parseInt(rankMatches[0][1]) : 0;
			}

			return IconData.create({
				id: id,
				name: json['name'],
				icon: json['icon'],
				hasBuff: json['buff'] !== '',
				rank: reportedRank,
			});
		} catch (e) {
			console.error('Error while fetching url: ' + url + '\n\n' + e);
			return IconData.create();
		}
	}

	public static mergeSimDatabases(db1: SimDatabase, db2: SimDatabase): SimDatabase {
		return SimDatabase.create({
			items: distinct(db1.items.concat(db2.items), (a, b) => a.id == b.id),
			randomSuffixes: distinct(db1.randomSuffixes.concat(db2.randomSuffixes), (a, b) => a.id == b.id),
			itemEffectRandPropPoints: distinct(db1.itemEffectRandPropPoints.concat(db2.itemEffectRandPropPoints), (a, b) => a.ilvl == b.ilvl),
			enchants: distinct(
				db1.enchants.concat(db2.enchants),
				(a, b) => a.effectId == b.effectId && a.type == b.type && a.enchantType == b.enchantType && arrayEquals(a.extraTypes, b.extraTypes),
			),
			spellEffects: distinct(db1.spellEffects.concat(db2.spellEffects), (a, b) => a.id == b.id),
			consumables: distinct(db1.consumables.concat(db2.consumables), (a, b) => a.id == b.id),
		});
	}
}
