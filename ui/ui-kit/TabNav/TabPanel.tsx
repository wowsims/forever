import { Tabs } from '@base-ui/react/tabs';
import type { ClassValue } from 'clsx';
import clsx from 'clsx';
import type { ReactNode } from 'react';

export interface TabPanelProps {
	value: string;
	id?: string;
	className?: ClassValue;
	testId?: string;
	children?: ReactNode;
}

export const TabPanel = ({ value, id, className, testId = 'tab-pane', children }: TabPanelProps) => (
	<Tabs.Panel
		value={value}
		id={id ?? value}
		keepMounted
		data-testid={testId}
		className={clsx('col-start-1 row-start-1 fade-in-out', className, 'data-ending-style:opacity-0 data-starting-style:opacity-0')}>
		{children}
	</Tabs.Panel>
);
