import i18n from '@i18n/config';
import { ToggleGroup } from '@ui-kit/ToggleGroup';

import type { ChartView } from './utils';
import { CHART_VIEWS } from './utils';

export interface ChartViewPickerProps {
	value: ChartView;
	onChange: (next: ChartView) => void;
	className?: string;
}

/** Rotation or the DPS/resources chart. The two are alternatives, and the rotation is the default. */
export const ChartViewPicker = ({ value, onChange, className }: ChartViewPickerProps) => (
	<ToggleGroup
		className={className}
		size="sm"
		value={value}
		onChange={onChange}
		options={CHART_VIEWS.map(view => ({ value: view, label: i18n.t(`results_tab.details.timeline.chart_types.${view}`) }))}
	/>
);
