import { useWowheadDataset } from '@sim/hooks/useWowheadDataset';
import type { Player } from '@sim/player/player';
import { equippedItemWowheadTooltipData } from '@sim/proto/action_id/tooltip_data';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { useMemo } from 'react';

export const useEquippedItemWowheadDataset = (player: Player<any>, item: EquippedItem | null | undefined) => {
	const resolve = useMemo(() => (item ? () => equippedItemWowheadTooltipData(player, item) : null), [player, item]);
	return useWowheadDataset(resolve);
};
