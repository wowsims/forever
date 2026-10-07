import type { IconEnumValueConfig } from '@features/inputs/configs';
import * as InputHelpers from '@features/inputs/input_helpers';
import { ConsumableType, ConsumesSpec } from '@generated/proto/common';
import { Consumable } from '@generated/proto/db';
import i18n from '@i18n/config';
import type { StoreField } from '@sim/hooks/useStoreField';
import { Player } from '@sim/player/player';
import { ActionId } from '@sim/proto/action_id';

export type ConsumableIdField =
	| 'potId'
	| 'conjuredId'
	| 'flaskId'
	| 'foodId'
	| 'explosiveId'
	| 'mhImbueId'
	| 'ohImbueId'
	| 'mhWeaponImbueId'
	| 'ohWeaponImbueId';

export interface ConsumableRelevance {
	relevant: (consumable: Consumable, player: Player<any>) => boolean;
}

export interface ConsumableInputOptions {
	consumesFieldName: ConsumableIdField;
	// A condition on top of relevance, for the one picker a consumable's kind shares with another.
	includeIf?: (consumable: Consumable, player: Player<any>) => boolean;
	showWhen?: (player: Player<any>) => boolean;
	tooltip?: string;
}

// The minimum level only trims what is offered: a consumable already picked stays offered below it.
const offeredTo =
	(consumable: Consumable, relevance: ConsumableRelevance, picked: (player: Player<any>) => boolean, includeIf?: ConsumableInputOptions['includeIf']) =>
	(player: Player<any>) =>
		(picked(player) || consumable.lvl >= player.sim.getMinConsumableLvl(consumable.type)) &&
		relevance.relevant(consumable, player) &&
		(!includeIf || includeIf(consumable, player));

const STORE_FIELDS: Array<StoreField> = ['consumables', 'gear', 'profession1', 'profession2', 'prof1Rank', 'prof2Rank', 'race', 'sim:filters'];

export const makeConsumableInput = (consumables: ReadonlyArray<Consumable>, relevance: ConsumableRelevance, options: ConsumableInputOptions) =>
	consumablePicker(
		consumables,
		relevance,
		options,
		consumes => consumes[options.consumesFieldName],
		(consumes, newValue) => ({ ...consumes, [options.consumesFieldName]: newValue }),
	);

// The spec's list a grouped consumable is picked into: an elixir's elixir ids, anything else's extra ids.
const idListOf = (consumable: Consumable): 'elixirIds' | 'extraIds' => (consumable.type === ConsumableType.ConsumableTypeElixir ? 'elixirIds' : 'extraIds');

// One exclusive group's pick: at most one of the group's consumables in the spec's elixir and extra ids.
export const makeConsumableGroupInput = (consumables: ReadonlyArray<Consumable>, relevance: ConsumableRelevance, tooltip: string) => {
	const ids = new Set(consumables.map(consumable => consumable.id));
	const lists = new Map(consumables.map(consumable => [consumable.id, idListOf(consumable)]));
	return consumablePicker(
		consumables,
		relevance,
		{ tooltip },
		consumes => consumes.elixirIds.find(id => ids.has(id)) ?? consumes.extraIds.find(id => ids.has(id)) ?? 0,
		(consumes, newValue) => {
			const picked = {
				...consumes,
				elixirIds: consumes.elixirIds.filter(id => !ids.has(id)),
				extraIds: consumes.extraIds.filter(id => !ids.has(id)),
			};
			if (newValue) picked[lists.get(newValue)!].push(newValue);
			return picked;
		},
	);
};

function consumablePicker(
	consumables: ReadonlyArray<Consumable>,
	relevance: ConsumableRelevance,
	options: Omit<ConsumableInputOptions, 'consumesFieldName'>,
	pickOf: (consumes: ConsumesSpec) => number,
	withPick: (consumes: ConsumesSpec, newValue: number) => ConsumesSpec,
): InputHelpers.TypedIconEnumPickerConfig<Player<any>, number> {
	const byLvl = [...consumables].sort((a, b) => b.lvl - a.lvl);
	const values: Array<IconEnumValueConfig<Player<any>, number>> = byLvl.map(consumable => ({
		value: consumable.id,
		actionId: ActionId.fromItemId(consumable.id),
		tooltip: consumable.name,
		showWhen: offeredTo(consumable, relevance, player => pickOf(player.getConsumes()) === consumable.id, options.includeIf),
	}));
	return {
		type: 'iconEnum',
		tooltip: options.tooltip,
		numColumns: (p: Player<any>) => {
			const offered = values.filter(value => value.showWhen!(p)).length;
			return offered > 10 ? 3 : offered > 5 ? 2 : 1;
		},
		values: [{ value: 0, iconUrl: '', tooltip: i18n.t('common.none') }, ...values],
		equals: (a: number, b: number) => a === b,
		zeroValue: 0,
		// `raid:partyBuffs` because the main-hand imbue's showWhen reads the weapon totems off the party.
		storeField: [...STORE_FIELDS, 'raid:partyBuffs'],
		showWhen: options.showWhen,
		getValue: (player: Player<any>) => pickOf(player.getConsumes()),
		setValue: (player: Player<any>, newValue: number) => {
			const consumes = player.getConsumes();
			if (pickOf(consumes) === newValue) return;
			player.setConsumes(withPick(consumes, newValue));
		},
	};
}

// A consumable picked any number at a time into the spec's extra ids.
export function makeIdListInput(consumable: Consumable, relevance: ConsumableRelevance): InputHelpers.TypedIconPickerConfig<Player<any>, boolean> {
	return {
		type: 'icon',
		actionId: ActionId.fromItemId(consumable.id),
		states: 2,
		storeField: STORE_FIELDS,
		showWhen: offeredTo(consumable, relevance, player => player.getConsumes().extraIds.includes(consumable.id)),
		getValue: (player: Player<any>) => player.getConsumes().extraIds.includes(consumable.id),
		setValue: (player: Player<any>, newValue: boolean) => {
			const consumes = player.getConsumes();
			const extraIds = consumes.extraIds.filter(id => id !== consumable.id);
			if (newValue) extraIds.push(consumable.id);
			player.setConsumes({ ...consumes, extraIds });
		},
	};
}
