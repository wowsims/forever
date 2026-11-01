import { useSavedPanel } from '@features/hooks/useSavedPanel';
import { SavedDataTransfer } from '@features/saved-data-transfer/components/SavedDataTransfer';
import i18n from '@i18n/config';
import { SavedDataPanel } from '@ui-kit/SavedDataPanel';

import { useSavedTalents } from '../../hooks/useSavedTalents';
import { useTalents } from '../../hooks/useTalents';
import { loadTalents, serializeTalents, talentPresets } from './utils';

export const SavedTalents = () => {
	const storage = useSavedTalents();
	const panel = useSavedPanel({
		label: i18n.t('talents_tab.saved_talents.label'),
		storage,
		current: useTalents(),
		serialize: serializeTalents,
		load: loadTalents,
		presets: talentPresets,
	});

	return (
		<SavedDataPanel
			title={i18n.t('talents_tab.saved_talents.title')}
			nameLabel={i18n.t('talents_tab.saved_talents.name_label')}
			saveButtonText={i18n.t('talents_tab.saved_talents.save_button')}
			deleteLabel={i18n.t('talents_tab.saved_talents.delete.tooltip')}
			deleteConfirmMessage={i18n.t('talents_tab.saved_talents.delete.confirm')}
			chooseNameAlert={i18n.t('talents_tab.saved_talents.alerts.choose_name')}
			nameExistsAlert={i18n.t('talents_tab.saved_talents.alerts.name_exists')}
			headerChildren={<SavedDataTransfer type="talents" label={panel.label} storage={storage} />}
			{...panel}
		/>
	);
};
