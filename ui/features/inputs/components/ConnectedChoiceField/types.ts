import type { ChoiceEntry, ChoiceFieldProps } from '@ui-kit/ChoiceField';
import type { EnumValueConfig } from '@ui-kit/MultiComboBox';

import type { InputConfig } from '../../input';

/** The `enum` input kind: a number field whose choices are the config's `values`. */
export type ChoiceFieldConfig<ModObject> = InputConfig<ModObject, number> & { id: string; values: Array<EnumValueConfig> };

export interface ConnectedChoiceFieldProps<M, T, V = T> extends Pick<
	ChoiceFieldProps<V>,
	'optionTree' | 'equals' | 'variant' | 'placeholder' | 'renderValue' | 'ariaLabel' | 'className' | 'testId'
> {
	modObject: M;
	/**
	 * `V` differs from `T` when the option type is not the stored type, bridged by the config's own
	 * `sourceToValue`/`valueToSource`. The APL action-id field stores an `ActionID` message and offers
	 * `ActionId` objects; its unit field stores a `UnitReference` and offers a `UnitValue` carrying
	 * that unit's icon and colour.
	 */
	config: InputConfig<M, T, V> & { id: string; values?: ReadonlyArray<EnumValueConfig> };
	/** The choices, for a config that does not list them as `values`. */
	options?: ReadonlyArray<ChoiceEntry<V>>;
}
