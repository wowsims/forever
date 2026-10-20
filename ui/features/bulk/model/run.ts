import { enchantPresetPool } from '@features/enchant-presets/model/enchant_presets';
import { BulkSettings, ProgressMetrics } from '@generated/proto/api';
import type { Encounter } from '@generated/proto/common';
import i18n from '@i18n/config';
import { BulkEncounterResults, BulkSimProgressConfig, BulkTalentsOverride } from '@sim/bulk/types';
import type { Player } from '@sim/player/player';
import { bulkState, patchBulkState } from '@sim/settings/bulk_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { RequestTypes } from '@sim/sim_signal_manager';
import { toastManager } from '@ui-kit/Toast';

import { trackEvent } from '../../../tracking/utils';
import { type CoreBulkSimResult, runCoreBulkSim as runCoreBulkSimImpl } from './core_sim';
import { bulkEncounterRuns, encounterLabel, encounterLength, formatEncounterSummary, pickableBulkEncounters } from './encounters';
import { BulkProgress, candidateGearProgress, simProgress } from './progress';
import { type BulkTalentRun, rankEncounterResults } from './results';
import { createBulkSettingsProto } from './settings';
import { bulkTalentRuns, pickableBulkTalents, talentsLabel } from './talents';

// The stored settings plus the enchant pool, which is derived from the presets and never stored.
const bulkRequestSettings = (player: Player<any>): BulkSettings => ({ ...createBulkSettingsProto(player), enchants: enchantPresetPool(player) });

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
}

interface BulkEncounterSteps {
	/** One per encounter and talent setup the batch runs, encounter outer, one array for the whole run. */
	labels: readonly string[];
	/** 0-based index into `labels` of the run in progress. */
	running: number;
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
			steps: { labels: [], running: 0 },
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
	emitProgress(player, simProgress(metrics, config, runOf(player).simStart, new Date().getTime()));
};

const calculateBulkCombinations = async (host: IndividualSimHost<any>) => {
	const { sim, player } = host;
	try {
		const bulkSettings = bulkRequestSettings(player);
		const combinationCountResult = await sim.getBulkCombinationCount(bulkSettings);

		if (combinationCountResult.error) {
			throw new Error(combinationCountResult.error.message || i18n.t('bulk_tab.notifications.combinations_failed'));
		}

		patchBulkState(player, { combinations: combinationCountResult.combinations, iterations: combinationCountResult.iterations });
	} catch (e) {
		host.handleCrash(e);
	}
};

export const refreshBulkCombinations = async (host: IndividualSimHost<any>) => {
	const run = runOf(host.player);
	const requestVersion = ++run.combinationsRequestVersion;
	patchBulkState(host.player, { combinationsPending: true });
	await calculateBulkCombinations(host);
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

const runCoreBulkSim = async (
	host: IndividualSimHost<any>,
	signal: AbortSignal,
	bulkSettings: BulkSettings,
	encounter: Encounter,
	talents?: BulkTalentsOverride,
): Promise<CoreBulkSimResult> => {
	const { player } = host;
	return runCoreBulkSimImpl(
		{
			simUI: host,
			throwIfBulkAborted: abortSignal => throwIfBulkAborted(player, abortSignal),
			runWithBulkAbort: (promise, abortSignal) => runWithBulkAbort(player, promise, abortSignal),
			setSimProgress: (metrics, config) => setSimProgress(player, metrics, config),
			debugOptimisationRound: (message, data) => console.debug(`[bulk-core] ${message}`, data ?? ''),
		},
		signal,
		bulkSettings,
		encounter,
		talents,
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

export const runBulkBatch = async (host: IndividualSimHost<any>) => {
	const { sim, player } = host;
	const run = runOf(player);
	if (bulkState(player).isRunning) return;

	trackEvent({
		action: 'sim',
		category: 'simulate',
		label: 'batch',
		value: bulkState(player).combinations,
	});

	run.isCancelling = false;
	const encounterRuns = bulkEncounterRuns(sim.encounter.toProto(), pickableBulkEncounters(host), bulkState(player).encounters, sim.getShowPowerTools());
	const talentRuns = bulkTalentRuns(
		player.getPlayerSpec(),
		player.getTalentsString(),
		pickableBulkTalents(host),
		bulkState(player).talents,
		sim.getShowPowerTools(),
	).map(talents => ({
		talents,
		override: talents.isCurrent
			? undefined
			: { talentsString: talents.talentsString, rotation: player.getResolvedAplRotationForTalents(talents.talentsString) },
	}));
	run.steps = {
		labels: encounterRuns.flatMap(encounterRun =>
			talentRuns.map(({ talents }) =>
				talentRuns.length === 1
					? `${encounterLabel(encounterRun)} · ${formatEncounterSummary(encounterRun.encounter)}`
					: `${encounterLabel(encounterRun)} · ${talentsLabel(talents)}`,
			),
		),
		running: 0,
	};
	// The progress dialog mounts on `isRunning` and takes the steps from the progress there, before the first stage reports.
	setCandidateGearProgress(player);
	patchBulkState(player, { isRunning: true, started: true });
	run.abortController = new AbortController();
	run.abortPromise = null;
	const abortSignal = run.abortController.signal;

	await sim.waitForInit();
	const bulkSettings = bulkRequestSettings(player);
	const entries: BulkEncounterResults[] = [];
	let runError: unknown = null;

	try {
		// Bulk owns its own request type now, so this has to name both: a batch still replaces an
		// in-flight single sim, and it still replaces a previous batch, which naming one would drop.
		await sim.signalManager.abortType(RequestTypes.IndividualSim | RequestTypes.BulkSim);
		run.simStart = new Date().getTime();
		const baseGear = player.getGear();
		patchBulkState(player, { runGear: baseGear, results: null });

		await refreshBulkCombinations(host);

		for (const [encounterIndex, encounterRun] of encounterRuns.entries()) {
			const finished: BulkTalentRun[] = [];
			for (const [setupIndex, { talents, override }] of talentRuns.entries()) {
				throwIfBulkAborted(player, abortSignal);
				run.steps = { ...run.steps, running: encounterIndex * talentRuns.length + setupIndex };
				setCandidateGearProgress(player);
				run.simStart = new Date().getTime();
				finished.push({ talents, result: await runCoreBulkSim(host, abortSignal, bulkSettings, encounterRun.encounter, override) });
			}
			// Pushed only once every setup finished, so a cancel or error mid-encounter drops the encounter.
			entries.push({
				name: encounterRun.name,
				isCurrent: encounterRun.isCurrent,
				...encounterLength(encounterRun.encounter),
				results: rankEncounterResults(finished, baseGear),
			});
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
		patchBulkState(player, entries.length ? { isRunning: false, results: entries } : { isRunning: false });
	}
};
