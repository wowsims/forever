import { fireEvent, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { EMPTY_SUGGESTIONS } from '../../model/log/search/indexes';
import { LogSearchGroup } from './LogSearchGroup';
import type { IdentifiedSearchGroup } from './utils';

const group = (extra: Partial<IdentifiedSearchGroup> = {}): IdentifiedSearchGroup => ({ id: 0, field: 'outcome', join: 'or', values: [], ...extra });

const suggestions = { ...EMPTY_SUGGESTIONS, spells: ['Cleave'], spellIcons: new Map([['Cleave', 'https://icons/cleave.jpg']]) };

const mount = (current: IdentifiedSearchGroup, onChange = vi.fn(), onRemove = vi.fn()) => ({
	onChange,
	onRemove,
	...render(<LogSearchGroup group={current} suggestions={suggestions} onChange={onChange} onRemove={onRemove} />),
});

const valueInput = (container: HTMLElement) => container.querySelector<HTMLInputElement>('[data-testid="log-search-group-input"]')!;
const submit = (container: HTMLElement) => container.querySelector<HTMLButtonElement>('[data-testid="input-group"] button')!;
const chips = (container: HTMLElement) =>
	[...container.querySelectorAll('[data-testid="log-search-chip"] [data-testid="saved-data-set-name"]')].map(chip => chip.textContent);

describe('LogSearchGroup', () => {
	it('names the field and marks the join in force', () => {
		const { container } = mount(group({ join: 'and' }));

		expect(container.querySelector('[data-testid="log-search-group-field"]')!.textContent).toBe('results_tab.details.logs.search_fields.outcome');
		const joins = [...container.querySelectorAll<HTMLButtonElement>('[data-testid="log-search-group-join"] button')];
		expect(joins.map(button => button.textContent)).toEqual(['results_tab.details.logs.search_join.and', 'results_tab.details.logs.search_join.or']);
		expect(joins.map(button => button.getAttribute('aria-pressed'))).toEqual(['true', 'false']);
	});

	it('switches the join, and stays quiet when the pressed one is already in force', () => {
		const { container, onChange } = mount(group({ join: 'or' }));
		const [and, or] = [...container.querySelectorAll<HTMLButtonElement>('[data-testid="log-search-group-join"] button')];

		fireEvent.click(or);
		expect(onChange).not.toHaveBeenCalled();
		fireEvent.click(and);
		expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ join: 'and' }));
	});

	it('shows one chip per picked value, with the label the field spells it as', () => {
		const { container } = mount(group({ values: ['critical-block', 'crit'] }));

		expect(chips(container)).toEqual(['results_tab.details.attack_types.critical_block', 'results_tab.details.attack_types.crit']);
	});

	it('drops the chip that was deleted and keeps the rest', () => {
		const { container, onChange } = mount(group({ values: ['crit', 'hit', 'miss'] }));

		fireEvent.click(container.querySelectorAll<HTMLButtonElement>('[data-testid="log-search-chip"] [data-testid="saved-data-set-delete"]')[1]);

		expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ values: ['crit', 'miss'] }));
	});

	it('removes the whole group from its head', () => {
		const { container, onRemove } = mount(group());

		fireEvent.click(container.querySelector<HTMLButtonElement>('[data-testid="log-search-group-remove"]')!);

		expect(onRemove).toHaveBeenCalledTimes(1);
	});

	// A picked field offers a menu; a numeric one offers a box, and never both.
	it('offers a picker for a picked field and a typed box for a numeric one', () => {
		expect(within(mount(group({ field: 'spell' })).container).queryByRole('combobox')).not.toBeNull();
		expect(mount(group({ field: 'spell' })).container.querySelector('[data-testid="log-search-group-input"]')).toBeNull();
		expect(
			mount(group({ field: 'time' }))
				.container.querySelector('[data-testid="log-search-group-input"]')!
				.getAttribute('placeholder'),
		).toBe('10-30');
		expect(within(mount(group({ field: 'amount' })).container).queryByRole('combobox')).toBeNull();
	});

	it('draws the spell icon the index resolved next to its option', async () => {
		const { container } = mount(group({ field: 'spell' }));
		await userEvent.click(within(container).getByRole('combobox'));

		expect(within(screen.getByRole('option')).getByTestId<HTMLImageElement>('log-search-suggestion-icon').src).toBe('https://icons/cleave.jpg');
	});

	describe('the typed value box', () => {
		it('commits a range on Enter and clears the box', () => {
			const { container, onChange } = mount(group({ field: 'time' }));

			fireEvent.change(valueInput(container), { target: { value: '10-30' } });
			fireEvent.keyDown(valueInput(container), { key: 'Enter' });

			expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ values: ['10-30'] }));
			expect(valueInput(container).value).toBe('');
		});

		it('commits from the button too', () => {
			const { container, onChange } = mount(group({ field: 'amount' }));

			fireEvent.change(valueInput(container), { target: { value: '>5000' } });
			fireEvent.click(submit(container));

			expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ values: ['>5000'] }));
		});

		// The index cannot filter on something that is not a number, so it never reaches the group.
		it('refuses anything that is not a numeric filter', () => {
			const { container, onChange } = mount(group({ field: 'time' }));

			for (const rejected of ['', '   ', 'soon', '10-']) {
				fireEvent.change(valueInput(container), { target: { value: rejected } });
				fireEvent.click(submit(container));
			}

			expect(onChange).not.toHaveBeenCalled();
		});

		it('refuses a duplicate of a value already picked', () => {
			const { container, onChange } = mount(group({ field: 'time', values: ['10-30'] }));

			fireEvent.change(valueInput(container), { target: { value: '10-30' } });
			fireEvent.click(submit(container));

			expect(onChange).not.toHaveBeenCalled();
		});
	});
});
