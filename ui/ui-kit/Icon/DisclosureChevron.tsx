import clsx from 'clsx';

import { Icon } from './Icon';

export interface DisclosureChevronProps {
	/** Points down while open and right while closed. */
	open: boolean;
	className?: string;
}

// The chevron at the end of a row that opens and closes what sits under it.
export const DisclosureChevron = ({ open, className }: DisclosureChevronProps) => (
	<Icon name="chevron-down" fixedWidth className={clsx('ui-disclosure-chevron', className)} data-open={open ? '' : undefined} />
);
