import i18n from '@i18n/config';
import type { ReactNode } from 'react';

export interface BulkResultTieGroupProps {
	/** The result rows whose means the paired error cannot tell apart. */
	children: ReactNode;
}

export const BulkResultTieGroup = ({ children }: BulkResultTieGroupProps) => (
	<div className="mb-6 flex flex-col rounded-sm border border-l-3 border-border border-l-warning px-4 py-2" data-testid="bulk-results-tie-group">
		<span className="mb-6">{i18n.t('bulk_tab.results.tied_group')}</span>
		{children}
	</div>
);
