import i18n from '@i18n/config';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import english from '../../../../../assets/locales/en/translation.json';
import { CombinationsCountView, type CombinationsCountViewProps } from './CombinationsCount';

const renderCount = (props: Partial<CombinationsCountViewProps> = {}) =>
	render(<CombinationsCountView pending={false} ready combinations={12} iterations={24000} iterationsLimit={50000} builds={null} {...props} />);

describe('CombinationsCountView', () => {
	it('says why the selection gives no combination, once the count is in', () => {
		renderCount({ combinations: 0, iterations: 0, error: 'no equippable pair of items available for grouped bulk slot 11' });
		expect(screen.getByTestId('bulk-combinations-error').textContent).toBe('no equippable pair of items available for grouped bulk slot 11');
		cleanup();
		renderCount({ pending: true, error: 'stale' });
		expect(screen.queryByTestId('bulk-combinations-error')).toBeNull();
	});

	it('shows a spinner in place of both numbers while the count is pending', () => {
		renderCount({ pending: true });

		expect(screen.getByTestId('bulk-combinations').querySelector('[data-testid="loader"]')).not.toBeNull();
		expect(screen.getByTestId('bulk-total-iterations').querySelector('[data-testid="loader"]')).not.toBeNull();
	});

	it('counts the combinations and the iterations they cost', () => {
		renderCount();

		expect(screen.getByTestId('bulk-combinations').textContent).toBe('12');
		expect(screen.getByTestId('bulk-total-iterations').textContent).toBe('24,000');
		expect(screen.getByTestId('bulk-total-iterations').hasAttribute('data-over-limit')).toBe(false);
		expect(screen.queryByTestId('warning')).toBeNull();
	});

	it('explains through a help icon beside the iterations label that the count is a minimum', async () => {
		renderCount();
		const help = screen.getByTestId('bulk-total-iterations-help');

		expect(help.parentElement!.textContent).toBe('bulk_tab.settings.total_iterations');
		fireEvent.focus(help);
		expect(await screen.findByText('bulk_tab.settings.total_iterations_tooltip')).toBeTruthy();
	});

	it('warns once the iterations pass the limit', () => {
		renderCount({ iterations: 60000 });
		const warning = screen.getByTestId('warning');

		expect(screen.getByTestId('bulk-total-iterations').hasAttribute('data-over-limit')).toBe(true);
		expect(warning.tagName).toBe('BUTTON');
		expect(warning.getAttribute('aria-label')).toBe('bulk_tab.warning.iterations_limit');
		expect(warning.hasAttribute('data-tooltip-id')).toBe(true);
	});

	it('ends the warning with a hint to lower Top results while the control shows', () => {
		renderCount({ iterations: 60000, topResultsHint: true });

		expect(screen.getByTestId('warning').getAttribute('aria-label')).toBe('bulk_tab.warning.iterations_limit bulk_tab.warning.lower_top_results');
	});

	it('names the bundles and encounters a Builds batch runs, and the runs they give', () => {
		const builds = { setups: 4, bundles: 3, encounters: 2, runs: 6, builds: 12, totalIterations: 120000, largestRequestIterations: 30000 };
		// The other cases read the keys; this one reads the English copy.
		i18n.addResourceBundle('en', 'translation', english, true, true);
		try {
			renderCount({ builds });
			expect(screen.getByTestId('bulk-runs-label').textContent).toBe('3 bundles × 2 encounters');
			expect(screen.getByTestId('bulk-runs-count').textContent).toBe('6 runs');
			cleanup();

			renderCount({ builds: { ...builds, bundles: 1, encounters: 1, runs: 1 } });
			expect(screen.getByTestId('bulk-runs-label').textContent).toBe('1 bundle × 1 encounter');
		} finally {
			i18n.removeResourceBundle('en', 'translation');
		}
	});

	it('asks for a second build while the batch sims one, and drops the hint once it sims two', () => {
		const builds = { setups: 1, bundles: 1, encounters: 1, runs: 1, builds: 1, totalIterations: 20000, largestRequestIterations: 20000 };
		renderCount({ builds });
		expect(screen.getByTestId('bulk-runs-hint').textContent).toBe('bulk_tab.settings.runs.second_build');
		cleanup();

		renderCount({ builds: { ...builds, encounters: 2, runs: 2, builds: 2 } });
		expect(screen.queryByTestId('bulk-runs-hint')).toBeNull();
	});
});
