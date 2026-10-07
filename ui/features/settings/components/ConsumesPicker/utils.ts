import type { TypedIconEnumPickerConfig, TypedIconPickerConfig } from '@features/inputs/input_helpers';
import { type ConsumableRelevance, makeConsumableGroupInput, makeConsumableInput, makeIdListInput } from '@features/settings/model/consumables';
import { ConsumableType, ItemSlot } from '@generated/proto/common';
import type { Consumable } from '@generated/proto/db';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import type { Database } from '@sim/proto/database';
import {
	consumablePseudoStatsOf,
	ConsumableSlot,
	consumableSlot,
	consumableStatsOf,
	explosiveUsedBeside,
	handTakesImbue,
	relevanceFor,
} from '@sim/settings/consumables';

export type ConsumeConfig = TypedIconEnumPickerConfig<Player<any>, number>;
export type ConsumeToggle = TypedIconPickerConfig<Player<any>, boolean>;

export interface ConsumeConfigs {
	potion: ConsumeConfig;
	conjured: ConsumeConfig;
	flask: ConsumeConfig;
	// One pick per exclusive group, in the order of the groups' keys: those holding an elixir, those
	// holding a scroll and no elixir, and the rest.
	elixirs: Array<ConsumeConfig>;
	scrolls: Array<ConsumeConfig>;
	groups: Array<ConsumeConfig>;
	food: ConsumeConfig;
	// A world event's consumables: one pick per exclusive group holding one, and the rest each on its own.
	holidayGroups: Array<ConsumeConfig>;
	holiday: Array<ConsumeToggle>;
	explosive: ConsumeConfig;
	explosiveExtras: Array<ConsumeToggle>;
	mhImbue: ConsumeConfig;
	ohImbue: ConsumeConfig;
	mhWeaponImbue: ConsumeConfig;
	ohWeaponImbue: ConsumeConfig;
	miscellaneous: Array<ConsumeToggle>;
}

type GroupRow = 'elixirs' | 'scrolls' | 'holiday' | 'groups';

export const consumeConfigs = (player: Player<any>, db: Database): ConsumeConfigs => {
	const all = db.getConsumables();
	const relevance: ConsumableRelevance = {
		relevant: relevanceFor(all, db, consumableStatsOf(player.getSpecConfig()), consumablePseudoStatsOf(player.getSpecConfig())),
	};
	const bySlot = new Map<ConsumableSlot, Array<Consumable>>();
	const byGroup = new Map<string, Array<Consumable>>();
	const add = <K>(map: Map<K, Array<Consumable>>, key: K, consumable: Consumable) => {
		const list = map.get(key);
		if (list) list.push(consumable);
		else map.set(key, [consumable]);
	};
	const holiday: Array<Consumable> = [];
	for (const consumable of all) {
		if (consumable.groupKey) add(byGroup, consumable.groupKey, consumable);
		else if (consumable.holiday) holiday.push(consumable);
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
					: 'groups';
	const groupInputs = (row: GroupRow) =>
		groups.filter(group => rowOf(group) === row).map(group => makeConsumableGroupInput(group, relevance, group[0].groupName));
	const explosives = inSlot(ConsumableSlot.Explosive);
	const fitsWeaponIn = (slot: ItemSlot) => (consumable: Consumable, p: Player<any>) => handTakesImbue(consumable, p.getEquippedItem(slot)?.item);

	return {
		potion: input(ConsumableSlot.Potion, { consumesFieldName: 'potId', tooltip: i18n.t('settings_tab.consumables.potions.combat') }),
		conjured: input(ConsumableSlot.Conjured, { consumesFieldName: 'conjuredId' }),
		flask: input(ConsumableSlot.Flask, { consumesFieldName: 'flaskId' }),
		elixirs: groupInputs('elixirs'),
		scrolls: groupInputs('scrolls'),
		groups: groupInputs('groups'),
		food: input(ConsumableSlot.Food, { consumesFieldName: 'foodId' }),
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
		miscellaneous: inSlot(ConsumableSlot.Extra).map(consumable => makeIdListInput(consumable, relevance)),
	};
};
