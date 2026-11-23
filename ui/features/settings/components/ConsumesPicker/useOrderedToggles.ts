import { useStoreField } from '@sim/hooks/useStoreField';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import type { Player } from '@sim/player/player';
import { CONSUMABLE_ORDER_FIELDS, consumableOrderOf, orderedByEp } from '@sim/settings/consumables';
import { useMemo } from 'react';

import type { ConsumeToggle } from './utils';

// The toggles by the EP their consumables add at the player's weights, again whenever the order's fields change.
export const useOrderedToggles = (player: Player<any>, toggles: ReadonlyArray<ConsumeToggle>) => {
	const order = useStoreSubscribe(useStoreField(CONSUMABLE_ORDER_FIELDS)!, () => consumableOrderOf(player));
	return useMemo(() => orderedByEp(toggles, toggle => toggle.consumable, order), [toggles, order]);
};
