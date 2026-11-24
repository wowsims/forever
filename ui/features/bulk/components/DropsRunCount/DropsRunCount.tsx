import i18n from '@i18n/config';
import { useSim } from '@sim/context/SimHostContext';
import { formatToNumber } from '@sim/utils/format';
import { Notice } from '@ui-kit/Notice';
import { Spinner } from '@ui-kit/Spinner';

import { useBulkState } from '../../hooks/useBulkState';
import { bulkIterationsLimit } from '../../model/run/limits';
import { bulkDropsIterations } from '../../model/selectors';
import { TotalIterations } from '../TotalIterations';

export interface DropsRunCountViewProps {
	pending: boolean;
	/** False until the sim is ready; the picked count comes from the database, so the numbers wait for it. */
	ready: boolean;
	/** The items the Setup pane lists as included and not worn. */
	picked: number;
	candidates: number;
	iterations: number;
	iterationsLimit: number;
	// Why the selection gives no candidate, shown under the numbers and apart from the notes.
	error: string | null;
}

const NOTES = ['one_change', 'slots', 'hands', 'no_drop_chances'] as const;

const runText = (key: string) => i18n.t(`bulk_tab.drops.run.${key}`);

export const DropsRunCountView = ({ pending, ready, picked, candidates, iterations, iterationsLimit, error }: DropsRunCountViewProps) => {
	const countPending = pending || !ready;
	const overLimit = !countPending && iterations > iterationsLimit;
	const warningText = i18n.t('bulk_tab.warning.iterations_limit', { limit: formatToNumber(iterationsLimit) });
	const count = (testId: string, value: number) => (
		<dd className="mb-0 text-right font-bold tabular-nums" data-testid={testId}>
			{countPending ? <Spinner size="sm" /> : formatToNumber(value)}
		</dd>
	);

	return (
		<div className="grid gap-4" data-testid="drops-run-count">
			<dl className="mb-0 grid grid-cols-[1fr_auto] items-center gap-x-4 gap-y-2">
				<dt className="font-normal">{runText('items_picked')}</dt>
				{count('drops-run-picked', picked)}
				<dt className="font-normal">{runText('candidates')}</dt>
				{count('drops-run-candidates', candidates)}
				<dt className="font-normal">{runText('variants')}</dt>
				{count('drops-run-variants', Math.max(0, candidates - picked))}
				<TotalIterations pending={countPending} iterations={iterations} overLimit={overLimit} warningText={warningText} />
			</dl>
			{!countPending && error && (
				<div role="alert">
					<Notice level="error" className="text-sm text-danger" testId="bulk-combinations-error">
						{error}
					</Notice>
				</div>
			)}
			<ul className="mb-0 grid list-disc gap-1 border-t border-border ps-4 pt-4 text-sm text-gray-400">
				{NOTES.map(note => (
					<li key={note} data-testid="drops-run-note">
						{runText(`notes.${note}`)}
					</li>
				))}
			</ul>
		</div>
	);
};

export const DropsRunCount = ({ ready, picked }: Pick<DropsRunCountViewProps, 'ready' | 'picked'>) => {
	const sim = useSim();
	const pending = useBulkState(slice => slice.combinationsPending);
	const candidates = useBulkState(slice => slice.combinations);
	const iterations = useBulkState(bulkDropsIterations);
	const error = useBulkState(slice => slice.combinationsError);

	return (
		<DropsRunCountView
			pending={pending}
			ready={ready}
			picked={picked}
			candidates={candidates}
			iterations={iterations}
			iterationsLimit={bulkIterationsLimit(sim.isNative)}
			error={error}
		/>
	);
};
