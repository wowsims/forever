import i18n from '@i18n/config';
import { act, fireEvent, render, screen } from '@testing-library/react';
import type { ResolvedIcon } from '@ui-kit/Field/types';
import { IconEnumPickerDirection } from '@ui-kit/IconEnumPicker';
import { describe, expect, it, vi } from 'vitest';

import { MultiIconPicker, type MultiIconPickerProps } from './MultiIconPicker';

const kingsIcon: ResolvedIcon = { iconUrl: 'kings.jpg', name: 'Kings', href: 'https://www.wowhead.com/spell=20217' };

const mount = (props: Partial<MultiIconPickerProps> = {}) =>
	render(
		<MultiIconPicker label="Stats" active={false} onClear={() => {}} {...props}>
			{props.children ?? [<span key="kings" data-testid="child" />, <span key="mark" data-testid="child" />, <span key="emperor" data-testid="child" />]}
		</MultiIconPicker>,
	);

const trigger = () => screen.getByTestId('multi-icon-picker-root') as HTMLAnchorElement;
const menu = () => screen.queryByTestId('multi-icon-picker-menu') as HTMLUListElement;
const clearOption = () => screen.getByTestId('icon-picker-clear') as HTMLAnchorElement;
// The menu mounts when it opens, so every test that reads an option has to open it first.
const open = () =>
	act(() => {
		fireEvent.mouseEnter(trigger());
		fireEvent.mouseMove(trigger());
	});
// happy-dom re-quotes the CSSOM value, so the icon is compared by name rather than by literal.
const backgroundIcon = () => trigger().style.backgroundImage.replace(/^url\(['"]?|['"]?\)$/g, '');

describe('MultiIconPicker', () => {
	it('lays the trigger and its caption out in an icon field, with no box around the trigger', () => {
		const { container } = mount();
		const field = container.firstElementChild as HTMLElement;

		expect(field.getAttribute('data-layout')).toBe('icon');
		expect(field.firstElementChild).toBe(trigger());
		expect(trigger().tagName).toBe('A');
		expect(field.querySelectorAll('div')).toHaveLength(0);
		expect(menu()).toBeNull();
	});

	it('renders the bare trigger without a label', () => {
		const { container } = mount({ label: undefined });
		expect(Array.from(container.children)).toEqual([trigger()]);
	});

	it('builds the option list: the clear option first, then one option per child', () => {
		mount();
		open();
		const options = Array.from(menu().children);
		expect(options).toHaveLength(4);
		expect(options[0].querySelector('a')).toBe(clearOption());
		expect(options.slice(1).every(option => option.querySelector('[data-testid="child"]'))).toBe(true);
	});

	it('names the group and the trigger with its label, and renders neither name without one', () => {
		const withLabel = mount();
		// A <span>, not a <label>: it names the icon group, which is not a form control.
		const group = screen.getByRole('group', { name: 'Stats' });
		expect(group.querySelector('label')).toBeNull();
		expect(trigger().getAttribute('aria-label')).toBe('Stats');
		withLabel.unmount();

		mount({ label: undefined });
		expect(screen.queryByRole('group')).toBeNull();
		expect(trigger().hasAttribute('aria-label')).toBe(false);
	});

	it('names the trigger by the label and opens the menu from a click on the label', () => {
		mount();
		const label = screen.getByTestId('form-label');
		expect(trigger().getAttribute('aria-labelledby')).toBe(label.id);

		fireEvent.click(label);
		expect(trigger().getAttribute('aria-expanded')).toBe('true');
	});

	it('paints the icon it is given and marks the trigger active', () => {
		const { rerender } = mount();
		expect(trigger().hasAttribute('data-active')).toBe(false);
		expect(backgroundIcon()).toBe('');

		rerender(
			<MultiIconPicker label="Stats" icon={kingsIcon} active onClear={() => {}}>
				<span />
			</MultiIconPicker>,
		);
		expect(trigger().hasAttribute('data-active')).toBe(true);
		expect(backgroundIcon()).toBe('kings.jpg');
	});

	it('never gives the button an href', () => {
		mount({ icon: kingsIcon, active: true });
		expect(trigger().hasAttribute('href')).toBe(false);
	});

	it('takes a testId for its trigger', () => {
		mount({ testId: 'misc-buffs' });
		expect(screen.getByTestId('misc-buffs').tagName).toBe('A');
	});

	it('clears from a right-click on the button and from the clear option, and not from a left-click', () => {
		const onClear = vi.fn();
		mount({ onClear });

		fireEvent.mouseDown(trigger(), { button: 0 });
		expect(onClear).not.toHaveBeenCalled();

		fireEvent.mouseDown(trigger(), { button: 2 });
		expect(onClear).toHaveBeenCalledTimes(1);

		open();
		fireEvent.click(clearOption());
		expect(onClear).toHaveBeenCalledTimes(2);
	});

	it('names the clear option by its locale string', () => {
		mount();
		open();
		expect(screen.getByLabelText(i18n.t('common.clear'))).toBe(clearOption());
	});

	it('prevents the context menu on the button', () => {
		mount();
		const event = new MouseEvent('contextmenu', { bubbles: true, cancelable: true });
		trigger().dispatchEvent(event);
		expect(event.defaultPrevented).toBe(true);
	});

	it('marks a kept trigger and anchors the kept tooltip on its corner badge, not on the trigger', () => {
		mount({ active: true, keptTooltip: 'Kept' });

		expect(trigger().parentElement!.hasAttribute('data-kept')).toBe(true);
		expect(trigger().hasAttribute('data-tooltip-id')).toBe(false);
		expect(screen.getByTestId('icon-swatch-kept').getAttribute('data-tooltip-content')).toBe('Kept');
	});

	it('opens its menu to the right as a row by default', () => {
		mount();
		open();
		expect(menu().className).toContain('grid-flow-col');
		expect(screen.getByTestId('multi-icon-picker-positioner').getAttribute('data-side')).toBe('right');
	});

	it('opens a vertical menu below, in the columns it is given', () => {
		mount({ direction: IconEnumPickerDirection.Vertical, numColumns: 2 });
		open();
		expect(menu().className).not.toContain('grid-flow-col');
		expect(menu().style.gridTemplateColumns).toBe('repeat(2, 1fr)');
		expect(screen.getByTestId('multi-icon-picker-positioner').getAttribute('data-side')).toBe('bottom');
	});

	it('carries its tooltip on the trigger, which it names when there is no label', () => {
		mount({ label: undefined, tooltip: 'Defensive' });
		expect(trigger().getAttribute('data-tooltip-content')).toBe('Defensive');
		expect(trigger().getAttribute('aria-label')).toBe('Defensive');
	});
});
