import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { useWowheadDataset } from '@sim/hooks/useWowheadDataset';
import { equippedItemTooltipProps } from '@sim/proto/action_id/tooltip_data';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { useMemo } from 'react';

// The tooltip marks the set pieces worn in every slot, so it follows the whole gear, not only this item.
export const useEquippedItemWowheadDataset = (item: EquippedItem | null | undefined) => {
	const gear = usePlayerStore('gear');
	const resolve = useMemo(() => (item ? () => equippedItemTooltipProps(gear, item) : null), [gear, item]);
	return useWowheadDataset(resolve);
};
