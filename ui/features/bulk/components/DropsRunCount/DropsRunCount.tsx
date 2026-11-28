import i18n from '@i18n/config';
import { usePlayer, useSim } from '@sim/context/SimHostContext';
import { formatToNumber } from '@sim/utils/format';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Notice } from '@ui-kit/Notice';
import { Spinner } from '@ui-kit/Spinner';

import { useBulkState } from '../../hooks/useBulkState';
import { setDropsCombinedGain } from '../../model/drops/settings';
import { bulkIterationsLimit } from '../../model/run/limits';
import { bulkDropsIterations } from '../../model/selectors';
import { TotalIterations } from '../TotalIterations';

export interface DropsRunCountViewProps {
	pending: boolean;
	/** False until the sim is ready; the picked count comes from the database, so the numbers wait for it. */
	ready: boolean;
	/** The items the Setup pane lists as included and not worn. */
	picked: number;
	/** The picks a worn two-hander skips, noted when above 0. */
	twoHanderSkipped: number;
	candidates: number;
	iterations: number;
	iterationsLimit: number;
	// Why the selection gives no candidate, shown under the numbers and apart from the notes.
	error: string | null;
	combinedGain: boolean;
	onCombinedGainChange: (on: boolean) => void;
}

const NOTES = ['one_change', 'slots', 'hands'] as const;

const runText = (key: string) => i18n.t(`bulk_tab.drops.run.${key}`);

export const DropsRunCountView = ({
	pending,
	ready,
	picked,
	twoHanderSkipped,
	candidates,
	iterations,
	iterationsLimit,
	error,
	combinedGain,
	onCombinedGainChange,
}: DropsRunCountViewProps) => {
	const countPending = pending || !ready;
	const overLimit = !countPending && iterations > iterationsLimit;
	const warningText = i18n.t('bulk_tab.warning.iterations_limit', { limit: formatToNumber(iterationsLimit) });
	const count = (testId: string, value: number) => (
		<dd className="mb-0 text-right font-bold tabular-nums" data-testid={testId}>
			{countPending ? <Spinner size="xs" /> : formatToNumber(value)}
		</dd>
	);

	return (
		<div className="grid gap-4" data-testid="drops-run-count">
			<dl className="mb-0 grid grid-cols-[1fr_auto] items-center gap-x-4 gap-y-2">
				<dt className="font-normal">{runText('items_picked')}</dt>
				{count('drops-run-picked', picked)}
				<dt className="font-normal">{runText('candidates')}</dt>
				{count('drops-run-candidates', candidates)}
				<TotalIterations pending={countPending} iterations={iterations} overLimit={overLimit} warningText={warningText} />
			</dl>
			<BooleanPicker
				id="drops-combined-gain"
				label={runText('combined_gain')}
				labelTooltip={runText('combined_gain_tooltip')}
				layout="inline"
				size="xs"
				testId="drops-combined-gain"
				value={combinedGain}
				onChange={onCombinedGainChange}
			/>
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
				{twoHanderSkipped > 0 && (
					<li data-testid="drops-run-two-hander-note">{i18n.t('bulk_tab.drops.run.two_hander_skipped', { count: twoHanderSkipped })}</li>
				)}
			</ul>
		</div>
	);
};

export const DropsRunCount = ({ ready, picked, twoHanderSkipped }: Pick<DropsRunCountViewProps, 'ready' | 'picked' | 'twoHanderSkipped'>) => {
	const sim = useSim();
	const player = usePlayer();
	const pending = useBulkState(slice => slice.combinationsPending);
	const candidates = useBulkState(slice => slice.combinations);
	const iterations = useBulkState(bulkDropsIterations);
	const error = useBulkState(slice => slice.combinationsError);
	const combinedGain = useBulkState(slice => slice.dropsCombinedGain);

	return (
		<DropsRunCountView
			pending={pending}
			ready={ready}
			picked={picked}
			twoHanderSkipped={twoHanderSkipped}
			candidates={candidates}
			iterations={iterations}
			iterationsLimit={bulkIterationsLimit(sim.isNative)}
			error={error}
			combinedGain={combinedGain}
			onCombinedGainChange={on => setDropsCombinedGain(player, on)}
		/>
	);
};
