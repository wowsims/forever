import { Button } from '@ui-kit/Button';

import { SIXTY_UPGRADES_IMPORTER } from '../../importers';
import { Importer } from './Importer';
import type { ImporterDialogProps } from './types';

export const SixtyUpgradesImporterDialog = ({ open, onOpenChange }: ImporterDialogProps) => (
	<Importer open={open} onOpenChange={onOpenChange} {...SIXTY_UPGRADES_IMPORTER}>
		<>
			<p>
				Import settings from{' '}
				<Button href="https://sixtyupgrades.com/forever" target="_blank" variant="unstyled">
					Sixty Upgrades
				</Button>
				.
			</p>
			<p>This feature imports gear, race, and (optionally) talents. It does NOT import buffs, debuffs, consumes, rotation, or custom stats.</p>
			<p>To import, paste the output from the site's export option below and click, 'Import'.</p>
		</>
	</Importer>
);
