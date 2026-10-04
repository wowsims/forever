import { stDevToConf90 } from '@sim/utils/math';
import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { StatWeightValue } from './StatWeightValue';

describe('StatWeightValue', () => {
	it('shows the value and its 90% confidence interval to two places', () => {
		const { container, getByTestId } = render(<StatWeightValue value={1.234} stdev={2} iterations={100} sign="negative" />);
		expect(getByTestId('results-avg').textContent).toBe('1.23');
		expect(getByTestId('results-avg').getAttribute('data-sign')).toBe('negative');
		expect(container.textContent).toContain(`(${stDevToConf90(2, 100).toFixed(2)})`);
	});
});
