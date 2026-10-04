import type { ActionId } from '@sim/proto/action_id';
import type { IconEnumOption, IconEnumPickerProps } from '@ui-kit/IconEnumPicker';

import type { InputConfig, StoreBinding } from '../../input';

export interface IconEnumValueConfig<ModObject, T> extends Pick<IconEnumOption<T>, 'value' | 'color' | 'iconUrl' | 'text' | 'tooltip'> {
	// Shows the icon for this id; takes precedence over color and iconUrl.
	actionId?: ActionId;
	showWhen?: (obj: ModObject) => boolean;
}

interface IconEnumPickerBase<ModObject, T> extends InputConfig<ModObject, T>, Pick<IconEnumPickerProps<T>, 'equals' | 'zeroValue' | 'tooltip' | 'direction'> {
	// A function is read on every render, for a picker whose shown values change.
	numColumns?: number | ((obj: ModObject) => number);
	values: Array<IconEnumValueConfig<ModObject, T>>;
	backupIconUrl?: (value: T) => ActionId;
	showWhen?: (obj: ModObject) => boolean;
}

// Required here: an icon enum picker has no parent that refreshes it.
export type IconEnumPickerConfig<ModObject, T> = IconEnumPickerBase<ModObject, T> & StoreBinding<ModObject>;
