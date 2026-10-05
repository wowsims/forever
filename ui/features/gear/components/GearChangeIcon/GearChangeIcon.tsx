import type { ItemSlot } from '@generated/proto/common';
import { useSimHost } from '@sim/context/SimHostContext';
import { useActionId } from '@sim/hooks/useActionId';
import { useActionIdWowheadDataset } from '@sim/hooks/useActionIdWowheadDataset';
import type { WowheadDatasetProps } from '@sim/hooks/useWowheadDataset';
import type { EquippedItem } from '@sim/proto/equipped_item';
import type { ResolvedIcon } from '@ui-kit/Field';
import { Icon } from '@ui-kit/Icon';
import { WowheadAnchor } from '@ui-kit/WowheadAnchor';
import { useMemo } from 'react';

import { useEquippedItemWowheadDataset } from '../../hooks/useEquippedItemWowheadDataset';
import { getEmptySlotIconUrl } from '../../model/empty_slot_icons';
import { enchantActionId } from '../../model/item_data';
import { enchantChanged as hasEnchantChanged } from './utils';

export interface GearChangeIconViewProps {
	slot: ItemSlot;
	/** False draws the empty frame, which is what an unfilled slot looks like. */
	equipped: boolean;
	iconUrl: string;
	href: string;
	wowhead?: WowheadDatasetProps;
	/** The item's enchant; an empty iconUrl falls back to a glyph. */
	enchant?: ResolvedIcon;
	enchantChanged?: boolean;
	enchantWowhead?: WowheadDatasetProps;
}

/** Parameterises the slot and the before/after pair; fixes the frame and the enchant change marker. */
export const GearChangeIconView = ({ slot, equipped, iconUrl, href, wowhead, enchant, enchantChanged, enchantWowhead }: GearChangeIconViewProps) => (
	<div className="ui-item-picker-root" data-testid="gear-change-icon">
		<div className="relative w-fit">
			<div
				className="ui-item-picker-icon-wrapper"
				data-testid="item-picker-icon-wrapper"
				style={{ backgroundImage: `url('${(equipped && iconUrl) || getEmptySlotIconUrl(slot)}')` }}
			/>
			<WowheadAnchor className="absolute inset-0" testId="gear-change-icon-link" href={equipped ? href || undefined : undefined} {...wowhead} />
			{enchant && (
				<WowheadAnchor
					className="ui-gear-change-icon-enchant"
					testId="gear-change-icon-enchant"
					href={enchant.href || undefined}
					aria-label={enchant.name}
					data-changed={enchantChanged ? '' : undefined}
					style={enchant.iconUrl ? { backgroundImage: `url('${enchant.iconUrl}')` } : undefined}
					{...enchantWowhead}>
					{!enchant.iconUrl && <Icon name="circle-exclamation" className="ml-px block text-danger" />}
				</WowheadAnchor>
			)}
		</div>
	</div>
);

export interface GearChangeIconProps {
	slot: ItemSlot;
	/** The slot after the run. Undefined renders the empty frame, which is what an unfilled slot looks like. */
	item?: EquippedItem;
	previousItem?: EquippedItem;
	/** Draws the item's enchant marker, bordered when the enchant changed. */
	showEnchant?: boolean;
}

export const GearChangeIcon = ({ slot, item, previousItem, showEnchant = false }: GearChangeIconProps) => {
	const player = useSimHost().player;

	const actionId = useMemo(() => item?.asActionId(), [item]);
	const { iconUrl, href } = useActionId(actionId);
	const wowhead = useEquippedItemWowheadDataset(player, item);

	const itemEnchant = showEnchant ? (item?.enchant ?? undefined) : undefined;
	const enchantActionIdOf = useMemo(() => (showEnchant && item?.enchant ? enchantActionId(item.enchant) : undefined), [showEnchant, item]);
	const { iconUrl: enchantIconUrl, href: enchantHref } = useActionId(enchantActionIdOf);
	const enchantWowhead = useActionIdWowheadDataset(enchantActionIdOf);
	const enchant = itemEnchant ? { name: itemEnchant.name, iconUrl: enchantIconUrl, href: enchantHref } : undefined;

	return (
		<GearChangeIconView
			slot={slot}
			equipped={!!item}
			iconUrl={iconUrl}
			href={href}
			wowhead={wowhead}
			enchant={enchant}
			enchantChanged={hasEnchantChanged(item, previousItem)}
			enchantWowhead={enchantWowhead}
		/>
	);
};
