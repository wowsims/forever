import { Accordion as BaseAccordion } from '@base-ui/react/accordion';
import { Box } from '@ui-kit/Box';
import { Icon } from '@ui-kit/Icon';
import type { ClassValue } from 'clsx';
import clsx from 'clsx';
import { type ReactNode, use } from 'react';

import { AccordionVariantContext } from './Accordion';

export interface AccordionItemProps {
	value: string;
	title: ReactNode;
	panelClassName?: ClassValue;
	testId?: string;
	children?: ReactNode;
}

export const AccordionItem = ({ value, title, panelClassName, testId, children }: AccordionItemProps) => (
	<BaseAccordion.Item
		className="ui-accordion-item"
		value={value}
		data-testid={testId}
		render={use(AccordionVariantContext) === 'boxed' ? <Box border="border-muted" padding="none" /> : undefined}>
		<BaseAccordion.Trigger className="ui-accordion-trigger" data-testid={testId ? `${testId}-trigger` : undefined}>
			{title}
			<Icon name="chevron-down" className="ui-accordion-chevron" />
		</BaseAccordion.Trigger>
		<BaseAccordion.Panel className={clsx('ui-accordion-panel', panelClassName)}>{children}</BaseAccordion.Panel>
	</BaseAccordion.Item>
);
