import { fireEvent, render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { EMPTY_SUGGESTIONS } from '../../model/log/search/indexes';
import { LogToolbar } from './LogToolbar';
import type { IdentifiedSearchGroup } from './utils';

vi.mock('./LogSearchBar', () => ({ LogSearchBar: () => <div className="log-search-bar" /> }));

const group = (field: IdentifiedSearchGroup['field'], values: Array<string>, id = 0): IdentifiedSearchGroup => ({ id, field, join: 'or', values });

const mount = (groups: Array<IdentifiedSearchGroup>, onChange = vi.fn()) => ({
	onChange,
	...render(
		<LogToolbar groups={groups} suggestions={EMPTY_SUGGESTIONS} onChange={onChange}>
			<button type="button" className="log-export" />
		</LogToolbar>,
	),
});

// The test setup loads no locale, so each label reads as its own key.
const PREVIEW = 'results_tab.details.logs.search_fields.outcome: results_tab.details.attack_types.crit';

const clear = (container: HTMLElement) => container.querySelector<HTMLButtonElement>('[data-testid="log-floating-action-bar-clear"]')!;

describe('LogToolbar', () => {
	// A group with no values is not a filter yet, so it must not be counted or previewed.
	it('counts and previews only the groups that carry a value', () => {
		const { container } = mount([group('outcome', ['crit']), group('spell', [], 1)]);

		expect(container.querySelector('[data-testid="log-floating-action-bar-summary"]')!.textContent).toBe(
			'results_tab.details.logs.floatingActionBar.active',
		);
		expect(container.querySelector('[data-testid="log-floating-action-bar-preview"]')!.textContent).toBe(PREVIEW);
	});

	it('elides the preview past three filters', () => {
		const { container } = mount([0, 1, 2, 3].map(index => group('outcome', ['crit'], index)));

		expect(container.querySelector('[data-testid="log-floating-action-bar-preview"]')!.textContent).toBe(`${PREVIEW}, ${PREVIEW}, ${PREVIEW}, …`);
	});

	it('says there are no filters and hides Clear while nothing is picked', () => {
		const { container } = mount([group('spell', [])]);

		expect(container.querySelector('[data-testid="log-floating-action-bar-summary"]')!.textContent).toBe('results_tab.details.logs.floatingActionBar.none');
		expect(container.querySelector('[data-testid="log-floating-action-bar-preview"]')!.textContent).toBe('');
		expect(clear(container).hidden).toBe(true);
	});

	it('shows Clear once a filter exists, and empties every group with it', () => {
		const { container, onChange } = mount([group('outcome', ['crit'])]);
		expect(clear(container).hidden).toBe(false);

		fireEvent.click(clear(container));

		expect(onChange).toHaveBeenCalledWith([]);
	});
});
