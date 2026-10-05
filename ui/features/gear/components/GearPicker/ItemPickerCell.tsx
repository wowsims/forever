import type { ItemSlot } from '@generated/proto/common';
import { useSimHost } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { useUiStore } from '@sim/hooks/useUiStore';
import { Tooltip } from '@ui-kit/Tooltip';
import { useCallback, useId } from 'react';

import { hasTouch } from '../../../../shared/pointer';
import { useOpenSelectorModal } from '../../hooks/useSelectorModal';
import { createGearData } from '../../model/gear_data';
import { SelectorModalTabs } from '../../types';
import { ItemDetailCell, type ItemDetailCellProps } from '../ItemCell';
import { QuickEnchantList } from './QuickEnchantList';

export interface ItemPickerCellProps {
	slot: ItemSlot;
	ready: boolean;
	align?: ItemDetailCellProps['align'];
}

export const ItemPickerCell = ({ slot, ready, align }: ItemPickerCellProps) => {
	const host = useSimHost();
	const openSelectorModal = useOpenSelectorModal();
	const player = host.player;
	const tooltipId = useId();

	const item = usePlayerStore('gear').getEquippedItem(slot);

	const showQuickSwapSetting = useUiStore('showQuickSwap');
	const showQuickSwap = !hasTouch() && showQuickSwapSetting;

	const open = useCallback(
		(tab: SelectorModalTabs) => {
			if (!ready) return;
			openSelectorModal(slot, tab, createGearData(player, slot));
		},
		[openSelectorModal, player, slot, ready],
	);

	// The list reads the store inside `Tooltip`'s children, which react-tooltip does not build until
	// the tooltip first opens — that is what keeps a `filters` subscriber per slot off the pane.
	const quickSwapTooltip =
		!item || !ready ? null : (
			<Tooltip
				id={`${tooltipId}-enchant`}
				className="tooltip-quick-swap cursor-default"
				maxWidth="max-w-none"
				widthClassName="w-54"
				align="start"
				padded={false}
				place="bottom"
				clickable
				hidden={!showQuickSwap}
				content={<QuickEnchantList slot={slot} onOpenDetail={() => open(SelectorModalTabs.Enchants)} />}
			/>
		);

	return <ItemDetailCell slot={slot} item={item} align={align} onOpen={open} enchantTooltipId={`${tooltipId}-enchant`} extraLabels={quickSwapTooltip} />;
};
