import { useSavedPanel } from '@features/hooks/useSavedPanel';
import i18n from '@i18n/config';
import { useSim } from '@sim/context/SimHostContext';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { subscribeEncounterChange } from '@sim/state/subscriptions';
import { SavedDataPanel } from '@ui-kit/SavedDataPanel';

import { useSavedEncounter } from '../../hooks/useSavedEncounter';
import { encounterData, encounterPresets, loadEncounter, serializeEncounter } from './utils';

export const SavedEncounter = () => {
	const { encounter } = useSim();
	const panel = useSavedPanel({
		label: i18n.t('settings_tab.saved_encounters.encounter'),
		storage: useSavedEncounter(),
		current: useStoreSubscribe(subscribeEncounterChange(encounter), () => encounterData(encounter)),
		serialize: serializeEncounter,
		load: loadEncounter,
		presets: encounterPresets,
	});

	return (
		<SavedDataPanel
			title={i18n.t('settings_tab.saved_encounters.title')}
			nameLabel={i18n.t('settings_tab.saved_encounters.encounter_name')}
			saveButtonText={i18n.t('settings_tab.saved_encounters.save_encounter')}
			{...panel}
		/>
	);
};
