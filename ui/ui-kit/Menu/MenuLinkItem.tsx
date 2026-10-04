import { Menu as BaseMenu } from '@base-ui/react/menu';
import clsx from 'clsx';
import type { ReactNode } from 'react';

const LAYOUT_CLASSES = {
	block: '',
	row: 'ui-menu-item-row',
} as const;

export interface MenuLinkItemProps {
	layout: 'block' | 'row';
	href: string;
	closeOnClick?: boolean;
	/** Skips the item's own padding and highlight, for content that brings both. */
	bare?: boolean;
	className?: string;
	testId?: string;
	children: ReactNode;
	[key: string]: unknown;
}

export const MenuLinkItem = ({ layout, href, closeOnClick, bare, className, testId, children, ...rest }: MenuLinkItemProps) => (
	<li role="none">
		<BaseMenu.LinkItem
			href={href}
			closeOnClick={closeOnClick}
			className={clsx('ui-menu-item', LAYOUT_CLASSES[layout], className)}
			data-bare={bare ? '' : undefined}
			data-testid={testId}
			{...rest}>
			{children}
		</BaseMenu.LinkItem>
	</li>
);
