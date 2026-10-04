import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { Table } from './Table';
import { TableCell } from './TableCell';
import { TableHeaderCell } from './TableHeaderCell';
import { TableRow } from './TableRow';

const mount = (props: Parameters<typeof Table>[0] = {}) =>
	render(
		<Table {...props}>
			<thead>
				<TableRow testId="header-row">
					<TableHeaderCell testId="header-cell">Name</TableHeaderCell>
					<TableHeaderCell>DPS</TableHeaderCell>
				</TableRow>
			</thead>
			<tbody>
				<TableRow>
					<TableCell testId="body-cell">Fireball</TableCell>
					<TableCell>1200</TableCell>
				</TableRow>
				<TableRow>
					<TableCell colSpan={2}>Total</TableCell>
				</TableRow>
			</tbody>
		</Table>,
	);

const FLAGS = ['striped', 'hover', 'dense', 'flush'] as const;

describe('Table', () => {
	it('renders a table of kit rows and cells', () => {
		mount();

		expect(screen.getByTestId('table').tagName).toBe('TABLE');
		expect(screen.getByTestId('header-row').tagName).toBe('TR');
		expect(screen.getByTestId('header-cell').tagName).toBe('TH');
		expect(screen.getByTestId('body-cell').tagName).toBe('TD');
	});

	it.each([
		['no variant by default', {}, null],
		['each variant it is given as a data attribute', { striped: true, hover: true, dense: true, flush: true }, ''],
	])('marks %s', (_, props, expected) => {
		mount({ ...props, testId: 'dps-table' });

		const table = screen.getByTestId('dps-table');
		for (const flag of FLAGS) expect(table.getAttribute(`data-${flag}`)).toBe(expected);
	});

	it('marks solid striping apart from the shaded kind', () => {
		mount({ striped: 'solid' });
		expect(screen.getByTestId('table').getAttribute('data-striped')).toBe('solid');
	});

	it('passes other attributes through', () => {
		mount({ 'aria-label': 'Damage' });

		expect(screen.getByTestId('table').getAttribute('aria-label')).toBe('Damage');
		expect(screen.getByText('Total').getAttribute('colspan')).toBe('2');
	});

	it('marks the plain variant, and nothing for the default one', () => {
		const { unmount } = mount({ variant: 'plain' });
		expect(screen.getByTestId('table').getAttribute('data-variant')).toBe('plain');
		unmount();

		mount({ variant: 'default' });
		expect(screen.getByTestId('table').hasAttribute('data-variant')).toBe(false);
	});

	it('keeps a header cell on one line unless it is told to wrap', () => {
		render(
			<Table>
				<thead>
					<TableRow>
						<TableHeaderCell testId="fixed">Damage per second</TableHeaderCell>
						<TableHeaderCell testId="wrapping" wrap>
							Casts per minute
						</TableHeaderCell>
					</TableRow>
				</thead>
			</Table>,
		);

		expect(screen.getByTestId('fixed').hasAttribute('data-wrap')).toBe(false);
		expect(screen.getByTestId('wrapping').getAttribute('data-wrap')).toBe('');
	});
});
