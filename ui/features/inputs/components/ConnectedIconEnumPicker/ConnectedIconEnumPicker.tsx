import { actionIdHref } from '@sim/hooks/useActionId';
import { useActionIds, useResolvedActionIds } from '@sim/hooks/useActionIds';
import { useResolvedIcon } from '@sim/hooks/useResolvedIcon';
import type { ActionId } from '@sim/proto/action_id';
import { actionIdWowheadTooltipData } from '@sim/proto/action_id/tooltip_data';
import { IconEnumPicker } from '@ui-kit/IconEnumPicker';
import { useState } from 'react';

import { useConnectedField } from '../../hooks/useConnectedField';
import { useHideZeroRestore } from '../../hooks/useHideZeroRestore';
import { keptTooltipOf } from '../../input';
import type { IconEnumPickerConfig, IconEnumValueConfig } from './types';
import { iconEnumPickerShown, iconEnumValues, valuePicked } from './utils';

export interface ConnectedIconEnumPickerProps<ModObject, T> {
	modObject: ModObject;
	config: IconEnumPickerConfig<ModObject, T>;
	testId?: string;
}

const buffDataset = (actionId: ActionId) =>
	actionIdWowheadTooltipData(actionId, { useBuffAura: true }).then(url => (url ? { 'data-wowhead': url } : undefined));

export const ConnectedIconEnumPicker = <ModObject, T>({ modObject, config, testId }: ConnectedIconEnumPickerProps<ModObject, T>) => {
	const { value, setValue, disabled, revision, frameProps } = useConnectedField(modObject, config);
	const shows = (valueConfig: IconEnumValueConfig<ModObject, T>) => !valueConfig.showWhen || valueConfig.showWhen(modObject);
	const values = iconEnumValues(config, modObject);

	// Keyed on which values carry an actionId, not on which icons have resolved: a notification landing before the icons arrive must not read as hidden and zero the source.
	const hidden = !iconEnumPickerShown(config, modObject);
	const { clearStored } = useHideZeroRestore(
		{ value, setValue, hidden, revision },
		config.zeroValue,
		current => !valuePicked(config, current),
		current => {
			if (values.some(valueConfig => valueConfig.value === current && !shows(valueConfig))) setValue(config.zeroValue);
		},
	);

	// Only the selection shows until the menu first opens, so the other options fill then, and a hidden
	// option never does. The selection keeps its own icon, so the swatch does not blank while the
	// options fill. Each opening renders again, so values given as a function are read afresh. Every shown
	// option's buff tooltip resolves from the start, so the first opening does not blank the selection's.
	const [openings, setOpenings] = useState(0);
	const selected = values.find(valueConfig => config.equals(valueConfig.value, value));
	const [selectedIcon] = useActionIds([selected && shows(selected) ? selected.actionId : undefined]);
	const icons = useActionIds(values.map(valueConfig => (openings > 0 && shows(valueConfig) ? valueConfig.actionId : undefined)));
	const buffDatasets = useResolvedActionIds(
		values.map(valueConfig => (!hidden && shows(valueConfig) ? valueConfig.buffActionId : undefined)),
		() => undefined,
		buffDataset,
	);
	const backupIcon = useResolvedIcon(selected ? undefined : config.backupIconUrl?.(value));

	if (hidden) return null;

	const iconOf = (valueConfig: IconEnumValueConfig<ModObject, T>, index: number) => {
		const icon = icons[index] ?? (valueConfig === selected ? selectedIcon : undefined);
		if (!icon || !valueConfig.buffActionId) return icon;
		return { ...icon, href: actionIdHref(valueConfig.buffActionId) || undefined };
	};

	return (
		<IconEnumPicker
			{...frameProps}
			testId={testId}
			options={values.map((valueConfig, index) => ({
				value: valueConfig.value,
				icon: iconOf(valueConfig, index),
				color: valueConfig.color,
				iconUrl: valueConfig.iconUrl,
				text: valueConfig.text,
				tooltip: valueConfig.tooltip,
				wowhead: buffDatasets[index],
				hidden: !shows(valueConfig),
			}))}
			equals={config.equals}
			zeroValue={config.zeroValue}
			backupIcon={backupIcon}
			tooltip={config.tooltip}
			keptTooltip={keptTooltipOf(config, modObject, value)}
			direction={config.direction}
			numColumns={typeof config.numColumns === 'function' ? config.numColumns(modObject) : config.numColumns}
			onOpenChange={open => {
				if (open) setOpenings(count => count + 1);
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
