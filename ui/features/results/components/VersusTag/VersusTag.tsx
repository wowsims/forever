import type { VersusLog } from '@sim/proto/combat_log';
import { Pill } from '@ui-kit/Pill';

import { versusTargetName } from '../../model/versus';

export interface VersusTagProps {
	versus: VersusLog;
}

/** The target kind a hit that took a versus bonus is tagged with. */
export const VersusTag = ({ versus }: VersusTagProps) => (
	<Pill tone="warning" size="xs" rounded="xs" testId="versus-tag">
		{versusTargetName(versus)}
	</Pill>
);
