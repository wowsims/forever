import { usePendingIcon, useResolvedIcon } from '@sim/hooks/useResolvedIcon';
import { IconPicker } from '@ui-kit/IconPicker';

import { useConnectedField } from '../../hooks/useConnectedField';
import { useHideZeroRestore } from '../../hooks/useHideZeroRestore';
import { keptTooltipOf } from '../../input';
import { valuePicked } from '../ConnectedIconEnumPicker/utils';
import type { IconPickerConfig } from './types';

export interface ConnectedIconPickerProps<ModObject, ValueType> {
	modObject: ModObject;
	config: IconPickerConfig<ModObject, ValueType>;
	testId?: string;
}

export const ConnectedIconPicker = <ModObject, ValueType>({ modObject, config, testId }: ConnectedIconPickerProps<ModObject, ValueType>) => {
	const { value, setValue, hidden: showWhenHidden, disabled, revision, frameProps } = useConnectedField(modObject, config);
	const hidden = config.actionId == null || showWhenHidden;

	// A two-state source stores a boolean. The hide/restore writes go through `write` as well, so a restored number is stored as a boolean too.
	const toSource = (n: number) => (config.states === 2 ? Boolean(n) : n) as ValueType;
	const write = (next: ValueType) => setValue(toSource(Number(next)));
	useHideZeroRestore({ value, setValue: write, hidden, revision }, toSource(0), current => !valuePicked(config, current));

	const icon = useResolvedIcon(config.actionId ?? undefined);
	// Pending, not resolved: a missing improved icon makes the picker write its counter, which would flash while it loads.
	const improvedIcon = usePendingIcon(config.improvedId);
	const improvedIcon2 = usePendingIcon(config.improvedId2);

	if (hidden) return null;

	return (
		<IconPicker
			{...frameProps}
			testId={testId}
			states={config.states}
			icon={icon}
			improvedIcon={improvedIcon}
			improvedIcon2={improvedIcon2}
			keptTooltip={keptTooltipOf(config, modObject, value)}
			value={Number(value)}
			onChange={n => setValue(toSource(n))}
			disabled={disabled}
		/>
	);
};
