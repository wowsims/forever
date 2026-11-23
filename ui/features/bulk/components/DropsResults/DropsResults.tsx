import type { Faction } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translateSlotName } from '@i18n/localization';
import type { BulkRankedResult, BulkResults, TopGearResult } from '@sim/bulk/types';
import { useSimHost } from '@sim/context/SimHostContext';
import { wowheadIconUrl } from '@sim/proto/wowhead';
import type { BulkDropsView } from '@sim/state/sim_store';
import { formatDeltaText, formatSignificance } from '@sim/utils/format';
import { stDevToConf95, zTest } from '@sim/utils/math';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { Pill } from '@ui-kit/Pill';
import { toastManager } from '@ui-kit/Toast';
import { ToggleGroup } from '@ui-kit/ToggleGroup';
import { type ReactNode, useMemo, useState } from 'react';

import { useBulkState } from '../../hooks/useBulkState';
import {
	dropsGap,
	dropsItemChains,
	dropsItemLookup,
	dropsNewItem,
	type DropsResultItem,
	dropsResultItems,
	dropsSourceGroups,
	dropsSummary,
	isDropsLowPrecision,
	topUpgrade,
} from '../../model/drops/results';
import { setDropsView } from '../../model/drops/settings';
import { dropsItemSourceKeys, type DropsNames, dropsNamesOf, dropsSourceLabel } from '../../model/drops/sources';
import { type BulkEquipped, equipBulkResult } from '../../model/results/results';
import { isTiedWithBaseline } from '../../model/results/tie_chains';
import { BulkResultTieGroup } from '../BulkResults/BulkResultTieGroup';
import { DropsBossCardView } from '../DropsBossCard';
import { DropsItemRowHeader, DropsItemRowView } from '../DropsItemRow';
import { DropsSummaryView } from '../DropsSummary';

const t = (key: string) => i18n.t(`bulk_tab.drops.results.${key}`);

export interface DropsResultsViewProps {
	summary: ReactNode;
	view: BulkDropsView;
	onViewChange: (view: BulkDropsView) => void;
	showAll: boolean;
	onShowAllChange: (showAll: boolean) => void;
	/** The source the Items view is narrowed to; null shows every item. */
	filter: { label: string; onClear: () => void } | null;
	cards: ReactNode;
	rows: ReactNode;
}

export const DropsResultsView = ({ summary, view, onViewChange, showAll, onShowAllChange, filter, cards, rows }: DropsResultsViewProps) => (
	<div className="grid content-start gap-4" data-testid="drops-results">
		{summary}
		<div className="flex flex-wrap items-center gap-2">
			<ToggleGroup
				size="sm"
				aria-label={t('view_label')}
				value={view}
				options={[
					{ value: 'bosses', label: t('view.bosses') },
					{ value: 'items', label: t('view.items') },
				]}
				onChange={onViewChange}
			/>
			{view === 'items' && (
				<>
					<BooleanPicker label={t('show_all_variations')} layout="inline" size="xs" value={showAll} onChange={onShowAllChange} />
					{filter && (
						<Pill testId="drops-source-filter" className="items-center gap-1">
							{t('filtered_by')} <span className="font-bold">{filter.label}</span>
							<Button variant="unstyled" aria-label={t('clear_filter')} onClick={filter.onClear}>
								<Icon name="xmark" />
							</Button>
						</Pill>
					)}
				</>
			)}
		</div>
		{view === 'bosses' ? (
			<div className="grid grid-cols-1 gap-2 md:grid-cols-2 xl:grid-cols-3">{cards}</div>
		) : (
			<div>
				<DropsItemRowHeader />
				{rows}
			</div>
		)}
	</div>
);

// A drop names its zone as the card's detail; any other source names its category, so a quest or vendor zone is not read as a drop zone.
const sourceText = (key: string, names: DropsNames): string => {
	const { label, detail } = dropsSourceLabel(key, names);
	return key.startsWith('drop:') ? label : `${label} (${detail})`;
};

const equippedToast = (equipped: BulkEquipped): string =>
	equipped === 'build' ? i18n.t('bulk_tab.results.build_equipped') : i18n.t('bulk_tab.results.gear_equipped');

// The z of a row's gap to the current gear: the paired one when the row has a paired error, else the two-sample one.
const gapSignificance = (row: TopGearResult, current: TopGearResult, gap: number, iterations: number): string => {
	const isDiff = !isTiedWithBaseline(row, current);
	if (row.pairedErrorToBaseline && row.pairedDpsDiffToBaseline !== undefined)
		return formatSignificance({ z: Math.abs(gap) / row.pairedErrorToBaseline, isDiff });
	const n = (result: TopGearResult) => result.dpsMetrics.aggregatorData?.n ?? iterations;
	const { z } = zTest(n(row), row.dpsMetrics.avg, row.dpsMetrics.stdev, n(current), current.dpsMetrics.avg, current.dpsMetrics.stdev);
	return formatSignificance({ z, isDiff });
};

interface DropsResultsPaneProps {
	results: BulkResults;
	runSources: readonly string[];
	runFaction: Faction;
	onActivateTab: (id: string) => void;
}

const DropsResultsPane = ({ results, runSources, runFaction, onActivateTab }: DropsResultsPaneProps) => {
	const host = useSimHost();
	const { player } = host;
	const view = useBulkState(slice => slice.dropsView);
	const [showAll, setShowAll] = useState(false);
	const [source, setSource] = useState<string | null>(null);
	const { originalGearResults: current, currentGear, iterations } = results;

	const db = player.sim.db;
	// The faction the run started on, and no phase filter: a run's items keep their cards after a phase switch.
	const itemById = useMemo(() => dropsItemLookup(db.getAllItems(), runFaction), [db, runFaction]);
	const names = useMemo(() => dropsNamesOf(db), [db]);
	const runKeys = useMemo(() => new Set(runSources), [runSources]);
	const items = useMemo(() => dropsResultItems(results), [results]);
	const groups = useMemo(() => dropsSourceGroups(items, runKeys, currentGear, itemById, names), [items, runKeys, currentGear, itemById, names]);
	const summary = useMemo(() => dropsSummary(results, items, groups), [results, items, groups]);
	const chains = useMemo(
		() => dropsItemChains(results, items, { showAll, source, runSources: runKeys, names, itemById }),
		[results, items, showAll, source, runKeys, names, itemById],
	);

	const pct = (gain: number) => (gain / current.dpsMetrics.avg) * 100;
	const iconOf = (id: number) => wowheadIconUrl(itemById(id)?.icon ?? '', 'medium');
	const nameOf = (id: number) => itemById(id)?.name ?? String(id);
	const selectCard = (key: string) => {
		setSource(key);
		setDropsView(player, 'items');
	};

	const cards = groups.map((group, index) => {
		const top = topUpgrade(group.items);
		return (
			<DropsBossCardView
				key={group.key}
				rank={group.upgrades > 0 ? index + 1 : null}
				label={group.label}
				strip={[
					...group.items.map(entry => ({
						itemId: entry.itemId,
						icon: iconOf(entry.itemId),
						name: nameOf(entry.itemId),
						gainPct: pct(entry.gain),
						upgrade: entry.upgrade,
					})),
					...group.wornIds.map(id => ({ itemId: id, icon: iconOf(id), name: nameOf(id), gainPct: null, upgrade: false })),
				]}
				best={top ? { dps: current.dpsMetrics.avg + top.gain, pct: pct(top.gain) } : null}
				upgrades={group.upgrades}
				total={group.total}
				selected={source === group.key}
				onSelect={() => selectCard(group.key)}
			/>
		);
	});

	const itemOfBest = new Map<BulkRankedResult, DropsResultItem>(items.map(entry => [entry.best, entry]));
	let rank = 0;
	const renderRow = (row: BulkRankedResult, key: number) => {
		const newItem = dropsNewItem(row.gear, currentGear);
		const isCurrent = row === current;
		if (!isCurrent) rank += 1;
		const slot = newItem ? row.gear.getItemSlots().find(candidate => row.gear.getEquippedItem(candidate)?.id === newItem.id) : undefined;
		const [first, ...more] = newItem
			? (itemById(newItem.id) ? dropsItemSourceKeys(itemById(newItem.id)!, names) : [])
					.filter(sourceKey => runKeys.has(sourceKey))
					.map(sourceKey => sourceText(sourceKey, names))
			: [];
		const gap = dropsGap(row, current);
		const lowPrecision = isDropsLowPrecision(row);
		return (
			<DropsItemRowView
				key={key}
				rank={isCurrent ? null : rank}
				item={isCurrent ? null : newItem}
				slot={slot === undefined ? '' : translateSlotName(slot)}
				source={first ?? ''}
				moreSources={more}
				avg={row.dpsMetrics.avg}
				plusMinus={lowPrecision ? null : stDevToConf95(row.dpsMetrics.stdev, row.dpsMetrics.aggregatorData?.n ?? iterations)}
				delta={formatDeltaText(current.dpsMetrics.avg, current.dpsMetrics.avg + gap, 2, undefined, isTiedWithBaseline(row, current), true)}
				significance={isCurrent ? '' : gapSignificance(row, current, gap, iterations)}
				hiddenVariations={showAll ? 0 : (itemOfBest.get(row)?.variants.length ?? 0)}
				onShowVariations={() => setShowAll(true)}
				onEquip={
					isCurrent
						? null
						: () => {
								const equipped = equipBulkResult(host, row);
								onActivateTab('gear-tab');
								toastManager.add({ variant: 'success', body: equippedToast(equipped) });
							}
				}
			/>
		);
	};

	let rowKey = 0;
	const rows = chains.map((chain, chainIdx) =>
		chain.length > 1 ? (
			<BulkResultTieGroup key={`group-${chainIdx}`}>{chain.map(row => renderRow(row, rowKey++))}</BulkResultTieGroup>
		) : (
			renderRow(chain[0], rowKey++)
		),
	);

	return (
		<DropsResultsView
			summary={
				<DropsSummaryView
					current={{ avg: current.dpsMetrics.avg, plusMinus: stDevToConf95(current.dpsMetrics.stdev, iterations) }}
					bestDrop={
						summary.best
							? {
									name: itemById(summary.best.itemId)?.name ?? `#${summary.best.itemId}`,
									dps: current.dpsMetrics.avg + summary.best.gain,
									pct: pct(summary.best.gain),
								}
							: null
					}
					bestBoss={
						summary.bestGroup
							? {
									label: summary.bestGroup.label.label,
									detail: summary.bestGroup.label.detail,
									pct: pct(topUpgrade(summary.bestGroup.items)!.gain),
								}
							: null
					}
				/>
			}
			view={view}
			onViewChange={next => setDropsView(player, next)}
			showAll={showAll}
			onShowAllChange={setShowAll}
			filter={source === null ? null : { label: sourceText(source, names), onClear: () => setSource(null) }}
			cards={cards}
			rows={rows}
		/>
	);
};

export interface DropsResultsProps {
	onActivateTab: (id: string) => void;
}

// The Results pane of a Drops run; before one it invites a run, and once one started and has no results yet it shows nothing,
// as `BulkResults` does.
export const DropsResults = ({ onActivateTab }: DropsResultsProps) => {
	const results = useBulkState(slice => slice.dropsResults);
	const started = useBulkState(slice => slice.dropsStarted);
	const runSources = useBulkState(slice => slice.dropsRunSources);
	const runFaction = useBulkState(slice => slice.dropsRunFaction);

	if (!results?.[0]) return started ? null : <div className="flex items-center justify-center p-6">{t('run_simulation')}</div>;
	return <DropsResultsPane results={results[0].results} runSources={runSources} runFaction={runFaction} onActivateTab={onActivateTab} />;
};
