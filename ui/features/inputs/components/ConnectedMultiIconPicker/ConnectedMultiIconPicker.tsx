import { usePendingIcon, useResolvedIcon } from '@sim/hooks/useResolvedIcon';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import type { Player } from '@sim/player/player';
import type { StoreSubscribe } from '@sim/state/subscriptions';
import { MultiIconPicker } from '@ui-kit/MultiIconPicker';

import { keptTooltipOf } from '../../input';
import { valuePicked } from '../ConnectedIconEnumPicker/utils';
import { ConnectedIconPicker } from '../ConnectedIconPicker';
import type { MultiIconPickerConfig } from './types';

export interface ConnectedMultiIconPickerProps<ModObject> {
	modObject: ModObject;
	config: MultiIconPickerConfig<ModObject>;
	subscribe: StoreSubscribe;
	onClear: () => void;
	testId?: string;
}

const firstActiveInput = <ModObject,>(config: MultiIconPickerConfig<ModObject>, modObject: ModObject) => {
	for (const input of config.inputs) {
		const value = Number(input.getValue(modObject));
		if (valuePicked(input, value)) return { input, value };
	}
	return null;
};

export const ConnectedMultiIconPicker = <ModObject,>({ modObject, config, subscribe, onClear, testId }: ConnectedMultiIconPickerProps<ModObject>) => {
	// Subscribed only to re-render on each notification: the reads below take `config` from the render, so a new
	// config is followed without waiting for one.
	useStoreSubscribe(subscribe, () => ({}));
	const picked = firstActiveInput(config, modObject);
	const hidden = !!config.showWhen && !config.showWhen(modObject as unknown as Player<any>);

	// With nothing on and no category icon, the first input's icon stands for the group.
	const icon = useResolvedIcon(config.categoryId ?? picked?.input.actionId ?? config.inputs[0]?.actionId);

	// The corner badges and counter belong to the input's own icon, so a category icon carries none.
	const shown = config.categoryId ? null : picked;
	const improvedIcon = usePendingIcon(shown?.input.improvedId);
	const improvedIcon2 = usePendingIcon(shown?.input.improvedId2);

	if (hidden) return null;

	return (
		<MultiIconPicker
			label={config.label}
			icon={icon}
			active={!!picked}
			state={shown ? { states: shown.input.states, value: shown.value, improvedIcon, improvedIcon2 } : undefined}
			direction={config.direction}
			tooltip={config.tooltip}
			numColumns={config.numColumns?.(modObject)}
			keptTooltip={keptTooltipOf(config, modObject, !!picked)}
			onClear={onClear}
			testId={testId}>
			{/* The menu is a row of bare swatches: each input's label would sit beside its icon, and the icon's tooltip names it. */}
			{config.inputs.map((input, index) => (
				<ConnectedIconPicker key={index} modObject={modObject} config={{ ...input, label: undefined }} />
			))}
		</MultiIconPicker>
	);
};
