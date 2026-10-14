// The store, the worker and the saved-data manager are stubbed — what is under test is the view,
// and every source it reads is driven directly.
import { ErrorOutcomeType, type StatWeightsResult, StatWeightsTargetMetric } from '@generated/proto/api';
import { Class, PseudoStat, Spec, Stat } from '@generated/proto/common';
import { DatabaseFilters } from '@generated/proto/ui';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { Stats, UnitStat } from '@sim/proto/stats';
import type { PrecisionSettings } from '@sim/settings/precision';
import { SimRuns } from '@sim/sim_runs';
import { createSimStore, patchKeyed, patchSlice, PLAYER_FIELDS, seedKeyed, type SimStore, zeroVersions } from '@sim/state/sim_store';
import { act, fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { PortalContainerContext } from '@ui-kit/hooks/usePortalContainer';
import { useState } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { emptyStatWeightsResult } from '../../model/ep_math';
import { EpWeightsDialog } from './EpWeightsDialog';

const source = vi.hoisted(() => {
	const listeners = new Map<string, Set<() => void>>();
	// One identity per key, like the real module: `useStoreSubscribe` resubscribes on every
	// source identity change, so a fake handing out a fresh function per call renders forever.
	const sources = new Map<string, ((onChange: () => void) => () => void) & { key: string }>();
	return {
		listeners,
		subscribe: (key: string) => {
			let existing = sources.get(key);
			if (!existing) {
				existing = Object.assign(
					(onChange: () => void) => {
						const set = listeners.get(key) ?? new Set<() => void>();
						listeners.set(key, set);
						set.add(onChange);
						return () => set.delete(onChange);
					},
					{ key },
				);
				sources.set(key, existing);
			}
			return existing;
		},
		notify: (key: string) => listeners.get(key)?.forEach(listener => listener()),
		count: () => [...listeners.values()].reduce((total, set) => total + set.size, 0),
	};
});

// The run flags live in the real store, so this one delegates instead of using the fake source:
// `useSimRun` reads them back through `useSyncExternalStore` and needs a notification that fires.
const runsStore = vi.hoisted(() => ({ store: null as { subscribe: (onChange: () => void) => () => void } | null }));

vi.mock('@sim/state/subscriptions', () => {
	const combined = new Map<string, (onChange: () => void) => () => void>();
	const runs = (onChange: () => void) => runsStore.store!.subscribe(onChange);
	return {
		// `useDisplayMetrics` folds its three ui fields into one source; the real `subscribeAll` collapses
		// same-store selectors into one, which a mock cannot, so it fans out and unsubscribes each.
		subscribeAll: (subs: Array<((onChange: () => void) => () => void) & { key?: string }>) => {
			const key = subs.map(sub => sub.key ?? '?').join('|');
			let existing = combined.get(key);
			if (!existing) {
				existing = (onChange: () => void) => {
					const unsubs = subs.map(sub => sub(onChange));
					return () => unsubs.forEach(unsub => unsub());
				};
				combined.set(key, existing);
			}
			return existing;
		},
		subscribePlayerField: (_player: unknown, field: string) => source.subscribe(`player:${field}`),
		subscribeUiField: (_sim: unknown, field: string) => source.subscribe(`ui:${field}`),
		subscribeSimField: (_sim: unknown, field: string) => source.subscribe(`sim:${field}`),
		subscribeStatWeightsChange: () => source.subscribe('statWeights'),
		subscribePlayerStatWeightsField: () => source.subscribe('statWeights'),
		subscribeRunState: () => runs,
	};
});

// Real stat names, because the vitest i18n stub answers with the enum key and the id-sanitising
// defect is about the spaces a real name has.
const STAT_NAMES: Partial<Record<Stat, string>> = {
	[Stat.StatStrength]: 'Strength',
	[Stat.StatAgility]: 'Agility',
	[Stat.StatCritRating]: 'Crit Rating',
};
vi.mock('@i18n/localization', () => ({
	translateStat: (stat: Stat) => STAT_NAMES[stat] ?? `Stat ${stat}`,
	translatePseudoStat: (pseudoStat: PseudoStat) => (pseudoStat === PseudoStat.PseudoStatSpellHitPercent ? 'Spell Hit Percent' : `Pseudo ${pseudoStat}`),
}));

vi.mock('../SavedEpWeights', () => ({ SavedEpWeights: () => null }));

// The vitest i18n stub has no resources, so `t` answers with the key and never interpolates — which
// would make the reference-stat half of a column tooltip invisible. This keeps the key and shows the
// interpolation beside it.
vi.mock('@i18n/config', () => ({
	default: { t: (key: string, options?: Record<string, unknown>) => (options?.refStatName ? `${key}[${options.refStatName}]` : key) },
}));

const toasts = vi.hoisted(() => [] as Array<{ variant: string; body: unknown }>);
vi.mock('@ui-kit/Toast', async importOriginal => ({
	...(await importOriginal<typeof import('@ui-kit/Toast')>()),
	toastManager: {
		add: (options: { variant: string; body: unknown }) => {
			toasts.push(options);
			return '';
		},
		close: () => {},
	},
}));
const trackEvent = vi.hoisted(() => vi.fn());
vi.mock('../../../../tracking/utils', () => ({ trackEvent }));

const STRENGTH = UnitStat.fromStat(Stat.StatStrength);
const AGILITY = UnitStat.fromStat(Stat.StatAgility);
const STAMINA = UnitStat.fromStat(Stat.StatStamina);

class FakePlayer {
	storeKey = 0;
	readonly sim: { store: SimStore };
	epRatios = [1, 0, 0, 0, 0, 0];
	readonly refStats: { dpsRefStat?: Stat; healRefStat?: Stat; tankRefStat?: Stat } = {};
	readonly playerClass = { classID: Class.ClassWarrior };
	playerSpec: { isTankSpec: boolean; isHealingSpec: boolean } | undefined = undefined;
	computeStatWeights = vi.fn<(...args: any[]) => Promise<StatWeightsResult>>();

	constructor(store: SimStore) {
		this.sim = { store };
		seedKeyed(store, 'players', this.storeKey, { epWeights: new Stats(), v: zeroVersions(PLAYER_FIELDS) } as never);
	}

	get epWeights(): Stats {
		return this.sim.store.getState().players[this.storeKey].epWeights;
	}
	set epWeights(next: Stats) {
		patchKeyed(this.sim.store, 'players', this.storeKey, { epWeights: next }, ['epWeights']);
	}

	getClass() {
		return Class.ClassWarrior;
	}
	getSpec() {
		return Spec.SpecUnknown;
	}
	getEpWeights() {
		return this.epWeights;
	}
	setEpWeights(weights: Stats) {
		this.epWeights = weights;
	}
	getEpRatios() {
		return this.epRatios.slice();
	}
	setEpRatios(ratios: number[]) {
		this.epRatios = ratios;
		source.notify('player:epRatios');
	}
	getRefStat(kind: 'dpsRefStat' | 'healRefStat' | 'tankRefStat') {
		return this.refStats[kind];
	}
	setRefStat(kind: 'dpsRefStat' | 'healRefStat' | 'tankRefStat', stat: Stat | undefined) {
		this.refStats[kind] = stat;
		source.notify('player:epRefStat');
	}
}

class FakeSettings {
	excludedStats: Stat[] = [];
	excludedPseudoStats: PseudoStat[] = [];
	precision: PrecisionSettings = { precisionMode: 'iterations', iterations: 0, targetCi: 0.01, maxIterationMultiplier: 0 };

	setPrecision(patch: Partial<PrecisionSettings>) {
		this.precision = { ...this.precision, ...patch };
		source.notify('statWeights');
	}
	isStatExcludedFromCalc(stat: Stat) {
		return this.excludedStats.includes(stat);
	}
	isPseudoStatExcludedFromCalc(pseudoStat: PseudoStat) {
		return this.excludedPseudoStats.includes(pseudoStat);
	}
	isUnitStatExcludedFromCalc(stat: UnitStat) {
		return stat.isStat() ? this.isStatExcludedFromCalc(stat.getStat()) : this.isPseudoStatExcludedFromCalc(stat.getPseudoStat());
	}
	setStatExcluded(stat: UnitStat, exclude: boolean) {
		const list = stat.isStat() ? this.excludedStats : this.excludedPseudoStats;
		const value = stat.isStat() ? stat.getStat() : stat.getPseudoStat();
		const at = list.indexOf(value as never);
		if (exclude && at === -1) list.push(value as never);
		if (!exclude && at !== -1) list.splice(at, 1);
		source.notify('statWeights');
	}
}

const rootElem = document.createElement('div');
rootElem.className = 'sim-ui';

let player: FakePlayer;
let settings: FakeSettings;
let host: any;
let abortType: ReturnType<typeof vi.fn>;

// The server returns `epValues` already normalised by `epReferenceStat`; `calculateEp` only
// re-normalises when the user has picked a different reference stat.
const weightsResult = (values: Array<[UnitStat, number]>): StatWeightsResult => {
	const result = emptyStatWeightsResult();
	const reference = values.find(([stat]) => stat.equalsStat(Stat.StatStrength))?.[1] ?? 1;
	for (const [stat, value] of values) {
		stat.setProtoValue(result.dps!.weights!, value);
		stat.setProtoValue(result.dps!.epValues!, value / reference);
	}
	return result;
};

const setup = () => {
	source.listeners.clear();
	toasts.length = 0;
	const store = createSimStore();
	player = new FakePlayer(store);
	settings = new FakeSettings();
	abortType = vi.fn().mockResolvedValue(undefined);
	runsStore.store = store;
	const runs = new SimRuns(store, { abortType: abortType as unknown as (mask: number) => Promise<void> });
	host = {
		rootElem,
		player,
		getStorageKey: (keyPart: string) => `test${keyPart}`,
		sim: {
			showThreatMetrics: false,
			getShowThreatMetrics() {
				return this.showThreatMetrics;
			},
			// `useDisplayMetrics` reads all three through one subscription; only threat varies here.
			getShowDamageMetrics: () => true,
			getShowHealingMetrics: () => true,
			getIterations: () => 1000,
			getFilters: () => DatabaseFilters.clone(store.getState().sim.filters),
			setFilters: (filters: DatabaseFilters) => patchSlice(store, 'sim', { filters: DatabaseFilters.clone(filters) }),
			store,
			runs,
			signalManager: { abortType },
		},
		individualConfig: {
			epStats: [Stat.StatStrength, Stat.StatAgility, Stat.StatCritRating],
			epPseudoStats: [PseudoStat.PseudoStatSpellHitPercent],
			epReferenceStat: Stat.StatStrength,
			defaults: { epWeights: new Stats().withStat(Stat.StatAgility, 7) },
		},
	};
};

// `SimApp` owns the open state; the harness stands in for it, and records what the dialog asks for.
const openChanges: boolean[] = [];
const DialogHarness = () => {
	const [open, setOpen] = useState(true);
	return (
		<>
			<button type="button" data-testid="toggle-open" onClick={() => setOpen(!open)} />
			<EpWeightsDialog
				open={open}
				onOpenChange={next => {
					openChanges.push(next);
					setOpen(next);
				}}
				settings={settings as never}
			/>
		</>
	);
};

const renderDialog = () => {
	document.body.appendChild(rootElem);
	openChanges.length = 0;
	return render(
		<PortalContainerContext value={rootElem}>
			<SimHostProvider host={host}>
				<DialogHarness />
			</SimHostProvider>
		</PortalContainerContext>,
	);
};

const popup = () => rootElem.querySelector('[data-testid="ep-weights-menu"]')!;
const table = () => popup().querySelector('[data-testid="results-ep-table"]')!;
const pickOption = async (id: string, index: number) => {
	await userEvent.click(popup().querySelector<HTMLElement>(`#${id}`)!);
	await userEvent.click(screen.getAllByRole('option')[index]);
};
const rowFor = (name: string) => [...table().querySelectorAll('tbody tr')].find(row => row.firstElementChild?.textContent === name)!;
const calculate = () => popup().querySelector<HTMLButtonElement>('[data-testid="calc-weights"]')!;
const statFilterTrigger = () => popup().querySelector<HTMLElement>('[data-testid="ep-stat-filter-button"]')!;
const openStatFilter = () =>
	act(async () => {
		fireEvent.click(statFilterTrigger());
	});
const statFilterBox = (stat: UnitStat) => rootElem.querySelector<HTMLElement>(`#ep-stat-filter-${stat.getKey()}`)!;
const precisionText = (key: string) => `common.precision.${key}`;

describe('EpWeightsDialog', () => {
	beforeEach(setup);

	it('builds the column set in order, each metric column tagged with the type its hide rule selects on', () => {
		host.sim.showThreatMetrics = true;
		renderDialog();
		const headers = [...table().querySelectorAll('thead tr:first-child th')];
		expect(headers.map(th => th.getAttribute('data-column-type'))).toEqual([
			null,
			null,
			'weight',
			'ep',
			'weight',
			'ep',
			'weight',
			'ep',
			'weight',
			'ep',
			'weight',
			'ep',
			'weight',
			'ep',
			null,
		]);
		expect(headers.map(th => th.textContent)).toEqual([
			'sidebar.buttons.stat_weights.modal.column_headers.stat',
			'sidebar.buttons.stat_weights.modal.column_headers.update',
			'sidebar.buttons.stat_weights.modal.dps_weight.label',
			'sidebar.buttons.stat_weights.modal.dps_ep.label',
			'sidebar.buttons.stat_weights.modal.hps_weight.label',
			'sidebar.buttons.stat_weights.modal.hps_ep.label',
			'sidebar.buttons.stat_weights.modal.tps_weight.label',
			'sidebar.buttons.stat_weights.modal.tps_ep.label',
			'sidebar.buttons.stat_weights.modal.dtps_weight.label',
			'sidebar.buttons.stat_weights.modal.dtps_ep.label',
			'sidebar.buttons.stat_weights.modal.tmi_weight.label',
			'sidebar.buttons.stat_weights.modal.tmi_ep.label',
			'sidebar.buttons.stat_weights.modal.death_weight.label',
			'sidebar.buttons.stat_weights.modal.death_ep.label',
			'sidebar.buttons.stat_weights.modal.current_ep.label',
		]);
	});

	// A gear planner never computes weights: no Calculate button, no computed columns, no Update
	// column, no reference picker. What is left is the editable Current EP column the item pickers sort by.
	it('gives a gear planner only the editable Current EP column', () => {
		host.simDisabled = true;
		host.sim.showThreatMetrics = true;
		renderDialog();

		expect(popup().querySelector('[data-testid="calc-weights"]')).toBeNull();
		expect(popup().querySelector('[data-testid="ep-reference-options"]')).toBeNull();
		const headers = [...table().querySelectorAll('thead tr:first-child th')];
		expect(headers.map(th => th.textContent)).toEqual([
			'sidebar.buttons.stat_weights.modal.column_headers.stat',
			'sidebar.buttons.stat_weights.modal.current_ep.label',
		]);
		expect(table().querySelectorAll('thead tr').length).toBe(1);
		expect(rowFor('Agility').querySelector('[data-testid="swcalc-include-toggle"]')).toBeNull();
		expect(rowFor('Agility').querySelector('[data-testid="current-ep"] input')).not.toBeNull();
	});

	// Vanilla's `.sim-type--tank .ep-weights-menu { display: none }` block is commented out, so a tank
	// gets the whole dialog there. Re-implementing that dead rule in TSX left the three tank specs
	// with no Calculate button at all and no way to compute weights.
	it('gives a tank the same dialog as everyone else', () => {
		player.playerSpec = { isTankSpec: true, isHealingSpec: false };
		host.sim.showThreatMetrics = true;
		renderDialog();

		expect(popup().querySelector('[data-testid="calc-weights"]')).not.toBeNull();
		expect(popup().querySelector('[data-testid="ep-reference-options"]')).not.toBeNull();
		expect(table().querySelectorAll('thead tr')[0].querySelectorAll('th').length).toBeGreaterThan(2);
	});

	it('shows the spec stats, and a stat ticked in the stat filter without an include toggle', async () => {
		renderDialog();
		const names = () => [...table().querySelectorAll('tbody tr')].map(row => row.firstElementChild!.textContent);
		expect(names()).toEqual(['Strength', 'Agility', 'Crit Rating', 'Spell Hit Percent']);

		await openStatFilter();
		act(() => {
			fireEvent.click(statFilterBox(STAMINA));
		});

		const unweighted = STAMINA.getFullName(Class.ClassWarrior);
		expect(names()).toEqual(['Strength', 'Agility', unweighted, 'Crit Rating', 'Spell Hit Percent']);
		expect(rowFor(unweighted).querySelector('[data-testid="swcalc-include-toggle"] input')).toBeNull();
		expect(rowFor('Strength').querySelector('[data-testid="swcalc-include-toggle"] input')).not.toBeNull();
	});

	it('drops the row of a spec stat hidden in the stat filter, and still weights that stat', async () => {
		player.computeStatWeights.mockResolvedValue(weightsResult([[STRENGTH, 2]]));
		renderDialog();

		await openStatFilter();
		act(() => {
			fireEvent.click(statFilterBox(AGILITY));
		});
		expect(rowFor('Agility')).toBeUndefined();

		await act(async () => {
			fireEvent.click(calculate());
		});
		expect(player.computeStatWeights.mock.calls[0][0]).toEqual([Stat.StatStrength, Stat.StatAgility, Stat.StatCritRating]);
	});

	it('switches the table between EP and weight columns, and keeps that and the stat filter in step across a close and reopen', async () => {
		const { getByTestId } = renderDialog();
		const select = () => popup().querySelector<HTMLElement>('#ep-type-select')!;
		expect(table().getAttribute('data-stats-type')).toBe('ep');
		await pickOption('ep-type-select', 1);
		expect(table().getAttribute('data-stats-type')).toBe('weight');
		await openStatFilter();
		act(() => {
			fireEvent.click(statFilterBox(STAMINA));
		});

		await act(async () => {
			fireEvent.click(getByTestId('toggle-open'));
		});
		expect(rootElem.querySelector('[data-testid="ep-weights-menu"]')).toBeNull();
		await act(async () => {
			fireEvent.click(getByTestId('toggle-open'));
		});

		expect(table().getAttribute('data-stats-type')).toBe('weight');
		expect(select().textContent).toBe('sidebar.buttons.stat_weights.modal.weights');
		expect(rowFor(STAMINA.getFullName(Class.ClassWarrior))).toBeTruthy();
	});

	// DEFECT FIXED. `makeEpRatioCell` was applied to the six EP cells and the six weight cells, both
	// indexed 0-5, so `ep-ratio-0`…`ep-ratio-5` each existed twice.
	it('gives every EP-ratio picker a unique id, and writes each ratio at its own index', () => {
		host.sim.showThreatMetrics = true;
		renderDialog();
		const ids = [...table().querySelectorAll('[data-testid="ep-ratios"] input')].map(input => input.id);
		expect(ids).toHaveLength(12);
		expect(new Set(ids).size).toBe(12);
		expect(ids.slice(0, 2)).toEqual(['ep-ratio-weight-0', 'ep-ratio-ep-0']);

		const input = table().querySelector<HTMLInputElement>('#ep-ratio-ep-2')!;
		act(() => {
			input.value = '0.5';
			fireEvent.change(input);
		});
		expect(player.epRatios).toEqual([1, 0, 0.5, 0, 0, 0]);
	});

	// DEFECT FIXED. The id was the untranslated full name, so it carried ASCII spaces — the line
	// below it in the same function already called `sanitizeId`.
	it('sanitises the include-toggle id, and disables the toggle on the reference stat', () => {
		renderDialog();
		expect(rowFor('Spell Hit Percent').querySelector('[data-testid="swcalc-include-toggle"] input')!.id).toBe('sw-stat-toggle-spellhitpercent');
		expect(rowFor('Crit Rating').querySelector('[data-testid="current-ep"] input')!.id).toBe('ep-weight-stat-crit');

		expect(rowFor('Strength').querySelector<HTMLInputElement>('[data-testid="swcalc-include-toggle"] input')!.disabled).toBe(true);
		expect(rowFor('Agility').querySelector<HTMLInputElement>('[data-testid="swcalc-include-toggle"] input')!.disabled).toBe(false);
	});

	// DEFECT FIXED. Every `<button>` in the panel was written without one, so any inside a form would
	// submit it.
	it('types every button it renders', () => {
		host.sim.showThreatMetrics = true;
		renderDialog();
		const buttons = [...rootElem.querySelectorAll('button')];
		expect(buttons.length).toBeGreaterThan(14);
		expect(buttons.filter(button => !button.getAttribute('type'))).toEqual([]);
	});

	// DEFECT FIXED. The three reference selects were named by an adjacent `<span>`, which is not an
	// accessible name at all.
	it('labels each reference select', () => {
		host.sim.showThreatMetrics = true;
		renderDialog();
		for (const id of ['ep-ref-stat-damage', 'ep-ref-stat-healing', 'ep-ref-stat-threat']) {
			const label = popup().querySelector<HTMLLabelElement>(`label[for="${id}"]`)!;
			expect(label).not.toBeNull();
			expect(label.textContent).not.toBe('');
			expect(popup().querySelector(`#${id}`)!.getAttribute('role')).toBe('combobox');
		}
	});

	// DEFECT FIXED. The `<option>`s carried no `value`, so the selection round-tripped through the
	// translated display string.
	it('carries the stat enum on each reference option, writes it, and follows it in the column tooltip', async () => {
		renderDialog();
		// The options run Strength, Agility, Crit Rating.
		await userEvent.click(popup().querySelector<HTMLElement>('#ep-ref-stat-damage')!);
		expect(screen.getAllByRole('option').map(option => option.getAttribute('aria-selected'))).toEqual(['true', 'false', 'false']);
		await userEvent.keyboard('{Escape}');

		const dpsEpLabel = () => table().querySelectorAll('thead tr:first-child th')[3].querySelector('span')!;
		expect(dpsEpLabel().getAttribute('data-tooltip-content')).toContain('Strength');

		await pickOption('ep-ref-stat-damage', 1);
		expect(player.getRefStat('dpsRefStat')).toBe(Stat.StatAgility);
		expect(dpsEpLabel().getAttribute('data-tooltip-content')).toContain('Agility');
		expect(dpsEpLabel().getAttribute('data-tooltip-content')).not.toContain('Strength');
	});

	it('copies a column into the current EP weights, leaving excluded stats alone', () => {
		host.sim.showThreatMetrics = true;
		host.individualConfig.defaults.epWeights = new Stats().withStat(Stat.StatAgility, 7).withStat(Stat.StatStrength, 11);
		renderDialog();
		act(() => {
			player.setEpWeights(new Stats().withStat(Stat.StatAgility, 3));
			settings.setStatExcluded(AGILITY, true);
		});

		act(() => {
			fireEvent.click(table().querySelectorAll('thead tr:first-child th')[14].querySelector('[data-testid="col-action"]')!);
		});

		expect(player.epWeights.getStat(Stat.StatStrength)).toBe(11);
		expect(player.epWeights.getStat(Stat.StatAgility)).toBe(3);
	});

	it('folds the ratios into the current EP weights on Update EP, per stats type', async () => {
		player.computeStatWeights.mockResolvedValue(
			weightsResult([
				[STRENGTH, 2],
				[AGILITY, 1],
			]),
		);
		renderDialog();
		await act(async () => {
			fireEvent.click(calculate());
		});
		act(() => player.setEpWeights(new Stats().withStat(Stat.StatStrength, 5)));

		act(() => {
			fireEvent.click(table().querySelector('[data-testid="ep-ratios"] [data-testid="compute-ep"]')!);
		});
		expect(player.epWeights.getStat(Stat.StatStrength)).toBe(1);
		expect(player.epWeights.getStat(Stat.StatAgility)).toBe(0.5);

		await pickOption('ep-type-select', 1);
		act(() => {
			fireEvent.click(table().querySelector('[data-testid="ep-ratios"] [data-testid="compute-ep"]')!);
		});
		expect(player.epWeights.getStat(Stat.StatStrength)).toBe(2);
		expect(player.epWeights.getStat(Stat.StatAgility)).toBe(1);
	});

	describe('calculating', () => {
		it('runs, shows progress, and fills the table, greying the columns whose EP ratio is zero', async () => {
			let report: ((metrics: any) => void) | null = null;
			let finish: ((result: StatWeightsResult) => void) | null = null;
			player.computeStatWeights.mockImplementation(
				(_stats, _pseudo, _ref, onProgress) =>
					new Promise(resolve => {
						report = onProgress;
						finish = resolve;
					}),
			);
			renderDialog();

			await act(async () => {
				fireEvent.click(calculate());
			});
			expect(trackEvent).toHaveBeenCalledWith({ action: 'sim', category: 'stat_weights', label: 'calculate' });
			expect(player.computeStatWeights).toHaveBeenCalledWith(
				[Stat.StatStrength, Stat.StatAgility, Stat.StatCritRating],
				[PseudoStat.PseudoStatSpellHitPercent],
				Stat.StatStrength,
				expect.any(Function),
				{ hpsReferenceStat: Stat.StatStrength },
			);

			const progress = rootElem.querySelector('[data-testid="progress-tracker-dialog"]')!;
			expect(progress.hasAttribute('hidden')).toBe(false);
			expect(popup().hasAttribute('hidden')).toBe(false);
			// Read off the *outer* popup: Base UI counts the dialogs nested inside it, and it only counts
			// this one because the progress dialog is rendered among the EP dialog's own children. A
			// sibling would leave it at 0 and make the inner backdrop an outside press on the outer.
			expect((popup() as HTMLElement).style.getPropertyValue('--nested-dialogs')).toBe('1');

			act(() => report!({ completedSims: 2, totalSims: 8, completedIterations: 250, totalIterations: 1000 }));
			expect(progress.querySelector('[data-testid="progress-tracker-modal-progress-text"]')!.textContent).toBe('250/1000');
			expect(progress.querySelector('[data-testid="progress-tracker-modal-progress-title"]')!.textContent).toContain('2 / 8');

			await act(async () =>
				finish!(
					weightsResult([
						[STRENGTH, 2],
						[AGILITY, 1],
					]),
				),
			);

			expect(rootElem.querySelector('[data-testid="progress-tracker-dialog"]')).toBeNull();
			const strength = rowFor('Strength');
			expect(strength.querySelector('[data-column-type="weight"] [data-testid="results-avg"]')!.textContent).toBe('2.00');
			expect(strength.querySelector('[data-column-type="ep"] [data-testid="results-avg"]')!.textContent).toBe('1.00');
			expect(rowFor('Agility').querySelector('[data-column-type="ep"] [data-testid="results-avg"]')!.textContent).toBe('0.50');

			const stdevCells = [...strength.querySelectorAll('[data-testid="stdev-cell"]')];
			expect(stdevCells[0].hasAttribute('data-unused')).toBe(false);
			expect(stdevCells[2].hasAttribute('data-unused')).toBe(true);
		});

		// DEFECT FIXED. The delta test was `if (…) epAvgElem;` — an expression statement that does
		// nothing, so a column at its current EP was styled as a change.
		it('marks the EP cell against the current weight with a live colour utility, and not at all when they match', async () => {
			player.computeStatWeights.mockResolvedValue(
				weightsResult([
					[STRENGTH, 2],
					[AGILITY, 1],
				]),
			);
			renderDialog();
			act(() => player.setEpWeights(new Stats().withStat(Stat.StatAgility, 0.5)));

			await act(async () => {
				fireEvent.click(calculate());
			});

			const epValue = (stat: string) => rowFor(stat).querySelector('[data-column-type="ep"] [data-testid="results-avg"]')!;
			expect(epValue('Strength').getAttribute('data-sign')).toBe('positive');
			expect([...epValue('Strength').classList]).toEqual(['text-success']);
			expect(epValue('Agility').getAttribute('data-sign')).toBeNull();
			expect([...epValue('Agility').classList]).toEqual([]);

			act(() => player.setEpWeights(new Stats().withStat(Stat.StatAgility, 9)));
			expect(epValue('Agility').getAttribute('data-sign')).toBe('negative');
			expect([...epValue('Agility').classList]).toEqual(['text-danger']);
		});

		it('shows N/A for a stat excluded from the calculation', async () => {
			player.computeStatWeights.mockResolvedValue(
				weightsResult([
					[STRENGTH, 2],
					[AGILITY, 1],
				]),
			);
			renderDialog();
			act(() => settings.setStatExcluded(AGILITY, true));

			await act(async () => {
				fireEvent.click(calculate());
			});
			expect(player.computeStatWeights.mock.calls[0][0]).toEqual([Stat.StatStrength, Stat.StatCritRating]);
			expect(rowFor('Agility').querySelector('[data-column-type="ep"] [data-notapplicable]')).not.toBeNull();
		});

		it('aborts through the progress dialog and leaves the EP dialog open', async () => {
			player.computeStatWeights.mockImplementation(() => new Promise(() => {}));
			renderDialog();
			await act(async () => {
				fireEvent.click(calculate());
			});
			abortType.mockClear();

			await act(async () => {
				fireEvent.click(rootElem.querySelector('button[data-testid="progress-tracker-modal-cancel-btn"]')!);
			});

			expect(abortType).toHaveBeenCalledTimes(1);
			expect(openChanges).toEqual([]);
			expect(popup().hasAttribute('hidden')).toBe(false);
		});

		// DEFECT FIXED. `isRunning` was set before the pre-run abort and the catch returned without
		// clearing it, so one rejection left the Calculate button inert for the life of the page —
		// and it was never disabled, so it still looked live.
		it('stays usable after the pre-run abort rejects', async () => {
			abortType.mockRejectedValueOnce(new Error('worker gone'));
			player.computeStatWeights.mockResolvedValue(weightsResult([[STRENGTH, 2]]));
			renderDialog();

			await act(async () => {
				fireEvent.click(calculate());
			});
			expect(player.computeStatWeights).not.toHaveBeenCalled();
			expect(calculate().disabled).toBe(false);

			await act(async () => {
				fireEvent.click(calculate());
			});
			expect(player.computeStatWeights).toHaveBeenCalledTimes(1);
		});

		it('aborts the run in flight when the dialog is closed', async () => {
			let settle!: (result: StatWeightsResult) => void;
			player.computeStatWeights.mockReturnValueOnce(new Promise<StatWeightsResult>(resolve => (settle = resolve)));
			renderDialog();

			await act(async () => {
				fireEvent.click(calculate());
			});
			expect(calculate().disabled).toBe(true);
			// The pre-run abort has already fired; this counts only the one the close triggers.
			abortType.mockClear();
			const closeButton = popup().querySelector<HTMLButtonElement>('button[data-testid="sim-dialog-close"]')!;

			await act(async () => {
				fireEvent.click(closeButton);
			});
			expect(abortType).toHaveBeenCalledTimes(1);

			await act(async () => {
				settle(weightsResult([[STRENGTH, 2]]));
			});
		});

		it('reports a thrown run, then an aborted one, and closes afterwards without aborting', async () => {
			player.computeStatWeights.mockRejectedValueOnce(new Error('sim exploded'));
			renderDialog();

			await act(async () => {
				fireEvent.click(calculate());
			});
			expect(toasts).toEqual([{ variant: 'error', body: 'sim exploded' }]);
			expect(calculate().disabled).toBe(false);
			expect(rootElem.querySelector('[data-testid="progress-tracker-dialog"]')).toBeNull();

			player.computeStatWeights.mockResolvedValue({ ...emptyStatWeightsResult(), error: { type: ErrorOutcomeType.ErrorOutcomeAborted, message: '' } });
			await act(async () => {
				fireEvent.click(calculate());
			});
			expect(toasts).toEqual([
				{ variant: 'error', body: 'sim exploded' },
				{ variant: 'info', body: 'sidebar.buttons.stat_weights.modal.notifications.cancelled' },
			]);
			expect(rowFor('Strength').querySelector('[data-column-type="ep"] [data-notapplicable]')).not.toBeNull();

			abortType.mockClear();
			await act(async () => {
				fireEvent.click(popup().querySelector('button[data-testid="sim-dialog-close"]')!);
			});
			expect(openChanges).toEqual([false]);
			expect(abortType).not.toHaveBeenCalled();
		});

		it('runs to the target ± in every shown metric, against the chosen DPS and healing reference stats, with no iteration count', async () => {
			player.computeStatWeights.mockResolvedValue(weightsResult([[STRENGTH, 2]]));
			player.refStats.dpsRefStat = Stat.StatAgility;
			player.refStats.healRefStat = Stat.StatIntellect;
			host.sim.showThreatMetrics = true;
			settings.precision = { precisionMode: 'target', iterations: 4000, targetCi: 0.02, maxIterationMultiplier: 4 };
			renderDialog();

			await act(async () => {
				fireEvent.click(calculate());
			});
			expect(player.computeStatWeights).toHaveBeenCalledWith(
				[Stat.StatStrength, Stat.StatAgility, Stat.StatCritRating],
				[PseudoStat.PseudoStatSpellHitPercent],
				Stat.StatAgility,
				expect.any(Function),
				{
					targetEpCi: 0.02,
					maxIterationMultiplier: 4,
					targetMetrics: [
						StatWeightsTargetMetric.StatWeightsTargetMetricDps,
						StatWeightsTargetMetric.StatWeightsTargetMetricHps,
						StatWeightsTargetMetric.StatWeightsTargetMetricTps,
					],
					hpsReferenceStat: Stat.StatIntellect,
				},
			);
			expect(player.computeStatWeights.mock.calls[0][4]).not.toHaveProperty('iterations');
		});

		it('disables Calculate in Target ± mode while the target is empty, and enables it in iterations mode', () => {
			settings.precision = { precisionMode: 'target', iterations: 0, targetCi: 0.02, maxIterationMultiplier: 0 };
			renderDialog();
			expect(calculate().disabled).toBe(false);

			fireEvent.change(popup().querySelector('#ep-precision-target')!, { target: { value: '' } });
			expect(settings.precision.targetCi).toBe(0);
			expect(calculate().disabled).toBe(true);

			fireEvent.click(screen.getByRole('button', { name: precisionText('mode_iterations') }));
			expect(calculate().disabled).toBe(false);
		});

		it('shows the iteration count with the sim setting as placeholder in iterations mode, and the target and cap alone in target mode', () => {
			renderDialog();
			const iterations = popup().querySelector<HTMLInputElement>('#ep-precision-iterations')!;
			expect([iterations.value, iterations.placeholder]).toEqual(['', String(host.sim.store.getState().sim.iterations)]);
			expect(popup().querySelector('#ep-precision-target')).toBeNull();

			fireEvent.change(iterations, { target: { value: '4000' } });
			expect(settings.precision.iterations).toBe(4000);

			fireEvent.click(screen.getByRole('button', { name: precisionText('mode_target') }));
			const cap = popup().querySelector<HTMLInputElement>('#ep-precision-max-multiplier')!;
			expect([cap.value, cap.placeholder]).toEqual(['', precisionText('max_multiplier_none')]);
			expect(popup().querySelector('#ep-precision-iterations')).toBeNull();

			fireEvent.change(cap, { target: { value: '4' } });
			expect(settings.precision.maxIterationMultiplier).toBe(4);
		});

		it('keeps four decimals of the target', () => {
			settings.precision = { precisionMode: 'target', iterations: 0, targetCi: 0.0025, maxIterationMultiplier: 0 };
			renderDialog();
			const target = popup().querySelector<HTMLInputElement>('#ep-precision-target')!;
			expect(target.value).toBe('0.0025');

			fireEvent.change(target, { target: { value: '0.0075' } });
			expect(settings.precision.targetCi).toBe(0.0075);
		});

		it('locks the precision mode and fields while a calculation runs', async () => {
			player.computeStatWeights.mockImplementation(() => new Promise(() => {}));
			settings.precision = { precisionMode: 'target', iterations: 0, targetCi: 0.02, maxIterationMultiplier: 0 };
			renderDialog();

			await act(async () => {
				fireEvent.click(calculate());
			});
			expect(screen.getByRole<HTMLButtonElement>('button', { name: precisionText('mode_iterations'), hidden: true }).disabled).toBe(true);
			expect(popup().querySelector<HTMLInputElement>('#ep-precision-target')!.disabled).toBe(true);
		});

		describe('HPS EP with no healing reference stat chosen', () => {
			// The sim's HPS EP for Agility is 9, which is not Agility's HPS weight over Strength's (0.5): a shown 9.00 is
			// the value as the sim sent it, a shown 0.50 is recomputed against Strength, the spec's reference stat.
			const hpsResult = () => {
				const result = weightsResult([
					[STRENGTH, 2],
					[AGILITY, 1],
				]);
				STRENGTH.setProtoValue(result.hps!.weights!, 2);
				STRENGTH.setProtoValue(result.hps!.epValues!, 1);
				AGILITY.setProtoValue(result.hps!.weights!, 1);
				AGILITY.setProtoValue(result.hps!.epValues!, 9);
				return result;
			};
			const hpsEp = (stat: string) => rowFor(stat).querySelectorAll('[data-column-type="ep"] [data-testid="results-avg"]')[1].textContent;

			beforeEach(() => {
				player.computeStatWeights.mockResolvedValue(hpsResult());
				player.refStats.dpsRefStat = Stat.StatAgility;
			});

			it.each(['iterations', 'target'] as const)(
				'runs in %s mode against the chosen DPS and the spec’s HPS reference stat, and shows HPS EP as the sim sent it',
				async precisionMode => {
					settings.precision = { precisionMode, iterations: 4000, targetCi: 0.02, maxIterationMultiplier: 4 };
					renderDialog();

					await act(async () => {
						fireEvent.click(calculate());
					});
					const [call] = player.computeStatWeights.mock.calls;
					expect(call[2]).toBe(Stat.StatAgility);
					expect(call[4].hpsReferenceStat).toBe(Stat.StatStrength);
					expect(hpsEp('Agility')).toBe('9.00');
				},
			);
		});
	});

	// DEFECT FIXED. The panel discarded the unsubscribe its table-level `subscribePlayerField`
	// returned, so the subscription outlived the panel.
	it('drops every subscription when it unmounts', () => {
		const { unmount } = renderDialog();
		expect(source.count()).toBeGreaterThan(0);
		unmount();
		expect(source.count()).toBe(0);
	});
});
