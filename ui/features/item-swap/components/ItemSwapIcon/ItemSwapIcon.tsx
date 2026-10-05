import { useEquippedItemWowheadDataset } from '@features/gear/hooks/useEquippedItemWowheadDataset';
import { useOpenSelectorModal } from '@features/gear/hooks/useSelectorModal';
import { getEmptySlotIconUrl } from '@features/gear/model/empty_slot_icons';
import { SelectorModalTabs } from '@features/gear/types';
import type { ItemSlot } from '@generated/proto/common';
import { translateSlotName } from '@i18n/localization';
import { useSimHost } from '@sim/context/SimHostContext';
import { useActionId } from '@sim/hooks/useActionId';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import type { WowheadDatasetProps } from '@sim/hooks/useWowheadDataset';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { WowheadAnchor } from '@ui-kit/WowheadAnchor';
import clsx from 'clsx';
import { useMemo } from 'react';

import { createItemSwapGearData } from '../../model/gear_data';

export interface ItemSwapIconViewProps {
	slot: ItemSlot;
	/** The swap item in the slot; `null` draws the slot's placeholder, greyed out. */
	item: EquippedItem | null;
	iconUrl: string;
	name: string;
	href: string;
	/** The swap item's wowhead tooltip, on the icon. */
	wowhead?: WowheadDatasetProps;
	onOpen: () => void;
}

export const ItemSwapIconView = ({ slot, item, iconUrl, name, href, wowhead, onOpen }: ItemSwapIconViewProps) => (
	<div className="relative items-center" data-testid="icon-picker-root">
		<WowheadAnchor
			className={clsx(
				'ui-icon-picker-swatch size-16 focus-visible:outline focus-visible:outline-1 focus-visible:outline-offset-1 focus-visible:outline-link',
				item ? 'filter-none' : 'grayscale',
			)}
			testId="icon-picker-button"
			data-active={item ? '' : undefined}
			aria-label={name || translateSlotName(slot) || undefined}
			href={href || undefined}
			onActivate={onOpen}
			style={{ backgroundImage: `url('${(item && iconUrl) || getEmptySlotIconUrl(slot)}')` }}
			{...wowhead}
		/>
	</div>
);

export interface ItemSwapIconProps {
	slot: ItemSlot;
}

export const ItemSwapIcon = ({ slot }: ItemSwapIconProps) => {
	const host = useSimHost();
	const openSelectorModal = useOpenSelectorModal();
	const player = host.player;

	const item = usePlayerStore('itemSwapGear').getEquippedItem(slot);

	const actionId = useMemo(() => item?.asActionId(), [item]);
	const { iconUrl, name, href } = useActionId(actionId);

	const wowheadProps = useEquippedItemWowheadDataset(player, item);

	return (
		<ItemSwapIconView
			slot={slot}
			item={item}
			iconUrl={iconUrl}
			name={name}
			href={href}
			wowhead={wowheadProps}
			onOpen={() => openSelectorModal(slot, SelectorModalTabs.Items, createItemSwapGearData(player, slot))}
		/>
	);
};
