import type { ItemSlot } from '@generated/proto/common';
import i18n from '@i18n/config';
import { getBulkSlotI18nKey } from '@i18n/entity_mapping';
import { translateBulkSlotName, translateSlotName } from '@i18n/localization';
import type { BulkPickerEntry } from '@sim/bulk/types';
import type { BulkSimItemSlot } from '@sim/bulk/utils';
import { Box } from '@ui-kit/Box';
import { Button } from '@ui-kit/Button';
import { ContentBlock } from '@ui-kit/ContentBlock';
import { Icon } from '@ui-kit/Icon';
import { tooltipAnchorProps } from '@ui-kit/Tooltip';

import { BulkItemPicker } from './BulkItemPicker';

export interface BulkItemPickerGroupProps {
	bulkSlot: BulkSimItemSlot;
	entries: readonly BulkPickerEntry[];
	addSlots: readonly ItemSlot[];
	onAddItem: (slot: ItemSlot) => void;
	addTooltipId: string;
}

export const BulkItemPickerGroup = ({ bulkSlot, entries, addSlots, onAddItem, addTooltipId }: BulkItemPickerGroupProps) => {
	const slotKey = getBulkSlotI18nKey(bulkSlot);
	const slotName = translateBulkSlotName(bulkSlot);

	return (
		<Box>
			<ContentBlock
				rootDataAttributes={{ 'data-gear-group': slotKey.replace(/_/g, '-') }}
				config={{ header: { title: slotName, className: 'border-b-0 mb-0 pb-0' }, bodyClassName: 'gap-2' }}
				headerChildren={
					addSlots.length === 1 ? (
						<Button
							iconOnly
							className="ml-auto self-center"
							aria-label={i18n.t('bulk_tab.picker.add_items_to', { slot: slotName })}
							onClick={() => onAddItem(addSlots[0])}
							{...tooltipAnchorProps(addTooltipId)}>
							<Icon name="plus" />
						</Button>
					) : (
						<div className="ml-auto flex gap-2 self-center">
							{addSlots.map(slot => (
								<Button key={slot} variant="link" size="sm" onClick={() => onAddItem(slot)} {...tooltipAnchorProps(addTooltipId)}>
									<Icon name="plus" className="mr-1" />
									{translateSlotName(slot)}
								</Button>
							))}
						</div>
					)
				}
				flush>
				{entries.length ? (
					entries.map(entry => <BulkItemPicker key={entry.index} bulkSlot={bulkSlot} index={entry.index} item={entry.item} />)
				) : (
					<span>{i18n.t('bulk_tab.picker.no_items')}</span>
				)}
			</ContentBlock>
		</Box>
	);
};
