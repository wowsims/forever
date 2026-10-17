import { Button, type ButtonVariant } from '@ui-kit/Button';
import { ButtonGroup, type ButtonGroupProps } from '@ui-kit/ButtonGroup';
import type { ReactNode } from 'react';

export interface ToggleGroupOption<T extends string> {
	value: T;
	label: ReactNode;
}

export type ToggleGroupTone = 'accent' | 'neutral';

export interface ToggleGroupProps<T extends string> {
	options: ReadonlyArray<ToggleGroupOption<T>>;
	value: T;
	/** Called with the clicked option; a click on the option already pressed reports nothing. */
	onChange: (next: T) => void;
	'aria-label'?: string;
	size?: 'sm' | 'xs';
	/** `neutral` draws the buttons in grey, for a switch that should not draw the eye. */
	tone?: ToggleGroupTone;
	orientation?: ButtonGroupProps['orientation'];
	disabled?: boolean;
	className?: string;
	testId?: string;
}

const VARIANTS: Record<ToggleGroupTone, { pressed: ButtonVariant; unpressed: ButtonVariant }> = {
	accent: { pressed: 'primary', unpressed: 'outline-primary' },
	neutral: { pressed: 'secondary', unpressed: 'outline-secondary' },
};

export const ToggleGroup = <T extends string>({
	options,
	value,
	onChange,
	'aria-label': ariaLabel,
	size,
	tone = 'accent',
	orientation,
	disabled,
	className,
	testId,
}: ToggleGroupProps<T>) => (
	<ButtonGroup className={className} size={size} orientation={orientation} aria-label={ariaLabel} testId={testId}>
		{options.map(option => (
			<Button
				key={option.value}
				size={size}
				variant={value === option.value ? VARIANTS[tone].pressed : VARIANTS[tone].unpressed}
				aria-pressed={value === option.value}
				disabled={disabled}
				onClick={() => {
					if (value !== option.value) onChange(option.value);
				}}>
				{option.label}
			</Button>
		))}
	</ButtonGroup>
);
