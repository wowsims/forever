import type { ControlledProps, FieldFrameProps } from '@ui-kit/Field';

export interface EnumValueConfig {
	name: string;
	value: number;
	tooltip?: string;
}

export interface MultiComboBoxProps extends FieldFrameProps, ControlledProps<Array<number>> {
	values: Array<EnumValueConfig>;
	placeholder?: string;
	emptyText?: string;
	removeLabel?: string;
	openLabel?: string;
	/** Sizes the field to its widest option name or placeholder instead of the row; the chips wrap inside that width. */
	fitOptions?: boolean;
}
