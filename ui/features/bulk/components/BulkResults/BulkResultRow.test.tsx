import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { BulkResultRowView, type BulkResultRowViewProps } from './BulkResultRow';

const renderRow = (props: Partial<BulkResultRowViewProps> = {}) => {
	const onEquip = vi.fn();
	render(
		<BulkResultRowView
			showDamage
			avg={1234.5}
			plusMinus={12.4}
			isBase={false}
			delta={{ text: '+34.50 (2.88%)', tone: 'positive' }}
			significance="Difference is significantly different (Z = 4.200)."
			onEquip={onEquip}
			{...props}>
			<span data-testid="gear-cells" />
		</BulkResultRowView>,
	);
	return onEquip;
};

describe('BulkResultRowView', () => {
	it('shows the average, its margin of error and the difference from the current gear', () => {
		renderRow();

		expect(screen.getByTestId('topline-result-avg').textContent).toBe('1,234.5');
		expect(screen.getByTestId('results-sim-dps').textContent).toContain('±12');
		const diff = screen.getByTestId('results-reference-diff');
		expect(diff.textContent).toBe('+34.50 (2.88%)');
		expect(diff.hasAttribute('data-tooltip-id')).toBe(true);
	});

	it('colours the difference by its tone', () => {
		renderRow({ delta: { text: '-10.00 (1.00%)', tone: 'negative' } });

		expect(screen.getByTestId('results-reference-diff').getAttribute('data-tone')).toBe('negative');
	});

	it('leaves the margin out when a run has no spread', () => {
		renderRow({ plusMinus: 0 });

		expect(screen.getByTestId('results-sim-dps').textContent).not.toContain('±');
	});

	it('labels the current gear instead of a difference and hides its equip button', () => {
		renderRow({ isBase: true });

		expect(screen.queryByTestId('results-reference-diff')).toBeNull();
		expect(screen.getByTestId('results-reference').textContent).toBe('bulk_tab.results.current_gear');
		expect(screen.getByTestId('bulk-equip-btn').hasAttribute('data-hidden')).toBe(true);
	});

	it('hides the dps block when damage is not displayed', () => {
		renderRow({ showDamage: false });

		expect(screen.queryByTestId('results-sim-dps')).toBeNull();
	});

	it('renders the gear combo it is given and equips it from its button', () => {
		const onEquip = renderRow();

		expect(screen.getByTestId('bulk-gear-combo').contains(screen.getByTestId('gear-cells'))).toBe(true);
		fireEvent.click(screen.getByTestId('bulk-equip-btn'));
		expect(onEquip).toHaveBeenCalledOnce();
	});
});
