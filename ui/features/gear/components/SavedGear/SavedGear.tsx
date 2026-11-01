import { useSavedPanel } from '@features/hooks/useSavedPanel';
import { SavedDataTransfer } from '@features/saved-data-transfer/components/SavedDataTransfer';
import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { subscribePlayerChange } from '@sim/state/subscriptions';
import { SavedDataPanel } from '@ui-kit/SavedDataPanel';

import { useSavedGear } from '../../hooks/useSavedGear';
import { gearSetData, gearSetPresets, loadGearSet, serializeGearSet } from './utils';

export const SavedGear = () => {
	const { player } = useSimHost();
	const storage = useSavedGear();
	const panel = useSavedPanel({
		label: i18n.t('gear_tab.gear_sets.gear_set'),
		storage,
		current: useStoreSubscribe(subscribePlayerChange(player), () => gearSetData(player)),
		serialize: serializeGearSet,
		load: loadGearSet,
		presets: gearSetPresets,
	});

	return (
		<SavedDataPanel
			title={i18n.t('gear_tab.gear_sets.title')}
			nameLabel={i18n.t('gear_tab.gear_sets.gear_set_name')}
			saveButtonText={i18n.t('gear_tab.gear_sets.save_gear_set')}
			headerChildren={<SavedDataTransfer type="gear" label={panel.label} storage={storage} />}
			{...panel}
		/>
	);
};
