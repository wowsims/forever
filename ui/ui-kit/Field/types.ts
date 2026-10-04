import type { ClassValue } from 'clsx';
import type { AriaRole, ReactElement, ReactNode, Ref, RefObject } from 'react';

export type FieldLayout = 'inline' | 'split' | 'icon';
export type FieldDensity = 'compact';
export type FieldTooltip = string | ReactElement;

/** What every field renders around its control. */
export interface FieldFrameProps {
	id?: string;
	label?: string;
	labelTooltip?: FieldTooltip;
	description?: string;
	layout?: FieldLayout;
	/** `compact` tightens this field and every field nested inside it. */
	density?: FieldDensity;
	className?: ClassValue;
	testId?: string;
}

export interface ControlledProps<V> {
	value: V;
	onChange: (next: V) => void;
	disabled?: boolean;
}

/** An icon already resolved by the caller; the kit never resolves one. */
export interface ResolvedIcon {
	iconUrl: string;
	name: string;
	href?: string;
}

export interface FieldClearProps {
	show: boolean;
	label?: string;
	className?: string;
	testId?: string;
	onClear: () => void;
}

export interface FieldProps extends FieldFrameProps {
	disabled?: boolean;
	/** `false` renders the root `flex-none` (ComboBox). */
	grow?: boolean;
	/** Renders the input wrapper with an absolute clear button; `show: false` keeps the wrapper and hides the button. */
	clear?: FieldClearProps;
	leading?: ReactNode;
	children?: ReactNode;
	after?: ReactNode;
	ref?: Ref<HTMLDivElement>;
	/** For a field that names a group of controls rather than one input; a `group` with a `label` and no `aria-labelledby` is named by that label. */
	role?: AriaRole;
	'aria-labelledby'?: string;
	/** Hides the field and keeps its children mounted. */
	hidden?: boolean;
	/** `BooleanPicker`'s label-before-the-control order; renders as `data-reverse`. */
	reverse?: boolean;
	/** Shares the label's id with a control the caller renders itself, such as an icon picker's swatch; defaults to one generated here. */
	labelId?: string;
	/** The control an icon field's label clicks; defaults to one generated here. */
	controlRef?: RefObject<HTMLElement | null>;
}

export interface FieldRootProps {
	/** The control's id: `Field.Label` points at it and `Field.Control` hands it to a native control. */
	id?: string;
	layout?: FieldLayout;
	density?: FieldDensity;
	disabled?: boolean;
	/** `false` renders the root and a `Field.Clear` inside it `flex-none`. */
	grow?: boolean;
	className?: ClassValue;
	testId?: string;
	/** `BooleanPicker`'s label-before-the-control order; renders as `data-reverse`. */
	reverse?: boolean;
	ref?: Ref<HTMLDivElement>;
	role?: AriaRole;
	'aria-labelledby'?: string;
	hidden?: boolean;
	children?: ReactNode;
	/** Shares the label's id with a control the caller renders itself, such as an icon picker's swatch; defaults to one generated here. */
	labelId?: string;
	/** The control an icon field's label clicks; defaults to one generated here. */
	controlRef?: RefObject<HTMLElement | null>;
}

export interface FieldLabelProps {
	/** `span` for a caption that names a group of controls, which a `<label>` cannot; the default inside a `role="group"` root. */
	as?: 'label' | 'span';
	id?: string;
	tooltip?: FieldTooltip;
	className?: ClassValue;
	testId?: string;
	children: string;
}

export interface FieldDescriptionProps {
	id?: string;
	className?: ClassValue;
	testId?: string;
	children: ReactNode;
}

export interface FieldControlProps {
	className?: ClassValue;
	testId?: string;
	children: ReactElement;
}

export interface FieldSlotProps {
	className?: ClassValue;
	testId?: string;
	children?: ReactNode;
}

export interface FieldClearPartProps extends FieldClearProps {
	children?: ReactNode;
}
