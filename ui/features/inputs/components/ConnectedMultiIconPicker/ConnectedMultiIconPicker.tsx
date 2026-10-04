import { useResolvedIcon } from '@sim/hooks/useResolvedIcon';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import type { Player } from '@sim/player/player';
import type { ActionId } from '@sim/proto/action_id';
import type { StoreSubscribe } from '@sim/state/subscriptions';
import { MultiIconPicker } from '@ui-kit/MultiIconPicker';

import { ConnectedIconPicker } from '../ConnectedIconPicker';
import type { MultiIconPickerConfig } from './types';

export interface ConnectedMultiIconPickerProps<ModObject> {
	modObject: ModObject;
	config: MultiIconPickerConfig<ModObject>;
	subscribe: StoreSubscribe;
	onClear: () => void;
	testId?: string;
}

const firstActiveActionId = <ModObject,>(config: MultiIconPickerConfig<ModObject>, modObject: ModObject): ActionId | null => {
	for (const input of config.inputs) {
		const value = Number(input.getValue(modObject));
		if (value === 0) continue;
		if (value === 2 && input.improvedId) return input.improvedId;
		if (value === 3 && input.improvedId2) return input.improvedId2;
		return input.actionId;
	}
	return null;
};

export const ConnectedMultiIconPicker = <ModObject,>({ modObject, config, subscribe, onClear, testId }: ConnectedMultiIconPickerProps<ModObject>) => {
	const { actionId, hidden } = useStoreSubscribe(subscribe, () => ({
		actionId: firstActiveActionId(config, modObject),
		hidden: !!config.showWhen && !config.showWhen(modObject as unknown as Player<any>),
	}));

	const icon = useResolvedIcon(config.categoryId ?? actionId ?? undefined);

	if (hidden) return null;

	return (
		<MultiIconPicker label={config.label} icon={icon} active={!!actionId} onClear={onClear} testId={testId}>
			{config.inputs.map((input, index) => (
				<ConnectedIconPicker key={index} modObject={modObject} config={input} />
			))}
		</MultiIconPicker>
	);
};
