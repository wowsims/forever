import { enchantPresetPool } from '@features/enchant-presets/model/enchant_presets';
import { readSavedSettings } from '@features/settings/model/saved_settings';
import { BulkSettings, ProgressMetrics } from '@generated/proto/api';
import { APLRotation } from '@generated/proto/apl';
import { type Encounter, type Faction, Profession } from '@generated/proto/common';
import i18n from '@i18n/config';
import { BulkDropsCombined, BulkEncounterResults, BulkEncounterSetup, BulkResults, BulkSimOverrides, BulkSimProgressConfig } from '@sim/bulk/types';
import type { Player } from '@sim/player/player';
import type { Gear } from '@sim/proto/gear';
import { getGearIdentityKey } from '@sim/proto/items';
import { bulkState, patchBulkState } from '@sim/settings/bulk_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { RequestTypes } from '@sim/sim_signal_manager';
import type { BulkMode } from '@sim/state/sim_store';
import { Z_95 } from '@sim/utils/math';
import { BULK_SIM_DEFAULT_TOP_RESULTS } from '@sim/wasm/bulk_sim/constants_auto_gen';
import { estimateBulkSimIterations } from '@sim/wasm/bulk_sim/estimate';
import { toastManager } from '@ui-kit/Toast';

import { trackEvent } from '../../../../tracking/utils';
import { batchEncounterSetup } from '../batch_encounter';
import { bulkBuildPlan, type BulkBuildStep, bulkBuildStep, bulkBundlePlans, bulkGearSetSetups, bulkPlanRuns } from '../builds/builds';
import { bulkBundleRuns, pickableBulkBundleParts } from '../builds/bundles';
import { bulkEnchantPickEntries, bulkEnchantPickRuns } from '../builds/enchant_picks';
import { bulkEncounterRuns, pickableBulkEncounters } from '../builds/encounters';
import { dropsCombinedSets } from '../drops/combined';
import { dropsGap } from '../drops/results';
import { dropsRequestSettings } from '../drops/settings';
import { createBulkSettingsProto } from '../gear/settings';
import { bulkRunMode } from '../mode';
import { type BulkFinishedRequest, rankEncounterResults } from '../results/results';
import { hasPairedDiffToBaseline } from '../results/tie_chains';
import { type CoreBulkSimResult, runCoreBulkSim as runCoreBulkSimImpl } from './core_sim';
import { BulkProgress, type BulkRequestIterations, type BulkStepsTiming, candidateGearProgress, requestIterationsProgress, simProgress } from './progress';
import { bulkTopResults } from './top_results';

// Gear sends the stored settings plus the enchant pool, which is derived from the presets and never stored. Builds sends
// none: its gear setups already carry its enchant presets, so Gear-mode picks stay out of its seed.
const bulkRequestSettingsFor = (runMode: BulkMode, player: Player<any>): BulkSettings => {
	if (runMode === 'builds') return BulkSettings.create();
	if (runMode === 'drops') return dropsRequestSettings(player);
	return { ...createBulkSettingsProto(player), enchants: enchantPresetPool(player) };
};

/**
 * A batch in flight, per player. None of it belongs in the store: an abort controller and a
 * promise are not state a selector can compare, and the progress ticks are kept out on purpose,
 * so a tick renders the one leaf that shows it instead of every reader of the slice.
 */
interface BulkRun {
	simStart: number;
	isCancelling: boolean;
	abortController: AbortController | null;
	abortPromise: Promise<void> | null;
	combinationsRequestVersion: number;
	progress: BulkRunProgress | null;
	listeners: Set<(progress: BulkRunProgress) => void>;
	steps: BulkEncounterSteps;
	/** The request in flight, with the iterations its step's earlier requests and all its requests are expected to sim. */
	request: BulkRequestIterations & { before: number; stepExpected: number };
}

interface BulkEncounterSteps extends BulkStepsTiming {
	/** One per encounter and bundle group the batch runs, encounter outer, one array for the whole run. */
	rows: readonly BulkBuildStep[];
}

/** A stage's progress with the batch's encounter steps; the run emits a first one before it is marked running. */
interface BulkRunProgress extends BulkProgress {
	steps: BulkEncounterSteps;
}

const runs = new WeakMap<Player<any>, BulkRun>();

const runOf = (player: Player<any>): BulkRun => {
	let run = runs.get(player);
	if (!run) {
		run = {
			simStart: 0,
			isCancelling: false,
			abortController: null,
			abortPromise: null,
			combinationsRequestVersion: 0,
			progress: null,
			listeners: new Set(),
			steps: { rows: [], groups: [], weights: [], seconds: [], running: 0, progress: 0 },
			request: { before: 0, expected: 0, stepExpected: 0, simmed: 0 },
		};
		runs.set(player, run);
	}
	return run;
};

export const bulkProgress = (player: Player<any>): BulkRunProgress | null => runOf(player).progress;

export const subscribeBulkProgress = (player: Player<any>, listener: (progress: BulkRunProgress) => void): (() => void) => {
	const { listeners } = runOf(player);
	listeners.add(listener);
	return () => {
		listeners.delete(listener);
	};
};

const emitProgress = (player: Player<any>, next: BulkProgress | null) => {
	if (!next) return;
	const run = runOf(player);
	const progress = { ...next, steps: run.steps };
	run.progress = progress;
	for (const listener of run.listeners) listener(progress);
};

const setCandidateGearProgress = (player: Player<any>) => {
	emitProgress(player, candidateGearProgress({ now: new Date().getTime() }));
};

const setSimProgress = (player: Player<any>, metrics: ProgressMetrics, config: BulkSimProgressConfig) => {
	const run = runOf(player);
	const { request } = run;
	const stepProgress = (request.before + request.expected * requestIterationsProgress(request, metrics)) / request.stepExpected;
	if (stepProgress > run.steps.progress) run.steps = { ...run.steps, progress: stepProgress };
	emitProgress(player, simProgress(metrics, config, run.simStart, new Date().getTime()));
};

// A reply to an older request writes nothing: the workers can answer two overlapping counts out of order.
const calculateBulkCombinations = async (host: IndividualSimHost<any>, requestVersion: number) => {
	const { sim, player } = host;
	try {
		const runMode = bulkRunMode(sim.getShowPowerTools(), bulkState(player).mode);
		// Builds counts off its plan; the stored count stays Gear's.
		const bulkSettings = bulkRequestSettingsFor(runMode === 'builds' ? 'gear' : runMode, player);
		const combinationCountResult = await sim.getBulkCombinationCount(bulkSettings);
		if (requestVersion !== runOf(player).combinationsRequestVersion) return;

		// A selection that gives no combination is the user's to change, shown beside the count, not a crash: the count
		// runs again on every edit, so a crash here would alert on each one.
		if (combinationCountResult.error) {
			patchBulkState(player, {
				combinations: 0,
				combinationsError: combinationCountResult.error.message || i18n.t('bulk_tab.notifications.combinations_failed'),
			});
			return;
		}

		patchBulkState(player, {
			combinations: combinationCountResult.combinations,
			combinationsError: null,
		});
	} catch (e) {
		host.handleCrash(e);
	}
};

export const refreshBulkCombinations = async (host: IndividualSimHost<any>) => {
	const run = runOf(host.player);
	const requestVersion = ++run.combinationsRequestVersion;
	patchBulkState(host.player, { combinationsPending: true });
	await calculateBulkCombinations(host, requestVersion);
	if (requestVersion !== run.combinationsRequestVersion) {
		return;
	}
	patchBulkState(host.player, { combinationsPending: false });
};

const throwIfBulkAborted = (player: Player<any>, signal: AbortSignal) => {
	if (signal.aborted || runOf(player).isCancelling) {
		throw new Error(i18n.t('bulk_tab.notifications.bulk_sim_aborted'));
	}
};

const runWithBulkAbort = async <T>(player: Player<any>, promise: Promise<T>, signal: AbortSignal): Promise<T> => {
	throwIfBulkAborted(player, signal);

	let abortHandler: (() => void) | null = null;
	const abortPromise = new Promise<never>((_, reject) => {
		abortHandler = () => reject(new Error(i18n.t('bulk_tab.notifications.bulk_sim_aborted')));
		signal.addEventListener('abort', abortHandler, { once: true });
	});

	try {
		return await Promise.race([promise, abortPromise]);
	} finally {
		if (abortHandler) {
			signal.removeEventListener('abort', abortHandler);
		}
	}
};

// `title` names every stage of the request on the progress bar in place of the stage's own name.
const runCoreBulkSim = async (
	host: IndividualSimHost<any>,
	signal: AbortSignal,
	bulkSettings: BulkSettings,
	overrides: BulkSimOverrides,
	title?: string,
): Promise<CoreBulkSimResult> => {
	const { player } = host;
	return runCoreBulkSimImpl(
		{
			simUI: host,
			throwIfBulkAborted: abortSignal => throwIfBulkAborted(player, abortSignal),
			runWithBulkAbort: (promise, abortSignal) => runWithBulkAbort(player, promise, abortSignal),
			setSimProgress: (metrics, config) => setSimProgress(player, metrics, title ? { ...config, title } : config),
			debugOptimisationRound: (message, data) => console.debug(`[bulk-core] ${message}`, data ?? ''),
		},
		signal,
		bulkSettings,
		overrides,
	);
};

const abortBulkSimWork = async (host: IndividualSimHost<any>) => {
	const run = runOf(host.player);
	if (run.abortPromise) {
		return run.abortPromise;
	}

	const abortController = run.abortController;
	if (!abortController) return;

	run.abortController = null;
	if (!abortController.signal.aborted) {
		abortController.abort();
	}

	run.abortPromise = (async () => {
		// Narrower than `All`: cancelling a batch must not also cancel a stat-weights or enchants
		// optimizer run, which is the whole point of bulk having its own type.
		await host.sim.signalManager.abortType(RequestTypes.BulkSim);
	})();

	try {
		await run.abortPromise;
	} finally {
		run.abortPromise = null;
	}
};

export const cancelBulkBatch = async (host: IndividualSimHost<any>) => {
	const run = runOf(host.player);
	if (!bulkState(host.player).isRunning || run.isCancelling) return;

	run.isCancelling = true;
	await abortBulkSimWork(host);
};

interface BulkRequestPlan {
	request: Omit<BulkFinishedRequest, 'result'>;
	overrides: BulkSimOverrides;
}

const requestEstimates = (requests: readonly BulkRequestPlan[]) =>
	requests.map(({ overrides }) => {
		const candidates = overrides.candidates?.length ?? 0;
		return estimateBulkSimIterations(candidates, (overrides.topResults ?? BULK_SIM_DEFAULT_TOP_RESULTS) || candidates);
	});

// An encounter and its steps, in run order. A step's requests are one run's splits.
interface BulkEncounterPlan {
	encounter: BulkEncounterSetup;
	/** `group` is the step's bundle group, as its index in the batch's groups. */
	steps: Array<{ row: BulkBuildStep; group: number; requests: BulkRequestPlan[] }>;
}

interface CurrentBuild {
	talentsString: string;
	rotation: APLRotation;
}

// Drops adds `cullToBaseline`, under which the sim plans for the default top results whatever it is sent.
const gearModePlan = (
	host: IndividualSimHost<any>,
	{ talentsString, rotation }: CurrentBuild,
	requestOverrides: Pick<BulkSimOverrides, 'topResults' | 'cullToBaseline'>,
): BulkEncounterPlan[] => {
	const encounter = batchEncounterSetup(host);
	const bundle = {
		talents: { name: '', talentsString, isCurrent: true },
		rotation: { name: '', isCurrent: true, rotation },
		settings: { name: '', isCurrent: true, settings: null },
	};
	return [
		{
			encounter,
			steps: [
				{
					row: bulkBuildStep({ encounter, bundle }),
					group: 0,
					requests: [
						{
							request: { reference: true, talents: bundle.talents, rotation: bundle.rotation, settings: null, setups: null },
							overrides: { encounter: encounter.encounter, ...requestOverrides },
						},
					],
				},
			],
		},
	];
};

// Empty with no bundle or no encounter to run. A request's setups are keyed by the gear as sent, which is the gear the
// sim's rows carry. Every request sends its candidates, an empty list too: given a baseline gear and no candidates, the
// sim would generate them. Every request asks for the batch's top results, every candidate with All, so a cap that covers
// a request's candidates sims them all at the high stage and culls none.
const buildsModePlan = (host: IndividualSimHost<any>, current: CurrentBuild, topResults: number): BulkEncounterPlan[] => {
	const { sim, player } = host;
	const slice = bulkState(player);
	const reference = {
		gear: player.getGear(),
		bonusStats: player.getBonusStats(),
		isEnchanter: player.hasProfession(Profession.Enchanting),
		settings: readSavedSettings(host),
	};
	const bundles = bulkBundleRuns(
		{ gear: reference.gear, bonusStats: reference.bonusStats, ...current, settings: reference.settings },
		pickableBulkBundleParts(host),
		slice.bundles,
		slice.currentBundleChecked,
	);
	const enchantPicks = bulkEnchantPickRuns(bulkEnchantPickEntries(slice.enchantPresets), slice.buildEnchantPresets);
	const plans = bulkBundlePlans(bundles, bulkGearSetSetups(player, enchantPicks));
	const encounters = bulkEncounterRuns(sim.encounter.toProto(), pickableBulkEncounters(host), slice.encounters);
	const runs = bulkPlanRuns(
		bulkBuildPlan(plans, reference, build => player.getResolvedAplRotationForBuild(build)),
		encounters,
	);
	return encounters.flatMap(encounter => {
		const steps = runs
			.filter(run => run.encounter === encounter)
			.map(run => ({
				row: bulkBuildStep(run),
				group: run.group,
				requests: run.requests.map((request): BulkRequestPlan => ({
					request: {
						reference: request.baselineGear === null,
						talents: run.bundle.talents,
						rotation: run.bundle.rotation,
						settings: run.bundle.settings,
						setups: {
							byGearKey: new Map(request.candidates.map(setup => [getGearIdentityKey(setup.gear.asSpec()), setup])),
							baseline: request.baselineSetup,
						},
					},
					overrides: {
						encounter: encounter.encounter,
						settings: run.bundle.settings.settings ?? undefined,
						build: request.override,
						candidates: request.candidates.map(setup => setup.gear),
						baselineGear: request.baselineGear ?? undefined,
						topResults,
					},
				})),
			}));
		return steps.length ? [{ encounter, steps }] : [];
	});
};

// One request for every Drops card's combined set, a set two cards share sent once, after the run's own steps. Every set
// counts among the top results, so the request goes straight to the high stage and culls none. It sims every set at the
// run's own iteration count, its finalists' when that stage ran, and skips the finalist stage, so a combined gain is as
// precise as the run's best rows. Its seed differs from the run's, so each gain is the paired difference to this
// request's own baseline. On the progress bar it follows the run's one step under its own title, the step's time left
// cleared.
const runDropsCombined = async (
	host: IndividualSimHost<any>,
	signal: AbortSignal,
	results: BulkResults,
	encounter: Encounter,
	runSources: readonly string[],
	runFaction: Faction,
): Promise<Map<string, BulkDropsCombined> | null> => {
	const { player } = host;
	const sets = dropsCombinedSets(player.sim.db, results, runSources, runFaction);
	const keyOf = (gear: Gear) => getGearIdentityKey(gear.asSpec());
	const candidates = [...new Map([...sets.values()].map(set => [keyOf(set.gear), set.gear])).values()];
	if (!candidates.length) return null;

	throwIfBulkAborted(player, signal);
	const run = runOf(player);
	const minIterations = results.iterations;
	const expected = minIterations * (candidates.length + 1);
	run.steps = { ...run.steps, running: run.steps.groups.length };
	run.request = { before: 0, expected, stepExpected: expected, simmed: 0 };
	run.simStart = new Date().getTime();
	emitProgress(player, run.progress);
	const result = await runCoreBulkSim(
		host,
		signal,
		BulkSettings.create(),
		{ encounter, candidates, topResults: 0, minIterations },
		i18n.t('bulk_tab.progress.combined_gain_rounds'),
	);

	const baseline = { gear: results.currentGear, dpsMetrics: result.referenceDpsMetrics };
	const rows = new Map(result.topGearResults.map(row => [keyOf(row.gear), row]));
	const combined = new Map<string, BulkDropsCombined>();
	for (const [groupKey, { gear, itemIds, leftOutIds }] of sets) {
		const row = rows.get(keyOf(gear));
		if (!row) continue;
		const margin = hasPairedDiffToBaseline(row) ? Z_95 * row.pairedErrorToBaseline : 0;
		combined.set(groupKey, { itemIds, leftOutIds, gain: dropsGap(row, baseline), margin, baselineDps: baseline.dpsMetrics.avg });
	}
	return combined;
};

export const runBulkBatch = async (host: IndividualSimHost<any>) => {
	const { sim, player } = host;
	const run = runOf(player);
	if (bulkState(player).isRunning) return;

	const powerTools = sim.getShowPowerTools();
	const runMode = bulkRunMode(powerTools, bulkState(player).mode);
	const builds = runMode === 'builds';
	const drops = runMode === 'drops';
	const topResults = bulkTopResults(powerTools, bulkState(player));
	const current = { talentsString: player.getTalentsString(), rotation: APLRotation.clone(player.aplRotation) };
	// Read before `waitForInit`: Simulate is enabled only once the sim is ready, so the database is loaded.
	const plan = builds
		? buildsModePlan(host, current, topResults)
		: gearModePlan(host, current, drops ? { topResults, cullToBaseline: true } : { topResults });
	if (!plan.length) return;

	trackEvent({
		action: 'sim',
		category: 'simulate',
		label: 'batch',
		value: builds ? plan.flatMap(encounterPlan => encounterPlan.steps.flatMap(step => step.requests)).length : bulkState(player).combinations,
	});

	run.isCancelling = false;
	const planSteps = plan.flatMap(encounterPlan => encounterPlan.steps);
	run.steps = {
		rows: planSteps.map(step => step.row),
		groups: planSteps.map(step => step.group),
		weights: planSteps.map(step => requestEstimates(step.requests).reduce((sum, count) => sum + count, 0)),
		seconds: [],
		running: 0,
		progress: 0,
	};
	// The progress dialog mounts on `isRunning` and takes the steps from the progress there, before the first stage reports.
	setCandidateGearProgress(player);
	const startedPatch = drops ? { dropsStarted: true } : builds ? { buildsStarted: true } : { gearStarted: true };
	patchBulkState(player, { isRunning: true, ...startedPatch });
	run.abortController = new AbortController();
	run.abortPromise = null;
	const abortSignal = run.abortController.signal;

	await sim.waitForInit();
	const bulkSettings = bulkRequestSettingsFor(runMode, player);
	// The ticks and the side the Drops request was built from, which its results group by.
	const dropsRunSources = bulkState(player).dropsSources;
	const dropsRunFaction = player.getFaction();
	const dropsCombinedGain = drops && bulkState(player).dropsCombinedGain;
	let dropsCombined: ReadonlyMap<string, BulkDropsCombined> | null = null;
	const resultsPatch = (results: BulkEncounterResults[] | null) => {
		if (drops) return { dropsResults: results, dropsRunSources, dropsRunFaction, dropsCombined };
		return builds ? { buildsResults: results } : { gearResults: results };
	};
	const entries: BulkEncounterResults[] = [];
	let runError: unknown = null;

	try {
		// Bulk owns its own request type now, so this has to name both: a batch still replaces an
		// in-flight single sim, and it still replaces a previous batch, which naming one would drop.
		await sim.signalManager.abortType(RequestTypes.IndividualSim | RequestTypes.BulkSim);
		run.simStart = new Date().getTime();
		const baseGear = player.getGear();
		patchBulkState(player, { runGear: baseGear, ...resultsPatch(null) });

		if (!builds) await refreshBulkCombinations(host);

		let stepIndex = 0;
		for (const encounterPlan of plan) {
			const finished: BulkFinishedRequest[] = [];
			for (const { requests } of encounterPlan.steps) {
				const expected = requestEstimates(requests);
				const stepExpected = expected.reduce((sum, count) => sum + count, 0);
				const stepStartedAt = Date.now();
				let before = 0;
				for (const [index, { request, overrides }] of requests.entries()) {
					throwIfBulkAborted(player, abortSignal);
					run.steps = { ...run.steps, running: stepIndex, progress: before / stepExpected };
					run.request = { before, expected: expected[index], stepExpected, simmed: 0 };
					before += expected[index];
					// The last frame with the moved steps, not a reset: an unmeasured frame drops the bar's readout line until
					// the request's first tick, which moves the dialog at every request.
					emitProgress(player, run.progress);
					run.simStart = new Date().getTime();
					finished.push({ ...request, result: await runCoreBulkSim(host, abortSignal, bulkSettings, overrides) });
				}
				run.steps = { ...run.steps, seconds: [...run.steps.seconds, (Date.now() - stepStartedAt) / 1000] };
				stepIndex++;
			}
			// Pushed only once every request finished, so a cancel or error mid-encounter drops the encounter.
			entries.push({
				name: encounterPlan.encounter.name,
				isCurrent: encounterPlan.encounter.isCurrent,
				duration: encounterPlan.encounter.encounter.duration,
				durationVariation: encounterPlan.encounter.encounter.durationVariation,
				// Drops ranks every row the sim returns, culled ones included.
				results: rankEncounterResults(finished, baseGear, drops ? 0 : topResults),
			});
		}
		// After the push, so a cancel or error here keeps the run's own results.
		if (dropsCombinedGain && entries.length) {
			dropsCombined = await runDropsCombined(host, abortSignal, entries[0].results, plan[0].encounter.encounter, dropsRunSources, dropsRunFaction);
		}
	} catch (error) {
		runError = error;
		console.error(error);
		const errorMessage = error instanceof Error ? error.message : typeof error === 'string' ? error : undefined;
		if (!run.isCancelling && errorMessage) {
			toastManager.add({
				variant: 'error',
				body: errorMessage,
			});
		}
	} finally {
		const wasCancelling = run.isCancelling;
		if (wasCancelling || runError) {
			await abortBulkSimWork(host);
		}
		await player.setGearAsync(bulkState(player).runGear!);
		if (wasCancelling) {
			toastManager.add({
				variant: 'error',
				body: i18n.t('bulk_tab.notifications.bulk_sim_cancelled'),
			});
		}
		run.isCancelling = false;
		patchBulkState(player, entries.length ? { isRunning: false, ...resultsPatch(entries) } : { isRunning: false });
	}
};
