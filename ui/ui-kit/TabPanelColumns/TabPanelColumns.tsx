import clsx from 'clsx';
import type { ElementType, ReactNode } from 'react';

export type TabPanelColumnsGap = 'default' | 'apl';

export interface TabPanelColumnsRootProps {
	as?: ElementType;
	gap?: TabPanelColumnsGap;
	fullWidth?: boolean;
	externalDisplay?: boolean;
	className?: string;
	testId?: string;
	children?: ReactNode;
	[key: string]: unknown;
}

const ROOT_GAP_CLASSES: Record<TabPanelColumnsGap, string> = {
	default: 'ui-columns',
	apl: 'ui-columns-apl',
};

const Root = ({ as: Component = 'div', gap = 'default', fullWidth, externalDisplay, className, testId, children, ...rest }: TabPanelColumnsRootProps) => (
	<Component
		className={clsx(!externalDisplay && 'flex', 'flex-wrap max-xl:flex-col', ROOT_GAP_CLASSES[gap], fullWidth && 'w-full', className)}
		data-testid={testId}
		{...rest}>
		{children}
	</Component>
);

export type TabPanelColumnsLeftVariant = 'default' | 'auto-columns' | 'settings-columns' | 'stacked';

export interface TabPanelColumnsLeftProps {
	className?: string;
	variant?: TabPanelColumnsLeftVariant;
	testId?: string;
	children?: ReactNode;
}

const LEFT_VARIANT_CLASSES: Record<TabPanelColumnsLeftVariant, string> = {
	default: 'ui-columns-left',
	'auto-columns': 'ui-columns-left-auto',
	'settings-columns': 'ui-columns-left-settings',
	stacked: 'ui-columns-left-stacked',
};

const Left = ({ className, variant = 'default', testId = 'tab-panel-left', children }: TabPanelColumnsLeftProps) => (
	<div className={clsx('w-full flex-4 gap-section', LEFT_VARIANT_CLASSES[variant], className)} data-testid={testId}>
		{children}
	</div>
);

export interface TabPanelColumnsRightProps {
	className?: string;
	testId?: string;
	children?: ReactNode;
}

const Right = ({ className, testId = 'tab-panel-right', children }: TabPanelColumnsRightProps) => (
	<div className={clsx('ui-columns-right', className)} data-testid={testId}>
		{children}
	</div>
);

export interface TabPanelColumnsColProps {
	className?: string;
	externalGap?: boolean;
	testId?: string;
	children?: ReactNode;
}

const Col = ({ className, externalGap, testId = 'tab-panel-col', children }: TabPanelColumnsColProps) => (
	<div className={clsx('ui-columns-col', !externalGap && 'gap-section', className)} data-testid={testId}>
		{children}
	</div>
);

export const TabPanelColumns = { Root, Left, Right, Col };
