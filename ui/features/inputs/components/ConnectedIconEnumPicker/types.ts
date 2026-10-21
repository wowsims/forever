import type { ActionId } from '@sim/proto/action_id';
import type { IconEnumOption, IconEnumPickerProps } from '@ui-kit/IconEnumPicker';

import type { InputConfig, KeptInput, StoreBinding } from '../../input';

export interface IconEnumValueConfig<ModObject, T> extends Pick<IconEnumOption<T>, 'value' | 'color' | 'iconUrl' | 'text' | 'tooltip'> {
	// Shows the icon for this id; takes precedence over color and iconUrl.
	actionId?: ActionId;
	// Shows the tooltip of the buff this id applies when hovering the option, under `actionId`'s icon.
	buffActionId?: ActionId;
	showWhen?: (obj: ModObject) => boolean;
}

interface IconEnumPickerBase<ModObject, T>
	extends InputConfig<ModObject, T>, KeptInput<ModObject, T>, Pick<IconEnumPickerProps<T>, 'equals' | 'zeroValue' | 'tooltip' | 'direction'> {
	// A function is read on every render, for a picker whose shown values change.
	numColumns?: number | ((obj: ModObject) => number);
	// A function is read on every render and on every opening of the menu, for a picker whose values change order.
	values: Array<IconEnumValueConfig<ModObject, T>> | ((obj: ModObject) => Array<IconEnumValueConfig<ModObject, T>>);
	backupIconUrl?: (value: T) => ActionId;
	showWhen?: (obj: ModObject) => boolean;
}

// Required here: an icon enum picker has no parent that refreshes it.
export type IconEnumPickerConfig<ModObject, T> = IconEnumPickerBase<ModObject, T> & StoreBinding<ModObject>;
