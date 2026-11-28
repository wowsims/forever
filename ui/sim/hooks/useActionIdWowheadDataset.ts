import type { ActionId } from '@sim/proto/action_id';
import { actionIdTooltipProps } from '@sim/proto/action_id/tooltip_data';
import { useMemo } from 'react';

import { useWowheadDataset } from './useWowheadDataset';

export const useActionIdWowheadDataset = (actionId: ActionId | null | undefined, useBuffAura?: boolean) => {
	const resolve = useMemo(() => (actionId ? () => actionIdTooltipProps(actionId, { useBuffAura }) : null), [actionId, useBuffAura]);
	return useWowheadDataset(resolve);
};
