import { act, cleanup, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { BulkBundleList } from '../BulkBundleList';
import { EncounterListView } from '../EncounterList';
import { BulkCheckList, type BulkCheckRow } from './BulkCheckList';

const row = (name: string, summary: string, preset: boolean, extra: Partial<BulkCheckRow> = {}): BulkCheckRow => ({
	ref: { name, preset },
	name,
	summary,
	checked: true,
	duplicate: false,
	...extra,
});

const descriptions = (element: Element) => (element.getAttribute('aria-describedby') ?? '').split(' ').map(id => document.getElementById(id)?.textContent);
const box = (name: string) => screen.getByRole('checkbox', { name });

describe('the bulk check lists', () => {
	it('describe each row by its summary, its duplicate note and the reason it is disabled, a bundle row’s buttons too', () => {
		render(
			<EncounterListView
				current={{ label: 'Current encounter', summary: '3:00 ±5s · 1 target', checked: true, onCheckedChange: vi.fn() }}
				presets={[row('Patchwerk', '5:00 ±5s · 1 target', true)]}
				saved={[row('Patchwerk copy', '5:00 ±5s · 1 target', false, { disabledReason: 'Same as Patchwerk' })]}
				onCheckedChange={vi.fn()}
			/>,
		);
		expect(descriptions(box('Current encounter'))).toEqual(['3:00 ±5s · 1 target']);
		expect(descriptions(box('Patchwerk'))).toEqual(['5:00 ±5s · 1 target']);
		expect(descriptions(box('Patchwerk copy'))).toEqual(['5:00 ±5s · 1 target', 'Same as Patchwerk']);
		cleanup();

		render(
			<BulkCheckList
				strings={{ saved: 'Saved', empty: 'No enchant presets yet.', duplicate: 'Same enchants as another preset, skipped' }}
				current={null}
				presets={[]}
				saved={[row('Agility cloak', '1 slot.', false), row('Agility cloak again', '1 slot.', false, { duplicate: true })]}
				onCheckedChange={vi.fn()}
				testIds={{ list: 'bulk-enchant-picks', row: 'bulk-enchant-pick' }}
			/>,
		);
		expect(descriptions(box('Agility cloak'))).toEqual(['1 slot.']);
		expect(descriptions(box('Agility cloak again'))).toEqual(['1 slot.', 'Same enchants as another preset, skipped']);
		cleanup();

		render(
			<BulkBundleList
				current={{ label: 'Current bundle', checked: true, onCheckedChange: vi.fn() }}
				rows={[
					{ key: '0', label: 'Fury', checked: true },
					{ key: '1', label: 'Fury copy', checked: true, disabledReason: 'Same as Fury' },
				]}
				onCheckedChange={vi.fn()}
				onEdit={vi.fn()}
				onDelete={vi.fn()}
				testIds={{ list: 'bulk-bundles', row: 'bulk-bundle', current: 'bulk-bundle-current' }}
			/>,
		);
		const [fury, copy] = screen.getAllByTestId('bulk-bundle');
		const controls = (bundle: HTMLElement) => [within(bundle).getByRole('checkbox'), ...within(bundle).getAllByRole('button')];
		expect(controls(fury).map(control => control.hasAttribute('aria-describedby'))).toEqual([false, false, false]);
		expect(controls(copy).map(descriptions)).toEqual([['Same as Fury'], ['Same as Fury'], ['Same as Fury']]);
		// Its reason tooltip has one anchor on the page, the row's own label.
		const copyLabel = within(copy).getByText('Fury copy');
		expect([...document.querySelectorAll(`[data-tooltip-id="${copyLabel.getAttribute('data-tooltip-id')}"]`)]).toEqual([copyLabel]);
	});

	it('take a disabled row’s box into the tab order, showing its reason right of its label on focus while a tick does nothing', async () => {
		const onCheckedChange = vi.fn();
		render(
			<EncounterListView
				current={{ label: 'Current encounter', summary: '3:00 ±5s · 1 target', checked: true, onCheckedChange: vi.fn() }}
				presets={[row('Patchwerk', '5:00 ±5s · 1 target', true)]}
				saved={[row('Patchwerk copy', '5:00 ±5s · 1 target', false, { disabledReason: 'Same as Patchwerk' })]}
				onCheckedChange={onCheckedChange}
			/>,
		);
		box('Patchwerk').focus();
		await userEvent.tab();

		expect(document.activeElement).toBe(box('Patchwerk copy'));
		const tooltip = await screen.findByRole('tooltip');
		expect(tooltip.textContent).toBe('Same as Patchwerk');
		expect(tooltip.className).toContain('react-tooltip__place-right');
		// The row's own label is the tooltip's one anchor, so it opens at that row, beside the label text.
		expect([...document.querySelectorAll(`[data-tooltip-id="${tooltip.id}"]`)]).toEqual([screen.getByText('Patchwerk copy')]);
		await userEvent.keyboard(' ');
		expect(onCheckedChange).not.toHaveBeenCalled();
	});

	it('show a disabled bundle’s reason below its row on focus, clear of the row’s pencil and ×', async () => {
		render(
			<BulkBundleList
				current={{ label: 'Current bundle', checked: true, onCheckedChange: vi.fn() }}
				rows={[{ key: '0', label: 'Fury copy', checked: true, disabledReason: 'Same as Current bundle' }]}
				onCheckedChange={vi.fn()}
				onEdit={vi.fn()}
				onDelete={vi.fn()}
				testIds={{ list: 'bulk-bundles', row: 'bulk-bundle', current: 'bulk-bundle-current' }}
			/>,
		);
		box('Current bundle').focus();
		await userEvent.tab();

		expect(document.activeElement).toBe(box('Fury copy'));
		const tooltip = await screen.findByRole('tooltip');
		expect(tooltip.textContent).toBe('Same as Current bundle');
		expect(tooltip.className).toContain('react-tooltip__place-bottom');
	});

	it('show a reason quoting a saved name as text', async () => {
		const reason = 'Same as <img src=x onerror=alert(1)>';
		render(
			<>
				<EncounterListView
					current={{ label: 'Current encounter', summary: '3:00 ±5s · 1 target', checked: true, onCheckedChange: vi.fn() }}
					presets={[]}
					saved={[row('Encounter copy', '5:00 ±5s · 1 target', false, { disabledReason: reason })]}
					onCheckedChange={vi.fn()}
				/>
				<BulkBundleList
					current={{ label: 'Current bundle', checked: true, onCheckedChange: vi.fn() }}
					rows={[{ key: '0', label: 'Bundle copy', checked: true, disabledReason: reason }]}
					onCheckedChange={vi.fn()}
					onEdit={vi.fn()}
					onDelete={vi.fn()}
					testIds={{ list: 'bulk-bundles', row: 'bulk-bundle', current: 'bulk-bundle-current' }}
				/>
			</>,
		);

		for (const name of ['Encounter copy', 'Bundle copy']) {
			act(() => box(name).focus());
			const tooltip = await screen.findByRole('tooltip');
			expect(tooltip.textContent).toBe(reason);
			expect(tooltip.querySelector('img')).toBeNull();
			act(() => box(name).blur());
			await waitFor(() => expect(screen.queryByRole('tooltip')).toBeNull());
		}
	});
});
