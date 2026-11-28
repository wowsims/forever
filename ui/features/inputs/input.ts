import i18n from '@i18n/config';
import type { StoreField } from '@sim/hooks/useStoreField';
import type { StoreSubscribe } from '@sim/state/subscriptions';
import type { FieldLayout } from '@ui-kit/Field';
import type { ReactElement } from 'react';

/** `icon` is out: the icon pickers set that layout themselves, never a config. */
export type PickerLayout = Exclude<FieldLayout, 'icon'>;

export type StoreBinding<ModObject> =
	| { storeSubscribe: (obj: ModObject) => StoreSubscribe; storeField?: StoreField | ReadonlyArray<StoreField> }
	| { storeSubscribe?: undefined; storeField: StoreField | ReadonlyArray<StoreField> };

/** What an input renders around its value. */
export interface InputFrame<ModObject> {
	label?: string;
	labelTooltip?: string | ReactElement;
	description?: string;
	layout?: PickerLayout;
	id?: string;
	extraClassNames?: Array<string>;

	// If set, will automatically disable the input when this evaluates to false.
	enableWhen?: (obj: ModObject) => boolean;

	// If set, will automatically hide the input when this evaluates to false.
	showWhen?: (obj: ModObject) => boolean;
}

export interface KeptInput<ModObject, T> {
	// Dims the swatch or trigger with the kept tooltip while it holds for the value the input shows: a selection the
	// filters hide, still offered because it is selected. A multi icon's value is whether any of its inputs is on.
	keptBy?: (obj: ModObject, value: T) => boolean;
}

export interface OverwrittenInput<ModObject, T> {
	// Badges the swatch or trigger with the tooltip it returns for the value: a selection another selection overwrites.
	// The swatch dims too unless `partly`, while the selection still adds something. An icon-enum picker asks it of the
	// value it shows and, while its menu is open, of each option, as if that option were selected in place of the value.
	overwritten?: (obj: ModObject, value: T) => { tooltip: string; partly: boolean } | undefined;
}

/** The kit pickers' `overwrittenTooltip` and `overwrittenPartly` for a value, both undefined while nothing overwrites it. */
export const overwrittenOf = <ModObject, T>(
	config: OverwrittenInput<ModObject, T>,
	obj: ModObject,
	value: T,
): { overwrittenTooltip?: string; overwrittenPartly?: boolean } => {
	const overwritten = config.overwritten?.(obj, value);
	return { overwrittenTooltip: overwritten?.tooltip, overwrittenPartly: overwritten?.partly };
};

/** The kit pickers' `keptTooltip` for the value the adapter read: the kept text while `keptBy` holds, otherwise none. */
export const keptTooltipOf = <ModObject, T>(config: KeptInput<ModObject, T>, obj: ModObject, value: T): string | undefined =>
	config.keptBy?.(obj, value) ? i18n.t('stat_filters.kept_tooltip') : undefined;

/**
 * Data for creating a new input UI element, reading its own value off the mod object.
 */
export interface InputConfig<ModObject, T, V = T> extends InputFrame<ModObject> {
	defaultValue?: T;

	// The input's change source. Omit for inputs that are re-synced
	// by their parent (nested APL pickers, UI-local toggles).
	storeSubscribe?: (obj: ModObject) => StoreSubscribe;

	// The same change source named as data: the store field(s) this input reads.
	// `storeSubscribe` wins where a config carries both.
	storeField?: StoreField | ReadonlyArray<StoreField>;

	// `getValue` is re-read on a notification, not on a render, so it must read the model off `obj`.
	getValue: (obj: ModObject) => T;
	setValue: (obj: ModObject, newValue: T) => void;

	// Convert between source value and input value types. These functions must be set if T != V.
	sourceToValue?: (src: T) => V;
	valueToSource?: (val: V) => T;
}
