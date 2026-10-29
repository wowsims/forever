import type { TypedIconEnumPickerConfig } from '@features/inputs/input_helpers';
import {
	consumableFilterScope,
	type ConsumableRelevance,
	type ConsumeToggle,
	makeConsumableGroupInput,
	makeConsumableInput,
	makeIdListInput,
} from '@features/settings/model/consumables';
import { DragonbreathChili } from '@features/settings/model/consumables_auto_gen';
import { shownStatsReader, type StatFilterScope } from '@features/stat-filters/model/stat_filters';
import { Class, ConsumableType, ItemSlot, PotionClass } from '@generated/proto/common';
import type { Consumable } from '@generated/proto/db';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import type { Database } from '@sim/proto/database';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { type StatSets, statSetsOf } from '@sim/proto/stats';
import { ConsumableSlot, consumableSlot, explosiveUsedBeside, handTakesImbue, relevanceFor } from '@sim/settings/consumables';
import { memoLast } from '@sim/utils/misc';

export type { ConsumeToggle };
export type ConsumeConfig = TypedIconEnumPickerConfig<Player<any>, number>;

export interface ConsumeConfigs {
	// One pick per potion class, each offering the potions of that class.
	offensivePotion: ConsumeConfig;
	defensivePotion: ConsumeConfig;
	resourcePotion: ConsumeConfig;
	conjured: ConsumeConfig;
	flask: ConsumeConfig;
	// One pick per exclusive group, in the order of the groups' keys: those holding an elixir, those
	// holding a scroll and no elixir, and the rest.
	elixirs: Array<ConsumeConfig>;
	scrolls: Array<ConsumeConfig>;
	// Those holding a food, picked beside the food (Rum), and those holding none (Zanza).
	foodGroups: Array<ConsumeConfig>;
	miscGroups: Array<ConsumeConfig>;
	food: ConsumeConfig;
	// A food that stacks with the food pick, offered to the specs that hit in melee, hunters among them.
	dragonbreathChili?: ConsumeToggle;
	// A world event's consumables: one pick per exclusive group holding one, and the rest each on its own.
	holidayGroups: Array<ConsumeConfig>;
	holiday: Array<ConsumeToggle>;
	explosive: ConsumeConfig;
	explosiveExtras: Array<ConsumeToggle>;
	mhImbue: ConsumeConfig;
	ohImbue: ConsumeConfig;
	mhWeaponImbue: ConsumeConfig;
	ohWeaponImbue: ConsumeConfig;
	// The swap weapons' picks, offered while item swap is on and the swap set holds a weapon in the hand.
	swapMhImbue: ConsumeConfig;
	swapOhImbue: ConsumeConfig;
	swapMhWeaponImbue: ConsumeConfig;
	swapOhWeaponImbue: ConsumeConfig;
	// The rest, split as the generator classes them: defensive where it is only that, offensive otherwise.
	miscOffensive: Array<ConsumeToggle>;
	miscDefensive: Array<ConsumeToggle>;
}

const defensiveOnly = (consumable: Consumable): boolean =>
	consumable.potionClasses.includes(PotionClass.PotionClassDefensive) && !consumable.potionClasses.includes(PotionClass.PotionClassOffensive);

type GroupRow = 'elixirs' | 'scrolls' | 'holiday' | 'food' | 'miscellaneous';

// The shown stats of the consumables slot, read from the store's filters on each call; the same object
// comes back while the slot holds the same stats, so relevance stays cached.
const shownConsumableStats = (player: Player<any>, scope: StatFilterScope): (() => StatSets) => {
	const readShown = shownStatsReader('consumables', scope);
	const setsOf = memoLast<StatSets>();
	return () => {
		const shown = readShown(player.sim.store.getState().sim.filters);
		return setsOf([shown], () => statSetsOf(shown));
	};
};

// One set per player and database, so every picker of the player shares the relevance cache and the order memos.
const configsByPlayer = new WeakMap<Player<any>, { db: Database; configs: ConsumeConfigs }>();

export const consumeConfigs = (player: Player<any>, db: Database): ConsumeConfigs => {
	const cached = configsByPlayer.get(player);
	if (cached?.db === db) return cached.configs;
	const configs = buildConsumeConfigs(player, db);
	configsByPlayer.set(player, { db, configs });
	return configs;
};

const buildConsumeConfigs = (player: Player<any>, db: Database): ConsumeConfigs => {
	const all = db.getConsumables();
	const relevance: ConsumableRelevance = {
		relevant: relevanceFor(all, db, shownConsumableStats(player, consumableFilterScope(player, db))),
	};
	const bySlot = new Map<ConsumableSlot, Array<Consumable>>();
	const byGroup = new Map<string, Array<Consumable>>();
	const add = <K>(map: Map<K, Array<Consumable>>, key: K, consumable: Consumable) => {
		const list = map.get(key);
		if (list) list.push(consumable);
		else map.set(key, [consumable]);
	};
	const holiday: Array<Consumable> = [];
	let chili: Consumable | undefined;
	for (const consumable of all) {
		if (consumable.groupKey) add(byGroup, consumable.groupKey, consumable);
		else if (consumable.holiday) holiday.push(consumable);
		else if (consumable.id === DragonbreathChili.id) chili = consumable;
		else add(bySlot, consumableSlot(consumable), consumable);
	}
	const inSlot = (slot: ConsumableSlot) => bySlot.get(slot) ?? [];
	const input = (slot: ConsumableSlot, options: Parameters<typeof makeConsumableInput>[2], consumables = inSlot(slot)) =>
		makeConsumableInput(consumables, relevance, options);

	const groups = [...byGroup.keys()].sort((a, b) => a.localeCompare(b, 'en', { numeric: true })).map(key => byGroup.get(key)!);
	const holds = (group: Array<Consumable>, type: ConsumableType) => group.some(consumable => consumable.type === type);
	const rowOf = (group: Array<Consumable>): GroupRow =>
		holds(group, ConsumableType.ConsumableTypeElixir)
			? 'elixirs'
			: holds(group, ConsumableType.ConsumableTypeScroll)
				? 'scrolls'
				: group.some(consumable => consumable.holiday)
					? 'holiday'
					: holds(group, ConsumableType.ConsumableTypeFood)
						? 'food'
						: 'miscellaneous';
	const groupInputs = (row: GroupRow) =>
		groups.filter(group => rowOf(group) === row).map(group => makeConsumableGroupInput(group, relevance, group[0].groupName));
	const explosives = inSlot(ConsumableSlot.Explosive);
	const fitsHandOf = (itemIn: (p: Player<any>, slot: ItemSlot) => EquippedItem | null) => (slot: ItemSlot) => (consumable: Consumable, p: Player<any>) =>
		handTakesImbue(consumable, itemIn(p, slot)?.item);
	const fitsWeaponIn = fitsHandOf((p, slot) => p.getEquippedItem(slot));
	const fitsSwapWeaponIn = fitsHandOf((p, slot) => p.itemSwapSettings.getItem(slot));
	const swapOn = (p: Player<any>) => p.itemSwapSettings.getEnableItemSwap();
	// A hunter weaves melee swings between shots.
	const hitsInMelee = (_: Consumable, p: Player<any>) =>
		p.getPlayerSpec().isMeleeDpsSpec || p.getPlayerSpec().isTankSpec || p.getClass() === Class.ClassHunter;
	// The potion already picked stays offered, so a pick outside the class can still be seen and cleared. A
	// potion that restores the class's own power is a resource potion alone to it (Mighty Rage Potion to a warrior).
	const ofPotionClass =
		(potionClass: PotionClass, field: 'offensivePotId' | 'defensivePotId' | 'resourcePotId') => (consumable: Consumable, p: Player<any>) =>
			(consumable.potionClasses.includes(potionClass) &&
				!(potionClass === PotionClass.PotionClassOffensive && consumable.resourceOnlyClasses.includes(p.getClass()))) ||
			p.getConsumes()[field] === consumable.id;

	return {
		offensivePotion: input(ConsumableSlot.Potion, {
			consumesFieldName: 'offensivePotId',
			tooltip: i18n.t('settings_tab.consumables.potions.offensive'),
			includeIf: ofPotionClass(PotionClass.PotionClassOffensive, 'offensivePotId'),
		}),
		defensivePotion: input(ConsumableSlot.Potion, {
			consumesFieldName: 'defensivePotId',
			tooltip: i18n.t('settings_tab.consumables.potions.defensive'),
			includeIf: ofPotionClass(PotionClass.PotionClassDefensive, 'defensivePotId'),
		}),
		resourcePotion: input(ConsumableSlot.Potion, {
			consumesFieldName: 'resourcePotId',
			tooltip: i18n.t('settings_tab.consumables.potions.resource'),
			includeIf: ofPotionClass(PotionClass.PotionClassResource, 'resourcePotId'),
		}),
		conjured: input(ConsumableSlot.Conjured, { consumesFieldName: 'conjuredId' }),
		flask: input(ConsumableSlot.Flask, { consumesFieldName: 'flaskId' }),
		elixirs: groupInputs('elixirs'),
		scrolls: groupInputs('scrolls'),
		foodGroups: groupInputs('food'),
		miscGroups: groupInputs('miscellaneous'),
		food: input(ConsumableSlot.Food, { consumesFieldName: 'foodId' }),
		// Its worth is damage, which no stat the filter shows speaks to.
		dragonbreathChili: chili && makeIdListInput(chili, { relevant: () => true }, hitsInMelee),
		holidayGroups: groupInputs('holiday'),
		holiday: holiday.map(consumable => makeIdListInput(consumable, relevance)),
		explosive: input(
			ConsumableSlot.Explosive,
			{ consumesFieldName: 'explosiveId', tooltip: i18n.t('settings_tab.consumables.engineering.explosives') },
			explosives.filter(consumable => !explosiveUsedBeside(consumable)),
		),
		explosiveExtras: explosives.filter(explosiveUsedBeside).map(consumable => makeIdListInput(consumable, relevance)),
		mhImbue: input(ConsumableSlot.Imbue, {
			consumesFieldName: 'mhImbueId',
			tooltip: i18n.t('settings_tab.consumables.imbue.mhImbue'),
			includeIf: fitsWeaponIn(ItemSlot.ItemSlotMainHand),
		}),
		ohImbue: input(ConsumableSlot.Imbue, {
			consumesFieldName: 'ohImbueId',
			tooltip: i18n.t('settings_tab.consumables.imbue.ohImbue'),
			includeIf: fitsWeaponIn(ItemSlot.ItemSlotOffHand),
		}),
		swapMhImbue: input(ConsumableSlot.Imbue, {
			consumesFieldName: 'swapMhImbueId',
			tooltip: i18n.t('settings_tab.consumables.imbue.swapMhImbue'),
			includeIf: fitsSwapWeaponIn(ItemSlot.ItemSlotMainHand),
			showWhen: swapOn,
		}),
		swapOhImbue: input(ConsumableSlot.Imbue, {
			consumesFieldName: 'swapOhImbueId',
			tooltip: i18n.t('settings_tab.consumables.imbue.swapOhImbue'),
			includeIf: fitsSwapWeaponIn(ItemSlot.ItemSlotOffHand),
			showWhen: swapOn,
		}),
		mhWeaponImbue: input(ConsumableSlot.WeaponImbue, {
			consumesFieldName: 'mhWeaponImbueId',
			tooltip: i18n.t('settings_tab.consumables.weaponImbue.mhWeaponImbue'),
			includeIf: fitsWeaponIn(ItemSlot.ItemSlotMainHand),
		}),
		ohWeaponImbue: input(ConsumableSlot.WeaponImbue, {
			consumesFieldName: 'ohWeaponImbueId',
			tooltip: i18n.t('settings_tab.consumables.weaponImbue.ohWeaponImbue'),
			includeIf: fitsWeaponIn(ItemSlot.ItemSlotOffHand),
		}),
		swapMhWeaponImbue: input(ConsumableSlot.WeaponImbue, {
			consumesFieldName: 'swapMhWeaponImbueId',
			tooltip: i18n.t('settings_tab.consumables.weaponImbue.swapMhWeaponImbue'),
			includeIf: fitsSwapWeaponIn(ItemSlot.ItemSlotMainHand),
			showWhen: swapOn,
		}),
		swapOhWeaponImbue: input(ConsumableSlot.WeaponImbue, {
			consumesFieldName: 'swapOhWeaponImbueId',
			tooltip: i18n.t('settings_tab.consumables.weaponImbue.swapOhWeaponImbue'),
			includeIf: fitsSwapWeaponIn(ItemSlot.ItemSlotOffHand),
			showWhen: swapOn,
		}),
		miscOffensive: inSlot(ConsumableSlot.Extra)
			.filter(consumable => !defensiveOnly(consumable))
			.map(consumable => makeIdListInput(consumable, relevance)),
		miscDefensive: inSlot(ConsumableSlot.Extra)
			.filter(defensiveOnly)
			.map(consumable => makeIdListInput(consumable, relevance)),
	};
};
