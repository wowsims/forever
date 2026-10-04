import { Accordion as BaseAccordion } from '@base-ui/react/accordion';
import type { ClassValue } from 'clsx';
import clsx from 'clsx';
import { createContext, type ReactNode } from 'react';

/** `divided` rules the items apart; `boxed` draws each item as its own bordered box, with a gap between them. */
export type AccordionVariant = 'divided' | 'boxed';

export const AccordionVariantContext = createContext<AccordionVariant>('divided');

export interface AccordionProps {
	value: string[];
	onValueChange: (value: string[]) => void;
	variant?: AccordionVariant;
	className?: ClassValue;
	testId?: string;
	children?: ReactNode;
}

export const Accordion = ({ value, onValueChange, variant = 'divided', className, testId = 'accordion', children }: AccordionProps) => (
	<BaseAccordion.Root
		className={clsx('ui-accordion', className)}
		multiple
		value={value}
		onValueChange={onValueChange}
		data-variant={variant}
		data-testid={testId}>
		<AccordionVariantContext value={variant}>{children}</AccordionVariantContext>
	</BaseAccordion.Root>
);
