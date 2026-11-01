import i18n from '@i18n/config';
import type { ReactNode } from 'react';

export interface BulkResultTieGroupProps {
	/** The result rows whose means the paired error cannot tell apart. */
	children: ReactNode;
}

/* From `lg` the box reaches into the page gutter by its own padding (8px), so the rows inside keep the list's content width and line up with the rows outside. Below `lg` the gutter is 8px, too narrow, so the box stays inside and indents its rows. */
export const BulkResultTieGroup = ({ children }: BulkResultTieGroupProps) => (
	<div
		className="relative mb-6 flex flex-col rounded-sm px-4 py-2 ring-1 ring-border before:absolute before:inset-y-0 before:left-0 before:w-1 before:rounded-l-sm before:bg-warning lg:-mx-4"
		data-testid="bulk-results-tie-group">
		<span className="mb-6">{i18n.t('bulk_tab.results.tied_group')}</span>
		{children}
	</div>
);
