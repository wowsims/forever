import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';

import { SIXTY_UPGRADES_IMPORTER } from '../../importers';
import { Importer } from './Importer';
import type { ImporterDialogProps } from './types';

export const SixtyUpgradesImporterDialog = ({ open, onOpenChange }: ImporterDialogProps) => (
	<Importer open={open} onOpenChange={onOpenChange} {...SIXTY_UPGRADES_IMPORTER}>
		<>
			<p>
				{i18n.t('import.sixty_upgrades.description')}{' '}
				<Button href="https://sixtyupgrades.com/forever" target="_blank" variant="unstyled">
					Sixty Upgrades
				</Button>
				.
			</p>
			<p>{i18n.t('import.sixty_upgrades.feature_description')}</p>
			<p>{i18n.t('import.sixty_upgrades.instructions')}</p>
		</>
	</Importer>
);
