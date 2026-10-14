import clsx from 'clsx';
import type { ReactNode } from 'react';

export interface DialogColumnsPartProps {
	className?: string;
	children?: ReactNode;
}

// At lg and up the row shrinks with a capped dialog body, so the body has nothing to scroll: the main column and the
// sidebar fill its height and scroll on their own. Below lg the sidebar stacks above the main column.
const Root = ({ className, children }: DialogColumnsPartProps) => <div className={clsx('ui-dialog-columns', className)}>{children}</div>;

const Main = ({ className, children }: DialogColumnsPartProps) => <div className={clsx('ui-dialog-columns-main', className)}>{children}</div>;

const Sidebar = ({ className, children }: DialogColumnsPartProps) => <div className={clsx('ui-dialog-columns-sidebar', className)}>{children}</div>;

export const DialogColumns = { Root, Main, Sidebar };
