import i18n from '@i18n/config';
import { useSim } from '@sim/context/SimHostContext';
import { formatToNumber } from '@sim/utils/format';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { Spinner } from '@ui-kit/Spinner';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { TooltipButton } from '@ui-kit/TooltipButton';
import clsx from 'clsx';
import { useId, useMemo } from 'react';

import { useBulkEncounters } from '../../hooks/useBulkEncounters';
import { useBulkState } from '../../hooks/useBulkState';
import { useBulkTalents } from '../../hooks/useBulkTalents';
import { bulkEncounterRuns } from '../../model/encounters';
import { bulkIterationsLimit } from '../../model/limits';
import { bulkTalentRuns } from '../../model/talents';

export interface CombinationsCountViewProps {
	pending: boolean;
	combinations: number;
	iterations: number;
	iterationsLimit: number;
	encounters: number;
	talentSetups: number;
}

export const CombinationsCountView = ({ pending, combinations, iterations, iterationsLimit, encounters, talentSetups }: CombinationsCountViewProps) => {
	const tooltipId = useId();
	const runs = encounters * talentSetups;
	const overLimit = !pending && iterations > iterationsLimit;
	const warningText = i18n.t('bulk_tab.warning.iterations_limit', { limit: formatToNumber(iterationsLimit) });

	return (
		<dl className="mb-0 grid grid-cols-[1fr_auto] items-center gap-x-4 gap-y-2" data-testid="bulk-combinations-count">
			<dt className="font-normal">{i18n.t('bulk_tab.settings.combinations')}</dt>
			<dd className="mb-0 text-right font-bold tabular-nums" data-testid="bulk-combinations">
				{pending ? <Spinner size="sm" /> : formatToNumber(combinations)}
			</dd>
			{runs > 1 && (
				<>
					<dt className="font-normal">
						{`${i18n.t('bulk_tab.settings.runs.encounters', { count: encounters })} × ${i18n.t('bulk_tab.settings.runs.talents', { count: talentSetups })}`}
					</dt>
					<dd className="mb-0 text-right font-bold tabular-nums" data-testid="bulk-runs-count">
						{i18n.t('bulk_tab.settings.runs.total', { count: runs, runs: formatToNumber(runs) })}
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
				{pending ? <Spinner size="sm" /> : formatToNumber(iterations * runs)}
			</dd>
		</dl>
	);
};

export const CombinationsCount = () => {
	const sim = useSim();
	const pending = useBulkState(slice => slice.combinationsPending);
	const combinations = useBulkState(slice => slice.combinations);
	const iterations = useBulkState(slice => slice.iterations);
	const { current, entries, checked, powerTools } = useBulkEncounters();
	const encounters = useMemo(() => bulkEncounterRuns(current, entries, checked, powerTools).length, [current, entries, checked, powerTools]);
	const { playerSpec, current: currentTalents, entries: talentEntries, checked: checkedTalents } = useBulkTalents();
	const talentSetups = useMemo(
		() => bulkTalentRuns(playerSpec, currentTalents, talentEntries, checkedTalents, powerTools).length,
		[playerSpec, currentTalents, talentEntries, checkedTalents, powerTools],
	);

	return (
		<CombinationsCountView
			pending={pending}
			combinations={combinations}
			iterations={iterations}
			iterationsLimit={bulkIterationsLimit(sim.isNative)}
			encounters={encounters}
			talentSetups={talentSetups}
		/>
	);
};
