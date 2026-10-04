import type { UIGem } from '@generated/proto/ui';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@sim/hooks/useActionId', () => ({ useActionId: () => ({ iconUrl: '', name: '', href: 'https://www.wowhead.com/item=7', ready: true }) }));

const { QuickSwapList } = await import('./QuickSwapList');

const GEM = { id: 7, name: 'Test Gem', quality: 3 } as UIGem;

describe('QuickSwapList', () => {
	it('swaps to the entry on a click and never follows its wowhead link', () => {
		const onItemClick = vi.fn();
		render(
			<QuickSwapList
				title="Gems"
				emptyMessage="None"
				entries={[{ item: GEM, active: false }]}
				onItemClick={onItemClick}
				footerButton={{ label: 'More', onClick: () => {} }}
			/>,
		);
		const anchor = screen.getByTestId('tooltip-quick-swap__anchor');
		const click = new MouseEvent('click', { bubbles: true, cancelable: true });
		anchor.dispatchEvent(click);

		expect(anchor.getAttribute('href')).toBe('https://www.wowhead.com/item=7');
		expect(click.defaultPrevented).toBe(true);
		expect(onItemClick).toHaveBeenCalledWith(GEM);
	});
});
