import i18n from '@i18n/config';
import { useSim } from '@sim/context/SimHostContext';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { Button } from '@ui-kit/Button';
import { ButtonGroup } from '@ui-kit/ButtonGroup';
import { Icon } from '@ui-kit/Icon';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import clsx from 'clsx';
import { useId } from 'react';

import { useResultFreshness } from '../../hooks/useResultFreshness';
import { type ReferenceDiffs, referenceDiffs as computeReferenceDiffs } from '../../model/reference_diffs';
import type { ResultFreshness } from '../../model/result_freshness';
import type { SimResultsManager } from '../../model/results_manager';
import { type ResultMetric, toplineResultMetrics } from '../../model/topline_metrics';
import { ResultMetricList } from '../ResultMetricList';

export interface SimResultSummaryViewProps {
	metrics: Array<ResultMetric>;
	/** Set exactly while a reference run is saved; each metric's delta against it. */
	referenceDiffs?: ReferenceDiffs;
	/** Whether the sim's inputs changed since the run that produced these metrics started. */
	freshness?: ResultFreshness;
	/** These metrics came in through a swap: they are the reference run's, not the last run's. */
	swapped?: boolean;
	onSetReference: () => void;
	onSwapReference: () => void;
	onClearReference: () => void;
	onOpenBreakdown?: () => void;
}

const FRESHNESS_TEXT: Record<NonNullable<ResultFreshness>, string> = {
	fresh: 'text-success',
	stale: 'text-link-warning',
};

/** The finished run: the primary metric large, the others compact under it, then the reference actions and the way to the full breakdown. */
export const SimResultSummaryView = ({
	metrics,
	referenceDiffs,
	freshness,
	swapped = false,
	onSetReference,
	onSwapReference,
	onClearReference,
	onOpenBreakdown,
}: SimResultSummaryViewProps) => {
	const tooltipId = useId();
	const hasReference = referenceDiffs !== undefined;
	const [primary, ...secondary] = metrics;

	return (
		<div data-testid="results-sim" className="flex w-full flex-col gap-2">
			<div className="flex items-center justify-between gap-2 text-sm" data-testid="results-last-run-header">
				<span data-testid="results-run-title" data-run={swapped ? 'reference' : 'last'}>
					{i18n.t(swapped ? 'sidebar.results.reference_run' : 'sidebar.results.last_run')}
				</span>
				{freshness && (
					<span className={clsx('flex items-center gap-2', FRESHNESS_TEXT[freshness])} data-testid="results-freshness" data-freshness={freshness}>
						<span className="size-2 rounded-full bg-current" aria-hidden="true" />
						{i18n.t(`sidebar.results.freshness.${freshness}`)}
					</span>
				)}
			</div>
			{/* One flex item per list, so the rows inside keep their own spacing instead of sharing this gap. */}
			<div className="flex flex-col items-start gap-2">
				{primary && <ResultMetricList metrics={[primary]} layout="list" referenceDiffs={referenceDiffs} />}
				{secondary.length > 0 && <ResultMetricList metrics={secondary} layout="list" size="sm" referenceDiffs={referenceDiffs} />}
			</div>
			<div
				className="flex flex-wrap items-center justify-between gap-2 font-normal"
				data-testid="results-sim-reference"
				data-has-reference={hasReference ? '' : undefined}>
				<Button
					variant="outline-primary"
					size="sm"
					hidden={hasReference}
					data-testid="results-sim-set-reference"
					onClick={onSetReference}
					{...tooltipAnchorProps(tooltipId, i18n.t('sidebar.results.reference.use_as_reference'))}>
					<Icon name="map-pin" style="base" />
					{i18n.t('sidebar.results.reference.save_as_reference')}
				</Button>
				<ButtonGroup size="sm" hidden={!hasReference} testId="results-sim-reference-actions">
					<Button
						variant="outline-primary"
						data-testid="results-sim-reference-swap"
						onClick={onSwapReference}
						{...tooltipAnchorProps(tooltipId, i18n.t('sidebar.results.reference.swap_reference_with_current'))}>
						<Icon name="arrows-rotate" />
						{i18n.t('sidebar.results.reference.swap')}
					</Button>
					<Button
						variant="outline-primary"
						data-testid="results-sim-reference-delete"
						onClick={onClearReference}
						{...tooltipAnchorProps(tooltipId, i18n.t('sidebar.results.reference.remove_reference'))}>
						<Icon name="xmark" style="base" />
						{i18n.t('sidebar.results.reference.unset')}
					</Button>
				</ButtonGroup>
				{onOpenBreakdown && (
					<Button variant="link" size="sm" data-testid="results-open-breakdown" onClick={onOpenBreakdown}>
						{i18n.t('results_tab.title')}
						<Icon name="arrow-right" />
					</Button>
				)}
			</div>
			<Tooltip id={tooltipId} />
		</div>
	);
};

export interface SimResultSummaryProps {
	results: SimResultsManager;
	onOpenBreakdown?: () => void;
}

/** Renders nothing until a run completes, so the panel's result zone is empty at load. */
export const SimResultSummary = ({ results, onOpenBreakdown }: SimResultSummaryProps) => {
	const displayMetrics = useDisplayMetrics(useSim());
	const { current, reference, swapped } = useStoreSubscribe(results.subscribe, results.getData);
	const freshness = useResultFreshness();

	if (!current) return null;

	return (
		<SimResultSummaryView
			metrics={toplineResultMetrics(current.simResult, undefined, { displayMetrics })}
			referenceDiffs={reference ? computeReferenceDiffs(current.simResult, reference.simResult) : undefined}
			freshness={freshness}
			swapped={swapped}
			onOpenBreakdown={onOpenBreakdown}
			onSetReference={() => results.setReference()}
			onSwapReference={() => results.swapReference()}
			onClearReference={() => results.clearReference()}
		/>
	);
};
