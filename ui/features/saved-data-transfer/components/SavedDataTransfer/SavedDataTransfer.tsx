import i18n from '@i18n/config';
import { useSimReady } from '@sim/hooks/useSimReady';
import { useUiStore } from '@sim/hooks/useUiStore';
import { Button } from '@ui-kit/Button';
import type { UseSavedDataResult } from '@ui-kit/hooks/useSavedData';
import { Icon } from '@ui-kit/Icon';
import { Menu, MenuItem } from '@ui-kit/Menu';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { useId, useState } from 'react';

import { trackEvent } from '../../../../tracking/utils';
import type { SavedDataType } from '../../model/transfer_text';
import { SavedDataExporter } from '../SavedDataExporter';
import { SavedDataImporter } from '../SavedDataImporter';

export interface SavedDataTransferProps<T> {
	type: SavedDataType;
	label: string;
	storage: UseSavedDataResult<T>;
}

type Open = 'export' | 'import' | null;

export const SavedDataTransfer = <T,>({ type, label, storage }: SavedDataTransferProps<T>) => {
	const shown = useUiStore('showPowerTools');
	// Both dialogs read the player, the presets and the database.
	const ready = useSimReady();
	const tooltipId = useId();
	const [open, setOpen] = useState<Open>(null);
	if (!shown) return null;

	const openChange = (dialog: Open) => (next: boolean) => setOpen(next ? dialog : null);
	const menuLabel = i18n.t('saved_data_transfer.menu_label');
	const nothingSaved = storage.entries.length === 0;
	const exportLabel = nothingSaved
		? i18n.t('saved_data_transfer.export_empty', { type: i18n.t(`saved_data_transfer.types.${type}`) })
		: i18n.t('saved_data_transfer.export');

	return (
		<>
			<Menu
				surface="menu"
				trigger={<Icon name="gear" />}
				triggerRender={<Button variant="link" size="sm" className="ml-auto text-muted hover:text-link-hover" />}
				triggerProps={{ 'aria-label': menuLabel, 'data-testid': 'saved-data-transfer-button', disabled: !ready, ...tooltipAnchorProps(tooltipId) }}
				side="bottom"
				align="end">
				<MenuItem
					layout="block"
					testId="saved-data-transfer-export"
					disabled={nothingSaved}
					onClick={() => {
						setOpen('export');
						trackEvent({ action: 'settings', category: 'export', label });
					}}>
					{exportLabel}
				</MenuItem>
				<MenuItem layout="block" testId="saved-data-transfer-import" onClick={() => setOpen('import')}>
					{i18n.t('saved_data_transfer.import')}
				</MenuItem>
			</Menu>
			<Tooltip id={tooltipId} content={menuLabel} />
			{/* Outside `Menu`, not inside its popup: clicking an item closes the menu, which unmounts the popup, and a dialog rendered in there would go with it. */}
			<SavedDataExporter type={type} storage={storage} open={open === 'export'} onOpenChange={openChange('export')} />
			<SavedDataImporter type={type} label={label} storage={storage} open={open === 'import'} onOpenChange={openChange('import')} />
		</>
	);
};
