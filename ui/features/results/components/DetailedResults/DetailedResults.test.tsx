import { Emitter } from '@sim/state/events';
import { createSimStore } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { TabPanel } from '@ui-kit/TabNav';
import { act } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ResultChannel } from '../../model/result_channel';
import type { SimResultData } from '../../model/result_data';
import { useDrStickySlot } from './context/DrStickySlotContext';
import type { DetailedResultsViewProps } from './DetailedResults';
import type { DetailedResultsTabConfig } from './utils';

// The picker is `ResultsFilter`'s test to cover; what this pane owns is the selection, so the filter is
// reduced to a button that reports one. Everything else in the module — `ALL_UNITS`, `hasTarget`,
// `simResultFilter` — is the real thing.
vi.mock('../ResultsFilter', async importOriginal => ({
	...(await importOriginal<typeof import('../ResultsFilter')>()),
	ResultsFilter: ({ target, onTargetChange }: { target: number; onTargetChange: (target: number) => void }) => (
		<button type="button" data-testid="results-filter-root" data-target={target} onClick={() => onTargetChange(1)} />
	),
}));
// `updateResults` builds one of these per run; the only thing read off it here is the target list
// the filter is validated against.
const run = vi.hoisted(() => ({ targets: 3 }));
vi.mock('@sim/proto/sim_result', async importOriginal => ({
	...(await importOriginal<typeof import('@sim/proto/sim_result')>()),
	SimResult: { fromProto: () => Promise.resolve({ getTargets: () => Array.from({ length: run.targets }, (_, index) => ({ index })) }) },
}));

// Every pane below is React now, so they are prop readers rather than islands: what this pane owes
// each is `active`, which is what their own deferral is built on.
const panes = vi.hoisted(() => ({ log: [] as Array<boolean>, timeline: [] as Array<boolean> }));
vi.mock('../LogRunner', () => ({
	LogRunner: ({ active }: { active: boolean }) => {
		panes.log.push(active);
		return <div className="log-runner-root" />;
	},
}));
vi.mock('../Timeline', () => ({
	Timeline: ({ active }: { active: boolean }) => {
		panes.timeline.push(active);
		return <div className="timeline-root" />;
	},
}));

vi.mock('../ToplineResults', () => ({ ToplineResults: () => <div data-testid="topline-results-root" /> }));
vi.mock('../DpsHistogram', () => ({ DpsHistogram: () => <div data-testid="dps-histogram-root" /> }));
vi.mock('../DamageMetricsTable', () => ({ DamageMetricsTable: () => <div data-testid="damage-metrics-root" /> }));
vi.mock('../HealingMetricsTable', () => ({ HealingMetricsTable: () => <div data-testid="healing-metrics-root" /> }));
vi.mock('../DtpsMetricsTable', () => ({ DtpsMetricsTable: () => <div data-testid="dtps-metrics-root" /> }));
vi.mock('../ThreatMetricsTable', () => ({ ThreatMetricsTable: () => <div data-testid="threat-metrics-root" /> }));
vi.mock('../CastMetricsTable', () => ({ CastMetricsTable: () => <div data-testid="cast-metrics-root" /> }));
vi.mock('../ResourceMetricsTable', () => ({ ResourceMetricsTable: () => <div data-testid="resource-metrics-root" /> }));
vi.mock('../AuraMetricsTable', () => ({
	AuraMetricsTable: ({ useDebuffs }: { useDebuffs: boolean }) => <div data-testid={useDebuffs ? 'debuff-metrics-root' : 'buff-metrics-root'} />,
}));

const { DetailedResults, DetailedResultsView } = await import('./DetailedResults');

const metrics = { damage: true, threat: false, healing: false, experimental: false };
let resultChannel: ResultChannel;
let currentChangeEmitter: Emitter<void>;

let runData: unknown = null;

const renderPane = () =>
	renderWithHost(<DetailedResults resultsManager={{ currentChangeEmitter, getRunData: () => runData } as never} />, {
		sim: {
			store: createSimStore(),
			getShowDamageMetrics: () => metrics.damage,
			getShowThreatMetrics: () => metrics.threat,
			getShowHealingMetrics: () => metrics.healing,
			getShowExperimental: () => metrics.experimental,
			getFixedRngSeed: () => 0,
			setFixedRngSeed: () => {},
		},
		resultChannel,
		disabled: false,
		runSingleIteration: () => Promise.resolve(undefined),
	});

const filterButton = (container: HTMLElement) => container.querySelector<HTMLButtonElement>('[data-testid="results-filter-root"]')!;

const tabButton = (container: HTMLElement, tabId: string) =>
	container.querySelector<HTMLButtonElement>(`[data-testid="dr-toolbar"] [role="tab"][aria-controls=${tabId}]`)!;

beforeEach(() => {
	panes.log.length = 0;
	panes.timeline.length = 0;
	resultChannel = new ResultChannel();
	currentChangeEmitter = new Emitter<void>();
	runData = null;
	run.targets = 3;
	metrics.damage = true;
	metrics.threat = false;
	metrics.healing = false;
	metrics.experimental = false;
});

describe('DetailedResults', () => {
	it('renders only the tabs whose metrics toggle is on, opens the damage pane and fades the rest out, and keeps each pane in the container its island was built into', () => {
		const { container } = renderPane();
		const buttons = [...container.querySelectorAll<HTMLButtonElement>('[data-testid="dr-toolbar"] [role="tab"]')];
		expect(buttons).toHaveLength(7);
		expect(buttons.map(button => button.getAttribute('aria-controls'))).toEqual([
			'damageTab',
			'buffsTab',
			'debuffsTab',
			'castsTab',
			'resourcesTab',
			'timelineTab',
			'logTab',
		]);
		expect(buttons.filter(button => button.getAttribute('aria-selected') === 'true')).toHaveLength(1);
		expect(buttons[0].getAttribute('aria-selected')).toBe('true');

		expect(container.querySelector<HTMLElement>('#damageTab')!.hidden).toBe(false);
		expect(container.querySelector('#damageTab')!.hasAttribute('data-starting-style')).toBe(false);
		expect(container.querySelector<HTMLElement>('#logTab')!.hidden).toBe(true);
		expect(container.querySelector('#noResultsTab')).toBeTruthy();

		expect(container.querySelectorAll('#logTab .log-runner-root')).toHaveLength(1);
		expect(container.querySelectorAll('#timelineTab .timeline-root')).toHaveLength(1);
	});

	it('holds every metrics table in its own tab, under its own root', () => {
		metrics.threat = true;
		metrics.healing = true;
		const { container } = renderPane();
		expect(container.querySelectorAll('#damageTab [data-testid="dr-row"] > [data-testid="damage-metrics-root"]')).toHaveLength(1);
		expect(container.querySelectorAll('#healingTab [data-testid="dr-row"] > [data-testid="healing-metrics-root"]')).toHaveLength(1);
		expect(container.querySelectorAll('#damageTakenTab [data-testid="dr-row"] > [data-testid="dtps-metrics-root"]')).toHaveLength(1);
		expect(container.querySelectorAll('#threatTab [data-testid="dr-row"] > [data-testid="threat-metrics-root"]')).toHaveLength(1);
		expect(container.querySelectorAll('#castsTab [data-testid="dr-row"] > [data-testid="cast-metrics-root"]')).toHaveLength(1);
		expect(container.querySelectorAll('#buffsTab [data-testid="dr-row"] > [data-testid="buff-metrics-root"]')).toHaveLength(1);
		expect(container.querySelectorAll('#debuffsTab [data-testid="dr-row"] > [data-testid="debuff-metrics-root"]')).toHaveLength(1);
		expect(container.querySelectorAll('#resourcesTab [data-testid="dr-row"] > [data-testid="resource-metrics-root"]')).toHaveLength(1);
		expect(container.querySelectorAll('[data-testid="dr-row-topline"] > [data-testid="topline-results-root"]')).toHaveLength(4);
		expect(container.querySelectorAll('#damageTab [data-testid="dr-row-dps-histogram"] > [data-testid="dps-histogram-root"]')).toHaveLength(1);
	});

	it('drops dr-no-results once a result reaches the channel', () => {
		const { container } = renderPane();
		expect(container.querySelector('[data-testid="dr-root"]')!.hasAttribute('data-no-results')).toBe(true);
		act(() => resultChannel.emit({ result: {}, filter: {} } as SimResultData));
		expect(container.querySelector('[data-testid="dr-root"]')!.hasAttribute('data-no-results')).toBe(false);
	});

	// What `onTabShown` was to an island, `active` is to every pane that defers its work.
	it.each([
		['log', 'logTab'],
		['timeline', 'timelineTab'],
	] as const)('tells the %s pane whether its tab is the open one', (pane, tabId) => {
		const { container } = renderPane();
		expect(panes[pane].at(-1)).toBe(false);
		fireEvent.click(tabButton(container, tabId));
		expect(panes[pane].at(-1)).toBe(true);
		fireEvent.click(tabButton(container, 'damageTab'));
		expect(panes[pane].at(-1)).toBe(false);
	});

	it('renders only the tabs whose toggle is on', () => {
		metrics.threat = true;
		const { container } = renderPane();
		expect(container.querySelector('#damageTakenTab')).toBeTruthy();
		expect(container.querySelector('#threatTab')).toBeTruthy();
		expect(container.querySelector('#healingTab')).toBeNull();
	});

	it('leaves the damage tab for healing when damage metrics are off', async () => {
		metrics.damage = false;
		metrics.healing = true;
		const { container } = renderPane();
		await waitFor(() => expect(container.querySelector<HTMLElement>('#healingTab')!.hidden).toBe(false));
		expect(container.querySelector('#damageTab')).toBeNull();
		expect(tabButton(container, 'healingTab').getAttribute('aria-selected')).toBe('true');
	});

	it('re-emits the last run under a newly picked target', async () => {
		const emitted: Array<SimResultData | null> = [];
		resultChannel.on(value => emitted.push(value));
		runData = { run: { request: { requestId: 'run-1' } } };
		const { container } = renderPane();

		await act(async () => currentChangeEmitter.emit());
		expect(emitted).toHaveLength(1);
		expect(emitted[0]!.filter).toEqual({ target: null });

		await act(async () => {
			fireEvent.click(filterButton(container));
		});
		expect(emitted).toHaveLength(2);
		expect(emitted[1]!.filter).toEqual({ target: 1 });
		expect(filterButton(container).dataset.target).toBe('1');
	});

	it('emits nothing on its own before a target is picked', async () => {
		const emitted: Array<SimResultData | null> = [];
		resultChannel.on(value => emitted.push(value));
		runData = { run: { request: { requestId: 'run-1' } } };
		renderPane();

		await act(async () => currentChangeEmitter.emit());
		await act(async () => currentChangeEmitter.emit());
		expect(emitted).toHaveLength(2);
	});

	it('keeps a selected target the next run still has, and drops one the next run no longer has before that run is emitted', async () => {
		const emitted: Array<SimResultData | null> = [];
		resultChannel.on(value => emitted.push(value));
		runData = { run: { request: { requestId: 'run-1' } } };
		const { container } = renderPane();

		await act(async () => currentChangeEmitter.emit());
		await act(async () => {
			fireEvent.click(filterButton(container));
		});
		expect(emitted.at(-1)!.filter).toEqual({ target: 1 });

		runData = { run: { request: { requestId: 'run-2' } } };
		await act(async () => currentChangeEmitter.emit());

		expect(emitted.at(-1)!.filter).toEqual({ target: 1 });
		expect(filterButton(container).dataset.target).toBe('1');

		const beforeDrop = emitted.length;
		run.targets = 1;
		runData = { run: { request: { requestId: 'run-3' } } };
		await act(async () => currentChangeEmitter.emit());

		expect(emitted.at(-1)!.filter).toEqual({ target: null });
		expect(filterButton(container).dataset.target).toBe('-1');
		// The reset rides on the run's own emit rather than queueing a second one.
		expect(emitted).toHaveLength(beforeDrop + 1);
	});

	it('keeps the death button disabled while no run has reported death seeds', async () => {
		const { container } = renderPane();
		const buttons = [...container.querySelectorAll<HTMLButtonElement>('[data-testid="detailed-results-controls-div"] button')];
		expect(buttons.map(button => button.getAttribute('type'))).toEqual(['button', 'button']);
		expect(buttons[0].disabled).toBe(false);
		expect(buttons[1].disabled).toBe(true);
		await act(async () => {
			currentChangeEmitter.emit();
		});
		expect(buttons[1].disabled).toBe(true);
	});
});

const TABS: Array<DetailedResultsTabConfig> = [
	{ id: 'damageTab', labelKey: 'results_tab.details.tabs.damage', groupsRanks: true },
	{ id: 'castsTab', labelKey: 'results_tab.details.tabs.casts', hidesPassives: true },
	{ id: 'logTab', labelKey: 'results_tab.details.tabs.log' },
];

const SlotReader = () => <span data-testid="slot-reader">{useDrStickySlot()?.getAttribute('data-testid') ?? 'none'}</span>;

const mount = (props: Partial<DetailedResultsViewProps> = {}) => {
	const handlers = {
		onActiveIdChange: vi.fn(),
		onSimulateIteration: vi.fn(),
		onSimulateDeath: vi.fn(),
		onGroupRanksChange: vi.fn(),
		onHidePassivesChange: vi.fn(),
	};
	const view = render(
		<DetailedResultsView
			tabs={TABS}
			activeId="damageTab"
			hasResults
			iterationDisabled={false}
			deathDisabled
			targetPicker={<span data-testid="filter-slot" />}
			groupRanks
			hidePassives
			{...handlers}
			{...props}>
			<TabPanel value="damageTab">damage pane</TabPanel>
			<TabPanel value="castsTab">casts pane</TabPanel>
			<TabPanel value="logTab">
				<SlotReader />
			</TabPanel>
		</DetailedResultsView>,
	);
	return { ...handlers, ...view };
};

const iterationButton = () => screen.getByTestId('detailed-results-1-iteration-button');
const deathButton = () => screen.getByRole('button', { name: 'results_tab.details.sim_1_death' });

describe('DetailedResultsView', () => {
	it('runs one iteration, and holds the death button back while it has no seed to replay', () => {
		const { onSimulateIteration, onSimulateDeath } = mount();

		fireEvent.click(iterationButton());
		fireEvent.click(deathButton());

		expect(onSimulateIteration).toHaveBeenCalledOnce();
		expect((deathButton() as HTMLButtonElement).disabled).toBe(true);
		expect(onSimulateDeath).not.toHaveBeenCalled();
	});

	it('runs a death once one is on offer, and disables the iteration button for a disabled sim', () => {
		const { onSimulateDeath } = mount({ deathDisabled: false, iterationDisabled: true });

		fireEvent.click(deathButton());

		expect(onSimulateDeath).toHaveBeenCalledOnce();
		expect((iterationButton() as HTMLButtonElement).disabled).toBe(true);
	});

	it('marks the root while there are no results, which is what shows the empty pane', () => {
		const empty = mount({ hasResults: false });
		expect(screen.getByTestId('dr-root').hasAttribute('data-no-results')).toBe(true);
		expect(screen.getByText('results_tab.details.no_results')).toBeTruthy();
		empty.unmount();

		mount();
		expect(screen.getByTestId('dr-root').hasAttribute('data-no-results')).toBe(false);
	});

	it('draws one tab per config, reports a pick, and puts the filter in the toolbar row', () => {
		const { onActiveIdChange } = mount();

		const tabs = screen.getAllByRole('tab');
		expect(tabs.map(tab => tab.id)).toEqual(['damageTab-nav', 'castsTab-nav', 'logTab-nav']);
		fireEvent.click(tabs[2]);

		expect(onActiveIdChange).toHaveBeenCalledWith('logTab');
		expect(screen.getByTestId('results-filter').contains(screen.getByTestId('filter-slot'))).toBe(true);
	});

	it('puts the group ranks toggle in the toolbar row, and reports a change', () => {
		const { onGroupRanksChange } = mount();

		const toggle = within(screen.getByTestId('results-filter')).getByRole('checkbox');
		expect(toggle.getAttribute('aria-checked')).toBe('true');
		fireEvent.click(toggle);

		expect(onGroupRanksChange).toHaveBeenCalledWith(false);
	});

	it('shows the group ranks toggle only on a tab whose table groups ranks', () => {
		mount({ activeId: 'logTab' });

		expect(within(screen.getByTestId('results-filter')).queryByRole('checkbox')).toBeNull();
	});

	it('shows the hide passives toggle only on a tab whose table hides passives, and reports a change', () => {
		const name = 'results_tab.details.hide_passives';
		const damage = mount();
		expect(screen.queryByRole('checkbox', { name })).toBeNull();
		damage.unmount();

		const { onHidePassivesChange } = mount({ activeId: 'castsTab' });
		const toggle = within(screen.getByTestId('results-filter')).getByRole('checkbox', { name });
		expect(toggle.getAttribute('aria-checked')).toBe('true');
		fireEvent.click(toggle);

		expect(onHidePassivesChange).toHaveBeenCalledWith(false);
	});

	it('hands the panes the sticky slot under the toolbar', () => {
		mount({ activeId: 'logTab' });

		expect(screen.getByTestId('slot-reader').textContent).toBe('dr-sticky-slot');
	});
});
