import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { downloadString } from '@ui-kit/utils/dom';

import { useBulkState } from '../../hooks/useBulkState';
import { BULK_RESULTS_CSV_FILE_NAME, bulkResultsCsv } from '../../model/results_csv';

export interface BulkResultsCsvButtonViewProps {
	onExport: () => void;
}

export const BulkResultsCsvButtonView = ({ onExport }: BulkResultsCsvButtonViewProps) => (
	<Button variant="outline-primary" size="sm" className="self-start" data-testid="bulk-results-export-csv" onClick={onExport}>
		<Icon name="download" />
		{i18n.t('bulk_tab.results.export_csv')}
	</Button>
);

// Exports every finished encounter of the last run; renders nothing before a run has finished one.
export const BulkResultsCsvButton = () => {
	const results = useBulkState(slice => slice.results);
	if (!results) return null;
	return <BulkResultsCsvButtonView onExport={() => downloadString(bulkResultsCsv(results), BULK_RESULTS_CSV_FILE_NAME, 'text/csv')} />;
};
