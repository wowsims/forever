import type { ActionIdState } from '@sim/hooks/useActionId';
import { ActionId } from '@sim/proto/action_id';
import { act, render, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { MetricsActionCell } from './MetricsActionCell';

const resolved = vi.hoisted(() => ({ state: { iconUrl: '', name: '', href: '', ready: true } as ActionIdState }));
const wowhead = vi.hoisted(() => ({ calls: [] as Array<unknown> }));

vi.mock('@sim/hooks/useActionId', () => ({ useActionId: () => resolved.state }));
vi.mock('@sim/proto/action_id/tooltip_data', () => ({
	actionIdWowheadTooltipData: (...args: Array<unknown>) => {
		wowhead.calls.push(args);
		return Promise.resolve('spell=34026');
	},
}));

const actionId = ActionId.empty('Kill Command');

// The wowhead tooltip resolves a microtask after mount, so the cell is not settled until then.
const cell = async (props: Partial<Parameters<typeof MetricsActionCell>[0]> = {}) => {
	const view = render(<MetricsActionCell name="Kill Command" actionId={actionId} expandable={false} expanded={false} onToggle={() => {}} {...props} />);
	await act(async () => {});
	return view;
};

describe('MetricsActionCell', () => {
	it('renders the icon anchor and the name, with no toggle on a row that cannot expand', async () => {
		const { container } = await cell();

		expect(container.querySelector('[data-testid="metrics-action"] a[data-testid="metrics-action-icon"]')).not.toBeNull();
		expect(container.querySelector('[data-testid="metrics-action-name"]')?.textContent).toBe('Kill Command');
		expect(container.querySelector('[data-testid="expand-toggle"]')).toBeNull();
	});

	it('carries the href and background of a resolved action id, and neither of an empty one', async () => {
		resolved.state = { iconUrl: '', name: '', href: '', ready: true };
		const empty = (await cell()).container.querySelector<HTMLAnchorElement>('a[data-testid="metrics-action-icon"]')!;
		expect(empty.getAttribute('href')).toBeNull();
		expect(empty.style.backgroundImage).toBe('');

		resolved.state = { iconUrl: 'https://icons/kc.jpg', name: 'Kill Command', href: 'https://wowhead/spell=34026', ready: true };
		const filled = (await cell()).container.querySelector<HTMLAnchorElement>('a[data-testid="metrics-action-icon"]')!;
		expect(filled.getAttribute('href')).toBe('https://wowhead/spell=34026');
		expect(filled.style.backgroundImage).toBe('url("https://icons/kc.jpg")');
		expect(filled.getAttribute('rel')).toBe('noopener noreferrer');
	});

	it('names the icon anchor, which is a link with no text of its own', async () => {
		resolved.state = { iconUrl: 'https://icons/kc.jpg', name: 'Kill Command', href: 'https://wowhead/spell=34026', ready: true };
		const icon = (await cell()).container.querySelector<HTMLAnchorElement>('a[data-testid="metrics-action-icon"]')!;

		expect(icon.textContent).toBe('');
		expect(icon.getAttribute('aria-label')).toBe('Kill Command');
		expect(screen.getByRole('link', { name: 'Kill Command' })).toBe(icon);
	});

	it('writes the wowhead tooltip dataset for its action id', async () => {
		wowhead.calls.length = 0;
		const { container } = await cell({ useBuffAura: true });

		expect(wowhead.calls).toHaveLength(1);
		expect(wowhead.calls[0]).toMatchObject([actionId, { useBuffAura: true }]);
		await waitFor(() => expect(container.querySelector<HTMLElement>('a[data-testid="metrics-action-icon"]')!.dataset.wowhead).toBe('spell=34026'));
	});

	it('exposes the toggle as a button reporting its state', async () => {
		await cell({ expandable: true, expanded: true });

		expect(screen.getByRole('button', { name: 'Kill Command' }).getAttribute('aria-expanded')).toBe('true');
	});
});
