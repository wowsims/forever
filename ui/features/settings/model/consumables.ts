import type { IconEnumValueConfig } from '@features/inputs/configs';
import * as InputHelpers from '@features/inputs/input_helpers';
import type { StatFilterScope } from '@features/stat-filters/model/stat_filters';
import { STAT_GROUPS } from '@features/stat-filters/model/stat_groups';
import { Class, ConsumableType, ConsumesSpec, PseudoStat, Spec, Stat } from '@generated/proto/common';
import { Consumable } from '@generated/proto/db';
import type { DruidTalents } from '@generated/proto/druid';
import i18n from '@i18n/config';
import { translatePseudoStat, translateStat } from '@i18n/localization';
import type { StoreField } from '@sim/hooks/useStoreField';
import { Player } from '@sim/player/player';
import { ActionId } from '@sim/proto/action_id';
import type { Database } from '@sim/proto/database';
import { UnitStat, unitStatsOf } from '@sim/proto/stats';
import { consumableAllowed, consumableOrderOf, consumablePseudoStatsOf, consumableStatsOf, consumableUniverse, orderedByEp } from '@sim/settings/consumables';
import { memoLast } from '@sim/utils/misc';

import { ArcaneIntellect, GENERATED_OVERWRITING_BUFFS, LeaderOfThePack, MoonkinAura } from './buffs_debuffs_auto_gen';

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
// weapons and their showWhen whether item swap is on. The toggles read only STORE_FIELDS, and the inputs with an
// overwritten badge OVERWRITTEN_PICKER_FIELDS.
export const CONSUMABLE_PICKER_FIELDS: ReadonlyArray<StoreField> = [...STORE_FIELDS, 'itemSwap'];

// The fields the inputs with an overwritten badge read: the elixir and scroll picks, the miscellaneous toggles of the
// consumables that count with the elixirs, and MultiToggle, whose trigger shows its toggle's badge. The badges follow
// the raid buffs' inputs, since a raid buff turned on overwrites the scrolls and elixirs of its stat, and the talents,
// which turn on a druid's own.
export const OVERWRITTEN_PICKER_FIELDS: ReadonlyArray<StoreField> = [...CONSUMABLE_PICKER_FIELDS, 'raid:buffs', 'talentsString'];

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
			return ordered([order.weights, order.mobType], () => [none, ...orderedByEp(values, value => byId.get(value.value)!, order)]);
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

// A consumable picked any number at a time into the spec's extra ids. It reads as picked from the elixir ids too, and
// leaves both lists when unpicked, since a saved setting may hold it there.
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
			if (newValue) lists.extraIds.push(consumable.id);
			player.setConsumes({ ...consumes, ...lists });
		},
	};
}

// The keys of the stats and percents the consumable grants. Its armor of either kind counts as Armor too, the row the
// character stats panel shows it in.
const grantedKeys = (consumable: Consumable): Set<string> => {
	const stats = new Set(consumable.stats.flatMap((value, stat) => (value > 0 ? [stat as Stat] : [])));
	if (stats.has(Stat.StatArmorNoMods) || stats.has(Stat.StatBonusArmor)) stats.add(Stat.StatArmor);
	const pseudoStats = new Set(consumable.pseudoStats.flatMap((value, pseudoStat) => (value > 0 ? [pseudoStat as PseudoStat] : [])));
	return new Set(unitStatsOf({ stats, pseudoStats }).map(stat => stat.getKey()));
};

// The groups in the order the spec's character stats panel shows the stat every consumable of a group grants: for a
// warrior a group of Strength elixirs early and the crit elixirs where melee crit sits, for a mage the crit elixirs
// where spell crit sits and the Strength ones late. The panel's headed groups (`STAT_GROUPS`) come first and its health
// and mana after them. A group whose consumables share no stat the spec shows goes last, and groups that tie keep their
// order.
export const inDisplayOrder = <T extends ReadonlyArray<Consumable>>(groups: ReadonlyArray<T>, displayStats: ReadonlyArray<UnitStat>): Array<T> => {
	const panel = [...STAT_GROUPS.filter(group => group.key !== 'primary'), ...STAT_GROUPS.filter(group => group.key === 'primary')];
	const shown = new Set(displayStats.map(stat => stat.getKey()));
	const order = new Map(
		panel
			.flatMap(group => group.stats)
			.filter(stat => shown.has(stat.getKey()))
			.map((stat, index) => [stat.getKey(), index]),
	);
	const rankOf = (group: T) => {
		const [first, ...rest] = group.map(grantedKeys);
		const shared = [...(first ?? [])].filter(key => rest.every(keys => keys.has(key)));
		return Math.min(Number.MAX_SAFE_INTEGER, ...shared.map(key => order.get(key) ?? Number.MAX_SAFE_INTEGER));
	};
	const ranks = new Map(groups.map(group => [group, rankOf(group)]));
	return [...groups].sort((a, b) => ranks.get(a)! - ranks.get(b)!);
};

export const countsWithTheElixirs = (consumable: Consumable): boolean =>
	consumable.type === ConsumableType.ConsumableTypeElixir || consumable.type === ConsumableType.ConsumableTypeScroll || consumable.countsWithElixirs;

interface Grants {
	name: string;
	// Each stat and percent the pick or buff grants, by stat or percent.
	amounts: Map<string, number>;
}

interface PickGrants extends Grants {
	consumable: Consumable;
}

const grantsOf = (consumable: Consumable): PickGrants => {
	const amounts = new Map<string, number>();
	consumable.stats.forEach((value, stat) => value > 0 && amounts.set(`stat ${stat}`, value));
	consumable.pseudoStats.forEach((value, pseudoStat) => value > 0 && amounts.set(`percent ${pseudoStat}`, value));
	return { consumable, name: consumable.name, amounts };
};

// The raid buffs that overwrite picks, each with its input and what it grants.
const OVERWRITING_BUFFS = GENERATED_OVERWRITING_BUFFS.map(buff => ({
	config: buff.config,
	grants: {
		name: buff.config.label ?? '',
		amounts: new Map([
			...Object.entries(buff.stats ?? {}).map(([stat, amount]): [string, number] => [`stat ${stat}`, amount!]),
			...Object.entries(buff.pseudoStats ?? {}).map(([pseudoStat, amount]): [string, number] => [`percent ${pseudoStat}`, amount!]),
		]),
	},
}));

// Whether the picks grant at least as much of every amount, false for no amounts at all.
const matchedBy = (amounts: Map<string, number>, picks: ReadonlyArray<Grants>): boolean =>
	amounts.size > 0 && [...amounts].every(([key, value]) => picks.some(pick => (pick.amounts.get(key) ?? 0) >= value));

// What a pick loses to one other pick or raid buff: the names of the stats and percents that one outbids.
interface Lost {
	by: Grants;
	stats: Array<string>;
}

const CRIT_PERCENTS = [PseudoStat.PseudoStatMeleeCritPercent, PseudoStat.PseudoStatSpellCritPercent, PseudoStat.PseudoStatRangedCritPercent];

// The names of the stats and percents behind amount keys, the three crit percents as one Crit.
const statNamesOf = (keys: ReadonlyArray<string>): Array<string> => {
	const names = keys.map(key => {
		const [kind, id] = key.split(' ');
		if (kind === 'stat') return translateStat(Number(id));
		return CRIT_PERCENTS.includes(Number(id)) ? i18n.t('common.stats.crit') : translatePseudoStat(Number(id));
	});
	return [...new Set(names)];
};

// What the amounts lose to the rivals, one entry per rival that outbids some of them, with those stats and percents,
// each under the rival granting the most of it, in the order the amounts list them. A rival with `ties` outbids an
// equal amount too.
function lostTo(amounts: Map<string, number>, rivals: ReadonlyArray<{ other: Grants; ties: boolean }>): Array<Lost> {
	const lost = new Map<Grants, Array<string>>();
	for (const [key, value] of amounts) {
		let by: Grants | undefined;
		for (const { other, ties } of rivals) {
			const amount = other.amounts.get(key) ?? 0;
			if ((amount > value || (ties && amount === value)) && (!by || amount > by.amounts.get(key)!)) by = other;
		}
		if (by) lost.set(by, [...(lost.get(by) ?? []), key]);
	}
	return [...lost].map(([by, keys]) => ({ by, stats: statNamesOf(keys) }));
}

const tying = (rivals: ReadonlyArray<Grants>) => rivals.map(other => ({ other, ties: true }));

// The sim counts each stat and percent the elixirs and scrolls grant at the largest amount among them and the raid
// buffs that overwrite them (R-0054), so a pick adds nothing where the other picks and those buffs grant as much of everything it
// grants. The picks are set aside one at a time, those granting the fewest stats and percents first and, among those,
// the one later in the spec's lists first (the elixir ids, then the extra ids), so of two that grant the same only one
// is. One granting nothing the picker can read is never set aside. Answers the picks set aside and the picks left that
// lose some of what they grant, by id, each with what it loses to whom. A pick left loses to a raid buff of an equal
// amount, and to a pick of an equal amount earlier in the lists.
function elixirOverwrites(
	picks: ReadonlyArray<PickGrants>,
	buffs: ReadonlyArray<Grants>,
): { full: Map<number, Array<Lost>>; partly: Map<number, Array<Lost>> } {
	const order = picks.map((entry, index) => ({ entry, index })).sort((a, b) => a.entry.amounts.size - b.entry.amounts.size || b.index - a.index);
	let counted = [...picks];
	const setAside: Array<PickGrants> = [];
	for (const { entry } of order) {
		const others = counted.filter(other => other !== entry);
		if (!matchedBy(entry.amounts, [...others, ...buffs])) continue;
		counted = others;
		setAside.push(entry);
	}
	const partly = new Map<number, Array<Lost>>();
	counted.forEach((entry, index) => {
		const rivals = [...counted.flatMap((other, i) => (other === entry ? [] : [{ other, ties: i < index }])), ...tying(buffs)];
		const lost = lostTo(entry.amounts, rivals);
		if (lost.length) partly.set(entry.consumable.id, lost);
	});
	const full = new Map(setAside.map(entry => [entry.consumable.id, lostTo(entry.amounts, tying([...counted, ...buffs]))]));
	return { full, partly };
}

const listOf = (names: ReadonlyArray<string>) => new Intl.ListFormat(i18n.language, { type: 'conjunction' }).format(names);

// One line per pick or buff that overwrites some of what the pick grants.
const overwrittenTooltipOf = (lost: ReadonlyArray<Lost>): string =>
	lost.map(({ by, stats }) => i18n.t('settings_tab.consumables.overwritten_tooltip', { stats: listOf(stats), name: by.name })).join('\n');

// The raid buffs the sim turns on whatever the inputs say (`AddRaidBuffs`): Arcane Intellect for every mage
// (sim/mage/mage.go), Moonkin Aura for a Balance druid with Moonkin Form and Leader of the Pack for a Feral cat or bear
// with the talent (sim/druid/druid.go).
const turnedOnBySim = (player: Player<any>, config: (typeof OVERWRITING_BUFFS)[number]['config']): boolean => {
	if (config === ArcaneIntellect) return player.getClass() === Class.ClassMage;
	const spec = player.getSpec();
	if (config === MoonkinAura) return spec === Spec.SpecBalanceDruid && (player.getTalents() as DruidTalents).moonkinForm;
	if (config === LeaderOfThePack) {
		return (spec === Spec.SpecFeralCatDruid || spec === Spec.SpecFeralBearDruid) && (player.getTalents() as DruidTalents).leaderOfThePack;
	}
	return false;
};

// The grants of the overwriting raid buffs that are on, by their input or because the sim turns them on.
const buffsOn = (player: Player<any>) =>
	OVERWRITING_BUFFS.filter(buff => buff.config.getValue(player) || turnedOnBySim(player, buff.config)).map(buff => buff.grants);

interface Overwritten {
	tooltip: string;
	// Whether the pick keeps some of what it grants while it loses the rest.
	partly: boolean;
}

// The overwritten badges of the elixir and scroll picks, read off the picks among `counted`, the elixirs, the scrolls
// and the other consumables that count with the elixirs, in the spec's elixir and extra ids and the raid buffs that
// are on. A badge answers for a consumable as if it were picked in place of its group's pick, so an option in a
// group's menu reads what it would add. Of each exclusive group only the first pick counts, as the sim selects them
// (`selectedConsumableIDs` in sim/core/consumes.go).
export function overwrittenTooltips(counted: ReadonlyArray<Consumable>) {
	const grants = new Map(counted.map(consumable => [consumable.id, grantsOf(consumable)]));
	const groupOf = new Map(counted.map(consumable => [consumable.id, consumable.groupKey]));
	// The elixir pass of each set of picks asked for since the picks or the raid buffs last changed.
	const memo = memoLast<Map<string, ReturnType<typeof elixirOverwrites>>>();
	const elixirPass = (player: Player<any>, id: number) => {
		const { elixirIds, extraIds } = player.getConsumes();
		const listed = [...new Set([...elixirIds, ...extraIds])].filter(pick => grants.has(pick));
		const buffs = buffsOn(player);
		const cache = memo([listed.join(), buffs.map(buff => buff.name).join()], () => new Map());
		const group = groupOf.get(id);
		const withId = listed.map(pick => (group && groupOf.get(pick) === group ? id : pick));
		if (!withId.includes(id)) withId.push(id);
		const picked = withId.filter((pick, index) => {
			const pickGroup = groupOf.get(pick);
			return !pickGroup || withId.findIndex(other => groupOf.get(other) === pickGroup) === index;
		});
		const key = picked.join();
		const picks = picked.flatMap(pick => grants.get(pick) ?? []);
		if (!cache.has(key)) cache.set(key, elixirOverwrites(picks, buffs));
		return cache.get(key)!;
	};
	const pick = (player: Player<any>, id: number): Overwritten | undefined => {
		if (!id) return undefined;
		const { full, partly } = elixirPass(player, id);
		const lost = full.get(id) ?? partly.get(id);
		return lost && { tooltip: overwrittenTooltipOf(lost), partly: partly.has(id) };
	};
	return {
		pick,
		// The same for a toggle of a consumable that counts with the elixirs, picked or not.
		toggle: (consumable: Consumable) => ({ overwritten: (player: Player<any>) => pick(player, consumable.id) }),
	};
}
