import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { SummaryTableRow } from './SummaryTableRow';

describe('SummaryTableRow', () => {
	it('takes its testId in place of the summary-table-row default', () => {
		render(<SummaryTableRow testId="gem-row" />);

		expect(screen.getByTestId('gem-row')).not.toBeNull();
		expect(screen.queryByTestId('summary-table-row')).toBeNull();
	});
});
