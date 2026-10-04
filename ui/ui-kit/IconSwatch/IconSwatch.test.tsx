import { act, fireEvent, render, screen } from '@testing-library/react';
import type { ResolvedIcon } from '@ui-kit/Field/types';
import { describe, expect, it, vi } from 'vitest';

import { IconMenu } from './IconMenu';
import { IconSwatch } from './IconSwatch';

const kings: ResolvedIcon = { iconUrl: 'kings.jpg', name: 'Kings', href: 'https://www.wowhead.com/spell=20217' };

// happy-dom re-quotes the CSSOM value, so icons are compared by name rather than by literal.
const iconOf = (element: HTMLElement) => element.style.backgroundImage.replace(/^url\(['"]?|['"]?\)$/g, '');

describe('IconSwatch', () => {
	it('renders one bare anchor, painted with the icon and linked where it is told', () => {
		const { container } = render(<IconSwatch icon={kings} href={kings.href} testId="swatch" />);
		const swatch = screen.getByTestId('swatch');

		expect(container.children).toHaveLength(1);
		expect(container.firstElementChild).toBe(swatch);
		expect(swatch.tagName).toBe('A');
		expect(swatch.childElementCount).toBe(0);
		expect(iconOf(swatch)).toBe('kings.jpg');
		expect(swatch.getAttribute('href')).toBe('https://www.wowhead.com/spell=20217');
		expect(swatch.getAttribute('rel')).toBe('noopener noreferrer');
	});

	it('carries the icon opt-out on a linked swatch, and the touch one only on a swatch that acts', () => {
		const { rerender } = render(<IconSwatch icon={kings} href={kings.href} testId="swatch" />);
		expect(screen.getByTestId('swatch').dataset.whtticon).toBe('false');
		expect(screen.getByTestId('swatch').hasAttribute('data-disable-wowhead-touch-tooltip')).toBe(false);

		rerender(<IconSwatch icon={kings} href={kings.href} testId="swatch" onActivate={() => {}} />);
		expect(screen.getByTestId('swatch').dataset.whtticon).toBe('false');
		expect(screen.getByTestId('swatch').dataset.disableWowheadTouchTooltip).toBe('true');

		rerender(<IconSwatch icon={kings} testId="swatch" onActivate={() => {}} />);
		expect(screen.getByTestId('swatch').hasAttribute('href')).toBe(false);
		expect(screen.getByTestId('swatch').hasAttribute('data-whtticon')).toBe(false);
		expect(screen.getByTestId('swatch').dataset.disableWowheadTouchTooltip).toBe('true');
	});

	it('marks the active and disabled states as attributes', () => {
		const { rerender } = render(<IconSwatch icon={kings} testId="swatch" />);
		expect(screen.getByTestId('swatch').hasAttribute('data-active')).toBe(false);
		expect(screen.getByTestId('swatch').hasAttribute('data-disabled')).toBe(false);

		rerender(<IconSwatch icon={kings} active disabled testId="swatch" />);
		expect(screen.getByTestId('swatch').hasAttribute('data-active')).toBe(true);
		expect(screen.getByTestId('swatch').hasAttribute('data-disabled')).toBe(true);
		expect(screen.getByTestId('swatch').hasAttribute('disabled')).toBe(false);
		expect(screen.getByTestId('swatch').getAttribute('aria-disabled')).toBe('true');
	});

	it('lets a style replace the icon background', () => {
		render(<IconSwatch icon={kings} style={{ backgroundColor: '#123456' }} testId="swatch" />);
		expect(iconOf(screen.getByTestId('swatch'))).toBe('');
		expect(screen.getByTestId('swatch').style.backgroundColor).toBe('#123456');
	});

	it('writes its text inside the anchor', () => {
		render(<IconSwatch icon={kings} text="3" testId="swatch" />);
		expect(screen.getByTestId('swatch').textContent).toBe('3');
		expect(screen.getByTestId('swatch').childElementCount).toBe(1);
	});

	it('lays badges beside the anchor in one box', () => {
		const { container } = render(
			<IconSwatch
				icon={kings}
				text="2"
				testId="swatch"
				badges={
					<>
						<IconSwatch icon={kings} href={kings.href} testId="badge-1" />
						<IconSwatch icon={kings} href={kings.href} testId="badge-2" />
					</>
				}
			/>,
		);
		const box = container.firstElementChild as HTMLElement;
		const swatch = screen.getByTestId('swatch');

		expect(box.tagName).toBe('DIV');
		expect(Array.from(box.children)).toEqual([swatch, screen.getByTestId('badge-1'), screen.getByTestId('badge-2')]);
		expect(box.querySelectorAll('div')).toHaveLength(0);
		expect(document.querySelectorAll('a a')).toHaveLength(0);
	});

	it('hands its anchor to a ref and passes events through', () => {
		const ref = { current: null as HTMLAnchorElement | null };
		const onClick = vi.fn();
		render(<IconSwatch icon={kings} testId="swatch" ref={ref} onClick={onClick} />);
		expect(ref.current).toBe(screen.getByTestId('swatch'));
		fireEvent.click(screen.getByTestId('swatch'));
		expect(onClick).toHaveBeenCalledOnce();
	});
});

describe('IconMenu', () => {
	const mount = (onOpenChange = vi.fn(), href?: string) => {
		render(
			<IconMenu
				side="bottom"
				onOpenChange={onOpenChange}
				trigger={{ icon: kings, href, testId: 'trigger' }}
				positionerProps={{ 'data-testid': 'positioner' }}
				popupProps={{ 'data-testid': 'popup' }}>
				<li data-testid="entry" />
			</IconMenu>,
		);
		return { onOpenChange, trigger: screen.getByTestId('trigger') };
	};

	it('makes the swatch itself the trigger, a focusable button with no wrapper', () => {
		const { trigger } = mount();
		expect(trigger.tagName).toBe('A');
		expect(trigger.getAttribute('role')).toBe('button');
		expect(trigger.getAttribute('tabindex')).toBe('0');
		expect(trigger.getAttribute('aria-haspopup')).toBe('menu');
		expect(trigger.parentElement).toBe(document.body.firstElementChild);
	});

	it('opens on hover, positioned against the viewport, and mounts its entries only while open', () => {
		const { trigger, onOpenChange } = mount();
		expect(screen.queryByTestId('entry')).toBeNull();

		act(() => {
			fireEvent.mouseEnter(trigger);
			fireEvent.mouseMove(trigger);
		});
		expect(onOpenChange.mock.lastCall?.[0]).toBe(true);
		expect(screen.getByTestId('entry')).toBeTruthy();
		expect(screen.getByTestId('positioner').style.position).toBe('fixed');
		expect(trigger.getAttribute('aria-expanded')).toBe('true');
	});

	it('opens on a click of a linked swatch without following the link, and closes on the next', async () => {
		const { trigger, onOpenChange } = mount(vi.fn(), kings.href);
		expect(trigger.getAttribute('href')).toBe(kings.href);

		const first = new MouseEvent('click', { bubbles: true, cancelable: true });
		act(() => {
			trigger.dispatchEvent(first);
		});
		expect(first.defaultPrevented).toBe(true);
		expect(onOpenChange.mock.calls.map(([open]) => open)).toEqual([true]);
		expect(screen.getByTestId('entry')).toBeTruthy();

		const second = new MouseEvent('click', { bubbles: true, cancelable: true });
		act(() => {
			trigger.dispatchEvent(second);
		});
		expect(second.defaultPrevented).toBe(true);
		expect(onOpenChange.mock.calls.map(([open]) => open)).toEqual([true, false]);

		await act(async () => {});
		expect(screen.queryByTestId('entry')).toBeNull();
	});

	it('opens once on Enter and once on Space from an unlinked swatch, never twice', () => {
		const { trigger, onOpenChange } = mount();
		const press = (key: string) => {
			act(() => {
				fireEvent.keyDown(trigger, { key });
			});
			act(() => {
				fireEvent.keyUp(trigger, { key });
			});
		};

		press('Enter');
		expect(onOpenChange.mock.calls.map(([open]) => open)).toEqual([true]);

		press(' ');
		expect(onOpenChange.mock.calls.map(([open]) => open)).toEqual([true, false]);
	});
});
