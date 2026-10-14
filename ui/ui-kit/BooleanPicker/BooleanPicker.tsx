import { Checkbox } from '@base-ui/react/checkbox';
import { type ControlledProps, Field, type FieldFrameProps, hasFieldFrame } from '@ui-kit/Field';
import { Icon } from '@ui-kit/Icon';
import clsx from 'clsx';

import type { BooleanPickerOptions } from './types';

export type BooleanPickerProps = FieldFrameProps & ControlledProps<boolean> & BooleanPickerOptions;

export const BooleanPicker = ({
	value,
	onChange,
	disabled,
	reverse,
	size,
	invalid,
	className,
	testId,
	'aria-describedby': describedBy,
	...frameProps
}: BooleanPickerProps) => {
	const hasFrame = hasFieldFrame(frameProps);

	const input = (
		<Checkbox.Root
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
		</Checkbox.Root>
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
