import { Checkbox } from '@base-ui/react/checkbox';
import { type ControlledProps, Field, type FieldFrameProps, hasFieldFrame, useControlSize } from '@ui-kit/Field';
import { FieldContext } from '@ui-kit/Field/context/FieldContext';
import { Icon } from '@ui-kit/Icon';
import clsx from 'clsx';
import { type KeyboardEvent, useContext } from 'react';

import type { BooleanPickerOptions } from './types';

export type BooleanPickerProps = FieldFrameProps & ControlledProps<boolean> & BooleanPickerOptions;

// Rendered inside the field root, so a focusable disabled box reaches its label's tooltip. Base UI's disabled tabIndex of
// -1 gives way to one passed here; an undefined tabIndex would override Base UI's 0 too, so it is spread only when set.
// Base UI drops its own key handlers on a disabled box, so the capture handler keeps Space from scrolling the page.
// The tooltip opens at the label, not at the box, so a side placement sits beside the label text.
const Box = ({ focusable, ...props }: Checkbox.Root.Props & { focusable: boolean }) => {
	const labelTooltip = useContext(FieldContext)?.labelTooltip;
	const focusableProps = focusable
		? {
				tabIndex: 0,
				onKeyDownCapture: (event: KeyboardEvent) => event.key === ' ' && event.preventDefault(),
				onFocus: () => labelTooltip?.current?.open(),
				onBlur: () => labelTooltip?.current?.close(),
			}
		: {};
	return <Checkbox.Root {...props} {...focusableProps} />;
};

export const BooleanPicker = ({
	value,
	onChange,
	disabled,
	reverse,
	invalid,
	className,
	testId,
	'aria-describedby': describedBy,
	focusableWhenDisabled,
	...frameProps
}: BooleanPickerProps) => {
	const hasFrame = hasFieldFrame(frameProps);
	const size = useControlSize(frameProps.size);

	const input = (
		<Box
			focusable={!!(disabled && focusableWhenDisabled)}
			id={frameProps.id}
			data-size={size}
			aria-invalid={invalid || undefined}
			aria-describedby={describedBy}
			className={clsx('ui-boolean-picker-input inline-flex items-center justify-center p-0', !hasFrame && className)}
			data-testid={hasFrame ? 'boolean-picker-input' : (testId ?? 'boolean-picker-input')}
			checked={value}
			disabled={disabled}
			onCheckedChange={checked => onChange(checked)}>
			<Checkbox.Indicator className="ui-boolean-picker-indicator">
				<Icon name="check" className="text-primary-foreground" />
			</Checkbox.Indicator>
		</Box>
	);

	if (!hasFrame) return input;

	return (
		<Field
			{...frameProps}
			className={['ui-boolean-picker', className]}
			testId={testId ?? 'boolean-picker-root'}
			disabled={disabled}
			reverse={reverse}
			leading={reverse ? undefined : input}>
			{reverse ? input : undefined}
		</Field>
	);
};
