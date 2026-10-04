import { Button, type ButtonAsButton, type ButtonVariant } from '@ui-kit/Button';
import { Spinner } from '@ui-kit/Spinner';
import clsx from 'clsx';
import { useContext } from 'react';

import { SidebarDisabledContext } from './context/sidebar_disabled';

export type SidebarActionButtonProps = Pick<ButtonAsButton, 'children' | 'className' | 'disabled' | 'loading'> &
	Required<Pick<ButtonAsButton, 'onClick'>> & {
		variant?: Extract<ButtonVariant, 'primary' | 'outline-primary' | 'cancel'>;
		testId?: string;
	};

export const SidebarActionButton = ({ children, className, onClick, variant = 'primary', disabled, loading, testId }: SidebarActionButtonProps) => {
	const sidebarDisabled = useContext(SidebarDisabledContext);

	return (
		<Button
			variant={variant}
			className={clsx('ui-sidebar-action w-full', className)}
			data-testid={testId}
			data-sidebar-action=""
			data-variant={variant}
			data-loading={loading ? '' : undefined}
			onClick={onClick}
			disabled={disabled || sidebarDisabled || loading}
			aria-busy={loading || undefined}>
			{children}
			{loading && <Spinner size="xs" className="ui-sidebar-action-spinner" />}
		</Button>
	);
};
