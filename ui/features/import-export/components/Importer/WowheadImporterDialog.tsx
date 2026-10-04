import i18n from '@i18n/config';
import { WOWHEAD_GEAR_PLANNER_URL } from '@sim/proto/wowhead';
import { Button } from '@ui-kit/Button';

import { WOWHEAD_GEAR_PLANNER_IMPORTER } from '../../importers';
import { Importer } from './Importer';
import type { ImporterDialogProps } from './types';

export const WowheadImporterDialog = ({ open, onOpenChange }: ImporterDialogProps) => (
	<Importer open={open} onOpenChange={onOpenChange} {...WOWHEAD_GEAR_PLANNER_IMPORTER}>
		<>
			<p>
				{i18n.t('import.wowhead.description')}{' '}
				<Button href={WOWHEAD_GEAR_PLANNER_URL} target="_blank" variant="unstyled">
					{i18n.t('import.wowhead.gear_planner_link')}
				</Button>
				.
			</p>
			<p>{i18n.t('import.wowhead.feature_description')}</p>
			<p>{i18n.t('import.wowhead.instructions')}</p>
		</>
	</Importer>
);
