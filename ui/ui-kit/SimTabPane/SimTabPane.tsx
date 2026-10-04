import { TabPanelColumns } from '@ui-kit/TabPanelColumns';
import clsx from 'clsx';
import type { ComponentPropsWithoutRef, ReactNode } from 'react';

export interface SimTabPaneProps extends Omit<ComponentPropsWithoutRef<'div'>, 'id' | 'className' | 'children'> {
	id: string;
	className?: string;
	/** Classes for the columns container inside the pane. */
	contentClassName?: string;
	testId?: string;
	children: ReactNode;
}

export const SimTabPane = ({ id, className, contentClassName, testId = 'sim-tab', children, ...rest }: SimTabPaneProps) => (
	<div id={id} className={clsx(id, '@container', className)} data-testid={testId} {...rest}>
		<TabPanelColumns.Root className={contentClassName} data-testid="tab-pane-content-container">
			{children}
		</TabPanelColumns.Root>
	</div>
);
