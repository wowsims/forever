import type { Faction } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translateSlotName } from '@i18n/localization';
import type { BulkDropsCombined, BulkRankedResult, BulkResults } from '@sim/bulk/types';
import { usePlayer, useSimHost } from '@sim/context/SimHostContext';
import { useSimReady } from '@sim/hooks/useSimReady';
import { wowheadIconUrl } from '@sim/proto/wowhead';
import { BULK_DROPS_ORDERS, type BulkDropsOrder, type BulkDropsView } from '@sim/state/sim_store';
import { formatDeltaText, formatSignificance } from '@sim/utils/format';
import { stDevToConf95 } from '@sim/utils/math';
import { Button } from '@ui-kit/Button';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { Icon } from '@ui-kit/Icon';
import { Pill } from '@ui-kit/Pill';
import { createToastManager, ToastArea, toastManager } from '@ui-kit/Toast';
import { ToggleGroup } from '@ui-kit/ToggleGroup';
import { type ReactNode, useEffect, useMemo, useState } from 'react';

import { useBulkMode } from '../../hooks/useBulkMode';
import { useBulkState } from '../../hooks/useBulkState';
import { useDropsSetup } from '../../hooks/useDropsSetup';
import { dropsPickedItemIds, wornIdSet } from '../../model/drops/items';
import {
	dropsGap,
	dropsItemChains,
	dropsItemLookup,
	dropsNewItem,
	dropsOrderedGroups,
	dropsResultItems,
	dropsSourceGroups,
	isProvenCombined,
	dropsSummary,
	runSourceKeys,
	topUpgrade,
} from '../../model/drops/results';
import { dropsLevelItemById, setDropsOrder, setDropsView } from '../../model/drops/settings';
import { type DropsNames, dropsNamesOf, dropsSourceLabel } from '../../model/drops/sources';
import { baseIlvl } from '../../model/gear/search';
import { equipAndToast } from '../../model/results/results';
import { baselineSignificance, isCulled, rowIterations } from '../../model/results/tie_chains';
import { runBulkBatch } from '../../model/run/run';
import { bulkRunLimits, canRunDropsBatch } from '../../model/selectors';
import { BulkRankingList } from '../BulkRankingList';
import { BulkResultRowView } from '../BulkResults/BulkResultRow';
import { BulkResultSlots } from '../BulkResults/BulkResultSlots';
import { BulkResultsCsvButton } from '../BulkResultsCsvButton';
import { DropsBossCardView } from '../DropsBossCard';
import { DropsItemLabel } from '../DropsItemLabel';
import { DropsSummaryView } from '../DropsSummary';

const t = (key: string) => i18n.t(`bulk_tab.drops.results.${key}`);

export interface DropsResultsViewProps {
	summary: ReactNode;
	view: BulkDropsView;
	onViewChange: (view: BulkDropsView) => void;
	/** The Bosses view's order and the orders it offers. */
	order: BulkDropsOrder;
	orders: readonly BulkDropsOrder[];
	onOrderChange: (order: BulkDropsOrder) => void;
	/** The source the Items view is narrowed to; null shows every item. */
	filter: { label: string; onClear: () => void } | null;
	cards: ReactNode;
	rows: ReactNode;
	/** At the end of the view switch's row, such as the CSV export. */
	actions?: ReactNode;
}

// A warning toast held in flow at the top of the results, with no close button: it stays as long as they do.
const DropsNextUpgradeNote = () => {
	const manager = useMemo(createToastManager, []);
	// Added once the area is mounted and subscribed, as a toast added earlier is dropped.
	useEffect(() => {
		manager.add({ variant: 'warning', title: t('next_upgrade_title'), body: t('next_upgrade_note'), autohide: false, canClose: false });
		return () => manager.close();
	}, [manager]);
	return <ToastArea manager={manager} inline testId="drops-next-upgrade-note" />;
};

export const DropsResultsView = ({ summary, view, onViewChange, order, orders, onOrderChange, filter, cards, rows, actions }: DropsResultsViewProps) => (
	<div className="grid content-start gap-4" data-testid="drops-results">
		<DropsNextUpgradeNote />
		{summary}
		<div className="flex flex-wrap items-center gap-2">
			<ToggleGroup
				size="sm"
				tone="neutral"
				aria-label={t('view_label')}
				value={view}
				options={[
					{ value: 'bosses', label: t('view.bosses') },
					{ value: 'items', label: t('view.items') },
				]}
				onChange={onViewChange}
			/>
			{view === 'bosses' && (
				<ChoiceField
					id="drops-order"
					label={t('order_label')}
					layout="inline"
					size="sm"
					testId="drops-order"
					options={orders.map(value => ({ value, label: t(`order.${value}`) }))}
					value={order}
					onChange={onOrderChange}
				/>
			)}
			{view === 'items' && filter && (
				<Pill testId="drops-source-filter" className="items-center gap-1">
					{t('filtered_by')} <span className="font-bold">{filter.label}</span>
					<Button variant="unstyled" aria-label={t('clear_filter')} onClick={filter.onClear}>
						<Icon name="xmark" />
					</Button>
				</Pill>
			)}
			{actions && <div className="ms-auto">{actions}</div>}
		</div>
		{view === 'bosses' ? <div className="grid grid-cols-1 gap-2 md:grid-cols-2 xl:grid-cols-3">{cards}</div> : <div>{rows}</div>}
	</div>
);

// A drop names its zone as the card's detail; any other source names its category, so a quest or vendor zone is not read as a drop zone.
const sourceText = (key: string, names: DropsNames): string => {
	const { label, detail } = dropsSourceLabel(key, names);
	return key.startsWith('drop:') ? label : `${label} (${detail})`;
};

// The rerun is the run button's own start, under its gate: off until the equip's recount is in, and off outside Drops.
const DropsRerunButton = ({ onRun }: { onRun: () => void }) => {
	const player = usePlayer();
	const ready = useSimReady();
	const { drops, powerTools } = useBulkMode();
	const picked = dropsPickedItemIds(useDropsSetup().list).length;
	const limits = bulkRunLimits(player.sim.isNative, powerTools);
	const canRun = useBulkState(slice => canRunDropsBatch(slice, picked, limits.combinations, limits.iterations));
	return (
		<Button size="sm" className="mt-2" data-testid="drops-run-again" disabled={!(ready && drops && canRun)} onClick={onRun}>
			{t('run_again')}
		</Button>
	);
};

// Long enough to reach the button.
const RERUN_TOAST_DELAY = 10000;

interface DropsResultsPaneProps {
	results: BulkResults;
	runSources: readonly string[];
	runFaction: Faction;
	combined: ReadonlyMap<string, BulkDropsCombined> | null;
}

const DropsResultsPane = ({ results, runSources, runFaction, combined }: DropsResultsPaneProps) => {
	const host = useSimHost();
	const { player } = host;
	const view = useBulkState(slice => slice.dropsView);
	const storedOrder = useBulkState(slice => slice.dropsOrder);
	const [source, setSource] = useState<string | null>(null);
	// A new run's results open unfiltered,
	const [shownResults, setShownResults] = useState(results);
	if (shownResults !== results) {
		setShownResults(results);
		setSource(null);
	}
	const { originalGearResults: current, currentGear, iterations } = results;

	const db = player.sim.db;
	// The faction the run started on, and no phase filter: a run's items keep their cards after a phase switch. The level
	// cutoff applies, so a stale tick on a source the tree hides makes no card.
	const itemById = useMemo(() => dropsItemLookup(dropsLevelItemById(db), runFaction), [db, runFaction]);
	const names = useMemo(() => dropsNamesOf(db), [db]);
	const runKeys = useMemo(() => new Set(runSources), [runSources]);
	const items = useMemo(() => dropsResultItems(results), [results]);
	const groups = useMemo(() => dropsSourceGroups(items, runKeys, currentGear, itemById, names), [items, runKeys, currentGear, itemById, names]);
	const summary = useMemo(() => dropsSummary(results, items, groups), [results, items, groups]);
	// A stored Combined gain order reads as Recommended on a run with none, and stays stored.
	const orders = BULK_DROPS_ORDERS.filter(value => value !== 'combined' || combined !== null);
	const order = orders.includes(storedOrder) ? storedOrder : 'recommended';
	// A copy: the summary takes its first card from `groups` in the Recommended order, whatever this one is.
	const ordered = useMemo(() => dropsOrderedGroups(groups, order, iterations, combined), [groups, order, iterations, combined]);
	const chains = useMemo(() => dropsItemChains(results, items, { source, groups }), [results, items, source, groups]);
	// Per row: the item it adds, the slot that item sits in and the run's sources that drop it.
	const rowDetails = useMemo(() => {
		const worn = wornIdSet(currentGear);
		return new Map(
			results.chains.flat().map(row => {
				const newItem = dropsNewItem(row.gear, worn);
				const slot = newItem ? row.gear.getItemSlots().find(candidate => row.gear.getEquippedItem(candidate)?.id === newItem.id) : undefined;
				const sources = newItem ? runSourceKeys(newItem.id, runKeys, names, itemById).map(sourceKey => sourceText(sourceKey, names)) : [];
				return [row, { newItem, slot, sources }];
			}),
		);
	}, [results, currentGear, runKeys, names, itemById]);

	const pct = (gain: number) => (gain / current.dpsMetrics.avg) * 100;
	const iconOf = (id: number) => wowheadIconUrl(itemById(id)?.icon ?? '', 'medium');
	const nameOf = (id: number) => itemById(id)?.name ?? String(id);
	const ilvlOf = (id: number) => {
		const item = itemById(id);
		return item && baseIlvl(item);
	};
	const selectCard = (key: string) => {
		setSource(key);
		setDropsView(player, 'items');
	};
	// The equip stays on the Upgrades side, and its announcement offers the same run again, which takes the gear it starts
	// on as its baseline.
	const equip = (row: BulkRankedResult) => {
		const toastId = equipAndToast(host, row, text => ({
			delay: RERUN_TOAST_DELAY,
			body: (
				<>
					<p className="mb-0">{text}</p>
					<DropsRerunButton
						onRun={() => {
							toastManager.close(toastId);
							void runBulkBatch(host);
						}}
					/>
				</>
			),
		}));
	};

	// The rank follows the Recommended order alone.
	const cards = ordered.map((group, index) => {
		const top = topUpgrade(group.items);
		const together = combined?.get(group.key);
		return (
			<DropsBossCardView
				key={group.key}
				rank={order === 'recommended' && group.upgrades > 0 ? index + 1 : null}
				label={group.label}
				strip={[
					...group.items.map(entry => ({
						itemId: entry.itemId,
						icon: iconOf(entry.itemId),
						name: nameOf(entry.itemId),
						ilvl: ilvlOf(entry.itemId),
						gainPct: pct(entry.gain),
						upgrade: entry.upgrade,
						onEquip: () => equip(entry.best),
					})),
					...group.wornIds.map(id => ({ itemId: id, icon: iconOf(id), name: nameOf(id), ilvl: ilvlOf(id), gainPct: null, upgrade: false })),
				]}
				best={top ? { dps: current.dpsMetrics.avg + top.gain, pct: pct(top.gain) } : null}
				combined={
					together && {
						dps: current.dpsMetrics.avg + together.gain,
						pct: (together.gain / together.baselineDps) * 100,
						marginPct: (together.margin / together.baselineDps) * 100,
						proven: isProvenCombined(together),
						items: together.itemIds.map(nameOf),
						leftOut: together.leftOutIds.map(nameOf),
					}
				}
				selected={source === group.key}
				onSelect={() => selectCard(group.key)}
			/>
		);
	});

	// Every Drops output is a DPS ranking, so a row shows its DPS whatever the display settings hide.
	const renderRow = (row: BulkRankedResult) => {
		const { newItem, slot, sources } = rowDetails.get(row)!;
		const [first, ...more] = sources;
		const isCurrent = row === current;
		const lowPrecision = isCulled(row);
		const gap = dropsGap(row, current);
		const significance = baselineSignificance(row, current, iterations);
		return (
			<BulkResultRowView
				showDamage
				avg={row.dpsMetrics.avg}
				plusMinus={lowPrecision ? 0 : stDevToConf95(row.dpsMetrics.stdev, rowIterations(row, iterations))}
				isBase={isCurrent}
				delta={formatDeltaText(current.dpsMetrics.avg, current.dpsMetrics.avg + gap, 2, undefined, !significance.isDiff, true)}
				significance={isCurrent ? '' : formatSignificance(significance)}
				lowPrecision={lowPrecision ? { note: t('low_precision'), tooltip: t('low_precision_tooltip') } : undefined}
				rootDataAttributes={{
					'data-testid': 'drops-result-row',
					...(isCurrent ? { 'data-current': '' } : newItem ? { 'data-item-id': String(newItem.id) } : {}),
					...(lowPrecision ? { 'data-precision': 'low' } : {}),
				}}
				onEquip={() => equip(row)}>
				{!isCurrent && (
					<BulkResultSlots
						gear={row.gear}
						currentGear={currentGear}
						changedOnly
						label={
							newItem && slot !== undefined ? (
								<DropsItemLabel item={newItem} slot={translateSlotName(slot)} source={first ?? ''} moreSources={more} />
							) : undefined
						}
					/>
				)}
			</BulkResultRowView>
		);
	};

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
					farmFirst={
						summary.farmFirst && {
							...summary.farmFirst.label,
							upgrades: summary.farmFirst.upgrades,
							total: summary.farmFirst.items.length + summary.farmFirst.wornIds.length,
						}
					}
				/>
			}
			view={view}
			onViewChange={next => setDropsView(player, next)}
			order={order}
			orders={orders}
			onOrderChange={next => setDropsOrder(player, next)}
			filter={source === null ? null : { label: sourceText(source, names), onClear: () => setSource(null) }}
			cards={cards}
			actions={<BulkResultsCsvButton />}
			rows={<BulkRankingList key={source ?? ''} chains={chains} renderRow={renderRow} />}
		/>
	);
};

// The Results pane of a Drops run; before one it invites a run, and once one started and has no results yet it shows nothing,
// as `BulkResults` does.
export const DropsResults = () => {
	const results = useBulkState(slice => slice.dropsResults);
	const started = useBulkState(slice => slice.dropsStarted);
	const runSources = useBulkState(slice => slice.dropsRunSources);
	const runFaction = useBulkState(slice => slice.dropsRunFaction);
	const combined = useBulkState(slice => slice.dropsCombined);

	if (!results?.[0]) return started ? null : <div className="flex items-center justify-center p-6">{t('run_simulation')}</div>;
	return <DropsResultsPane results={results[0].results} runSources={runSources} runFaction={runFaction} combined={combined} />;
};
