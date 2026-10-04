import type { ItemSlot } from '@generated/proto/common';
import { translateSlotName } from '@i18n/localization';
import { usePlayer } from '@sim/context/SimHostContext';
import { useActionId } from '@sim/hooks/useActionId';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { WowheadAnchor } from '@ui-kit/WowheadAnchor';
import { useMemo } from 'react';

import { useEquippedItemWowheadDataset } from '../../hooks/useEquippedItemWowheadDataset';
import { getEmptySlotIconUrl } from '../../model/empty_slot_icons';

export interface SlotRailIconProps {
	slot: ItemSlot;
	item: EquippedItem | null;
	active: boolean;
	tooltipId: string;
	onOpen: () => void;
}

export const SlotRailIcon = ({ slot, item, active, tooltipId, onOpen }: SlotRailIconProps) => {
	const player = usePlayer();
	const actionId = useMemo(() => item?.asActionId(), [item]);
	const { iconUrl, href } = useActionId(actionId);

	const wowheadProps = useEquippedItemWowheadDataset(player, item);

	return (
		<div className="ui-slot-rail-icon-wrapper" data-testid="item-picker-icon-wrapper" data-active={active ? '' : undefined} data-slot={slot}>
			<WowheadAnchor
				className="ui-item-picker-icon"
				testId="item-picker-icon"
				href={href || undefined}
				onActivate={onOpen}
				data-tooltip-id={tooltipId}
				data-slot-label={translateSlotName(slot) ?? ''}
				style={{ backgroundImage: `url('${(item && iconUrl) || getEmptySlotIconUrl(slot)}')` }}
				{...wowheadProps}
			/>
		</div>
	);
};
