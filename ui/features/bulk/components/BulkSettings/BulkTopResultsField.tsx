import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { BULK_SIM_DEFAULT_TOP_RESULTS } from '@sim/wasm/bulk_sim/constants_auto_gen';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { NumberField } from '@ui-kit/TextField';

import { useBulkMode } from '../../hooks/useBulkMode';
import { useBulkState } from '../../hooks/useBulkState';
import { bulkCombinationsLimit } from '../../model/run/limits';
import { setBulkTopResults } from '../../model/run/top_results';

export interface BulkTopResultsFieldViewProps {
	/** 0 is All. */
	value: number;
	/** The largest number the field takes. */
	max: number;
	onChange: (topResults: number) => void;
}

// The number field holds 1 up to `max`, which also keeps it inside the request's int32; unticking All starts again from 5.
export const BulkTopResultsFieldView = ({ value, max, onChange }: BulkTopResultsFieldViewProps) => {
	const all = value === 0;
	return (
		<div className="flex items-end gap-4" data-testid="bulk-top-results">
			<NumberField
				id="bulk-top-results"
				label={i18n.t('bulk_tab.settings.top_results.label')}
				labelTooltip={i18n.t('bulk_tab.settings.top_results.tooltip')}
				positive
				showZeroes={false}
				disabled={all}
				value={value}
				onChange={entered => onChange(Math.min(max, Math.max(1, entered)))}
			/>
			<BooleanPicker
				id="bulk-top-results-all"
				label={i18n.t('bulk_tab.settings.top_results.all')}
				layout="inline"
				value={all}
				onChange={checked => onChange(checked ? 0 : BULK_SIM_DEFAULT_TOP_RESULTS)}
			/>
		</div>
	);
};

// The running mode's own cap, at most the combinations limit.
export const BulkTopResultsField = () => {
	const player = usePlayer();
	const { builds } = useBulkMode();
	const value = useBulkState(slice => (builds ? slice.buildsTopResults : slice.gearTopResults));
	return (
		<BulkTopResultsFieldView
			value={value}
			max={bulkCombinationsLimit(player.sim.isNative)}
			onChange={topResults => setBulkTopResults(player, builds ? 'builds' : 'gear', topResults)}
		/>
	);
};
