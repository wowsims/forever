import type { IconEnumValueConfig } from '@features/inputs/configs';
import * as InputHelpers from '@features/inputs/input_helpers';
import type { StatFilterScope } from '@features/stat-filters/model/stat_filters';
import { ConsumableType, ConsumesSpec } from '@generated/proto/common';
import { Consumable } from '@generated/proto/db';
import i18n from '@i18n/config';
import type { StoreField } from '@sim/hooks/useStoreField';
import { Player } from '@sim/player/player';
import { ActionId } from '@sim/proto/action_id';
import type { Database } from '@sim/proto/database';
import { UnitStat, unitStatsOf } from '@sim/proto/stats';
import { consumableAllowed, consumableOrderOf, consumablePseudoStatsOf, consumableStatsOf, consumableUniverse, orderedByEp } from '@sim/settings/consumables';
import { memoLast } from '@sim/utils/misc';

export type ConsumableIdField =
	| 'resourcePotId'
	| 'defensivePotId'
	| 'offensivePotId'
	| 'conjuredId'
	| 'flaskId'
	| 'foodId'
	| 'explosiveId'
	| 'mhImbueId'
	| 'ohImbueId'
	| 'mhWeaponImbueId'
	| 'ohWeaponImbueId'
	| 'swapMhImbueId'
	| 'swapOhImbueId'
	| 'swapMhWeaponImbueId'
	| 'swapOhWeaponImbueId';

export interface ConsumableRelevance {
	// The stat rule, which the consumables stat filter moves.
	relevant: (consumable: Consumable, player: Player<any>) => boolean;
}

export interface ConsumableInputOptions {
	consumesFieldName: ConsumableIdField;
	// A condition on top of relevance, for the one picker a consumable's kind shares with another.
	includeIf?: (consumable: Consumable, player: Player<any>) => boolean;
	showWhen?: (player: Player<any>) => boolean;
	tooltip?: string;
}

// The class, race and profession decide whether a consumable is offered at all, picked or not. The
// minimum level and the stat rule only trim what is offered: a consumable already picked stays
// offered below the level and with its stats hidden.
const offeredTo =
	(consumable: Consumable, relevance: ConsumableRelevance, picked: (player: Player<any>) => boolean, includeIf?: ConsumableInputOptions['includeIf']) =>
	(player: Player<any>) =>
		consumableAllowed(consumable, player) &&
		((consumable.lvl >= player.sim.getMinConsumableLvl(consumable.type) && relevance.relevant(consumable, player)) || picked(player)) &&
		(!includeIf || includeIf(consumable, player));

// Defaults are the stats and percents the spec weighs consumables by and the stats it displays, as the
// buffs' filter counts them (a pseudo-stat by its root rating); the universe adds every one a
// consumable grants. A default no consumable grants (a resource stat the restore rule reads) stays
// in the universe, so the defaults show in full. What the consumables grant is worked out once per
// database, as it walks every consumable, and the scope once per player and database, so the
// pickers and the filters popover share it.
const grantedByDatabase = new WeakMap<Database, UnitStat[]>();
const scopeByPlayer = new WeakMap<Player<any>, { db: Database; scope: StatFilterScope }>();

export const consumableFilterScope = (player: Player<any>, db: Database): StatFilterScope => {
	const cached = scopeByPlayer.get(player);
	if (cached?.db === db) return cached.scope;
	let granted = grantedByDatabase.get(db);
	if (!granted) {
		granted = consumableUniverse(db.getConsumables(), db);
		grantedByDatabase.set(db, granted);
	}
	const specConfig = player.getSpecConfig();
	const displayed = (specConfig.displayStats ?? []).filter(unitStat => unitStat.hasRootStat()).map(unitStat => unitStat.getRootStat());
	const stats = consumableStatsOf({ epStats: [...specConfig.epStats, ...displayed], consumableStats: specConfig.consumableStats });
	const defaults = unitStatsOf({ stats, pseudoStats: consumablePseudoStatsOf(specConfig) });
	const pool = new Set([...defaults, ...granted].map(stat => stat.getKey()));
	const scope = { defaults, universe: UnitStat.getAll().filter(stat => pool.has(stat.getKey())) };
	scopeByPlayer.set(player, { db, scope });
	return scope;
};

const STORE_FIELDS: Array<StoreField> = ['consumables', 'gear', 'profession1', 'profession2', 'prof1Rank', 'prof2Rank', 'race', 'sim:filters'];

// Every field a consumable picker reads: `itemSwap` because the swap imbues' includeIf reads the swap
// weapons and their showWhen whether item swap is on. The toggles read only STORE_FIELDS.
export const CONSUMABLE_PICKER_FIELDS: ReadonlyArray<StoreField> = [...STORE_FIELDS, 'itemSwap'];

// The columns a consumable menu lays its offers out in, by how many there are.
export const consumableColumns = (offered: number): number => (offered > 10 ? 3 : offered > 5 ? 2 : 1);

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
	const byId = new Map(consumables.map(consumable => [consumable.id, consumable]));
	const values: Array<IconEnumValueConfig<Player<any>, number>> = consumables.map(consumable => ({
		value: consumable.id,
		actionId: ActionId.fromItemId(consumable.id),
		tooltip: consumable.name,
		showWhen: offeredTo(consumable, relevance, player => pickOf(player.getConsumes()) === consumable.id, options.includeIf),
	}));
	const none: IconEnumValueConfig<Player<any>, number> = { value: 0, iconUrl: '', tooltip: i18n.t('common.none') };
	const ordered = memoLast<Array<IconEnumValueConfig<Player<any>, number>>>();
	return {
		type: 'iconEnum',
		tooltip: options.tooltip,
		numColumns: (p: Player<any>) => consumableColumns(values.filter(value => value.showWhen!(p)).length),
		values: (player: Player<any>) => {
			const order = consumableOrderOf(player);
			return ordered([order.weights, order.mobType, order.zoneId], () => [none, ...orderedByEp(values, value => byId.get(value.value)!, order)]);
		},
		equals: (a: number, b: number) => a === b,
		zeroValue: 0,
		storeField: CONSUMABLE_PICKER_FIELDS,
		showWhen: options.showWhen,
		keptBy: (player: Player<any>, value: number) => {
			const picked = byId.get(value);
			return !!picked && !relevance.relevant(picked, player);
		},
		getValue: (player: Player<any>) => pickOf(player.getConsumes()),
		setValue: (player: Player<any>, newValue: number) => {
			const consumes = player.getConsumes();
			if (pickOf(consumes) === newValue) return;
			player.setConsumes(withPick(consumes, newValue));
		},
	};
}

// The consumable rides along so a row of these can be ordered by the EP it adds.
export type ConsumeToggle = InputHelpers.TypedIconPickerConfig<Player<any>, boolean> & { consumable: Consumable };

// A consumable picked any number at a time into the spec's extra ids.
export function makeIdListInput(consumable: Consumable, relevance: ConsumableRelevance, includeIf?: ConsumableInputOptions['includeIf']): ConsumeToggle {
	const picked = (player: Player<any>) => player.getConsumes().extraIds.includes(consumable.id);
	return {
		consumable,
		type: 'icon',
		actionId: ActionId.fromItemId(consumable.id),
		states: 2,
		storeField: STORE_FIELDS,
		showWhen: offeredTo(consumable, relevance, picked, includeIf),
		keptBy: (player: Player<any>, value: boolean) => value && !relevance.relevant(consumable, player),
		getValue: picked,
		setValue: (player: Player<any>, newValue: boolean) => {
			const consumes = player.getConsumes();
			const extraIds = consumes.extraIds.filter(id => id !== consumable.id);
			if (newValue) extraIds.push(consumable.id);
			player.setConsumes({ ...consumes, extraIds });
		},
	};
}
