import { useSavedPanel } from '@features/hooks/useSavedPanel';
import type { SavedEPWeights } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import type { SavedDataPanelEntry } from '@ui-kit/SavedDataPanel';
import { SavedDataPanel } from '@ui-kit/SavedDataPanel';
import type { ClassValue } from 'clsx';
import { useMemo } from 'react';

import { useSavedEpWeights } from '../../hooks/useSavedEpWeights';
import { epWeightsData, epWeightsPresets, loadEpWeights, serializeEpWeights } from './utils';

const NO_USER_DATA: Array<SavedDataPanelEntry<SavedEPWeights>> = [];

export interface SavedEpWeightsProps {
	className?: ClassValue;
	loadOnly?: boolean;
	// Presets only, which also means load only: the user's own saved sets are neither listed nor written.
	presetsOnly?: boolean;
}

export const SavedEpWeights = ({ className, loadOnly, presetsOnly }: SavedEpWeightsProps) => {
	const epWeights = usePlayerStore('epWeights');
	const panel = useSavedPanel({
		label: i18n.t('sidebar.buttons.stat_weights.modal.ep'),
		storage: useSavedEpWeights(),
		current: useMemo(() => epWeightsData(epWeights), [epWeights]),
		serialize: serializeEpWeights,
		load: loadEpWeights,
		presets: epWeightsPresets,
	});

	return (
		<SavedDataPanel
			className={className}
			title={i18n.t('sidebar.buttons.stat_weights.saved_ep_weights.title')}
			nameLabel={i18n.t('sidebar.buttons.stat_weights.title')}
			loadOnly={loadOnly || presetsOnly}
			{...panel}
			userData={presetsOnly ? NO_USER_DATA : panel.userData}
		/>
	);
};
