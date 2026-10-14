import i18n from '@i18n/config';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { ReforgeStatCapRowView } from './ReforgeStatCapRow';
import { ReforgeStatCapsView } from './ReforgeStatCaps';

const renderCaps = () => {
	const onReset = vi.fn();
	render(
		<ReforgeStatCapsView onReset={onReset}>
			<ReforgeStatCapRowView statName="Hit" percentage={<span data-testid="hit-percentage" />} undershoot={<span data-testid="hit-undershoot" />} />
		</ReforgeStatCapsView>,
	);
	return onReset;
};

describe('ReforgeStatCapsView', () => {
	it('titles the table and heads its stat, percentage and undershoot columns', () => {
		renderCaps();

		expect(screen.getByRole('heading').textContent).toBe('sidebar.buttons.suggest_reforges.edit_stat_caps');
		expect(screen.getAllByRole('columnheader').map(header => header.textContent)).toEqual([
			'sidebar.buttons.suggest_reforges.stat',
			'%',
			'sidebar.buttons.suggest_reforges.max_column',
		]);
	});

	it('names the caps explanation and the reset as the only two buttons', () => {
		renderCaps();

		expect(screen.getByTestId('reforge-stat-caps-help').hasAttribute('data-tooltip-id')).toBe(true);
		expect(screen.getAllByRole('button').map(button => button.getAttribute('aria-label'))).toEqual([
			i18n.t('common.help'),
			i18n.t('sidebar.buttons.suggest_reforges.reset_to_defaults'),
		]);
	});

	it('puts the rows it is given in the table body', () => {
		renderCaps();
		const [, body] = screen.getAllByRole('rowgroup');

		expect(within(body).getByText('Hit')).toBeTruthy();
		expect(within(body).getByTestId('hit-percentage')).toBeTruthy();
	});

	it('resets the caps to the defaults', () => {
		const onReset = renderCaps();

		fireEvent.click(screen.getByRole('button', { name: 'sidebar.buttons.suggest_reforges.reset_to_defaults' }));

		expect(onReset).toHaveBeenCalledOnce();
	});
});
