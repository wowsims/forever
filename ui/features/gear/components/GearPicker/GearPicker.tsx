import { TabPanelColumns } from '@ui-kit/TabPanelColumns';

import { LEFT_ITEM_SLOTS, RIGHT_ITEM_SLOTS } from '../../model/gear_data';
import { ItemPickerCell } from './ItemPickerCell';

export interface GearPickerProps {
	ready: boolean;
}

export const GearPicker = ({ ready }: GearPickerProps) => (
	<div className="ui-gear-picker grid grid-cols-2 gap-section max-lg:gap-2 max-sm:grid-cols-1" data-testid="gear-picker-root">
		<TabPanelColumns.Col className="gap-2 *:flex-row *:text-left [&>*:nth-child(6)]:mb-section" externalGap>
			{LEFT_ITEM_SLOTS.map(slot => (
				<ItemPickerCell key={slot} slot={slot} ready={ready} />
			))}
		</TabPanelColumns.Col>
		<TabPanelColumns.Col className="gap-2 *:flex-row-reverse max-md:*:flex-row" externalGap>
			{RIGHT_ITEM_SLOTS.map(slot => (
				<ItemPickerCell key={slot} slot={slot} ready={ready} align="end" />
			))}
		</TabPanelColumns.Col>
	</div>
);
