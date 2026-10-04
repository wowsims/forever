import { act, fireEvent, render, screen, within } from '@testing-library/react';
import type { ResolvedIcon } from '@ui-kit/Field/types';
import { type Ref, useImperativeHandle, useState } from 'react';
import { describe, expect, it, vi } from 'vitest';

import { IconEnumPicker, type IconEnumPickerProps } from './IconEnumPicker';
import { type IconEnumOption, IconEnumPickerDirection } from './types';

const frostIcon: ResolvedIcon = { iconUrl: 'frost.jpg', name: 'Frost Armor', href: 'https://www.wowhead.com/spell=7302' };
const moltenIcon: ResolvedIcon = { iconUrl: 'molten.jpg', name: 'Molten Armor', href: 'https://www.wowhead.com/spell=30482' };
// The hunter's "No Pet" entry: an icon with no wowhead page.
const noPetIcon: ResolvedIcon = { iconUrl: 'questionmark.jpg', name: 'No Pet' };

const OPTIONS: Array<IconEnumOption<number>> = [
	{ value: 0, tooltip: 'No Armor' },
	{ value: 1, icon: frostIcon },
	{ value: 2, icon: moltenIcon, tooltip: 'Molten' },
];

interface HarnessHandle {
	set: (next: number) => void;
}

type HarnessProps = Partial<Omit<IconEnumPickerProps<number>, 'value' | 'onChange'>> & {
	initial?: number;
	onChange?: (next: number) => void;
	ref?: Ref<HarnessHandle>;
};

/** Holds the value the way a caller does, and lets a test move it underneath the picker. */
const Harness = ({ initial = 0, onChange, ref, ...props }: HarnessProps) => {
	const [value, setValue] = useState(initial);
	useImperativeHandle(ref, () => ({ set: setValue }), []);
	return (
		<IconEnumPicker
			options={OPTIONS}
			equals={(a, b) => a === b}
			zeroValue={0}
			{...props}
			value={value}
			onChange={next => {
				setValue(next);
				onChange?.(next);
			}}
		/>
	);
};

const mount = (props: HarnessProps = {}) => {
	const handle: { current: HarnessHandle | null } = { current: null };
	const onChange = vi.fn();
	const view = render(<Harness onChange={onChange} ref={handle} {...props} />);
	return { ...view, onChange, set: (next: number) => act(() => handle.current!.set(next)) };
};

const button = () => screen.getByTestId('icon-enum-picker-root') as HTMLAnchorElement;
const menu = () => (screen.queryByTestId('icon-enum-picker-menu') ?? null) as HTMLUListElement | null;
const items = () => Array.from(menu()!.children) as HTMLLIElement[];
const optionAnchor = (index: number) => within(items()[index]).getByTestId('icon-enum-option') as HTMLAnchorElement;
// The menu mounts when it opens, so every test that reads an option has to open it first.
const open = (trigger: HTMLElement = button()) =>
	act(() => {
		fireEvent.mouseEnter(trigger);
		fireEvent.mouseMove(trigger);
	});
const iconText = () => button().textContent || null;
// happy-dom re-quotes the CSSOM value, so icons are compared by name rather than by literal.
const iconOf = (element: HTMLElement) => element.style.backgroundImage.replace(/^url\(['"]?|['"]?\)$/g, '');

describe('IconEnumPicker', () => {
	it('renders the bare trigger swatch, and mounts the options against the viewport when the menu opens', async () => {
		const { container } = mount();

		expect(Array.from(container.children)).toEqual([button()]);
		expect(button().tagName).toBe('A');

		expect(screen.queryByTestId('icon-enum-picker-portal')).toBeNull();
		expect(menu()).toBeNull();

		await open();
		expect(screen.getByTestId('icon-enum-picker-positioner').style.position).toBe('fixed');

		expect(menu()!.hasAttribute('hidden')).toBe(false);
		expect(items()).toHaveLength(3);
		expect(items().map(item => item.getAttribute('role'))).toEqual(['none', 'none', 'none']);
		expect(items().map(item => optionAnchor(items().indexOf(item)).getAttribute('role'))).toEqual(['menuitem', 'menuitem', 'menuitem']);
	});

	it('lays a labelled trigger out in an icon field and takes a testId for its trigger', () => {
		const { container } = mount({ id: 'armor', label: 'Armor', testId: 'armor-root' });
		const field = container.firstElementChild as HTMLElement;
		const trigger = screen.getByTestId('armor-root');
		expect(field.getAttribute('data-layout')).toBe('icon');
		expect(field.firstElementChild).toBe(trigger);
		expect(trigger.tagName).toBe('A');
		expect(within(field).getByTestId('form-label').textContent).toBe('Armor');
	});

	it('names the trigger by the label, and opens the menu from a click on the label', () => {
		mount({ id: 'armor', label: 'Armor' });
		const label = screen.getByTestId('form-label');
		expect(button().getAttribute('aria-labelledby')).toBe(label.id);
		expect(screen.getByRole('button', { name: 'Armor' })).toBe(button());

		fireEvent.click(label);
		expect(button().getAttribute('aria-expanded')).toBe('true');
	});

	it('opens from the keyboard, moves through the options and picks one', async () => {
		const { onChange } = mount();
		act(() => button().focus());
		act(() => {
			fireEvent.keyDown(button(), { key: 'ArrowDown' });
		});
		await act(async () => {});
		expect(button().getAttribute('aria-expanded')).toBe('true');
		expect(items()).toHaveLength(3);

		const highlighted = () => document.activeElement as HTMLElement;
		act(() => {
			fireEvent.keyDown(highlighted(), { key: 'ArrowDown' });
		});
		act(() => {
			fireEvent.keyDown(highlighted(), { key: 'ArrowDown' });
		});
		expect(highlighted()).toBe(optionAnchor(2));
		act(() => {
			fireEvent.keyDown(highlighted(), { key: 'ArrowUp' });
		});
		act(() => {
			fireEvent.keyDown(highlighted(), { key: 'ArrowUp' });
		});
		expect(highlighted()).toBe(optionAnchor(0));

		act(() => {
			fireEvent.keyDown(highlighted(), { key: 'Enter' });
		});
		expect(onChange).toHaveBeenCalledTimes(1);
		expect(onChange).toHaveBeenCalledWith(0);
	});

	it('takes the button’s icon, href and active class from the selected value', () => {
		const { set } = mount();

		// The zero value carries neither an icon nor an iconUrl nor a colour, so there is no image and
		// no link. React refuses to render a `javascript:` URL, so the button carries no `href` at all.
		expect(iconOf(button())).toBe('');
		expect(button().hasAttribute('href')).toBe(false);
		expect(button().hasAttribute('data-active')).toBe(false);

		set(2);
		expect(iconOf(button())).toBe('molten.jpg');
		expect(button().getAttribute('href')).toBe('https://www.wowhead.com/spell=30482');
		expect(button().hasAttribute('data-active')).toBe(true);

		set(0);
		expect(iconOf(button())).toBe('');
		expect(button().hasAttribute('href')).toBe(false);
		expect(button().hasAttribute('data-active')).toBe(false);
	});

	it('leaves the button unlinked for an icon with no wowhead page', () => {
		mount({ initial: 1, options: [{ value: 0 }, { value: 1, icon: noPetIcon }] });
		expect(iconOf(button())).toBe('questionmark.jpg');
		expect(button().hasAttribute('href')).toBe(false);
	});

	it('keeps the button focusable without an href', () => {
		mount();
		// An <a> with no href is not tabbable on its own; `nativeButton={false}` is what puts it back.
		expect(button().getAttribute('tabindex')).toBe('0');
		expect(button().getAttribute('role')).toBe('button');
	});

	it('grays an iconUrl value and paints a colour value', () => {
		const { set } = mount({
			initial: 1,
			options: [
				{ value: 0, color: '#123456' },
				{ value: 1, iconUrl: 'raw.jpg' },
				{ value: 2, icon: frostIcon },
			],
		});

		expect(iconOf(button())).toBe('raw.jpg');
		expect(button().style.filter).toBe('grayscale(1)');

		set(0);
		expect(button().style.backgroundColor).toBe('#123456');
		expect(button().style.filter).toBe('');
	});

	it('writes the value’s text over the swatch and over its option, and nothing for a value without one', () => {
		const { set } = mount({ options: [{ value: 0 }, { value: 1, icon: frostIcon, text: '3' }] });

		expect(iconText()).toBeNull();

		set(1);
		expect(iconText()).toBe('3');

		open();
		expect(optionAnchor(0).textContent).toBe('');
		expect(optionAnchor(1).textContent).toBe('3');
	});

	it('hands the chosen value to onChange and closes the menu', () => {
		const { onChange } = mount();

		open();
		expect(button().getAttribute('aria-expanded')).toBe('true');

		act(() => {
			fireEvent.click(optionAnchor(2));
		});
		expect(onChange).toHaveBeenCalledWith(2);
		expect(button().getAttribute('aria-expanded')).toBe('false');
	});

	it('reports the menu opening and closing through onOpenChange', () => {
		const onOpenChange = vi.fn();
		mount({ onOpenChange });
		expect(onOpenChange).not.toHaveBeenCalled();

		open();
		expect(onOpenChange.mock.lastCall?.[0]).toBe(true);

		act(() => {
			fireEvent.click(optionAnchor(1));
		});
		expect(onOpenChange.mock.lastCall?.[0]).toBe(false);
	});

	it('cancels the anchor’s navigation when an option is chosen, selects it once and closes the menu', () => {
		const { onChange } = mount();
		open();
		expect(optionAnchor(1).getAttribute('href')).toBe(frostIcon.href);
		const event = new MouseEvent('click', { bubbles: true, cancelable: true });
		act(() => {
			optionAnchor(1).dispatchEvent(event);
		});
		expect(event.defaultPrevented).toBe(true);
		expect(onChange.mock.calls).toEqual([[1]]);
		expect(button().getAttribute('aria-expanded')).toBe('false');
	});

	it('selects an option with no link once on Enter and once on Space', () => {
		const { onChange, set } = mount();
		const press = (key: string) => {
			act(() => {
				fireEvent.keyDown(optionAnchor(0), { key });
			});
			if (menu())
				act(() => {
					fireEvent.keyUp(optionAnchor(0), { key });
				});
		};

		set(1);
		open();
		expect(optionAnchor(0).hasAttribute('href')).toBe(false);
		press('Enter');
		expect(onChange.mock.calls).toEqual([[0]]);

		set(1);
		open();
		press(' ');
		expect(onChange.mock.calls).toEqual([[0], [0]]);
	});

	it('opts every option out of the touch tooltip, since a tap selects', () => {
		mount();
		open();
		expect(items().map((_item, index) => optionAnchor(index).dataset.disableWowheadTouchTooltip)).toEqual(['true', 'true', 'true']);
	});

	it('cancels the trigger’s navigation on a click, so a tap opens the menu instead of the link', () => {
		mount({ initial: 1 });
		expect(button().getAttribute('href')).toBe(frostIcon.href);
		const event = new MouseEvent('click', { bubbles: true, cancelable: true });
		act(() => {
			button().dispatchEvent(event);
		});
		expect(event.defaultPrevented).toBe(true);
		expect(menu()).not.toBeNull();
	});

	it('gives each option its own icon and wowhead link, and none of them the active class', () => {
		mount({ initial: 1 });
		open();

		expect(items().map((_item, index) => iconOf(optionAnchor(index)))).toEqual(['', 'frost.jpg', 'molten.jpg']);
		expect(items().map((_item, index) => optionAnchor(index).getAttribute('href'))).toEqual([
			null,
			'https://www.wowhead.com/spell=7302',
			'https://www.wowhead.com/spell=30482',
		]);
		expect(items().some((_item, index) => optionAnchor(index).hasAttribute('data-active'))).toBe(false);
	});

	it('leaves a hidden option out of the menu', () => {
		const { rerender } = render(
			<IconEnumPicker
				options={[...OPTIONS.slice(0, 2), { ...OPTIONS[2], hidden: true }]}
				equals={(a, b) => a === b}
				zeroValue={0}
				value={0}
				onChange={() => {}}
			/>,
		);
		open();
		expect(items()).toHaveLength(2);

		rerender(<IconEnumPicker options={OPTIONS} equals={(a, b) => a === b} zeroValue={0} value={0} onChange={() => {}} />);
		expect(items()).toHaveLength(3);
		expect(optionAnchor(2).getAttribute('href')).toBe('https://www.wowhead.com/spell=30482');
	});

	it('shows a selected value whose option is hidden as no selection', () => {
		mount({ initial: 2, options: [{ value: 0 }, { value: 2, icon: moltenIcon, text: '2', hidden: true }] });
		expect(iconOf(button())).toBe('');
		expect(button().hasAttribute('href')).toBe(false);
		expect(iconText()).toBeNull();
	});

	it('marks the trigger disabled and aria-disabled, and its menu stays closed from a click on the trigger or the label', () => {
		const { unmount } = mount({ disabled: true });

		expect(button().hasAttribute('data-disabled')).toBe(true);
		expect(button().getAttribute('aria-disabled')).toBe('true');
		act(() => {
			button().dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true }));
		});
		expect(button().getAttribute('aria-expanded')).toBe('false');
		expect(menu()).toBeNull();
		unmount();

		const { container } = mount({ disabled: true, label: 'Armor' });
		expect((container.firstElementChild as HTMLElement).hasAttribute('data-disabled')).toBe(true);
		fireEvent.click(screen.getByTestId('form-label'));
		expect(button().getAttribute('aria-expanded')).toBe('false');
		expect(menu()).toBeNull();
	});

	it('never marks the button active while it is disabled', () => {
		mount({ initial: 2, disabled: true });
		expect(button().hasAttribute('data-active')).toBe(false);
	});

	it('lays the menu out from numColumns, and turns it sideways for a horizontal picker', () => {
		mount({ numColumns: 5 });
		open();
		expect(menu()!.style.gridTemplateColumns).toBe('repeat(5, 1fr)');
		expect(menu()!.style.gridAutoFlow).toBe('');

		render(<Harness direction={IconEnumPickerDirection.Horizontal} />);
		const horizontalButton = screen.getAllByTestId('icon-enum-picker-root')[1];
		open(horizontalButton);
		const horizontalMenu = document.getElementById(horizontalButton.getAttribute('aria-controls')!) as HTMLUListElement;
		expect(horizontalMenu.style.gridAutoFlow).toBe('column');
	});

	it('carries every tooltip on its anchor, and renders none until one is asked for', () => {
		mount({ tooltip: 'Lethal Poison' });

		const tooltipId = button().getAttribute('data-tooltip-id');
		expect(tooltipId).toBeTruthy();
		expect(button().getAttribute('data-tooltip-content')).toBe('Lethal Poison');
		// Before the menu is opened, because opening it is a hover on the button and a hover is what
		// asks for a tooltip.
		expect(screen.queryByRole('tooltip')).toBeNull();

		open();
		expect(optionAnchor(0).getAttribute('data-tooltip-content')).toBe('No Armor');
		expect(optionAnchor(0).getAttribute('data-tooltip-id')).toBe(tooltipId);
		// The value with no tooltip of its own opts out rather than anchoring an empty one.
		expect(optionAnchor(1).hasAttribute('data-tooltip-id')).toBe(false);
	});

	it('renders no tooltip anchor at all when nothing configures one', () => {
		const { container } = mount({ options: [{ value: 0 }, { value: 1, icon: frostIcon }] });
		expect(container.querySelectorAll('[data-tooltip-id]')).toHaveLength(0);
	});

	it('falls back to backupIcon for a value no option carries', () => {
		mount({ initial: 99, backupIcon: moltenIcon });

		expect(iconOf(button())).toBe('molten.jpg');
		expect(button().getAttribute('href')).toBe('https://www.wowhead.com/spell=30482');
		// A non-zero value is active whether or not an option carries it.
		expect(button().hasAttribute('data-active')).toBe(true);
	});
});
