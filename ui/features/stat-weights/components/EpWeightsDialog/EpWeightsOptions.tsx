import i18n from '@i18n/config';
import { ChoiceField } from '@ui-kit/ChoiceField';

import { StatsType } from './types';

export interface EpWeightsOptionsProps {
	statsType: StatsType;
	onStatsTypeChange: (statsType: StatsType) => void;
}

export const EpWeightsOptions = ({ statsType, onStatsTypeChange }: EpWeightsOptionsProps) => (
	<ChoiceField
		id="ep-type-select"
		className="[&_.ui-field-label]:font-bold"
		label={i18n.t('sidebar.buttons.stat_weights.modal.show_as')}
		options={[
			{ value: StatsType.Ep, label: i18n.t('sidebar.buttons.stat_weights.modal.ep') },
			{ value: StatsType.Weight, label: i18n.t('sidebar.buttons.stat_weights.modal.weights') },
		]}
		value={statsType}
		onChange={onStatsTypeChange}
	/>
);
