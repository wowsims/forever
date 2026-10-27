import { readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { inputPicked } from '@features/inputs/components/ConnectedIconEnumPicker/utils';
import { type ConsumeConfig, consumeConfigs, type ConsumeToggle } from '@features/settings/components/ConsumesPicker/utils';
import { BUFFS_CONFIG, DEBUFFS_CONFIG, DEBUFFS_MISC_CONFIG, PARTY_BUFFS_CONFIG } from '@features/settings/model/buffs_debuffs';
import { DragonbreathChili } from '@features/settings/model/consumables_auto_gen';
import { isMultiIconOption, type PickerStatOptions, relevantStatOptions, statOptionScope } from '@features/settings/model/stat_options';
import { shownStats, type StatFilterSlot } from '@features/stat-filters/model/stat_filters';
import { ConsumableType, type PseudoStat, Stat } from '@generated/proto/common';
import type { Consumable } from '@generated/proto/db';
import { ResourceType } from '@generated/proto/spell';
import { DatabaseFilters, UIDatabase } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import { Database } from '@sim/proto/database';
import { consumablePseudoStatsOf, consumableStatsOf } from '@sim/settings/consumables';
import type { SpecDefinition } from '@sim/spec_config';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

const specs = Object.entries(import.meta.glob<SpecDefinition<any>>('./*/*/spec.ts', { eager: true, import: 'default' }));

const sections: Array<[StatFilterSlot, PickerStatOptions[]]> = [
	['buffs', [...BUFFS_CONFIG, ...PARTY_BUFFS_CONFIG]],
	['debuffs', [...DEBUFFS_CONFIG, ...DEBUFFS_MISC_CONFIG]],
];

const listed = (list: ReadonlyArray<unknown>, row: PickerStatOptions) => list.includes(row.config) || row.stats.some(stat => list.includes(stat));

const membersOf = (row: PickerStatOptions) => (isMultiIconOption(row) ? row.members : undefined);

// The rows the spec's exclusions leave: a group keeps the members they leave, tagged with those members' stats, and is
// gone with the last of them.
const survivors = (rows: ReadonlyArray<PickerStatOptions>, spec: SpecDefinition<any>): PickerStatOptions[] =>
	rows.flatMap(row => {
		const members = membersOf(row);
		if (!members) return listed(spec.excludeBuffDebuffInputs, row) ? [] : [row];
		const left = members.filter(member => !listed(spec.excludeBuffDebuffInputs, member));
		return left.length === 0 ? [] : [{ ...row, members: left, stats: [...new Set(left.flatMap(member => member.stats))] }];
	});

// The spec's own relevance rule, written out in full: the filter at its defaults must offer exactly the surviving rows
// that pass it.
const relevantTo = (spec: SpecDefinition<any>) => {
	const displayed = new Set(spec.displayStats.filter(stat => stat.hasRootStat()).map(stat => stat.getRootStat()));
	return (row: PickerStatOptions) =>
		row.stats.length === 0 ||
		row.stats.some(stat => displayed.has(stat) || spec.epStats.includes(stat)) ||
		listed(spec.includeBuffDebuffInputs, row) ||
		!!membersOf(row)?.some(member => spec.includeBuffDebuffInputs.includes(member.config));
};

const specRule = (rows: ReadonlyArray<PickerStatOptions>, spec: SpecDefinition<any>) => survivors(rows, spec).filter(relevantTo(spec));

const defaultShown = (slot: StatFilterSlot, rows: PickerStatOptions[], spec: SpecDefinition<any>) =>
	new Set(shownStats(DatabaseFilters.create(), slot, statOptionScope(rows, spec)).map(stat => stat.getStat()));

describe('the buff and debuff stat filters at their defaults', () => {
	it.each(specs.flatMap(([file, spec]) => sections.map(([slot, rows]) => [file, slot, spec, rows] as const)))(
		'%s: the %s filter offers the rows the spec lists',
		(_file, slot, spec, rows) => {
			const offered = relevantStatOptions(rows, spec, defaultShown(slot, rows, spec)).filter(row => !row.config.keptBy);

			expect(offered.map(shapeOf)).toEqual(specRule(rows, spec).map(shapeOf));
		},
	);
});

// The kept mark an adapter shows, for the value it reads: the input's own, or for a multi icon whether any input is picked.
const marked = (config: PickerStatOptions['config'], player: Player<any>) =>
	!!config.keptBy?.(player, ('inputs' in config ? inputPicked(config, player) : config.getValue(player)) as never);

const labelOf = (row: PickerStatOptions) => row.config.label ?? ('actionId' in row.config ? String(row.config.actionId.spellId) : '?');

// A row by its label, and a group by its members' labels too: a narrowed group is a new row with new closures.
const shapeOf = (row: PickerStatOptions) => [labelOf(row), ...(membersOf(row)?.map(labelOf) ?? [])];

// Whether the row holds a pick; for a group, one of the members still in it.
const picked = (row: PickerStatOptions, player: Player<any>) =>
	membersOf(row)?.some(member => inputPicked(member.config, player)) ?? inputPicked(row.config, player);

const presetPlayer = async (spec: SpecDefinition<any>) => {
	const { sim, player } = await makeIndividualSim({ spec: spec.spec, config: spec });
	player.setConsumes(spec.defaults.consumables);
	player.setBuffs(spec.defaults.individualBuffs);
	player.getParty()!.setBuffs(spec.defaults.partyBuffs);
	player.getRaid()!.setBuffs(spec.defaults.raidBuffs);
	sim.raid.setDebuffs(spec.defaults.debuffs);
	return player;
};

// Types whose consumables exist to grant a stat: one granting none of the spec's does nothing for it.
const STAT_ONLY_TYPES = [
	ConsumableType.ConsumableTypeUnknown,
	ConsumableType.ConsumableTypeConjured,
	ConsumableType.ConsumableTypeFlask,
	ConsumableType.ConsumableTypeFood,
	ConsumableType.ConsumableTypeScroll,
	ConsumableType.ConsumableTypeOther,
	ConsumableType.ConsumableTypeElixir,
];

// The consumables' stat rule, written out in full against the stats and percents the spec weighs
// consumables by and the stats it displays: the filter at its defaults must keep exactly the picks
// that fail it. Dragonbreath Chili's worth is damage, so it always passes.
const consumableRule = (consumable: Consumable, player: Player<any>) => {
	if (consumable.id === DragonbreathChili.id) return true;
	const specConfig = player.getSpecConfig();
	const displayed = (specConfig.displayStats ?? []).filter(stat => stat.hasRootStat()).map(stat => stat.getRootStat());
	const stats = consumableStatsOf({ epStats: [...specConfig.epStats, ...displayed], consumableStats: specConfig.consumableStats });
	const pseudoStats = consumablePseudoStatsOf(specConfig);
	const conditionalStats = [...consumable.mobTypeStats, ...consumable.zoneStats].flatMap(({ stats }) =>
		Object.entries(stats).flatMap(([stat, value]) => (value > 0 ? [Number(stat) as Stat] : [])),
	);
	const grantedStats = [...consumable.stats.flatMap((value, stat) => (value > 0 ? [stat as Stat] : [])), ...conditionalStats];
	const grantedPseudoStats = consumable.pseudoStats.flatMap((value, pseudoStat) => (value > 0 ? [pseudoStat as PseudoStat] : []));
	const restores = consumable.effectIds.flatMap(effectId => {
		const miscValue = Database.getSync().getSpellEffect(effectId)?.miscValue0;
		return miscValue?.oneofKind === 'resourceType' ? [miscValue.resourceType] : [];
	});
	const spends = (resource: ResourceType) => {
		if (resource === ResourceType.ResourceTypeMana) return [Stat.StatMana, Stat.StatMP5, Stat.StatIntellect].some(stat => stats.has(stat));
		if (resource === ResourceType.ResourceTypeHealth) return stats.has(Stat.StatHealth);
		return stats.has(Stat.StatAttackPower);
	};
	if (
		consumable.buffsMainStat ||
		grantedStats.some(stat => stats.has(stat)) ||
		grantedPseudoStats.some(pseudoStat => pseudoStats.has(pseudoStat)) ||
		restores.some(spends)
	)
		return true;
	return !grantedStats.length && !grantedPseudoStats.length && !restores.length && !STAT_ONLY_TYPES.includes(consumable.type);
};

// The consumable a picker holds: an enum picker's value is its id, a toggle's consumable is its icon's item.
const pickedConsumable = (config: ConsumeConfig | ConsumeToggle, player: Player<any>): Consumable | undefined => {
	if (config.type === 'icon') return config.getValue(player) ? Database.getSync().getConsumable(config.actionId.itemId) : undefined;
	return Database.getSync().getConsumable(config.getValue(player));
};

describe("a spec's default picks at the default stat filters", () => {
	const players = new Map<SpecDefinition<any>, Player<any>>();

	beforeAll(async () => {
		const dbPath = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../assets/database/db.json');
		serveDatabase(UIDatabase.fromJson(JSON.parse(readFileSync(dbPath, 'utf8'))));
		for (const [, spec] of specs) players.set(spec, await presetPlayer(spec));
	});

	const keptAndExpected = (spec: SpecDefinition<any>, slot: StatFilterSlot, rows: PickerStatOptions[]) => {
		const player = players.get(spec)!;
		const left = survivors(rows, spec);
		const offered = relevantStatOptions(rows, spec, defaultShown(slot, rows, spec));
		const relevant = relevantTo(spec);
		return {
			kept: left.filter((_, index) => marked(offered[index].config, player)).map(labelOf),
			expected: left.filter(row => picked(row, player) && !relevant(row)).map(labelOf),
		};
	};

	it.each(specs.flatMap(([file, spec]) => sections.map(([slot, rows]) => [file, slot, spec, rows] as const)))(
		'%s: the %s filter keeps exactly the picked rows the spec does not list',
		(_file, slot, spec, rows) => {
			const { kept, expected } = keptAndExpected(spec, slot, rows);

			expect(kept).toEqual(expected);
		},
	);

	it('keeps a default pick on some spec, so the rule above is exercised', () => {
		const kept = specs.flatMap(([, spec]) => sections.flatMap(([slot, rows]) => keptAndExpected(spec, slot, rows).kept));

		expect(kept.length).toBeGreaterThan(0);
	});

	const consumablePicks = (spec: SpecDefinition<any>) => {
		const player = players.get(spec)!;
		const configs = Object.entries(consumeConfigs(player, Database.getSync())).flatMap(([field, value]) =>
			(Array.isArray(value) ? value : [value]).map((config, index) => [`${field}[${index}]`, config] as const),
		);
		const picks = configs.flatMap(([field, config]) => {
			const consumable = pickedConsumable(config, player);
			return consumable ? [{ field, consumable }] : [];
		});
		return {
			picks,
			kept: configs
				.filter(([, config]) =>
					config.type === 'icon' ? config.keptBy?.(player, config.getValue(player)) : config.keptBy?.(player, config.getValue(player)),
				)
				.map(([field]) => field),
			expected: picks.filter(({ consumable }) => !consumableRule(consumable, player)).map(({ field }) => field),
		};
	};

	it.each(specs)('%s: the consumables filter keeps exactly the picked consumables the stat rule leaves out', (_file, spec) => {
		const { kept, expected } = consumablePicks(spec);

		expect(kept).toEqual(expected);
	});

	it('picks a default consumable on some spec, so the rule above is exercised', () => {
		const picks = specs.flatMap(([, spec]) => consumablePicks(spec).picks);

		expect(picks.length).toBeGreaterThan(0);
	});
});
