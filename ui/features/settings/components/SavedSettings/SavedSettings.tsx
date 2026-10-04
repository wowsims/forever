import { useSavedPanel } from '@features/hooks/useSavedPanel';
import { readSavedSettings } from '@features/settings/model/saved_settings';
import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { useReadyStoreSubscribe } from '@sim/hooks/useReadyStoreSubscribe';
import { useSimReady } from '@sim/hooks/useSimReady';
import { subscribeAll, subscribePartyBuffs, subscribePlayerField, subscribeRaidField } from '@sim/state/subscriptions';
import { SavedDataPanel } from '@ui-kit/SavedDataPanel';

import { useSavedSettings } from '../../hooks/useSavedSettings';
import { loadSettings, serializeSettings, settingsPresets } from './utils';

export const SavedSettings = () => {
	const host = useSimHost();
	const ready = useSimReady();

	const settings = useReadyStoreSubscribe(
		subscribeAll([
			subscribeRaidField(host.sim.raid, 'buffs'),
			subscribeRaidField(host.sim.raid, 'debuffs'),
			subscribePartyBuffs(host.player.getParty()!),
			subscribePlayerField(host.player, 'buffs'),
			subscribePlayerField(host.player, 'consumables'),
			subscribePlayerField(host.player, 'race'),
			subscribePlayerField(host.player, 'profession1'),
			subscribePlayerField(host.player, 'profession2'),
			subscribePlayerField(host.player, 'itemSwap'),
			subscribePlayerField(host.player, 'reactionTime'),
			subscribePlayerField(host.player, 'channelClipDelay'),
			subscribePlayerField(host.player, 'inFrontOfTarget'),
			subscribePlayerField(host.player, 'distanceFromTarget'),
			subscribePlayerField(host.player, 'healingModel'),
		]),
		// Reaches `player.getConsumes()`, so it dereferences a null `sim.db` before the sim is ready.
		() => readSavedSettings(host),
		ready,
	);

	const panel = useSavedPanel({
		label: i18n.t('settings_tab.saved_settings.settings'),
		storage: useSavedSettings(),
		current: settings,
		serialize: serializeSettings,
		load: loadSettings,
		presets: settingsPresets,
	});

	return (
		<SavedDataPanel
			title={i18n.t('settings_tab.saved_settings.title')}
			nameLabel={i18n.t('settings_tab.saved_settings.settings_name')}
			saveButtonText={i18n.t('settings_tab.saved_settings.save_settings')}
			{...panel}
		/>
	);
};
