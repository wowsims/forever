import i18n from '@i18n/config';
import { useSim } from '@sim/context/SimHostContext';
import { formatToNumber } from '@sim/utils/format';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { Spinner } from '@ui-kit/Spinner';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { TooltipButton } from '@ui-kit/TooltipButton';
import clsx from 'clsx';
import { useId } from 'react';

import { useBulkMode } from '../../hooks/useBulkMode';
import { useBulkState } from '../../hooks/useBulkState';
import type { BulkBuildsCount } from '../../model/builds';
import { bulkIterationsLimit } from '../../model/limits';
import { bulkGearIterations } from '../../model/selectors';
import { bulkTopResults } from '../../model/top_results';

type CombinationsCountBuilds = Pick<BulkBuildsCount, 'setups' | 'bundles' | 'encounters' | 'runs' | 'builds' | 'totalIterations' | 'largestRequestIterations'>;

export interface CombinationsCountViewProps {
	pending: boolean;
	/** False until the sim is ready; Builds counts its gear setups from the database, so its numbers wait for it. */
	ready: boolean;
	combinations: number;
	iterations: number;
	iterationsLimit: number;
	/** Whether the limit warning ends with a hint to lower Top results; true while the Top results control shows. */
	topResultsHint?: boolean;
	/** Null in Gear mode, which counts the combinations and iterations of its one run. */
	builds: CombinationsCountBuilds | null;
	// Why the selection gives no combination, shown under the numbers.
	error?: string | null;
}

const untickedColumns = (builds: CombinationsCountBuilds) =>
	(
		[
			['builds', builds.bundles],
			['encounters', builds.encounters],
		] as const
	).flatMap(([column, count]) => (count === 0 ? [column] : []));

export const CombinationsCountView = ({
	pending,
	ready,
	combinations,
	iterations,
	iterationsLimit,
	topResultsHint = false,
	builds,
	error,
}: CombinationsCountViewProps) => {
	const tooltipId = useId();
	const countPending = builds ? !ready : pending;
	// A Builds batch sends its runs one at a time, so the limit is per run.
	const overLimit = builds ? builds.largestRequestIterations > iterationsLimit : !pending && iterations > iterationsLimit;
	const limitText = i18n.t('bulk_tab.warning.iterations_limit', { limit: formatToNumber(iterationsLimit) });
	const warningText = topResultsHint ? `${limitText} ${i18n.t('bulk_tab.warning.lower_top_results')}` : limitText;

	return (
		<dl className="mb-0 grid grid-cols-[1fr_auto] items-center gap-x-4 gap-y-2" data-testid="bulk-combinations-count">
			{builds ? (
				<>
					<dt className="font-normal">{i18n.t('bulk_tab.settings.gear_setups')}</dt>
					<dd className="mb-0 text-right font-bold tabular-nums" data-testid="bulk-gear-setups">
						{countPending ? <Spinner size="sm" /> : formatToNumber(builds.setups)}
					</dd>
				</>
			) : (
				<>
					<dt className="font-normal">{i18n.t('bulk_tab.settings.combinations')}</dt>
					<dd className="mb-0 text-right font-bold tabular-nums" data-testid="bulk-combinations">
						{pending ? <Spinner size="sm" /> : formatToNumber(combinations)}
					</dd>
				</>
			)}
			{builds && builds.setups > 0 && (
				<>
					<dt className="font-normal" data-testid="bulk-runs-label">
						{i18n.t('bulk_tab.settings.runs.builds', {
							bundles: i18n.t('bulk_tab.settings.runs.bundles', { count: builds.bundles }),
							encounters: i18n.t('bulk_tab.settings.runs.encounters', { count: builds.encounters }),
						})}
					</dt>
					<dd className="mb-0 text-right font-bold tabular-nums" data-testid="bulk-runs-count">
						{i18n.t('bulk_tab.settings.runs.total', { count: builds.runs, runs: formatToNumber(builds.runs) })}
					</dd>
				</>
			)}
			<dt className="flex items-center gap-1 font-normal">
				{i18n.t('bulk_tab.settings.total_iterations')}
				<TooltipButton
					tooltip={i18n.t('bulk_tab.settings.total_iterations_tooltip')}
					label={i18n.t('common.help')}
					testId="bulk-total-iterations-help"
				/>
			</dt>
			<dd
				className={clsx('mb-0 flex items-center justify-end gap-2 font-bold tabular-nums', overLimit && 'text-danger')}
				data-testid="bulk-total-iterations"
				data-over-limit={overLimit ? '' : undefined}>
				{overLimit && (
					<>
						<Button
							variant="unstyled"
							aria-label={warningText}
							className="text-link-warning"
							data-testid="warning"
							{...tooltipAnchorProps(tooltipId)}>
							<Icon name="triangle-exclamation" />
						</Button>
						<Tooltip id={tooltipId} place="left" content={warningText} />
					</>
				)}
				{countPending ? <Spinner size="sm" /> : formatToNumber(builds ? builds.totalIterations : iterations)}
			</dd>
			{builds && !countPending && builds.runs === 0 && (
				<dd className="col-span-2 mb-0 text-sm text-gray-400" data-testid="bulk-runs-hint">
					{untickedColumns(builds).map(column => (
						<p key={column} className="mb-0">
							{i18n.t('bulk_tab.settings.runs.tick', { column: i18n.t(`bulk_tab.builds.columns.${column}.title`) })}
						</p>
					))}
				</dd>
			)}
			{/* Simulate Batch waits for a second build to compare, as `canRunBuildsBatch` does. */}
			{builds && !countPending && builds.runs > 0 && builds.builds <= 1 && (
				<dd className="col-span-2 mb-0 text-sm text-gray-400" data-testid="bulk-runs-hint">
					{i18n.t('bulk_tab.settings.runs.second_build')}
				</dd>
			)}
			{!builds && !pending && error && (
				<dd className="col-span-2 mb-0 text-danger" role="alert" data-testid="bulk-combinations-error">
					{error}
				</dd>
			)}
		</dl>
	);
};

export const CombinationsCount = ({ builds, ready }: Pick<CombinationsCountViewProps, 'builds' | 'ready'>) => {
	const sim = useSim();
	const pending = useBulkState(slice => slice.combinationsPending);
	const combinations = useBulkState(slice => slice.combinations);
	const { powerTools } = useBulkMode();
	const iterations = useBulkState(slice => bulkGearIterations(slice, bulkTopResults(powerTools, slice)));
	const error = useBulkState(slice => slice.combinationsError);

	return (
		<CombinationsCountView
			pending={pending}
			ready={ready}
			combinations={combinations}
			iterations={iterations}
			iterationsLimit={bulkIterationsLimit(sim.isNative)}
			topResultsHint={powerTools}
			builds={builds}
			error={error}
		/>
	);
};
