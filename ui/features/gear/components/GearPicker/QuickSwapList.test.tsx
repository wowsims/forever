import type { UIEnchant } from '@generated/proto/ui';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@sim/hooks/useActionId', () => ({ useActionId: () => ({ iconUrl: '', name: '', href: 'https://www.wowhead.com/spell=7', ready: true }) }));

const { QuickSwapList } = await import('./QuickSwapList');

const ENCHANT = { effectId: 7, spellId: 7, type: 0, name: 'Test Enchant', quality: 3 } as UIEnchant;

describe('QuickSwapList', () => {
	it('swaps to the entry on a click and never follows its wowhead link', () => {
		const onItemClick = vi.fn();
		render(
			<QuickSwapList
				title="Enchants"
				emptyMessage="None"
				entries={[{ item: ENCHANT, active: false }]}
				onItemClick={onItemClick}
				footerButton={{ label: 'More', onClick: () => {} }}
			/>,
		);
		const anchor = screen.getByTestId('tooltip-quick-swap__anchor');
		const click = new MouseEvent('click', { bubbles: true, cancelable: true });
		anchor.dispatchEvent(click);

		expect(anchor.getAttribute('href')).toBe('https://www.wowhead.com/spell=7');
		expect(click.defaultPrevented).toBe(true);
		expect(onItemClick).toHaveBeenCalledWith(ENCHANT);
	});
});
