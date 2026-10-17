import { Button } from '@ui-kit/Button';
import { ButtonGroup } from '@ui-kit/ButtonGroup';
import type { ReactNode } from 'react';

export interface ToggleGroupOption<T extends string> {
	value: T;
	label: ReactNode;
}

export interface ToggleGroupProps<T extends string> {
	options: ReadonlyArray<ToggleGroupOption<T>>;
	value: T;
	/** Called with the clicked option; a click on the option already pressed reports nothing. */
	onChange: (next: T) => void;
	'aria-label'?: string;
	size?: 'sm' | 'xs';
	disabled?: boolean;
	className?: string;
	testId?: string;
}

export const ToggleGroup = <T extends string>({
	options,
	value,
	onChange,
	'aria-label': ariaLabel,
	size,
	disabled,
	className,
	testId,
}: ToggleGroupProps<T>) => (
	<ButtonGroup className={className} size={size} aria-label={ariaLabel} testId={testId}>
		{options.map(option => (
			<Button
				key={option.value}
				size={size}
				variant={value === option.value ? 'primary' : 'outline-primary'}
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
