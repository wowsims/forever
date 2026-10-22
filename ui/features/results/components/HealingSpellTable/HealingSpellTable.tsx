import { HitKind } from '@generated/proto/api';
import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { useActionId } from '@sim/hooks/useActionId';
import { ActionId } from '@sim/proto/action_id';
import { formatToNumber } from '@sim/utils/format';
import { Accordion, AccordionItem } from '@ui-kit/Accordion';
import { Table, TableCell, TableHeaderCell, TableRow } from '@ui-kit/Table';
import { TooltipButton } from '@ui-kit/TooltipButton';
import { WowheadIcon } from '@ui-kit/WowheadIcon';
import { useMemo, useState } from 'react';

import { useSimResult } from '../../hooks/useSimResult';

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

const displayAmount = (value: number | undefined) => formatToNumber(value ?? 0, { maximumFractionDigits: 1 });
const displayOptionalAmount = (value: number | undefined) => (value === undefined ? '–' : displayAmount(value));
const HealingSpellRankCell = ({ spell }: { spell: HealingSpellRow }) => {
	const { iconUrl, href } = useActionId(spell.id);
	const rankLabel = spell.rank
		? i18n.t('results_tab.details.healing_spells.rank_label', { rank: spell.rank })
		: i18n.t('results_tab.details.healing_spells.rank');

	return (
		<TableCell>
			<span className="inline-flex items-center gap-2">
				{iconUrl && <WowheadIcon href={href} iconUrl={iconUrl} label={spell.name} className="size-4 rounded-sm" />}
				<span>{rankLabel}</span>
			</span>
		</TableCell>
	);
};

const HealingSpellGroupTitle = ({ spell }: { spell: HealingSpellRow }) => {
	const { iconUrl } = useActionId(spell.id);
	return (
		<span className="inline-flex items-center gap-2">
			{iconUrl && <WowheadIcon as="div" iconUrl={iconUrl} label={spell.name} className="h-6 w-6 shrink-0 rounded-sm" />}
			<span>{spell.name}</span>
		</span>
	);
};

export const HealingSpellTableView = ({ spells, procUptimes = [], spellOrder = [] }: HealingSpellTableViewProps) => {
	const [expanded, setExpanded] = useState<Array<string>>([]);
	const groups = useMemo(() => {
		const byName = new Map<string, Array<HealingSpellRow>>();
		for (const spell of spells) {
			const group = byName.get(spell.name) ?? [];
			group.push(spell);
			byName.set(spell.name, group);
		}
		const order = new Map(spellOrder.map((name, index) => [name, index]));
		return [...byName.entries()]
			.map(([name, ranks]) => [name, ranks.sort((a, b) => a.rank - b.rank)] as const)
			.sort(([nameA], [nameB]) => {
				const orderA = order.get(nameA);
				const orderB = order.get(nameB);
				if (orderA !== undefined && orderB !== undefined) return orderA - orderB;
				if (orderA !== undefined) return -1;
				if (orderB !== undefined) return 1;
				return nameA.localeCompare(nameB);
			});
	}, [spells, spellOrder]);

	if (!groups.length) return <p className="p-4 text-sm opacity-70">{i18n.t('results_tab.details.healing_spells.no_spells')}</p>;

	return (
		<div className="p-2">
			<Accordion value={expanded} onValueChange={setExpanded} variant="boxed" testId="healing-spell-accordion">
				{groups.map(([name, ranks]) => {
					const first = ranks[0];
					const isAbsorb = first.absorb;
					const isPeriodic = first.periodic;
					const visibleCustomMetrics = [...new Set(ranks.flatMap(spell => Object.keys(spell.customMetrics ?? {})))];
					return (
						<AccordionItem
							key={name}
							value={name}
							title={<HealingSpellGroupTitle spell={first} />}
							testId={`healing-spell-${name}`}
							panelClassName="overflow-x-auto p-2">
							<Table dense striped className="w-full table-fixed text-center" testId={`healing-spell-table-${name}`}>
								<thead>
									<TableRow>
										<TableHeaderCell rowSpan={2}>
											<span className="inline-flex items-center justify-center">
												{i18n.t('results_tab.details.healing_spells.rank')}
												<TooltipButton tooltip={i18n.t('results_tab.details.healing_spells.periodic_tooltip')} className="ml-1" />
											</span>
										</TableHeaderCell>
										{isAbsorb ? (
											<TableHeaderCell colSpan={3}>{i18n.t('results_tab.details.healing_spells.absorb')}</TableHeaderCell>
										) : isPeriodic ? (
											<TableHeaderCell colSpan={2}>{i18n.t('results_tab.details.healing_spells.per_tick')}</TableHeaderCell>
										) : (
											<>
												<TableHeaderCell colSpan={3}>{i18n.t('results_tab.details.healing_spells.normal')}</TableHeaderCell>
												<TableHeaderCell colSpan={3}>{i18n.t('results_tab.details.healing_spells.critical')}</TableHeaderCell>
											</>
										)}
										{!isAbsorb && <TableHeaderCell rowSpan={2}>{i18n.t('results_tab.details.healing_spells.crit_chance')}</TableHeaderCell>}
										{visibleCustomMetrics.map(key => (
											<TableHeaderCell key={key} colSpan={3}>
												{i18n.t(`results_tab.details.custom_metrics.${key}`, { defaultValue: key })}
											</TableHeaderCell>
										))}
									</TableRow>
									<TableRow>
										{!isAbsorb &&
											(isPeriodic ? ['normal_tick', 'crit_tick'] : ['min', 'mean', 'max']).map(column => (
												<TableHeaderCell key={`normal-${column}`}>
													{i18n.t(`results_tab.details.healing_spells.${column}`)}
												</TableHeaderCell>
											))}
										{!isAbsorb &&
											!isPeriodic &&
											['min', 'mean', 'max'].map(column => (
												<TableHeaderCell key={`crit-${column}`}>
													{i18n.t(`results_tab.details.healing_spells.${column}`)}
												</TableHeaderCell>
											))}
										{isAbsorb &&
											['min', 'mean', 'max'].map(column => (
												<TableHeaderCell key={`absorb-${column}`}>
													{i18n.t(`results_tab.details.healing_spells.${column}`)}
												</TableHeaderCell>
											))}
										{visibleCustomMetrics.flatMap(key =>
											['min', 'mean', 'max'].map(column => (
												<TableHeaderCell key={`${key}-${column}`}>
													{i18n.t(`results_tab.details.healing_spells.${column}`)}
												</TableHeaderCell>
											)),
										)}
									</TableRow>
								</thead>
								<tbody>
									{ranks.map(spell => {
										return (
											<TableRow key={spell.id.toString()}>
												<HealingSpellRankCell spell={spell} />
												{(isAbsorb || !isPeriodic
													? [spell.normal?.min, spell.normal?.avg, spell.normal?.max]
													: [spell.normal?.avg, spell.critical?.avg]
												).map((value, index) => (
													<TableCell key={`normal-${index}`}>{displayAmount(value)}</TableCell>
												))}
												{!isAbsorb &&
													!isPeriodic &&
													[spell.critical?.min, spell.critical?.avg, spell.critical?.max].map((value, index) => (
														<TableCell key={`crit-${index}`}>{displayAmount(value)}</TableCell>
													))}
												{!isAbsorb && <TableCell>{displayAmount(spell.critChance)}%</TableCell>}
												{visibleCustomMetrics.flatMap(key => {
													const range = spell.customMetrics?.[key];
													return [range?.min, range?.avg, range?.max].map((value, index) => (
														<TableCell key={`${key}-${index}`}>{displayOptionalAmount(value)}</TableCell>
													));
												})}
											</TableRow>
										);
									})}
								</tbody>
							</Table>
						</AccordionItem>
					);
				})}
			</Accordion>
			{procUptimes.length > 0 && (
				<div className="mt-4">
					<h3 className="px-2 pb-2 text-sm font-semibold">{i18n.t('results_tab.details.healing_spells.proc_uptimes')}</h3>
					<Table dense striped testId="healing-proc-uptime-table">
						<thead>
							<TableRow>
								<TableHeaderCell>{i18n.t('results_tab.details.healing_spells.proc')}</TableHeaderCell>
								<TableHeaderCell>{i18n.t('results_tab.details.healing_spells.target')}</TableHeaderCell>
								<TableHeaderCell>{i18n.t('results_tab.details.healing_spells.uptime')}</TableHeaderCell>
							</TableRow>
						</thead>
						<tbody>
							{procUptimes.map((proc, index) => (
								<TableRow key={`${proc.name}-${proc.target}-${index}`}>
									<TableCell>{proc.name}</TableCell>
									<TableCell>{proc.target}</TableCell>
									<TableCell>{`${proc.uptime.toFixed(2)}%`}</TableCell>
								</TableRow>
							))}
						</tbody>
					</Table>
				</div>
			)}
			<p className="pt-2 text-xs opacity-70">{i18n.t('results_tab.details.healing_spells.uptime_note')}</p>
		</div>
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
		const players = resultData.result.getRaidIndexedPlayers(resultData.filter);
		const player = players.find(candidate => candidate.spec?.isHealingSpec) ?? players[0];
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
					name: action.name,
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
