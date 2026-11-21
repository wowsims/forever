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
	// A hidden picker writes its zero value and brings the pick back when it shows again. With this set it leaves
	// its value alone while hidden, except that a value whose own option `showWhen` hides is still reset. It is for a
	// picker whose `getValue` reads a stored zero as a default, which the restore cannot tell from a pick, and
	// whatever reads the field has to ignore it while the picker is hidden.
	keepWhenHidden?: boolean;
}

// Required here: an icon enum picker has no parent that refreshes it.
export type IconEnumPickerConfig<ModObject, T> = IconEnumPickerBase<ModObject, T> & StoreBinding<ModObject>;
