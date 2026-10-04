import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';

import { ADDON_IMPORTER } from '../../importers';
import { Importer } from './Importer';
import type { ImporterDialogProps } from './types';

export const AddonImporterDialog = ({ open, onOpenChange }: ImporterDialogProps) => (
	<Importer open={open} onOpenChange={onOpenChange} {...ADDON_IMPORTER}>
		<>
			<p>
				{i18n.t('import.addon.description')}{' '}
				<Button href="https://www.curseforge.com/wow/addons/wowsimsexporter" target="_blank" variant="unstyled">
					{i18n.t('import.addon.addon_link')}
				</Button>
				.
			</p>
			<p>{i18n.t('import.addon.feature_description')}</p>
			<p>{i18n.t('import.addon.instructions')}</p>
		</>
	</Importer>
);
