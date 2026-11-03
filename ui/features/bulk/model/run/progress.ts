import { BulkSimStage, ProgressMetrics } from '@generated/proto/api';
import i18n from '@i18n/config';
import type { BulkSimProgressConfig } from '@sim/bulk/types';
import { timeRemainingAtRate } from '@sim/utils/math';

export interface BulkRequestIterations {
	/** What the request is expected to sim, `estimateBulkSimIterations` of its candidates. */
	expected: number;
	simmed: number;
	phase?: { stage: BulkSimStage; total: number; completed: number };
}

// A stage reports in phases, each with its own counts: the baseline probe, the main pass, each adaptive pass, and each
// finalist pass. A tick whose stage or total differs, or whose count went down, starts a phase, and only sets where it
// starts: the main pass starts at the probe's iterations, which are counted. A finalist stage with nothing to separate
// reports nothing.
/** How far the request is, 0 to 1: its simmed iterations over the estimate, or over what the phase in flight still adds when that is more. */
export const requestIterationsProgress = (request: BulkRequestIterations, metrics: ProgressMetrics): number => {
	const { phase } = request;
	const samePhase = phase?.stage === metrics.bulkStage && phase.total === metrics.totalIterations && metrics.completedIterations >= phase.completed;
	if (samePhase) request.simmed += metrics.completedIterations - phase.completed;
	request.phase = { stage: metrics.bulkStage, total: metrics.totalIterations, completed: metrics.completedIterations };
	const phaseLeft = Math.max(0, metrics.totalIterations - metrics.completedIterations);
	return Math.min(1, request.simmed / Math.max(request.expected, request.simmed + phaseLeft, 1));
};

export interface BulkStepsTiming {
	/** Each step's bundle group, as its index in the batch's groups. */
	groups: readonly number[];
	/** The iterations each step's requests are expected to sim. */
	weights: readonly number[];
	/** How long each finished step took, in seconds, in step order. */
	seconds: readonly number[];
	/** 0-based index of the step in progress. */
	running: number;
	/** How far the running step is, 0 to 1, each of its requests weighted by the iterations it is expected to sim. It never moves back within a step. */
	progress: number;
}

// A step costs what its bundle group cost under the latest encounter, scaled by their planned iterations: in a real
// batch one group took up to four times as long as another with the same plan.
/** The seconds left, or undefined until every bundle group has finished a step, which is after the first encounter. */
export const batchSecondsRemaining = ({ groups, weights, seconds, running, progress }: BulkStepsTiming): number | undefined => {
	const finished = seconds.length;
	if (!finished || running >= groups.length) return undefined;
	if (new Set(groups.slice(0, finished)).size < new Set(groups).size) return undefined;
	const predict = (step: number) => {
		const measured = groups.lastIndexOf(groups[step], finished - 1);
		return seconds[measured] * (weights[step] / weights[measured]);
	};
	let left = predict(running) * (1 - progress);
	for (let step = running + 1; step < groups.length; step++) left += predict(step);
	return left;
};

export interface BulkProgress {
	stage: string;
	title: string;
	current?: number;
	total?: number;
	secondsRemaining?: number;
	/** Only the sim stages carry these; the candidate stages report a time estimate alone. */
	iterations?: { completed: number; total: number };
}

export interface CandidateGearProgressInput {
	completed?: number;
	total?: number;
	title?: string;
	stage?: string;
	startedAt?: number;
	now: number;
}

export const candidateGearProgress = ({
	completed,
	total,
	title = i18n.t('bulk_tab.progress.building_candidate_gear_sets'),
	stage = 'preparing',
	startedAt,
	now,
}: CandidateGearProgressInput): BulkProgress => {
	if (completed === undefined || total === undefined) return { stage, title };

	const secondsRemaining = startedAt !== undefined ? timeRemainingAtRate((now - startedAt) / 1000, completed, total) : undefined;
	return { stage, title, current: completed, total, secondsRemaining };
};

/** Null where there is no update to report: an unusable estimate leaves the last frame up. */
export const simProgress = (progress: ProgressMetrics, config: BulkSimProgressConfig, simStart: number, now: number): BulkProgress | null => {
	const stageCurrentRound = config.stageCurrentRound ?? config.currentRound;
	const stageRounds = config.stageRounds ?? config.totalRounds;
	const isBaselineRound = stageCurrentRound === 1;
	const totalElapsedSeconds = (now - (config.aggregateStartedAt ?? simStart)) / 1000;
	const completedIterations = config.aggregateCompletedIterations ?? progress.completedIterations;
	const totalIterations = config.aggregateTotalIterations ?? progress.totalIterations;
	const completedRoundsFromIterations = Math.max(
		0,
		config.aggregateTotalIterations && config.aggregateTotalIterations > 0
			? (completedIterations / config.aggregateTotalIterations) * stageRounds
			: stageCurrentRound - 1 + (progress.totalIterations > 0 ? progress.completedIterations / progress.totalIterations : 0),
	);
	const completedSimsFromIterations =
		config.useSimCountProgress && progress.totalSims > 0 && progress.totalIterations > 0
			? (progress.completedIterations / progress.totalIterations) * progress.totalSims
			: 0;
	const completedRounds =
		config.useSimCountProgress && progress.totalSims > 0 ? Math.max(progress.completedSims, completedSimsFromIterations) : completedRoundsFromIterations;
	const totalRounds = config.useSimCountProgress && progress.totalSims > 0 ? progress.totalSims : stageRounds;
	const secondsRemaining = timeRemainingAtRate(totalElapsedSeconds, completedRounds, totalRounds) ?? 0;

	if (isNaN(Number(secondsRemaining))) return null;

	return {
		stage: 'sim',
		title: config.title ?? (isBaselineRound ? i18n.t('bulk_tab.progress.baseline_round') : i18n.t('bulk_tab.progress.refining_rounds')),
		current: completedRounds,
		total: totalRounds,
		secondsRemaining,
		iterations: { completed: completedIterations, total: totalIterations },
	};
};
