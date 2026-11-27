import { useWowheadDataset } from '@sim/hooks/useWowheadDataset';
import { equippedItemTooltipProps } from '@sim/proto/action_id/tooltip_data';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { useMemo } from 'react';

export const useEquippedItemWowheadDataset = (item: EquippedItem | null | undefined) => {
	const resolve = useMemo(() => (item ? () => equippedItemTooltipProps(item) : null), [item]);
	return useWowheadDataset(resolve);
};
