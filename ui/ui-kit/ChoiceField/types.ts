import type { Select } from '@base-ui/react/select';
import type { ControlledProps, FieldFrameProps, ResolvedIcon } from '@ui-kit/Field';
import type { IconName } from '@ui-kit/Icon';
import type { ClassValue } from 'clsx';
import type { ReactNode } from 'react';

export interface ChoiceOption<V> {
	value: V;
	/** Rendered after `icon`, in the option and on the trigger while that option is selected. */
	label: ReactNode;
	/** Hover text for this option alone, rendered as HTML by the one tooltip the list shares. */
	tooltip?: string;
	icon?: ResolvedIcon | IconName;
	/** Lands on the option **and** on the trigger while it is selected. */
	className?: ClassValue;
	/** Options filed under this one; the option stays selectable and also opens them. */
	submenu?: ReadonlyArray<ChoiceEntry<V>>;
}

/** A submenu with no value of its own: a category the options inside it are filed under. */
export interface ChoiceGroup<V> {
	label: ReactNode;
	submenu: ReadonlyArray<ChoiceEntry<V>>;
}

/** A heading over options that stay in the same list, where a `ChoiceGroup` opens a submenu. */
export interface ChoiceSection<V> {
	label: ReactNode;
	options: ReadonlyArray<ChoiceOption<V>>;
	submenu?: never;
}

export type ChoiceEntry<V> = ChoiceOption<V> | ChoiceGroup<V> | ChoiceSection<V>;

/** `select` is the bordered form control; `link` is a bare text trigger with a caret. */
export type ChoiceVariant = 'select' | 'link';

/**
 * A field renders a `Field` root only for a label, description or layout; without any of them it
 * is the bare trigger. `className` and `testId` land on the outermost element either way.
 */
export interface ChoiceFieldProps<V> extends FieldFrameProps, ControlledProps<V> {
	options: ReadonlyArray<ChoiceEntry<V>>;
	/**
	 * Builds the nested menu, called only while it is open and once per opening, so a large tree costs
	 * nothing until it is needed. `options` stay the flat list the trigger finds the selection in, and
	 * the field opens a menu, not a select. Keep its identity stable, or an open menu rebuilds on every render.
	 */
	optionTree?: () => ReadonlyArray<ChoiceEntry<V>>;
	/** Two values are the same selection. Defaults to `Object.is`; a caller whose values are messages passes its own. */
	equals?(a: V, b: V): boolean;
	variant?: ChoiceVariant;
	/** Marks the trigger `aria-invalid`; the `select` look draws a danger border. */
	invalid?: boolean;
	/** The trigger's content while no option matches `value`. */
	placeholder?: ReactNode;
	/** The trigger's content for the selected option, in place of its icon and label. */
	renderValue?(selected: ChoiceOption<V>): ReactNode;
	/** Names a trigger that has no `label`. */
	ariaLabel?: string;
	/** Ids of extra text that describes the trigger, such as a message under it; a field `description` is added to them. */
	'aria-describedby'?: string;
	side?: Select.Positioner.Props['side'];
	/**
	 * `'fixed'` takes the list out of flow so an overflow-clipped ancestor cannot cut it off. It only
	 * holds while nothing between the list and the viewport carries a `transform`, `filter` or
	 * `contain`.
	 */
	positionMethod?: Select.Positioner.Props['positionMethod'];
}
