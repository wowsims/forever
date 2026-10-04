import { useActionIds } from '@sim/hooks/useActionIds';
import { useResolvedIcon } from '@sim/hooks/useResolvedIcon';
import { IconEnumPicker } from '@ui-kit/IconEnumPicker';
import { useState } from 'react';

import { useConnectedField } from '../../hooks/useConnectedField';
import { useHideZeroRestore } from '../../hooks/useHideZeroRestore';
import type { IconEnumPickerConfig, IconEnumValueConfig } from './types';
import { iconEnumPickerShown } from './utils';

export interface ConnectedIconEnumPickerProps<ModObject, T> {
	modObject: ModObject;
	config: IconEnumPickerConfig<ModObject, T>;
	testId?: string;
}

export const ConnectedIconEnumPicker = <ModObject, T>({ modObject, config, testId }: ConnectedIconEnumPickerProps<ModObject, T>) => {
	const { value, setValue, disabled, revision, frameProps } = useConnectedField(modObject, config);
	const shows = (valueConfig: IconEnumValueConfig<ModObject, T>) => !valueConfig.showWhen || valueConfig.showWhen(modObject);

	// Keyed on which values carry an actionId, not on which icons have resolved: a notification landing before the icons arrive must not read as hidden and zero the source.
	const hidden = !iconEnumPickerShown(config, modObject);
	const { clearStored } = useHideZeroRestore(
		{ value, setValue, hidden, revision },
		config.zeroValue,
		current => config.equals(current, config.zeroValue),
		current => {
			if (config.values.some(valueConfig => !shows(valueConfig) && valueConfig.value === current)) setValue(config.zeroValue);
		},
	);

	// Only the selection shows until the menu first opens, so the other options fill then, and a hidden
	// option never does. The selection keeps its own icon, so the swatch does not blank while the
	// options fill.
	const [opened, setOpened] = useState(false);
	const selected = config.values.find(valueConfig => config.equals(valueConfig.value, value));
	const [selectedIcon] = useActionIds([selected && shows(selected) ? selected.actionId : undefined]);
	const icons = useActionIds(config.values.map(valueConfig => (opened && shows(valueConfig) ? valueConfig.actionId : undefined)));
	const backupIcon = useResolvedIcon(selected ? undefined : config.backupIconUrl?.(value));

	if (hidden) return null;

	return (
		<IconEnumPicker
			{...frameProps}
			testId={testId}
			options={config.values.map((valueConfig, index) => ({
				value: valueConfig.value,
				icon: icons[index] ?? (valueConfig === selected ? selectedIcon : undefined),
				color: valueConfig.color,
				iconUrl: valueConfig.iconUrl,
				text: valueConfig.text,
				tooltip: valueConfig.tooltip,
				hidden: !shows(valueConfig),
			}))}
			equals={config.equals}
			zeroValue={config.zeroValue}
			backupIcon={backupIcon}
			tooltip={config.tooltip}
			direction={config.direction}
			numColumns={typeof config.numColumns === 'function' ? config.numColumns(modObject) : config.numColumns}
			onOpenChange={open => {
				if (open) setOpened(true);
			}}
			value={value}
			onChange={next => {
				clearStored();
				setValue(next);
			}}
			disabled={disabled}
		/>
	);
};
