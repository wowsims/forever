import i18n from '@i18n/config';
import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ReforgeBreakpointLimitsView } from './ReforgeBreakpointLimits';

const ROWS = [
	{ key: 'hit', statName: 'Hit', picker: <span data-testid="hit-limit" /> },
	{ key: 'haste', statName: 'Haste', picker: <span data-testid="haste-limit" /> },
];

describe('ReforgeBreakpointLimitsView', () => {
	it('titles the limits and names their explanation button', () => {
		render(<ReforgeBreakpointLimitsView rows={ROWS} />);

		expect(screen.getByRole('heading').textContent).toBe('sidebar.buttons.suggest_reforges.breakpoint_limit');
		const button = screen.getByTestId('reforge-breakpoint-limit-help');
		expect(button.hasAttribute('data-tooltip-id')).toBe(true);
		expect(screen.getByRole('button', { name: i18n.t('common.help') })).toBe(button);
	});

	it('lists one stat per row beside its limit select', () => {
		render(<ReforgeBreakpointLimitsView rows={ROWS} />);
		const rows = screen.getAllByRole('row');

		expect(rows.map(row => row.children[0].textContent)).toEqual(['Hit', 'Haste']);
		expect(rows[1].children[1].contains(screen.getByTestId('haste-limit'))).toBe(true);
	});
});
