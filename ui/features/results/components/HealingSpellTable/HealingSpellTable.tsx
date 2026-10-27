import { HitKind } from '@generated/proto/api';
import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { ActionId } from '@sim/proto/action_id';
import { formatToNumber } from '@sim/utils/format';
import { useMemo } from 'react';

import { useSimResult } from '../../hooks/useSimResult';
import { indexMetricRows, type MetricRow } from '../../model/grouping';
import { MetricsCombinedTooltip } from '../MetricsCombinedTooltip';
import { createMetricsColumnHelper, metricForAnchor, MetricsActionCell, MetricsTable, MetricTooltip } from '../MetricsTable';

export interface HealingSpellTableViewProps {
	spells: ReadonlyArray<HealingSpellRow>;
	procUptimes?: ReadonlyArray<{ name: string; target: string; uptime: number }>;
	spellOrder?: ReadonlyArray<string>;
}

export interface HealingSpellRow {
	id: ActionId;
	name: string;
	rank: number;
	periodic: boolean;
	absorb: boolean;
	normal?: HealingSpellRange;
	critical?: HealingSpellRange;
	critChance: number;
	customMetrics?: Readonly<Record<string, HealingSpellRange>>;
}

export interface HealingSpellRange {
	min: number;
	avg: number;
	max: number;
}

type SpellMetric = HealingSpellRow & { order: number; group: boolean };
type ProcUptime = NonNullable<HealingSpellTableViewProps['procUptimes']>[number];
const helper = createMetricsColumnHelper<SpellMetric>();
const procHelper = createMetricsColumnHelper<ProcUptime>();
const label = (key: string) => i18n.t(`results_tab.details.healing_spells.${key}`);
const displayAmount = (value: number | undefined) => (value === undefined ? '–' : formatToNumber(value, { maximumFractionDigits: 1 }));
const rangeTooltip = (spell: SpellMetric, range: HealingSpellRange | undefined, name: string) => {
	if (spell.group || !range) return null;
	return (
		<>
			{spell.periodic && <p className="pb-2">{label('periodic_tooltip')}</p>}
			<MetricsCombinedTooltip
				headerValues={[undefined, label('mean')]}
				groups={[{ totalPercentage: 100, data: [{ name, value: range.avg, percentage: 100, min: range.min, average: range.avg, max: range.max }] }]}
			/>
		</>
	);
};

export const HealingSpellTableView = ({ spells, procUptimes = [], spellOrder = [] }: HealingSpellTableViewProps) => {
	const rows = useMemo((): Array<MetricRow<SpellMetric>> => {
		const byName = new Map<string, Array<HealingSpellRow>>();
		for (const spell of spells) {
			const group = byName.get(spell.name) ?? [];
			group.push(spell);
			byName.set(spell.name, group);
		}
		const order = new Map(spellOrder.map((name, index) => [name, index]));
		return [...byName.entries()]
			.sort(([a], [b]) => (order.get(a) ?? Infinity) - (order.get(b) ?? Infinity) || a.localeCompare(b))
			.map(([, ranks], index) => {
				ranks.sort((a, b) => a.rank - b.rank);
				const metric = { ...ranks.at(-1)!, order: index, group: ranks.length > 1 };
				return ranks.length === 1 ? { metric } : { metric, subRows: ranks.map(spell => ({ metric: { ...spell, order: spell.rank, group: false } })) };
			});
	}, [spells, spellOrder]);
	const customKeys = useMemo(() => [...new Set(spells.flatMap(spell => Object.keys(spell.customMetrics ?? {})))], [spells]);
	const columns = useMemo(
		() =>
			helper.columns([
				helper.accessor(row => -row.metric.order, {
					id: 'name',
					header: i18n.t('results_tab.details.columns.name'),
					cell: info => {
						const spell = info.row.original.metric;
						const name =
							spell.group || !spell.rank
								? spell.name
								: `${spell.name} (${i18n.t('results_tab.details.healing_spells.rank_label', { rank: spell.rank })})`;
						return (
							<MetricsActionCell
								name={name}
								actionId={spell.id}
								expandable={info.row.getCanExpand()}
								expanded={info.row.getIsExpanded()}
								onToggle={info.row.getToggleExpandedHandler()}
							/>
						);
					},
				}),
				...(['normal', 'critical'] as const).map(kind =>
					helper.accessor(row => row.metric[kind]?.avg, {
						id: kind,
						header: label(kind),
						meta: { tooltipId: `healing-spell-${kind}` },
						cell: info => (info.row.original.metric.group ? '–' : displayAmount(info.getValue())),
					}),
				),
				helper.accessor(row => row.metric.critChance, {
					id: 'crit-chance',
					header: label('crit_chance'),
					cell: info => (info.row.original.metric.group || info.row.original.metric.absorb ? '–' : `${displayAmount(info.getValue())}%`),
				}),
				...customKeys.map(key =>
					helper.accessor(row => row.metric.customMetrics?.[key]?.avg, {
						id: key,
						header: i18n.t(`results_tab.details.custom_metrics.${key}`, { defaultValue: key }),
						meta: { tooltipId: `healing-spell-custom-${key}` },
						cell: info => (info.row.original.metric.group ? '–' : displayAmount(info.getValue())),
					}),
				),
			]),
		[customKeys],
	);
	const metricsByRowId = useMemo(() => indexMetricRows(rows), [rows]);
	const forAnchor = metricForAnchor(metricsByRowId);
	const procRows = useMemo(() => procUptimes.map(metric => ({ metric })), [procUptimes]);
	const procColumns = useMemo(
		() =>
			procHelper.columns([
				procHelper.accessor(row => row.metric.name, { id: 'name', header: label('proc'), cell: info => info.getValue() }),
				procHelper.accessor(row => row.metric.target, { id: 'target', header: label('target'), cell: info => info.getValue() }),
				procHelper.accessor(row => row.metric.uptime, { id: 'uptime', header: label('uptime'), cell: info => `${displayAmount(info.getValue())}%` }),
			]),
		[],
	);

	if (!rows.length) return <p className="p-4 text-sm opacity-70">{label('no_spells')}</p>;
	return (
		<>
			<MetricsTable rootTestId="healing-spells-table" columns={columns} rows={rows} sortColumnId="name" hasResult />
			<MetricTooltip
				id="healing-spell-normal"
				forAnchor={forAnchor}
				body={spell => rangeTooltip(spell, spell.normal, label(spell.absorb ? 'absorb' : spell.periodic ? 'normal_tick' : 'normal'))}
			/>
			<MetricTooltip
				id="healing-spell-critical"
				forAnchor={forAnchor}
				body={spell => rangeTooltip(spell, spell.critical, label(spell.periodic ? 'crit_tick' : 'critical'))}
			/>
			{customKeys.map(key => (
				<MetricTooltip
					key={key}
					id={`healing-spell-custom-${key}`}
					forAnchor={forAnchor}
					body={spell => rangeTooltip(spell, spell.customMetrics?.[key], i18n.t(`results_tab.details.custom_metrics.${key}`, { defaultValue: key }))}
				/>
			))}
			{procRows.length > 0 && (
				<div className="mt-4">
					<h3 className="px-2 pb-2 text-sm font-semibold">{label('proc_uptimes')}</h3>
					<MetricsTable rootTestId="healing-proc-uptime-table" columns={procColumns} rows={procRows} sortColumnId="uptime" hasResult />
				</div>
			)}
			<p className="pt-2 text-xs opacity-70">{label('uptime_note')}</p>
		</>
	);
};

export const HealingSpellTable = () => {
	const host = useSimHost();
	const resultData = useSimResult();
	const procAuraIds = host.individualConfig.healingProcAuras;
	const procUptimes = useMemo(() => {
		if (!resultData) return [];
		const duration = resultData.result.result.avgIterationDuration || 1;
		const procAuraIdSet = new Set(procAuraIds ?? []);
		return (resultData.result.result.raidMetrics?.parties ?? [])
			.flatMap(party => party.players)
			.filter(player => player.name.startsWith('Target Dummy'))
			.flatMap(player =>
				player.auras.flatMap(aura => {
					if (!aura.id) return [];
					const actionId = ActionId.fromProto(aura.id);
					return procAuraIdSet.has(actionId.spellId)
						? [{ name: actionId.name, target: player.name, uptime: ((aura.uptimeSecondsAvg || 0) / duration) * 100 }]
						: [];
				}),
			)
			.sort((a, b) => a.name.localeCompare(b.name) || a.target.localeCompare(b.target));
	}, [procAuraIds, resultData]);
	const actions = useMemo(() => {
		if (!resultData) return [];
		const player = resultData.result.getRaidIndexedPlayers(resultData.filter)[0];
		if (!player) return [];
		return player
			.getHealingActions()
			.map(action => action.forTarget(resultData.filter))
			.filter(action => !action.isPassiveAction || action.actionId.tag === 0)
			.filter(action => action.healing > 0 || action.shielding > 0)
			.map(action => {
				const periodic = action.targets.some(target => target.ticks > 0);
				const absorb = action.healing === action.shielding && action.shielding > 0;
				const { critical, landed } = action.targets.reduce(
					(total, target) => ({
						critical: total.critical + target.crits + target.critTicks,
						landed: total.landed + target.hits + target.crits + target.ticks + target.critTicks,
					}),
					{ critical: 0, landed: 0 },
				);
				const customMetrics = Object.fromEntries(
					action.customMetrics.map(metric => [metric.key, { avg: metric.average, min: metric.min ?? 0, max: metric.max ?? 0 }]),
				);
				const normal = action.range(periodic ? HitKind.HitKindHealTick : absorb ? HitKind.HitKindShield : HitKind.HitKindHeal);
				const criticalRange = action.range(periodic ? HitKind.HitKindCritHealTick : HitKind.HitKindCritHeal);
				return {
					id: action.actionId,
					name: action.actionId.baseName || action.name,
					rank: action.actionId.rank,
					periodic,
					absorb,
					normal: normal ? { min: normal.min, avg: normal.avg, max: normal.max } : undefined,
					critical: criticalRange ? { min: criticalRange.min, avg: criticalRange.avg, max: criticalRange.max } : undefined,
					critChance: landed > 0 ? (critical / landed) * 100 : 0,
					customMetrics,
				};
			});
	}, [resultData]);
	return <HealingSpellTableView spells={actions} procUptimes={procUptimes} spellOrder={host.individualConfig.healingSpellOrder} />;
};
