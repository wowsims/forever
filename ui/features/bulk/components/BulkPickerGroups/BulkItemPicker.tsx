import { ItemDetailCell } from '@features/gear/components/ItemCell';
import { useOpenSelectorModal } from '@features/gear/hooks/useSelectorModal';
import type { SelectorModalTabs } from '@features/gear/types';
import type { ItemSlot } from '@generated/proto/common';
import i18n from '@i18n/config';
import { BULK_SIM_ITEM_SLOT_TO_ITEM_SLOT_PAIRS, BulkSimItemSlot } from '@sim/bulk/utils';
import { usePlayer } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { getEligibleItemSlots } from '@sim/proto/items';
import type { BoxBorder } from '@ui-kit/Box';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { useId } from 'react';

import { useBulkState } from '../../hooks/useBulkState';
import { createBulkGearData } from '../../model/gear_data';
import { removeBulkItemByIndex } from '../../model/items';
import { frozenItemSlot } from '../../model/picker_groups';

export type BulkItemPickerState = 'frozen' | 'equipped' | 'editable';

export interface BulkItemPickerViewProps {
	slot: ItemSlot;
	item: EquippedItem;
	state: BulkItemPickerState;
	onOpen: (tab: SelectorModalTabs) => void;
	/** Omitted for an item that cannot be removed from the batch. */
	onRemove?: () => void;
}

const FRAME: Record<BulkItemPickerState, BoxBorder> = {
	frozen: 'frozen',
	equipped: 'brand',
	editable: 'border',
};

const ROOT_ATTRIBUTES: Record<BulkItemPickerState, Record<string, string>> = {
	frozen: { 'data-frozen': '' },
	equipped: { 'data-equipped': '' },
	editable: {},
};

export const BulkItemPickerView = ({ slot, item, state, onOpen, onRemove }: BulkItemPickerViewProps) => {
	const tooltipId = useId();
	const frozenTooltipId = useId();
	return (
		<ItemDetailCell
			slot={slot}
			item={item}
			testId="bulk-item-picker"
			rootDataAttributes={ROOT_ATTRIBUTES[state]}
			border={FRAME[state]}
			padding="sm"
			className="ui-bulk-item-cell mb-0"
			onOpen={onOpen}
			action={
				<div className="ml-2 grid shrink-0 grid-flow-col [align-items:start] gap-1">
					{state === 'frozen' && (
						<>
							<Icon
								name="lock"
								className="text-xs text-frozen"
								title={i18n.t('bulk_tab.picker.frozen_tooltip')}
								testId="item-picker-frozen"
								{...tooltipAnchorProps(frozenTooltipId)}
							/>
							<Tooltip id={frozenTooltipId} content={i18n.t('bulk_tab.picker.frozen_tooltip')} />
						</>
					)}
					{onRemove && (
						<>
							<Button
								iconOnly
								aria-label={i18n.t('bulk_tab.picker.remove_tooltip')}
								variant="link-danger"
								className="leading-none"
								data-testid="item-picker-actions-btn"
								onClick={onRemove}
								{...tooltipAnchorProps(tooltipId)}>
								<Icon name="xmark" />
							</Button>
							<Tooltip id={tooltipId} content={i18n.t('bulk_tab.picker.remove_tooltip')} />
						</>
					)}
				</div>
			}
		/>
	);
};

export interface BulkItemPickerProps {
	bulkSlot: BulkSimItemSlot;
	/** Below zero for the two equipped slots the group covers; those report what is worn rather than offering a choice. */
	index: number;
	item: EquippedItem;
}

/** The physical slot an equipped entry stands for. Null for a user-added one. */
const equippedSlotOf = (bulkSlot: BulkSimItemSlot, index: number): ItemSlot | null => {
	if (index >= 0) return null;
	const slots = BULK_SIM_ITEM_SLOT_TO_ITEM_SLOT_PAIRS.get(bulkSlot);
	if (!slots) return null;
	return index === -1 ? slots[0] : slots[1];
};

export const BulkItemPicker = ({ bulkSlot, index, item }: BulkItemPickerProps) => {
	const openSelectorModal = useOpenSelectorModal();
	const player = usePlayer();

	const gear = usePlayerStore('gear');
	const frozenItems = useBulkState(slice => slice.frozenItems);
	const frozenWeaponSlot = useBulkState(slice => slice.frozenWeaponSlot);

	const ownSlot = equippedSlotOf(bulkSlot, index);
	const frozenBulkSlot = frozenItemSlot(gear, BULK_SIM_ITEM_SLOT_TO_ITEM_SLOT_PAIRS.get(bulkSlot), frozenItems.get(bulkSlot));
	const isCurrentlyEquipped = bulkSlot !== BulkSimItemSlot.ItemSlotHandWeapon && player.getEquippedItems().some(equipped => equipped?.id === item.id);
	const isEditable = index >= 0 && !isCurrentlyEquipped;
	const isFrozen = !ownSlot ? false : ownSlot === frozenWeaponSlot ? (gear.getEquippedItem(ownSlot)?.equals(item) ?? false) : ownSlot === frozenBulkSlot;

	const slot = getEligibleItemSlots(item.item)[0];

	return (
		<BulkItemPickerView
			slot={slot}
			item={item}
			state={isFrozen ? 'frozen' : isEditable ? 'editable' : 'equipped'}
			onOpen={(tab: SelectorModalTabs) => {
				if (!isEditable) return;
				openSelectorModal(slot, tab, createBulkGearData(player, bulkSlot, index));
			}}
			onRemove={isEditable ? () => removeBulkItemByIndex(player, index) : undefined}
		/>
	);
};
