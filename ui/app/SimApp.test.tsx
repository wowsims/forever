import { LaunchStatus, Phase } from '@sim/constants/other';
import { createSimStore } from '@sim/state/sim_store';
import { memoryEnv } from '@testing/sim/memory_env';
import { act, render } from '@testing-library/react';
import type { ReactElement, ReactNode } from 'react';
import { StrictMode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

// The shell is stubbed on purpose. What is under test is the construct-once gate, not the shell —
// and constructing the real one would need a Database and a worker.
const constructions: Array<{ root: HTMLElement }> = [];
// Read off `dom.root` inside the constructor, which runs before the render that hands the shell its
// host — the only way to prove the header and `main` it names are the ones still in the document later.
const constructedHeaders: Array<Element | null> = [];
const constructedMains: Array<Element | null> = [];
vi.mock('./individual_sim_ui', async () => {
	const { SidebarRegistry } = await import('@app/sidebar_registry');
	const { CrashReportOpener } = await import('./crash_report_opener');
	return {
		SimHostObject: class {
			// `NoticeNativeSim` asks the host's sim whether this is a local build; a native one raises no notice.
			readonly sim = { waitForInit: () => Promise.resolve(), isNative: true };
			readonly sidebar = new SidebarRegistry();
			readonly disabled = false;
			readonly simDisabled = false;
			readonly individualConfig = { displayStats: [], epReferenceStat: 0 };
			// The sidebar panel is React now; the shell only owns the store it drives and the registry
			// the warnings zone reads.
			readonly resultsPanel = {} as never;
			readonly warnings = {} as never;
			readonly raidSimResultsManager = {} as never;
			readonly crashReport = new CrashReportOpener();
			constructor(dom: { root: HTMLElement }) {
				constructions.push(dom);
				constructedHeaders.push(dom.root.querySelector('[data-testid="sim-header"]'));
				constructedMains.push(dom.root.querySelector('[data-testid="sim-main"]'));
			}
		},
	};
});

// The six real panes need a Database and a worker; what is under test is that a pane renders as an
// ordinary child of the shell's main column, so the stub wraps one fake tab through the real `SimTabs`
// — mirroring the real delegation on `host`, so the host's arrival is the same defs-appear transition.
vi.mock('./SimTabsSection', async () => {
	const { createElement } = await import('react');
	const { SimTabDef } = await import('./SimTabDef');
	const { SimTabs } = await import('./SimTabs');
	return {
		SimTabsSection: ({ host, children }: { host: unknown; children: ReactNode }) =>
			createElement(SimTabs, {
				defs: host
					? createElement(SimTabDef, {
							id: 'gear-tab',
							title: 'Gear',
							children: createElement('div', { className: 'gear-tab-left' }),
						})
					: null,
				children,
			}),
	};
});
// Every import and export dialog reads the item database.
vi.mock('./SimImportExport', () => ({ SimImportExport: () => <div className="import-export-menus" /> }));
// The tour reads the results panel, which the stub host does not have.
vi.mock('./SimOnboarding', () => ({ SimOnboarding: () => null }));
// Both reach through the host for the run facade; SidebarActions.test.tsx asserts the sidebar's gate.
vi.mock('@features/results/components/SimulateAction', () => ({ SimulateAction: () => <button className="dps-action" /> }));
vi.mock('@features/stat-weights/components/StatWeightsAction', () => ({ StatWeightsAction: () => <button className="ep-weights-action" /> }));

// The real one needs a Player with a live store; what is under test here is the portal, not it.
vi.mock('@features/character-stats', () => ({ CharacterStats: () => <div className="character-stats-root" /> }));
vi.mock('@features/results/components/SimResultsPanel', () => ({
	SimResultsPanel: () => <div data-testid="results-viewer" />,
	SimWarnings: () => <div data-testid="warning-zone" />,
}));
vi.mock('@features/stat-weights/components/EpWeightsDialog', () => ({ EpWeightsDialog: () => <div className="ep-weights-dialog-root" /> }));
// Six pickers over a real Sim and a Base UI portal into `host.rootElem`; SettingsDialog.test.tsx is where those are asserted.
vi.mock('./SettingsDialog', () => ({ SettingsDialog: () => <div className="settings-menu-root" /> }));
// Needs the real spec registry to list every class; what is under test here is the shell's gate.
vi.mock('./header/SimTitleDropdown', () => ({ SimTitleDropdown: () => <div className="sim-title-dropdown-root" /> }));

// The toolbar asks a local sim host whether it is outdated, and happy-dom's hostname is localhost,
// so it takes that branch. Left in flight, the request is aborted at teardown and the rejection is
// printed; refused outright it takes the toolbar's own `.catch(noop)`, which is the real path.
vi.stubGlobal('fetch', () => Promise.reject(new Error('no sim host')));

const { SimApp } = await import('./SimApp');

// The shell reads the metric toggles off a real store — `subscribeUiField` selects from it — and the
// spec's shape decides the `sim-type--*` class, so both have to be genuine rather than empty casts.
const sim = {
	store: createSimStore(),
	env: memoryEnv(),
	// The tab bodies gate their own content on this.
	waitForInit: () => Promise.resolve(),
	getShowDamageMetrics: () => true,
	getShowThreatMetrics: () => false,
	getShowHealingMetrics: () => false,
	getShowExperimental: () => false,
	// The shell renders the iterations picker itself, so this stub is its source too, not just the host's.
	getIterations: () => 3000,
};
// `launch` is real because the shell derives the known-issues list from it — a launched spec earns
// no status notice, so the toolbar's link ships hidden.
const spec = {
	isHealingSpec: false,
	isTankSpec: false,
	isMeleeDpsSpec: true,
	isRangedDpsSpec: false,
	launch: { phase: Phase.Phase1, status: LaunchStatus.Launched },
	simLink: '/forever/warrior/dps/',
};
const player = { sim, getPlayerSpec: () => spec } as never;
const def = { className: 'arms-warrior-sim-ui', encounterPicker: { showExecuteProportion: true } } as never;

const mount = async (element: ReactElement) => {
	const rendered = render(element);
	await act(async () => {});
	return rendered;
};

describe('SimApp', () => {
	beforeEach(() => {
		constructions.length = 0;
		constructedHeaders.length = 0;
		constructedMains.length = 0;
	});

	it('constructs the shell once', async () => {
		const { container } = await mount(<SimApp player={player} def={def} />);
		expect(constructions).toHaveLength(1);
		expect(container.querySelectorAll('[data-testid="sim-ui"]')).toHaveLength(1);
	});

	it('still constructs once under StrictMode, whose effects run twice', async () => {
		// This is the whole reason the gate exists. Constructing the shell subscribes autosave with no
		// unsubscribe and queues work on sim.waitForInit(), so a second construction cannot be undone
		// by a cleanup function — it has to not happen.
		const { container } = await mount(
			<StrictMode>
				<SimApp player={player} def={def} />
			</StrictMode>,
		);
		expect(constructions).toHaveLength(1);
		expect(container.querySelectorAll('[data-testid="sim-ui"]')).toHaveLength(1);
	});

	it.each([
		['the sidebar stats', '.character-stats-root'],
		['the results panel', '[data-testid="results-viewer"]'],
	])('renders %s', async (_name, selector) => {
		const { container } = await mount(<SimApp player={player} def={def} />);
		expect(container.querySelectorAll(selector)).toHaveLength(1);
	});

	it('renders each registered pane inside the shell’s main column', async () => {
		const { container } = await mount(<SimApp player={player} def={def} />);
		expect(container.querySelectorAll('[data-testid="sim-main"] .gear-tab-left')).toHaveLength(1);
	});

	it('renders it exactly once under StrictMode', async () => {
		await mount(
			<StrictMode>
				<SimApp player={player} def={def} />
			</StrictMode>,
		);
		expect(document.querySelectorAll('.character-stats-root')).toHaveLength(1);
	});

	// The failure this guards against is silent: React re-renders when `simUI` is set, and if the
	// skeleton were recreated in that second render, every element the shell imperatively filled
	// during construction would be discarded with the old nodes.
	it('keeps the same skeleton nodes when the constructed shell arrives', async () => {
		const { container } = await mount(<SimApp player={player} def={def} />);
		const dom = constructions[0];
		const sidebarActions = dom.root.querySelector('[data-testid="sim-sidebar"]')!;
		const marker = document.createElement('span');
		marker.className = 'built-imperatively';
		sidebarActions.appendChild(marker);

		// `setSimUI` has already re-rendered by now; the nodes must have survived it.
		expect(container.querySelector('[data-testid="sim-ui"]')).toBe(dom.root);
		expect(sidebarActions.isConnected).toBe(true);
		expect(sidebarActions.querySelector('.built-imperatively')).toBe(marker);
	});

	it('keeps the header and main mounted when the host arrives and the tabs get their defs', async () => {
		const { container } = await mount(<SimApp player={player} def={def} />);
		expect(constructedHeaders[0]).not.toBeNull();
		expect(constructedMains[0]).not.toBeNull();
		expect(container.querySelector('[data-testid="sim-header"]')).toBe(constructedHeaders[0]);
		expect(container.querySelector('[data-testid="sim-main"]')).toBe(constructedMains[0]);
	});
});
