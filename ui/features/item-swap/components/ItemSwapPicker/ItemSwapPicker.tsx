import { ConnectedBooleanPicker } from '@features/inputs/components/ConnectedBooleanPicker';
import type { BooleanPickerConfig } from '@features/inputs/configs';
import type { ItemSlot, Spec } from '@generated/proto/common';
import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import type { Player } from '@sim/player/player';
import { Button } from '@ui-kit/Button';
import { Field } from '@ui-kit/Field';
import { HelpText } from '@ui-kit/FormControl';
import { Icon } from '@ui-kit/Icon';
import { IconRow } from '@ui-kit/IconRow';
import { LocaleHtml, Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { type ReactNode, useId, useMemo } from 'react';

import { swapWithGear } from '../../model/swap_with_gear';
import { ItemSwapIcon } from '../ItemSwapIcon';

export interface ItemSwapPickerViewProps {
	note?: string;
	onSwap: () => void;
	/** One icon per swap slot. */
	children: ReactNode;
}

export const ItemSwapPickerView = ({ note, onSwap, children }: ItemSwapPickerViewProps) => {
	const swapId = useId();
	const swapTooltip = i18n.t('settings_tab.other.item_swap.tooltip');

	return (
		<>
			<Field
				layout="split"
				role="group"
				label={i18n.t('settings_tab.other.item_swap.label')}
				className="flex-wrap gap-2"
				testId="input-item-swap-container">
				<Button variant="unstyled" data-testid="gear-swap-icon" aria-label={swapTooltip} {...tooltipAnchorProps(swapId)} onClick={onSwap}>
					<Icon name="arrows-rotate" className="me-1" />
				</Button>
				<Tooltip id={swapId} content={swapTooltip} />
				<IconRow className="items-center justify-end" testId="item-swap-icon-grid">
					{children}
				</IconRow>
			</Field>
			{note && <HelpText as="p">{note}</HelpText>}
		</>
	);
};

export interface ItemSwapPickerProps {
	itemSlots: ReadonlyArray<ItemSlot>;
	note?: string;
}

export const ItemSwapPicker = <SpecType extends Spec>({ itemSlots, note }: ItemSwapPickerProps) => {
	const host = useSimHost();
	const player = host.player as Player<SpecType>;
	const enabled = usePlayerStore('itemSwapEnabled');

	const enableConfig = useMemo(
		(): BooleanPickerConfig<Player<SpecType>> => ({
			id: 'enable-item-swap',
			reverse: true,
			label: i18n.t('settings_tab.other.enable_item_swap.label'),
			labelTooltip: <LocaleHtml html={i18n.t('settings_tab.other.enable_item_swap.tooltip')} />,
			layout: 'split',
			storeField: 'itemSwap',
			getValue: (subject: Player<SpecType>) => subject.itemSwapSettings.getEnableItemSwap(),
			setValue: (subject: Player<SpecType>, newValue: boolean) => subject.itemSwapSettings.setEnableItemSwap(newValue),
		}),
		[],
	);

	return (
		<div className="grid gap-2" data-testid="item-swap-picker-root">
			<ConnectedBooleanPicker modObject={player} config={enableConfig} />
			{enabled && (
				<ItemSwapPickerView note={note} onSwap={() => swapWithGear(player, itemSlots)}>
					{itemSlots.map(itemSlot => (
						<ItemSwapIcon key={itemSlot} slot={itemSlot} />
					))}
				</ItemSwapPickerView>
			)}
		</div>
	);
};
