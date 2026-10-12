import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { CombinationsCountView, type CombinationsCountViewProps } from './CombinationsCount';

const renderCount = (props: Partial<CombinationsCountViewProps> = {}) =>
	render(<CombinationsCountView pending={false} combinations={12} iterations={24000} iterationsLimit={50000} durations={1} {...props} />);

describe('CombinationsCountView', () => {
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

	it.each([
		[1, null],
		[3, '× 3'],
	])('shows the duration multiplier only for more than one duration (%i durations)', (durations, multiplier) => {
		renderCount({ durations });

		expect(screen.queryByTestId('bulk-durations-count')?.textContent ?? null).toBe(multiplier);
	});

	it('warns once the iterations pass the limit', () => {
		renderCount({ iterations: 60000 });
		const warning = screen.getByTestId('warning');

		expect(screen.getByTestId('bulk-total-iterations').hasAttribute('data-over-limit')).toBe(true);
		expect(warning.tagName).toBe('BUTTON');
		expect(warning.getAttribute('aria-label')).toBe('bulk_tab.warning.iterations_limit');
		expect(warning.hasAttribute('data-tooltip-id')).toBe(true);
	});
});
