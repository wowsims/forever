import type { ActionId } from '@sim/proto/action_id';
import type { IconPickerProps } from '@ui-kit/IconPicker';

import type { InputConfig, StoreBinding } from '../../input';

// One icon input, e.g. one for arcane brilliance and another for kings. ModObject is the object being
// modified (Sim, Player or Target); ValueType is a number or a boolean.
interface IconPickerBase<ModObject, ValueType> extends InputConfig<ModObject, ValueType>, Pick<IconPickerProps, 'states'> {
	actionId: ActionId;
	// Only used if states >= 3.
	improvedId?: ActionId;
	// Only used if states >= 4.
	improvedId2?: ActionId;
}

// Required here: an icon picker has no parent that refreshes it.
export type IconPickerConfig<ModObject, ValueType> = IconPickerBase<ModObject, ValueType> & StoreBinding<ModObject>;
