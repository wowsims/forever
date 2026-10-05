import type { ItemSlot } from '@generated/proto/common';
import { translateProtoStatName, translateSlotName } from '@i18n/localization';
import { usePlayer } from '@sim/context/SimHostContext';
import { useActionId } from '@sim/hooks/useActionId';
import type { WowheadDatasetProps } from '@sim/hooks/useWowheadDataset';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { ItemCell, type ItemCellProps } from '@ui-kit/ItemCell';
import { itemQualityClassName } from '@ui-kit/utils/css';
import { WowheadAnchor } from '@ui-kit/WowheadAnchor';
import type { ClassValue } from 'clsx';
import clsx from 'clsx';
import type { ReactNode } from 'react';
import { useMemo } from 'react';

import { useEquippedItemWowheadDataset } from '../../hooks/useEquippedItemWowheadDataset';
import { MISSING_RANDOM_SUFFIX_WARNING } from '../../item_notices';
import { getEmptySlotIconUrl } from '../../model/empty_slot_icons';
import { SelectorModalTabs } from '../../types';
import { EnchantLabel } from '../GearPicker/EnchantLabel';
import { ItemNoticeIcon } from '../GearPicker/ItemNoticeIcon';
import { NameDescriptionLabel } from './NameDescriptionLabel';

export interface ItemDetailCellViewProps extends Pick<ItemCellProps, 'border' | 'padding'> {
	/** Drives the empty-state icon and the empty-state name. */
	slot: ItemSlot;
	item: EquippedItem | null;
	iconUrl: string;
	href: string;
	/** The equipped item's wowhead tooltip, on the icon and the name. */
	wowhead?: WowheadDatasetProps;
	/** Where the name and labels sit against the icon; `end` is the gear picker's right column. */
	align?: 'start' | 'end';
	className?: ClassValue;
	/** Omitted for a cell nothing can be picked from — the batch results are read-only. */
	onOpen?: (tab: SelectorModalTabs) => void;
	/** Rendered beside the name. */
	notice?: ReactNode;
	action?: ReactNode;
	/** Anchors the enchant label at a tooltip — the gear picker's quick-swap popover. */
	enchantTooltipId?: string;
	/** Rendered after the enchant label. */
	extraLabels?: ReactNode;
	testId?: string;
	rootDataAttributes?: Record<string, string>;
}

/**
 * One equipped item as the gear vocabulary renders it: icon, item level, name and enchant, with
 * the empty state for a slot that holds nothing.
 */
export const ItemDetailCellView = ({
	slot,
	item,
	iconUrl,
	href,
	wowhead,
	align = 'start',
	border,
	padding,
	className,
	onOpen,
	notice,
	action,
	enchantTooltipId,
	extraLabels,
	testId,
	rootDataAttributes,
}: ItemDetailCellViewProps) => (
	<ItemCell
		align={align}
		border={border}
		padding={padding}
		className={className}
		testId={testId}
		rootDataAttributes={rootDataAttributes}
		ilvl={item ? item.ilvl.toString() : null}
		icon={
			<WowheadAnchor
				className="ui-item-picker-icon"
				testId="item-picker-icon"
				href={href || undefined}
				onActivate={onOpen && (() => onOpen(SelectorModalTabs.Items))}
				style={{ backgroundImage: `url('${(item && iconUrl) || getEmptySlotIconUrl(slot)}')` }}
				{...wowhead}
			/>
		}
		name={
			<>
				<WowheadAnchor
					className={clsx('ui-item-picker-name-container', itemQualityClassName(item?.item.quality))}
					testId="item-picker-name-container"
					href={href || undefined}
					onActivate={onOpen && (() => onOpen(SelectorModalTabs.Items))}
					{...wowhead}>
					{item ? (
						<>
							<span className="ui-item-picker-name tracking-normal" data-testid="item-picker-name">
								{item.item.name}
								{!!item.randomSuffix && ` ${translateProtoStatName(item.randomSuffix.name)}`}
							</span>
							{!!item.item.nameDescription && <NameDescriptionLabel nameDescription={item.item.nameDescription} className="tracking-normal" />}
						</>
					) : (
						translateSlotName(slot)
					)}
				</WowheadAnchor>
				{notice}
			</>
		}
		labels={
			<>
				<EnchantLabel
					className="ui-item-picker-label-muted tracking-normal"
					testId="item-picker-enchant"
					enchant={item?.enchant}
					onActivate={onOpen && (() => onOpen(SelectorModalTabs.Enchants))}
					tooltipId={enchantTooltipId}
				/>
				{extraLabels}
			</>
		}
		action={action}
	/>
);

export type ItemDetailCellProps = Omit<ItemDetailCellViewProps, 'iconUrl' | 'href' | 'wowhead' | 'notice'>;

/** Resolves the equipped item's icon, link, wowhead tooltip and notice for `ItemDetailCellView`. */
export const ItemDetailCell = (props: ItemDetailCellProps) => {
	const { item } = props;
	const player = usePlayer();

	const actionId = useMemo(() => item?.asActionId(), [item]);
	const { iconUrl, href } = useActionId(actionId);
	const wowhead = useEquippedItemWowheadDataset(player, item);

	return (
		<ItemDetailCellView
			{...props}
			iconUrl={iconUrl}
			href={href}
			wowhead={wowhead}
			notice={
				!!item && (
					<ItemNoticeIcon
						itemId={item.item.id}
						enchantId={item.enchant?.effectId}
						additionalNotice={item.hasRandomSuffixOptions() && !item.randomSuffix ? MISSING_RANDOM_SUFFIX_WARNING : undefined}
					/>
				)
			}
		/>
	);
};
