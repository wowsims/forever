// The dialog over a real store holding three finished encounters. With no locale loaded every string is its key, so sets
// are told apart by their gains.
//
// The fixture, in display order: Council (1:30) ranks S1 1,030, S2 1,020, S3 990; Patchwerk (5:00) S2 1,100, S4 1,090,
// S5 1,080; Ragnaros (health-based) S6 12,500, S1 1,150, S2 1,120. The current gear does 1,000 in each, and every result
// spreads ±10 a fight.
//
// The talent fixture runs the current talents and Arms in Council and Patchwerk: S1 is helm 2 with Arms, S2 the current
// gear with Arms, S3 helm 3 with the current talents.
import type { BulkEncounterResults } from '@sim/bulk/types';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { BULK_COMPARE_TALENT_ENTRIES, makeBulkCompareHost } from '@testing/storybook/fixtures/bulk-compare';
import { render, screen, within } from '@testing-library/react';
import { PortalContainerContext } from '@ui-kit/hooks/usePortalContainer';
import { describe, expect, it } from 'vitest';

import { BulkCompareDialog } from './BulkCompareDialog';

const rootElem = document.createElement('div');

const mount = (entries?: BulkEncounterResults[]) => {
	document.body.appendChild(rootElem);
	render(
		<PortalContainerContext value={rootElem}>
			<SimHostProvider host={makeBulkCompareHost(entries)}>
				<BulkCompareDialog open onOpenChange={() => {}} />
			</SimHostProvider>
		</PortalContainerContext>,
	);
};

const dialog = () => within(screen.getByRole('dialog'));
const bodyRows = () => dialog().getAllByRole('row').slice(1);
const cells = (row: HTMLElement) => within(row).getAllByRole('cell');
const shown = (row: HTMLElement) =>
	cells(row).map(cell => (cell.hasAttribute('data-gap') ? 'gap' : `${cell.firstElementChild!.textContent}${cell.hasAttribute('data-best') ? ' best' : ''}`));

describe('BulkCompareDialog', () => {
	it('shows the table straight away, with no view switch', () => {
		mount();

		expect(dialog().getByRole('table')).toBeTruthy();
		expect(screen.getByRole('dialog').querySelector('[aria-pressed]')).toBeNull();
	});

	it('gives every encounter a column by length, the health-based one last', () => {
		mount();

		const headers = within(dialog().getAllByRole('row')[0]).getAllByRole('columnheader');
		expect(headers.map(header => header.firstChild!.textContent)).toEqual(['bulk_tab.compare.table_set', 'Council', 'Patchwerk', 'Ragnaros']);
	});

	it('lists the current gear, then every set by number, with gaps and each encounter’s best marked', () => {
		mount();

		const rows = bodyRows();
		expect(rows.map(row => within(row).getByRole('rowheader').firstChild!.textContent)).toEqual([
			'bulk_tab.results.current_gear',
			...Array.from({ length: 6 }, () => 'bulk_tab.compare.short_set'),
		]);
		const [current, s1, , , , , s6] = rows;
		expect(shown(current)).toEqual(['0.00%', '0.00%', '0.00%']);
		expect(shown(s1)).toEqual(['+3.00% best', 'gap', '+15.00%']);
		expect(shown(s6)).toEqual(['gap', 'gap', '+1,150.00% best']);
	});

	it('names the items a set changes under its label', () => {
		mount();

		const [, s1] = bodyRows();
		expect(within(s1).getByText('Helm 3')).toBeTruthy();
	});

	it('gives each cell its rank, spread and gain as a tooltip, and the current gear’s its DPS and spread only', () => {
		mount();

		const [current, s1, , s3] = bodyRows();
		expect(cells(s1)[0].getAttribute('data-tooltip-content')).toBe('bulk_tab.compare.readout_rank bulk_tab.compare.readout_spread · +30.00 (+3.00%)');
		expect(cells(s3)[0].getAttribute('data-tooltip-content')).toBe('bulk_tab.compare.readout_rank bulk_tab.compare.readout_spread · -10.00 (-1.00%)');
		expect(cells(current)[0].getAttribute('data-tooltip-content')).toBe('bulk_tab.compare.readout_current bulk_tab.compare.readout_spread');
	});

	it('names the talent setup under each row’s label when the batch ran more than one', () => {
		mount(BULK_COMPARE_TALENT_ENTRIES);

		const rows = bodyRows();
		expect(rows.map(row => within(row).getByRole('rowheader').textContent)).toEqual([
			'bulk_tab.results.current_gearbulk_tab.results.current_talents',
			'bulk_tab.compare.short_setArmsHelm 2',
			'bulk_tab.compare.short_setArmsbulk_tab.results.gear_unchanged',
			'bulk_tab.compare.short_setbulk_tab.results.current_talentsHelm 3',
		]);
		expect(shown(rows[2])).toEqual(['+3.00%', 'gap']);
	});

	it('names no talent setup when every encounter ran one', () => {
		mount();

		const [current, s1] = bodyRows();
		expect(within(current).getByRole('rowheader').textContent).toBe('bulk_tab.results.current_gear');
		expect(within(s1).getByRole('rowheader').textContent).toBe('bulk_tab.compare.short_setHelm 3');
	});
});
