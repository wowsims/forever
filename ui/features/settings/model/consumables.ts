import type { IconEnumValueConfig } from '@features/inputs/configs';
import * as InputHelpers from '@features/inputs/input_helpers';
import type { StatFilterScope } from '@features/stat-filters/model/stat_filters';
import { ConsumableType, ConsumesSpec, PseudoStat, Stat } from '@generated/proto/common';
import { Consumable } from '@generated/proto/db';
import i18n from '@i18n/config';
import { translatePseudoStat, translateStat } from '@i18n/localization';
import type { StoreField } from '@sim/hooks/useStoreField';
import { Player } from '@sim/player/player';
import { ActionId } from '@sim/proto/action_id';
import type { Database } from '@sim/proto/database';
import { UnitStat, unitStatsOf } from '@sim/proto/stats';
import { consumableAllowed, consumableOrderOf, consumablePseudoStatsOf, consumableStatsOf, consumableUniverse, orderedByEp } from '@sim/settings/consumables';
import { memoLast } from '@sim/utils/misc';

import { BoglingRoot, JujuMight, JujuPower } from './consumables_auto_gen';

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
// offered below the level or with its stats hidden, marked as kept.
const trimmed = (consumable: Consumable, relevance: ConsumableRelevance, player: Player<any>) =>
	consumable.lvl < player.sim.getMinConsumableLvl(consumable.type) || !relevance.relevant(consumable, player);

const offeredTo =
	(consumable: Consumable, relevance: ConsumableRelevance, picked: (player: Player<any>) => boolean, includeIf?: ConsumableInputOptions['includeIf']) =>
	(player: Player<any>) =>
		consumableAllowed(consumable, player) && (!trimmed(consumable, relevance, player) || picked(player)) && (!includeIf || includeIf(consumable, player));

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

// The spec's list a consumable is picked into: an elixir's elixir ids, anything else's extra ids.
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
			return !!picked && trimmed(picked, relevance, player);
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

// A consumable picked any number at a time into the spec's elixir ids for an elixir, its extra ids otherwise. It reads
// as picked from either list, since a saved setting may hold it in the other.
export function makeIdListInput(consumable: Consumable, relevance: ConsumableRelevance, includeIf?: ConsumableInputOptions['includeIf']): ConsumeToggle {
	const picked = (player: Player<any>) => {
		const { elixirIds, extraIds } = player.getConsumes();
		return elixirIds.includes(consumable.id) || extraIds.includes(consumable.id);
	};
	return {
		consumable,
		type: 'icon',
		actionId: ActionId.fromItemId(consumable.id),
		states: 2,
		storeField: STORE_FIELDS,
		showWhen: offeredTo(consumable, relevance, picked, includeIf),
		keptBy: (player: Player<any>, value: boolean) => value && trimmed(consumable, relevance, player),
		getValue: picked,
		setValue: (player: Player<any>, newValue: boolean) => {
			const consumes = player.getConsumes();
			const lists = { elixirIds: consumes.elixirIds.filter(id => id !== consumable.id), extraIds: consumes.extraIds.filter(id => id !== consumable.id) };
			if (newValue) lists[idListOf(consumable)].push(consumable.id);
			player.setConsumes({ ...consumes, ...lists });
		},
	};
}

// Consumables the game files do not type as elixirs whose stats the sim counts with the elixirs' (statsCountWithTheElixirs
// in sim/common/shared/consumable_stats.go).
const COUNT_WITH_THE_ELIXIRS = new Set([BoglingRoot.id, JujuPower.id, JujuMight.id]);

export const countsWithTheElixirs = (consumable: Consumable): boolean =>
	consumable.type === ConsumableType.ConsumableTypeElixir || COUNT_WITH_THE_ELIXIRS.has(consumable.id);

interface Grants {
	consumable: Consumable;
	// Each stat and percent the consumable grants, by stat or percent.
	amounts: Map<string, number>;
}

const grantsOf = (consumable: Consumable): Grants => {
	const amounts = new Map<string, number>();
	consumable.stats.forEach((value, stat) => value > 0 && amounts.set(`stat ${stat}`, value));
	consumable.pseudoStats.forEach((value, pseudoStat) => value > 0 && amounts.set(`percent ${pseudoStat}`, value));
	return { consumable, amounts };
};

// The picks granting the most of each amount, where every amount has a pick granting at least as much. Undefined
// where one has none, and for no amounts at all.
const matchedBy = (amounts: Map<string, number>, picks: ReadonlyArray<Grants>): Array<Consumable> | undefined => {
	if (!amounts.size) return undefined;
	const matching = new Set<Consumable>();
	for (const [key, value] of amounts) {
		let best: Grants | undefined;
		for (const pick of picks) {
			const amount = pick.amounts.get(key) ?? 0;
			if (amount >= value && (!best || amount > best.amounts.get(key)!)) best = pick;
		}
		if (!best) return undefined;
		matching.add(best.consumable);
	}
	return [...matching];
};

interface ElixirOverwrites {
	// The picks that add nothing, each with the picks that grant as much of everything it grants.
	overwritten: Map<Consumable, Array<Consumable>>;
	// The picks that add something.
	counted: Array<Consumable>;
}

// The sim counts each stat and percent the elixirs grant at the largest amount among them only (R-0053), so a pick
// adds nothing where the others grant as much of everything it grants. The picks are set aside one at a time, those
// granting the fewest stats and percents first and, among those, the one later in the spec's lists first (the elixir
// ids, then the extra ids), so of two that grant the same only one is. One granting nothing the picker can read is
// never set aside.
function elixirOverwrites(picks: ReadonlyArray<Consumable>): ElixirOverwrites {
	const grants = picks.map(grantsOf);
	const order = grants.map((entry, index) => ({ entry, index })).sort((a, b) => a.entry.amounts.size - b.entry.amounts.size || b.index - a.index);
	let counted = grants;
	const setAside: Array<Grants> = [];
	for (const { entry } of order) {
		const others = counted.filter(other => other !== entry);
		if (!matchedBy(entry.amounts, others)) continue;
		counted = others;
		setAside.push(entry);
	}
	return {
		overwritten: new Map(setAside.map(entry => [entry.consumable, matchedBy(entry.amounts, counted)!])),
		counted: counted.map(entry => entry.consumable),
	};
}

// The sim counts a scroll's stat only while it outbids the elixirs' (R-0053), so the elixirs that grant as much of it
// overwrite it. Undefined while it outbids them. The raid buffs that overwrite a scroll are not read here.
const scrollOverwrittenBy = (scroll: Consumable, overwrites: ElixirOverwrites): Array<Consumable> | undefined =>
	matchedBy(grantsOf(scroll).amounts, overwrites.counted.map(grantsOf));

const overwrittenTooltipOf = (by: ReadonlyArray<Consumable>): string =>
	i18n.t('settings_tab.consumables.overwritten_tooltip', {
		names: new Intl.ListFormat(i18n.language, { type: 'conjunction' }).format(by.map(consumable => consumable.name)),
	});

// The overwritten tooltips of the elixir toggles and the scroll picks, read off the picks among `elixirs` in the
// spec's elixir and extra ids and worked out once per set of picks.
export function overwrittenTooltips(elixirs: ReadonlyArray<Consumable>) {
	const byId = new Map(elixirs.map(consumable => [consumable.id, consumable]));
	const overwritesFor = memoLast<ElixirOverwrites>();
	const overwrites = (player: Player<any>) => {
		const { elixirIds, extraIds } = player.getConsumes();
		const ids = [...new Set([...elixirIds, ...extraIds])].filter(id => byId.has(id));
		return overwritesFor([ids.join()], () => elixirOverwrites(ids.map(id => byId.get(id)!)));
	};
	return {
		elixir:
			(consumable: Consumable) =>
			(player: Player<any>, picked: boolean): string | undefined => {
				const by = picked ? overwrites(player).overwritten.get(consumable) : undefined;
				return by && overwrittenTooltipOf(by);
			},
		scroll:
			(scrolls: ReadonlyArray<Consumable>) =>
			(player: Player<any>, id: number): string | undefined => {
				const scroll = scrolls.find(consumable => consumable.id === id);
				const by = scroll && scrollOverwrittenBy(scroll, overwrites(player));
				return by && overwrittenTooltipOf(by);
			},
	};
}

// The stats the elixir groups go in, in the order they are listed. A stat outside it follows them, in its enum order.
const ELIXIR_STAT_ORDER = [
	Stat.StatStrength,
	Stat.StatAgility,
	Stat.StatStamina,
	Stat.StatIntellect,
	Stat.StatSpirit,
	Stat.StatAttackPower,
	Stat.StatRangedAttackPower,
	Stat.StatSpellDamage,
	Stat.StatArcaneDamage,
	Stat.StatFireDamage,
	Stat.StatFrostDamage,
	Stat.StatHolyDamage,
	Stat.StatNatureDamage,
	Stat.StatShadowDamage,
	Stat.StatArmor,
	Stat.StatArmorNoMods,
	Stat.StatHealth,
	Stat.StatMP5,
	Stat.StatHealingPower,
	Stat.StatPhysicalDamage,
];
const statRank = (stat: Stat) => {
	const index = ELIXIR_STAT_ORDER.indexOf(stat);
	return index >= 0 ? index : ELIXIR_STAT_ORDER.length + stat;
};
// The stats whose group takes a name of its own, which the stat's own name reads poorly as.
const ELIXIR_GROUP_NAMES: Partial<Record<Stat, string>> = {
	[Stat.StatArmor]: 'armor',
	[Stat.StatArmorNoMods]: 'armor',
	[Stat.StatSpellDamage]: 'spell_power',
	[Stat.StatMP5]: 'mana_regeneration',
	[Stat.StatHealingPower]: 'healing',
	[Stat.StatPhysicalDamage]: 'physical_damage',
};
const CRIT_PERCENTS = [PseudoStat.PseudoStatMeleeCritPercent, PseudoStat.PseudoStatSpellCritPercent, PseudoStat.PseudoStatRangedCritPercent];

export interface ElixirGroup<T> {
	key: string;
	label: string;
	items: Array<T>;
}

const groupName = (name: string) => i18n.t(`settings_tab.consumables.elixirs.groups.${name}`);

// The group an elixir is listed under, and where the group goes: Critical Strike first for one granting a crit,
// then a percent it grants, then its first stat in ELIXIR_STAT_ORDER, and Other last for one granting nothing the
// picker reads.
const elixirGroupOf = (consumable: Consumable): { key: string; rank: number; label: () => string } => {
	if (CRIT_PERCENTS.some(percent => consumable.pseudoStats[percent] > 0))
		return { key: 'critical_strike', rank: -2, label: () => groupName('critical_strike') };
	const percent = consumable.pseudoStats.findIndex(value => value > 0);
	if (percent >= 0) return { key: `percent ${percent}`, rank: -1, label: () => translatePseudoStat(percent) };
	const stats = consumable.stats.flatMap((value, stat) => (value > 0 ? [stat as Stat] : []));
	if (!stats.length) return { key: 'other', rank: Number.MAX_SAFE_INTEGER, label: () => groupName('other') };
	const stat = stats.reduce((first, next) => (statRank(next) < statRank(first) ? next : first));
	const name = ELIXIR_GROUP_NAMES[stat];
	return name
		? { key: name, rank: statRank(stat), label: () => groupName(name) }
		: { key: `stat ${stat}`, rank: statRank(stat), label: () => translateStat(stat) };
};

// The items by the group of their consumable, each group keeping the items' order. A group sits where its first
// stat does; two stats that share a name (the two armors) share a group.
export function elixirGroups<T>(items: ReadonlyArray<T>, consumableOf: (item: T) => Consumable): Array<ElixirGroup<T>> {
	const groups = new Map<string, { rank: number; group: ElixirGroup<T> }>();
	for (const item of items) {
		const { key, rank, label } = elixirGroupOf(consumableOf(item));
		const entry = groups.get(key);
		if (entry) {
			entry.rank = Math.min(entry.rank, rank);
			entry.group.items.push(item);
		} else groups.set(key, { rank, group: { key, label: label(), items: [item] } });
	}
	return [...groups.values()].sort((a, b) => a.rank - b.rank).map(({ group }) => group);
}
