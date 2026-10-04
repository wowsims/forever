import { ItemType } from '@generated/proto/common';
import { type UIItem, UIItem_FactionRestriction } from '@generated/proto/ui';
import { render } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { HREF } = vi.hoisted(() => ({ HREF: 'https://www.wowhead.com/mop-classic/item=1' }));
let href: string = HREF;

vi.mock('@sim/hooks/useActionId', () => ({
	useActionId: () => ({ iconUrl: 'icon.jpg', name: 'Item', href, ready: true }),
}));

import { BulkItemSearchRow } from './BulkItemSearchRow';

const item = () =>
	({
		id: 1,
		name: 'Item',
		quality: 4,
		ilvl: 500,
		type: ItemType.ItemTypeHead,
		nameDescription: '',
		factionRestriction: UIItem_FactionRestriction.UNSPECIFIED,
	}) as unknown as UIItem;

const anchor = (container: HTMLElement) => container.querySelector<HTMLAnchorElement>('a[data-item-id]')!;

describe('BulkItemSearchRow', () => {
	beforeEach(() => {
		href = HREF;
	});

	it('keeps the row out of the tab order, since it is the whole content of a role="option"', () => {
		const { container } = render(<BulkItemSearchRow item={item()} />);

		expect(anchor(container).getAttribute('tabindex')).toBe('-1');
	});

	it("still carries the wowhead href, the row's only tooltip attachment", () => {
		const { container } = render(<BulkItemSearchRow item={item()} />);

		expect(anchor(container).getAttribute('href')).toBe(HREF);
		expect(anchor(container).hasAttribute('target')).toBe(false);
	});

	it('cancels the link on a click and lets it reach the option around the row, which selects', () => {
		const onSelect = vi.fn();
		const { container } = render(
			<div role="option" aria-selected={false} onClick={onSelect}>
				<BulkItemSearchRow item={item()} />
			</div>,
		);
		const click = new MouseEvent('click', { bubbles: true, cancelable: true });
		anchor(container).dispatchEvent(click);

		expect(click.defaultPrevented).toBe(true);
		expect(onSelect).toHaveBeenCalledOnce();
	});

	it('carries no role and no touch opt-out on an item with no wowhead page, inside a role="option"', () => {
		href = '';
		const { container } = render(<BulkItemSearchRow item={item()} />);

		expect(anchor(container).hasAttribute('href')).toBe(false);
		expect(anchor(container).hasAttribute('role')).toBe(false);
		expect(anchor(container).hasAttribute('data-disable-wowhead-touch-tooltip')).toBe(false);
	});
});
