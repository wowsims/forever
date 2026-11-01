import { useSavedPanel } from '@features/hooks/useSavedPanel';
import { SavedDataTransfer } from '@features/saved-data-transfer/components/SavedDataTransfer';
import type { SavedSettings as SavedSettingsProto } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { withLiveSpecOptions } from '@sim/proto/spec_functions';
import type { SavedDataPanelEntry } from '@ui-kit/SavedDataPanel';
import { SavedDataPanel } from '@ui-kit/SavedDataPanel';
import { useCallback } from 'react';

import { useCurrentSettings } from '../../hooks/useCurrentSettings';
import { useSavedSettings } from '../../hooks/useSavedSettings';
import { loadSettings, serializeSettings, settingsPresets } from './utils';

export const SavedSettings = () => {
	const settings = useCurrentSettings();
	const storage = useSavedSettings();

	const panel = useSavedPanel({
		label: i18n.t('settings_tab.saved_settings.settings'),
		storage,
		current: settings,
		serialize: serializeSettings,
		load: loadSettings,
		presets: settingsPresets,
	});

	// An entry without spec options for the live spec matches on everything else, as loading it keeps the live ones.
	const { currentJson } = panel;
	const isActive = useCallback(
		(entry: SavedDataPanelEntry<SavedSettingsProto>) => settings !== null && serializeSettings(withLiveSpecOptions(entry.data, settings)) === currentJson,
		[settings, currentJson],
	);

	return (
		<SavedDataPanel
			title={i18n.t('settings_tab.saved_settings.title')}
			nameLabel={i18n.t('settings_tab.saved_settings.settings_name')}
			saveButtonText={i18n.t('settings_tab.saved_settings.save_settings')}
			isActive={isActive}
			headerChildren={<SavedDataTransfer type="settings" label={panel.label} storage={storage} />}
			{...panel}
		/>
	);
};
