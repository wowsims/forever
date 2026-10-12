import { BulkSettings, DistributionMetrics, ProgressMetrics } from '@generated/proto/api';
import i18n from '@i18n/config';
import { BulkSimProgressConfig, TopGearResult } from '@sim/bulk/types';
import { bulkSimStageToOptimisationStage, cleanBulkDpsMetrics, getCoreBulkSimTrackingMetrics } from '@sim/bulk/utils';
import type { IndividualSimHost } from '@sim/sim_host';

export interface CoreBulkSimContext {
	simUI: IndividualSimHost<any>;
	throwIfBulkAborted: (signal: AbortSignal) => void;
	runWithBulkAbort: <T>(promise: Promise<T>, signal: AbortSignal) => Promise<T>;
	setSimProgress: (progress: ProgressMetrics, config: BulkSimProgressConfig) => void;
	debugOptimisationRound: (message: string, data?: unknown) => void;
}

export interface CoreBulkSimResult {
	referenceDpsMetrics: DistributionMetrics;
	topGearResults: TopGearResult[];
	/** Iterations per displayed result, from the last stage that ran; 0 when none did. */
	iterations: number;
	metrics: Record<string, string | number>;
}

export interface CoreBulkSimOptions {
	encounterOverride?: { duration: number; durationVariation: number };
	/** Wraps every progress title, e.g. with the duration being run. */
	decorateTitle?: (title: string) => string;
}

export async function runCoreBulkSim(
	context: CoreBulkSimContext,
	signal: AbortSignal,
	bulkSettings: BulkSettings,
	{ encounterOverride, decorateTitle = title => title }: CoreBulkSimOptions = {},
): Promise<CoreBulkSimResult> {
	context.throwIfBulkAborted(signal);
	context.debugOptimisationRound('core bulk sim started');

	let currentProgressStage: ProgressMetrics['bulkStage'] | undefined;
	let currentProgressStageStartedAt = new Date().getTime();
	const updateProgress = (progress: ProgressMetrics) => {
		if (progress.totalIterations <= 0) return;
		if (progress.bulkStage !== currentProgressStage) {
			currentProgressStage = progress.bulkStage;
			currentProgressStageStartedAt = new Date().getTime();
		}

		const stageName = bulkSimStageToOptimisationStage(progress.bulkStage);
		context.setSimProgress(progress, {
			currentRound: 1,
			totalRounds: 1,
			title: decorateTitle(stageName ? i18n.t(`bulk_tab.progress.${stageName}_iteration_rounds`) : i18n.t('bulk_tab.progress.refining_rounds')),
			aggregateStartedAt: currentProgressStageStartedAt,
			useSimCountProgress: true,
		});
	};

	const result = await context.runWithBulkAbort(context.simUI.sim.runBulkSim(updateProgress, bulkSettings, signal, encounterOverride), signal);
	if (!result || (result && 'type' in result)) {
		throw new Error(result?.message);
	}
	if (!result.baseline?.dpsMetrics) {
		throw new Error('Bulk sim did not return baseline results.');
	}

	const topGearResults = result.topResults
		.filter(topResult => topResult.gear && topResult.dpsMetrics)
		.map((topResult, backendRank) => ({
			gear: context.simUI.sim.db.lookupEquipmentSpec(topResult.gear!),
			dpsMetrics: cleanBulkDpsMetrics(topResult.dpsMetrics!),
			backendRank,
			pairedErrorToNextResult: topResult.pairedErrorToNextResult,
			pairedErrorToBaseline: topResult.pairedErrorToBaseline,
		}));

	context.debugOptimisationRound('core bulk sim complete', {
		durationSeconds: result.timings?.totalSeconds ?? 0,
		stageMetrics: result.stageMetrics,
		topGearResults: topGearResults.map((topResult, index) => ({
			rank: index + 1,
			avg: topResult.dpsMetrics.avg,
			stdev: topResult.dpsMetrics.stdev,
		})),
	});

	return {
		referenceDpsMetrics: result.baseline.dpsMetrics,
		topGearResults,
		iterations: result.stageMetrics.at(-1)?.iterations ?? 0,
		metrics: getCoreBulkSimTrackingMetrics(result),
	};
}
