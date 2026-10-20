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
			talents={null}
			gearUnchanged={false}
			onEquip={onEquip}
			{...props}>
			<span data-testid="gear-cells" />
		</BulkResultRowView>,
	);
	return onEquip;
};

describe('BulkResultRowView', () => {
	it('shows the average with its unit and margin of error, then the difference from the current gear as a pill', () => {
		renderRow();

		expect(screen.getByTestId('topline-result-avg').textContent).toBe('1,234.50 sidebar.results.metrics.dps.label');
		const margin = screen.getByTestId('topline-result-stdev');
		expect(margin.textContent).toBe('(12)');
		expect(margin.hasAttribute('data-tooltip-id')).toBe(true);
		const diff = screen.getByTestId('results-reference-diff');
		expect(diff.textContent).toBe('+34.50 (2.88%)');
		expect(diff.getAttribute('data-tooltip-content')).toBe('Difference is significantly different (Z = 4.200).');
		expect(screen.getByTestId('results-reference').textContent).toBe('+34.50 (2.88%)');
	});

	it('tones the difference pill by its sign and significance', () => {
		renderRow({ delta: { text: '-10.00 (1.00%)', tone: 'negative' } });
		expect(screen.getByTestId('results-reference').getAttribute('data-tone')).toBe('negative');
		expect(screen.getByTestId('results-reference-diff').getAttribute('data-sign')).toBe('negative');
	});

	it('draws a not significant difference in the neutral tone', () => {
		renderRow({ delta: { text: '+0.41 (0.03%)', tone: null } });
		expect(screen.getByTestId('results-reference').getAttribute('data-tone')).toBe('neutral');
	});

	it('leaves the margin out when a run has no spread', () => {
		renderRow({ plusMinus: 0 });

		expect(screen.queryByTestId('topline-result-stdev')).toBeNull();
	});

	it('labels the current gear above its value, with no difference pill, and hides its equip button', () => {
		renderRow({ isBase: true });

		expect(screen.queryByTestId('results-reference')).toBeNull();
		expect(screen.getByTestId('results-sim-dps').firstElementChild!.textContent).toBe('bulk_tab.results.current_gear');
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

	it('names the row’s talent setup above the gear and as a talent preset', () => {
		renderRow({ talents: { label: 'Arms', talentsString: '3030500130', isCurrent: false } });

		const line = screen.getByTestId('bulk-result-talents');
		expect(line.textContent).toBe('bulk_tab.results.talent_preset');
		expect(screen.getByTestId('bulk-gear-combo').firstElementChild).toBe(line);
		expect(screen.getByTestId('gear-cells')).toBeTruthy();
	});

	it('names the current talents plainly, without the talent preset wording', () => {
		renderRow({ talents: { label: 'bulk_tab.results.current_talents', talentsString: '3030500130', isCurrent: true } });

		expect(screen.getByTestId('bulk-result-talents').textContent).toBe('bulk_tab.results.current_talents');
	});

	it('shows no talent line when the encounter ran one setup', () => {
		renderRow({ talents: null });

		expect(screen.queryByTestId('bulk-result-talents')).toBeNull();
	});

	it('says the gear is unchanged for the current gear with another talent setup, and offers Equip', () => {
		renderRow({ gearUnchanged: true, talents: { label: 'Arms', talentsString: '3030500130', isCurrent: false } });

		expect(screen.getByTestId('bulk-result-gear-unchanged').textContent).toBe('bulk_tab.results.gear_unchanged');
		expect(screen.queryByTestId('gear-cells')).toBeNull();
		expect(screen.getByTestId('results-reference')).toBeTruthy();
		expect(screen.getByTestId('bulk-equip-btn').hasAttribute('data-hidden')).toBe(false);
	});
});
