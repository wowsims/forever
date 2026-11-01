import { useSavedPanel } from '@features/hooks/useSavedPanel';
import { SavedDataTransfer } from '@features/saved-data-transfer/components/SavedDataTransfer';
import { SavedRotation as SavedRotationProto } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { isEqualAPLRotation } from '@sim/proto/apl_utils';
import { subscribeAll, subscribePlayerField } from '@sim/state/subscriptions';
import type { SavedDataPanelEntry } from '@ui-kit/SavedDataPanel';
import { SavedDataPanel } from '@ui-kit/SavedDataPanel';
import { useCallback } from 'react';

import { useSavedRotation } from '../../hooks/useSavedRotation';
import { loadRotation, rotationPresets, serializeRotation } from './utils';

export const SavedRotation = () => {
	const player = usePlayer();
	const current = useStoreSubscribe(subscribeAll([subscribePlayerField(player, 'rotation'), subscribePlayerField(player, 'talentsString')]), () =>
		SavedRotationProto.create({ rotation: player.getResolvedAplRotation() }),
	);

	// An Auto and an APL rotation can serialise differently and still be the same rotation, so this
	// slot is the one that needs the panel's semantic override rather than its JSON comparison.
	const isActive = useCallback(
		(entry: SavedDataPanelEntry<SavedRotationProto>) => isEqualAPLRotation(player, current.rotation, entry.data.rotation),
		[player, current],
	);

	const storage = useSavedRotation();
	const panel = useSavedPanel({
		label: i18n.t('rotation_tab.saved_rotations.label'),
		storage,
		current,
		serialize: serializeRotation,
		load: loadRotation,
		presets: rotationPresets,
	});

	return (
		<SavedDataPanel
			title={i18n.t('rotation_tab.saved_rotations.title')}
			nameLabel={i18n.t('rotation_tab.saved_rotations.name_label')}
			saveButtonText={i18n.t('rotation_tab.saved_rotations.save_button')}
			deleteLabel={i18n.t('rotation_tab.saved_rotations.delete.tooltip')}
			deleteConfirmMessage={i18n.t('rotation_tab.saved_rotations.delete.confirm')}
			chooseNameAlert={i18n.t('rotation_tab.saved_rotations.alerts.choose_name')}
			nameExistsAlert={i18n.t('rotation_tab.saved_rotations.alerts.name_exists')}
			isActive={isActive}
			headerChildren={<SavedDataTransfer type="rotation" label={panel.label} storage={storage} presets={panel.presets} />}
			{...panel}
		/>
	);
};
