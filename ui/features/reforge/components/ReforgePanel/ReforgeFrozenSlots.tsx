import { ConnectedBooleanPicker } from '@features/inputs/components/ConnectedBooleanPicker';
import { ItemSlot } from '@generated/proto/common';
import { translateSlotName } from '@i18n/localization';
import type { Player } from '@sim/player/player';
import type { ReforgeSettings } from '@sim/settings/reforge_settings';
import { type ReactNode, useMemo } from 'react';

import { useReforgeIdPrefix } from './context/ReforgeIdPrefixContext';

export interface ReforgeFrozenSlotsViewProps {
	hidden: boolean;
	pickers: ReadonlyArray<ReactNode>;
}

export const ReforgeFrozenSlotsView = ({ hidden, pickers }: ReforgeFrozenSlotsViewProps) => (
	<div className="mb-2 grid grid-cols-2" data-testid="reforge-frozen-slots" hidden={hidden}>
		{pickers}
	</div>
);

export interface ReforgeFrozenSlotsProps {
	settings: ReforgeSettings;
	player: Player<any>;
	freezeItemSlots: boolean;
}

/** One toggle per slot, in the gear's own slot order. The slot list is the whole slot enum, so it is memoised on the player. */
export const ReforgeFrozenSlots = ({ settings, player, freezeItemSlots }: ReforgeFrozenSlotsProps) => {
	const idPrefix = useReforgeIdPrefix();
	const slots = useMemo(() => player.getGear().getItemSlots(), [player]);

	return (
		<ReforgeFrozenSlotsView
			hidden={!freezeItemSlots}
			pickers={slots.map(slot => (
				<ConnectedBooleanPicker
					key={slot}
					modObject={player}
					config={{
						id: `${idPrefix}-freeze-${ItemSlot[slot]}`,
						label: translateSlotName(slot),
						layout: 'inline',
						storeField: 'reforge:freezeItemSlots',
						getValue: () => settings.getFrozenItemSlot(slot) || false,
						setValue: (_player, newValue) => settings.setFrozenItemSlot(slot, newValue),
					}}
				/>
			))}
		/>
	);
};
