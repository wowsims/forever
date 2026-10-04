import { ConnectedInput } from '@features/inputs/components/ConnectedInput';
import { ItemSwapPicker } from '@features/item-swap';
import type { ItemSlot } from '@generated/proto/common';
import type { Player } from '@sim/player/player';
import type { InputConfig } from '@sim/spec_config';

export interface OtherSettingsProps {
	inputs: ReadonlyArray<InputConfig<Player<any>>>;
	itemSlots: ReadonlyArray<ItemSlot>;
}

export const OtherSettings = ({ inputs, itemSlots }: OtherSettingsProps) => (
	<>
		{inputs.map(config => (
			<ConnectedInput key={config.id} config={{ ...config, layout: 'split' }} />
		))}
		{itemSlots.length > 0 && <ItemSwapPicker itemSlots={itemSlots} />}
	</>
);
