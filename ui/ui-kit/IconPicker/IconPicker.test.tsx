import { act, fireEvent, render, screen, within } from '@testing-library/react';
import type { ResolvedIcon } from '@ui-kit/Field/types';
import { type Ref, useImperativeHandle, useState } from 'react';
import { describe, expect, it, vi } from 'vitest';

import { IconPicker, type IconPickerProps } from './IconPicker';

const buffIcon: ResolvedIcon = { iconUrl: 'buff.jpg', name: 'Buff', href: 'https://www.wowhead.com/spell=1' };
const improvedIcon: ResolvedIcon = { iconUrl: 'improved.jpg', name: 'Improved Buff', href: 'https://www.wowhead.com/spell=2' };
const improvedIcon2: ResolvedIcon = { iconUrl: 'greater.jpg', name: 'Greater Buff', href: 'https://www.wowhead.com/spell=3' };

interface HarnessHandle {
	set: (next: number) => void;
}

type HarnessProps = Omit<IconPickerProps, 'value' | 'onChange'> & {
	initial?: number;
	onChange?: (next: number) => void;
	ref?: Ref<HarnessHandle>;
};

/** Holds the value the way a caller does, and lets a test move it underneath the picker. */
const Harness = ({ initial = 0, onChange, ref, ...props }: HarnessProps) => {
	const [value, setValue] = useState(initial);
	useImperativeHandle(ref, () => ({ set: setValue }), []);
	return (
		<IconPicker
			{...props}
			value={value}
			onChange={next => {
				setValue(next);
				onChange?.(next);
			}}
		/>
	);
};

const mount = (props: Partial<HarnessProps> = {}) => {
	const handle: { current: HarnessHandle | null } = { current: null };
	const onChange = vi.fn();
	const view = render(<Harness states={2} icon={buffIcon} onChange={onChange} ref={handle} {...props} />);
	return { ...view, onChange, set: (next: number) => act(() => handle.current!.set(next)) };
};

// By role would miss them: an <a> without an href has no `link` role.
const allAnchors = () => Array.from(document.querySelectorAll('a'));
const mainAnchor = () => screen.getByTestId('icon-picker-root');

describe('IconPicker', () => {
	it('paints the icon as the anchor background, names the anchor after it and links it', () => {
		mount();
		expect(mainAnchor().style.backgroundImage).toContain('buff.jpg');
		expect(mainAnchor().getAttribute('aria-label')).toBe('Buff');
		expect(mainAnchor().getAttribute('href')).toBe('https://www.wowhead.com/spell=1');
	});

	it('renders a bare, unlinked swatch without an icon', () => {
		mount({ icon: undefined });
		expect(mainAnchor().style.backgroundImage).toBe('');
		expect(mainAnchor().hasAttribute('aria-label')).toBe(false);
		expect(mainAnchor().hasAttribute('href')).toBe(false);
	});

	it('renders a bare swatch without a label: the anchor is all there is', () => {
		const { container } = mount();
		expect(Array.from(container.children)).toEqual([mainAnchor()]);
		expect(mainAnchor().tagName).toBe('A');
		expect(screen.queryByTestId('form-label')).toBeNull();
	});

	it('lays a labelled swatch out in an icon field, the swatch first', () => {
		const { container } = mount({ id: 'buff-icon', label: 'Buff' });
		const field = container.firstElementChild as HTMLElement;
		expect(field.getAttribute('data-layout')).toBe('icon');
		expect(field.firstElementChild).toBe(mainAnchor());
		expect(within(field).getByTestId('form-label').textContent).toBe('Buff');
	});

	it('names the swatch by the label, beating the icon name', () => {
		mount({ id: 'battle-shout', label: 'Battle Shout' });
		expect(screen.getByRole('link', { name: 'Battle Shout' })).toBe(mainAnchor());
		expect(screen.queryByRole('link', { name: 'Buff' })).toBeNull();
	});

	it('clicks the swatch when its label is clicked, cycling the value exactly once', () => {
		const { onChange } = mount({ states: 3, label: 'Buff' });
		fireEvent.click(screen.getByTestId('form-label'));
		expect(onChange).toHaveBeenCalledOnce();
		expect(onChange).toHaveBeenLastCalledWith(1);
	});

	it('leaves an unlabelled swatch without an aria-labelledby', () => {
		mount();
		expect(mainAnchor().hasAttribute('aria-labelledby')).toBe(false);
	});

	it('lays out an icon field whatever layout the caller asks for', () => {
		const { container } = mount({ label: 'Buff', layout: 'inline' });
		expect((container.firstElementChild as HTMLElement).getAttribute('data-layout')).toBe('icon');
	});

	it('sits in an icon field for a layout alone, with no label', () => {
		const { container } = mount({ layout: 'inline' });
		const field = container.firstElementChild as HTMLElement;
		expect(field.getAttribute('data-layout')).toBe('icon');
		expect(field.firstElementChild).toBe(mainAnchor());
		expect(screen.queryByTestId('form-label')).toBeNull();
	});

	it('takes a testId for its swatch', () => {
		mount({ testId: 'custom-root', label: 'Buff' });
		expect(screen.getByTestId('custom-root').tagName).toBe('A');
	});

	it('grows the anchor count as each improved anchor mounts once its own states gate is met', () => {
		const { rerender } = render(<IconPicker states={2} icon={buffIcon} improvedIcon={improvedIcon} value={0} onChange={() => {}} />);
		expect(allAnchors()).toHaveLength(1);

		rerender(<IconPicker states={3} icon={buffIcon} improvedIcon={improvedIcon} value={0} onChange={() => {}} />);
		expect(allAnchors()).toHaveLength(2);

		rerender(<IconPicker states={4} icon={buffIcon} improvedIcon={improvedIcon} improvedIcon2={improvedIcon2} value={0} onChange={() => {}} />);
		expect(allAnchors()).toHaveLength(3);
	});

	it('mounts an improved anchor with its href and background filled once its states gate is met', () => {
		mount({ states: 3, improvedIcon });
		const improved = allAnchors()[1];
		expect(improved.getAttribute('href')).toBe('https://www.wowhead.com/spell=2');
		expect(improved.style.backgroundImage).toContain('improved.jpg');
	});

	it.each([
		['', 2, 1],
		['', 3, 1],
		['', 3, 2],
		['3', 4, 3],
		['1', 0, 1],
	])('shows counter %j for states %i at value %i, without an improved icon', (counter, states, initial) => {
		mount({ states, initial });
		expect(mainAnchor().textContent).toBe(counter);
	});

	it('left-clicks to the next state and rolls over to 0 at the top', () => {
		const { onChange } = mount({ states: 3 });

		fireEvent.click(mainAnchor());
		expect(onChange).toHaveBeenLastCalledWith(1);

		fireEvent.click(mainAnchor());
		expect(onChange).toHaveBeenLastCalledWith(2);

		fireEvent.click(mainAnchor());
		expect(onChange).toHaveBeenLastCalledWith(0);
	});

	it('counts up without a top at states 0', () => {
		const { onChange } = mount({ states: 0, initial: 7 });
		fireEvent.click(mainAnchor());
		expect(onChange).toHaveBeenLastCalledWith(8);
	});

	it('right-clicks via mousedown(button 2) to the previous state and rolls over to the top', () => {
		const { onChange } = mount({ states: 3, initial: 1 });

		fireEvent.mouseDown(mainAnchor(), { button: 2 });
		expect(onChange).toHaveBeenLastCalledWith(0);

		fireEvent.mouseDown(mainAnchor(), { button: 2 });
		expect(onChange).toHaveBeenLastCalledWith(2);
	});

	it('does not change the value on a plain mousedown (button 0)', () => {
		const { onChange } = mount({ states: 3, initial: 1 });
		fireEvent.mouseDown(mainAnchor(), { button: 0 });
		expect(onChange).not.toHaveBeenCalled();
	});

	it('suppresses the browser context menu', () => {
		mount();
		// fireEvent returns false when preventDefault() was called.
		expect(fireEvent.contextMenu(mainAnchor())).toBe(false);
	});

	it.each([
		[1, true],
		[0, false],
	])('marks the main anchor active only above zero: value %i is %s', (initial, active) => {
		mount({ states: 3, initial });
		expect(mainAnchor().hasAttribute('data-active')).toBe(active);
	});

	// A quadstate icon carries two proto fields, and the second badge is the only affordance that a
	// fourth state exists — so both badges stay visible in opposite corners and only `active` moves.
	it('keeps both improved anchors visible, marking only the ones the value has reached', () => {
		const { set } = mount({ states: 4, improvedIcon, improvedIcon2, initial: 3 });
		const [, improved1, improved2] = allAnchors();
		expect(improved1.hidden).toBe(false);
		expect(improved2.hidden).toBe(false);
		expect(improved1.hasAttribute('data-active')).toBe(true);
		expect(improved2.hasAttribute('data-active')).toBe(true);

		set(1);
		expect(improved1.hidden).toBe(false);
		expect(improved2.hidden).toBe(false);
		expect(improved1.hasAttribute('data-active')).toBe(false);
		expect(improved2.hasAttribute('data-active')).toBe(false);
	});

	it('renders an improved anchor only for an improved icon it was given', () => {
		mount({ states: 4, initial: 2, improvedIcon2 });
		expect(screen.queryByTestId('icon-input-improved1')).toBeNull();
		expect(screen.getByTestId('icon-input-improved2').hidden).toBe(false);
	});

	it('carries the wowhead opt-outs and names no target window', () => {
		mount({ states: 3, improvedIcon });
		const [main, improved1] = allAnchors();
		expect(main.target).toBe('');
		expect(main.dataset.whtticon).toBe('false');
		expect(main.dataset.disableWowheadTouchTooltip).toBe('true');
		expect(improved1.dataset.whtticon).toBe('false');
		expect(improved1.dataset.disableWowheadTouchTooltip).toBe('true');
	});

	it('marks the first improved anchor active only above value 1', () => {
		const { set } = mount({ states: 3, improvedIcon, initial: 1 });
		expect(allAnchors()[1].hasAttribute('data-active')).toBe(false);

		set(2);
		expect(allAnchors()[1].hasAttribute('data-active')).toBe(true);
	});

	it('activates the second improved anchor at value 3, and drops it entirely below states 4', () => {
		const { rerender } = render(
			<IconPicker states={4} icon={buffIcon} improvedIcon={improvedIcon} improvedIcon2={improvedIcon2} value={2} onChange={() => {}} />,
		);
		expect(allAnchors()[2].hasAttribute('data-active')).toBe(false);

		rerender(<IconPicker states={4} icon={buffIcon} improvedIcon={improvedIcon} improvedIcon2={improvedIcon2} value={3} onChange={() => {}} />);
		expect(allAnchors()[2].hasAttribute('data-active')).toBe(true);

		rerender(<IconPicker states={3} icon={buffIcon} improvedIcon={improvedIcon} value={3} onChange={() => {}} />);
		expect(allAnchors()).toHaveLength(2);
		expect(allAnchors()[1].hidden).toBe(false);
	});

	// Dimming a swatch does not stop pointer events.
	it('changes nothing on a left or right click while disabled', () => {
		const { onChange } = mount({ states: 3, initial: 1, disabled: true });

		fireEvent.click(mainAnchor());
		fireEvent.mouseDown(mainAnchor(), { button: 2 });
		expect(onChange).not.toHaveBeenCalled();
	});

	it('marks the swatch disabled, and the field around a labelled one', () => {
		const { unmount } = mount({ disabled: true });
		expect(mainAnchor().hasAttribute('data-disabled')).toBe(true);
		// The swatch links to wowhead, so it would otherwise be a natively focusable anchor.
		expect(mainAnchor().getAttribute('aria-disabled')).toBe('true');
		expect(mainAnchor().getAttribute('tabindex')).toBe('-1');
		unmount();

		const { container } = mount({ disabled: true, label: 'Buff' });
		expect(mainAnchor().hasAttribute('data-disabled')).toBe(true);
		expect((container.firstElementChild as HTMLElement).hasAttribute('data-disabled')).toBe(true);
	});

	it('lays the improved anchors beside the swatch in one box, so no anchor sits inside another', () => {
		const { container } = mount({ states: 4, improvedIcon, improvedIcon2, initial: 3 });
		const box = mainAnchor().parentElement!;

		expect(box).toBe(container.firstElementChild);
		expect(box.tagName).toBe('DIV');
		expect(Array.from(box.children)).toEqual([mainAnchor(), screen.getByTestId('icon-input-improved1'), screen.getByTestId('icon-input-improved2')]);
		expect(screen.getByTestId('icon-input-improved1').getAttribute('href')).toBe('https://www.wowhead.com/spell=2');
		expect(screen.getByTestId('icon-input-improved2').getAttribute('href')).toBe('https://www.wowhead.com/spell=3');
		expect(document.querySelectorAll('a a')).toHaveLength(0);
	});

	it('keeps the box inside a labelled field', () => {
		const { container } = mount({ states: 3, improvedIcon, label: 'Mark' });
		const field = container.firstElementChild as HTMLElement;

		expect(field.getAttribute('data-layout')).toBe('icon');
		expect(mainAnchor().parentElement!.parentElement).toBe(field);
	});

	// The improved icons are anchors of their own, so without the handlers a click on one would follow its wowhead link and leave the value alone.
	it('left-clicks, right-clicks and suppresses the context menu from an improved anchor as well', () => {
		const { onChange } = mount({ states: 3, improvedIcon, initial: 1 });
		const improved = screen.getByTestId('icon-input-improved1');

		// fireEvent returns false when preventDefault() was called.
		expect(fireEvent.click(improved)).toBe(false);
		expect(onChange).toHaveBeenLastCalledWith(2);

		fireEvent.mouseDown(improved, { button: 2 });
		expect(onChange).toHaveBeenLastCalledWith(1);

		expect(fireEvent.contextMenu(improved)).toBe(false);
		expect(onChange).toHaveBeenCalledTimes(2);
	});
});
