import { kebabCase } from '@sim/utils/format';
import clsx, { type ClassValue } from 'clsx';
import type { ReactNode } from 'react';

export interface MenuSectionProps {
	name: string;
	/** Lays out the section's controls; a column with a gap when omitted. */
	className?: ClassValue;
	children: ReactNode;
}

export const MenuSection = ({ name, className, children }: MenuSectionProps) => (
	<div className={clsx('flex max-sm:block', `${kebabCase(name)}-section`)}>
		<div className="flex-1">
			<h6 data-testid="menu-section-title">{name}</h6>
		</div>
		<div className={clsx('flex-3', className ?? 'flex flex-col gap-2')}>{children}</div>
	</div>
);
