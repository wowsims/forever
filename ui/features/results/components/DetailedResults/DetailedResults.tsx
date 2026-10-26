import { Tabs } from '@base-ui/react/tabs';
import { SimRun, SimRunData } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { LOCAL_STORAGE_PREFIX } from '@sim/constants/other';
import { useSimHost } from '@sim/context/SimHostContext';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import { SimResult } from '@sim/proto/sim_result';
import { subscribeSimSettingsChange } from '@sim/state/subscriptions';
import { isDevMode } from '@sim/utils/env';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { useLocalStorageState } from '@ui-kit/hooks/useLocalStorageState';
import { useStickyToolbar } from '@ui-kit/hooks/useStickyToolbar';
import { TabPanels } from '@ui-kit/TabNav';
import clsx from 'clsx';
import { type ReactNode, useCallback, useEffect, useMemo, useRef, useState } from 'react';

import { trackEvent } from '../../../../tracking/utils';
import { GroupRanksContext } from '../../context/GroupRanksContext';
import { ShowApmBreakdownContext } from '../../context/ShowApmBreakdownContext';
import { useSimResult } from '../../hooks/useSimResult';
import type { SimResultsManager } from '../../model/results_manager';
import { AuraMetricsTable } from '../AuraMetricsTable';
import { CastMetricsTable } from '../CastMetricsTable';
import { DamageMetricsTable } from '../DamageMetricsTable';
import { DtpsMetricsTable } from '../DtpsMetricsTable';
import { HealingMetricsTable } from '../HealingMetricsTable';
import { LogRunner } from '../LogRunner';
import { ResourceMetricsTable } from '../ResourceMetricsTable';
import { ALL_UNITS, hasTarget, ResultsFilter, simResultFilter } from '../ResultsFilter';
import { ThreatMetricsTable } from '../ThreatMetricsTable';
import { Timeline } from '../Timeline';
import { DrStickySlotContext } from './context/DrStickySlotContext';
import { DetailedResultsPane } from './DetailedResultsPane';
import { DetailedResultsTabs } from './DetailedResultsTabs';
import { DEFAULT_DETAILED_RESULTS_TAB, DETAILED_RESULTS_TABS, type DetailedResultsTabConfig } from './utils';

// The viewer's choice, shared by every spec page.
const GROUP_RANKS_STORAGE_KEY = `${LOCAL_STORAGE_PREFIX}_results-group-ranks.v1`;
const SHOW_APM_BREAKDOWN_STORAGE_KEY = `${LOCAL_STORAGE_PREFIX}_results-show-apm-breakdown.v1`;

const parseFlag = (value: unknown): boolean | undefined => (typeof value === 'boolean' ? value : undefined);

export interface DetailedResultsViewProps {
	tabs: ReadonlyArray<DetailedResultsTabConfig>;
	activeId: string;
	onActiveIdChange: (id: string) => void;
	hasResults: boolean;
	iterationDisabled: boolean;
	deathDisabled: boolean;
	onSimulateIteration: () => void;
	onSimulateDeath: () => void;
	targetPicker: ReactNode;
	/** Whether the tables of the tabs marked `groupsRanks` join the ranks of a spell under one row. */
	groupRanks: boolean;
	onGroupRanksChange: (groupRanks: boolean) => void;
	/** Whether the tables of the tabs marked `breaksDownApm` list only the player's presses, which APM counts. */
	showApmBreakdown: boolean;
	onShowApmBreakdownChange: (showApmBreakdown: boolean) => void;
	/** One `DetailedResultsPane` per tab, its value the tab's id. */
	children: ReactNode;
}

export const DetailedResultsView = ({
	tabs,
	activeId,
	onActiveIdChange,
	hasResults,
	iterationDisabled,
	deathDisabled,
	onSimulateIteration,
	onSimulateDeath,
	targetPicker,
	groupRanks,
	onGroupRanksChange,
	showApmBreakdown,
	onShowApmBreakdownChange,
	children,
}: DetailedResultsViewProps) => {
	const { ref: toolbarRef, stuck, className: stickyToolbarClassName } = useStickyToolbar<HTMLDivElement>();
	const [stickySlot, setStickySlot] = useState<HTMLElement | null>(null);
	const activeTab = tabs.find(tab => tab.id === activeId);

	return (
		<div data-testid="detailed-results-manager-root" className="flex flex-col *:min-h-0">
			<div data-testid="detailed-results-controls-div" className="mb-2 flex">
				<Button data-testid="detailed-results-1-iteration-button" disabled={iterationDisabled} onClick={onSimulateIteration}>
					{i18n.t('results_tab.details.sim_1_iteration')}
				</Button>
				<Button className="ml-2" disabled={deathDisabled} onClick={onSimulateDeath}>
					{i18n.t('results_tab.details.sim_1_death')}
				</Button>
			</div>
			<DrStickySlotContext.Provider value={stickySlot}>
				<Tabs.Root
					className="group/dr flex scroll-mt-sim-header flex-col"
					data-testid="dr-root"
					data-no-results={!hasResults ? '' : undefined}
					value={activeId}
					onValueChange={next => onActiveIdChange(String(next))}>
					<div ref={toolbarRef} data-testid="dr-toolbar" className={clsx(stickyToolbarClassName, 'flex-col')} data-stuck={stuck ? '' : undefined}>
						<div data-testid="dr-toolbar-row" className="flex min-h-0 w-full items-center">
							<div data-testid="results-filter" className="flex min-h-0 items-center">
								{targetPicker}
								{activeTab?.groupsRanks && (
									<BooleanPicker
										id="results-group-ranks"
										className="ml-4 group-data-no-results/dr:hidden"
										label={i18n.t('results_tab.details.group_ranks')}
										layout="inline"
										value={groupRanks}
										onChange={onGroupRanksChange}
									/>
								)}
								{activeTab?.breaksDownApm && (
									<BooleanPicker
										id="results-show-apm-breakdown"
										className="ml-4 group-data-no-results/dr:hidden"
										label={i18n.t('results_tab.details.show_apm_breakdown')}
										layout="inline"
										value={showApmBreakdown}
										onChange={onShowApmBreakdownChange}
									/>
								)}
							</div>
							<div className="min-h-0 grow" />
							<DetailedResultsTabs tabs={tabs} />
						</div>
						<div ref={setStickySlot} data-testid="dr-sticky-slot" />
					</div>
					<TabPanels data-testid="dr-tab-content">
						<div
							id="noResultsTab"
							data-active
							className="col-start-1 row-start-1 flex items-center justify-center p-6 text-base opacity-100 transition-opacity duration-150 ease-linear group-not-data-no-results/dr:hidden">
							{i18n.t('results_tab.details.no_results')}
						</div>
						<GroupRanksContext.Provider value={groupRanks}>
							<ShowApmBreakdownContext.Provider value={showApmBreakdown}>{children}</ShowApmBreakdownContext.Provider>
						</GroupRanksContext.Provider>
					</TabPanels>
				</Tabs.Root>
			</DrStickySlotContext.Provider>
		</div>
	);
};

export interface DetailedResultsProps {
	resultsManager: SimResultsManager;
}

export const DetailedResults = ({ resultsManager }: DetailedResultsProps) => {
	const host = useSimHost();
	const sim = host.sim;
	const resultsEmitter = host.resultChannel;

	const settingsSubscribe = subscribeSimSettingsChange(sim);
	const { damage: showDamage, threat: showThreat, healing: showHealing } = useDisplayMetrics(sim);

	const visibleTabs = useMemo(
		() =>
			DETAILED_RESULTS_TABS.filter(tab => {
				if (tab.id === 'damageTab') return showDamage;
				if (tab.id === 'healingTab') return showHealing;
				if (tab.id === 'threatTab' || tab.id === 'damageTakenTab') return showThreat;
				return true;
			}),
		[showDamage, showThreat, showHealing],
	);

	const [activeId, setActiveId] = useState<string>(DEFAULT_DETAILED_RESULTS_TAB);
	const [deathDisabled, setDeathDisabled] = useState(true);
	const [target, setTarget] = useState(ALL_UNITS);
	const [groupRanks, setGroupRanks] = useLocalStorageState(GROUP_RANKS_STORAGE_KEY, parseFlag, () => true);
	const [showApmBreakdown, setShowApmBreakdown] = useLocalStorageState(SHOW_APM_BREAKDOWN_STORAGE_KEY, parseFlag, () => false);
	const hasResults = useSimResult() !== null;

	// `updateResults` is bound to the emitter, not to the filter, so the selection it reads is a ref.
	const targetRef = useRef(target);
	// What the last emit already carried, so the reset below does not queue a second one.
	const emittedTarget = useRef(target);

	const latestRun = useRef<SimRunData | null>(null);
	const currentSimResult = useRef<SimResult | null>(null);
	const latestDeathSeeds = useRef<Array<bigint>>([]);
	const recentlyEditedSeed = useRef(false);

	const updateResults = useCallback(
		async (simRunData: SimRunData | null) => {
			if (simRunData?.run?.request?.requestId !== latestRun.current?.run?.request?.requestId) {
				latestRun.current = simRunData;
				currentSimResult.current = await SimResult.fromProto(simRunData?.run || SimRun.create());
			}

			const playerMetrics = latestRun.current?.run?.result?.raidMetrics?.parties.map(party => party.players).flat();

			if (isDevMode() && playerMetrics) {
				console.log('Found player metrics:');
				console.log(playerMetrics);
			}

			if (playerMetrics?.length) {
				const deathSeeds = playerMetrics[0].deathSeeds;

				if (isDevMode() && !!deathSeeds.length) {
					console.log('Found death seeds:');
					console.log(deathSeeds);
				}

				if (deathSeeds.length > 1 || latestDeathSeeds.current.length == 0) {
					latestDeathSeeds.current = deathSeeds;
				}
			}

			if (currentSimResult.current == null) {
				resultsEmitter.emit(null);
			} else {
				// A run with fewer targets than the last one drops the selection, and it happens here
				// rather than in the filter: the emit below is what every metrics table reads, and a
				// component-level effect would let one render through pointing at a target this
				// result does not have.
				if (!hasTarget(currentSimResult.current, targetRef.current)) {
					targetRef.current = ALL_UNITS;
					emittedTarget.current = ALL_UNITS;
					setTarget(ALL_UNITS);
				}
				resultsEmitter.emit({ result: currentSimResult.current, filter: simResultFilter(targetRef.current) });
			}
		},
		[resultsEmitter],
	);

	// The seed the "Simulate a Death" button pinned is released by the next settings change or run.
	const releaseEditedSeed = useCallback(() => {
		if (!recentlyEditedSeed.current) return;
		sim.setFixedRngSeed(0);
		recentlyEditedSeed.current = false;
	}, [sim]);

	useEffect(() => settingsSubscribe(releaseEditedSeed), [settingsSubscribe, releaseEditedSeed]);

	useEffect(
		() =>
			resultsManager.currentChangeEmitter.on(async () => {
				const runData = resultsManager.getRunData();
				if (runData) {
					releaseEditedSeed();
					await updateResults(runData);
				}
				setDeathDisabled(latestDeathSeeds.current.length < 2);
			}),
		[resultsManager, releaseEditedSeed, updateResults],
	);

	// Picking a target re-emits the last run under the new filter. Guarded on the value rather than on a mount flag, so a change
	// of `updateResults` cannot fire it on its own.
	useEffect(() => {
		if (emittedTarget.current === target) return;
		emittedTarget.current = target;
		void updateResults(latestRun.current);
	}, [target, updateResults]);

	useEffect(() => {
		if (visibleTabs.some(tab => tab.id === activeId)) return;
		setActiveId(visibleTabs[0]?.id ?? DEFAULT_DETAILED_RESULTS_TAB);
	}, [visibleTabs, activeId]);

	const onSimulateDeath = () => {
		trackEvent({ action: 'sim', category: 'simulate', label: 'death' });

		if (latestDeathSeeds.current.length > 1) {
			sim.setFixedRngSeed(Number(latestDeathSeeds.current.pop()));
			recentlyEditedSeed.current = true;

			if (isDevMode()) {
				console.log('Setting fixed seed:');
				console.log(sim.getFixedRngSeed());
			}
		}

		void host.runSingleIteration();
	};

	return (
		<DetailedResultsView
			tabs={visibleTabs}
			activeId={activeId}
			onActiveIdChange={setActiveId}
			hasResults={hasResults}
			iterationDisabled={host.disabled}
			deathDisabled={deathDisabled}
			onSimulateIteration={() => {
				trackEvent({ action: 'sim', category: 'simulate', label: 'once' });
				void host.runSingleIteration();
			}}
			onSimulateDeath={onSimulateDeath}
			targetPicker={
				<ResultsFilter
					target={target}
					onTargetChange={next => {
						targetRef.current = next;
						setTarget(next);
					}}
				/>
			}
			groupRanks={groupRanks}
			onGroupRanksChange={setGroupRanks}
			showApmBreakdown={showApmBreakdown}
			onShowApmBreakdownChange={setShowApmBreakdown}>
			{showDamage && (
				<DetailedResultsPane id="damageTab" className="[&_.ui-metrics-table]:text-xs" topline histogram>
					<DamageMetricsTable />
				</DetailedResultsPane>
			)}
			{showThreat && (
				<DetailedResultsPane id="threatTab" className="[&_.ui-metrics-table]:text-xs" topline>
					<ThreatMetricsTable />
				</DetailedResultsPane>
			)}
			{showHealing && (
				<DetailedResultsPane id="healingTab" className="[&_.ui-metrics-table]:text-xs" topline>
					<HealingMetricsTable />
				</DetailedResultsPane>
			)}
			{showThreat && (
				<DetailedResultsPane id="damageTakenTab" topline>
					<DtpsMetricsTable />
				</DetailedResultsPane>
			)}
			<DetailedResultsPane id="buffsTab">
				<AuraMetricsTable useDebuffs={false} />
			</DetailedResultsPane>
			<DetailedResultsPane id="debuffsTab">
				<AuraMetricsTable useDebuffs={true} />
			</DetailedResultsPane>
			<DetailedResultsPane id="castsTab">
				<CastMetricsTable />
			</DetailedResultsPane>
			<DetailedResultsPane id="resourcesTab">
				<ResourceMetricsTable />
			</DetailedResultsPane>
			<DetailedResultsPane id="timelineTab" filling>
				<Timeline active={activeId === 'timelineTab'} />
			</DetailedResultsPane>
			<DetailedResultsPane id="logTab">
				<LogRunner active={activeId === 'logTab'} />
			</DetailedResultsPane>
		</DetailedResultsView>
	);
};
