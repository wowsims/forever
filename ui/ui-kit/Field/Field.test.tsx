import { fireEvent, render, screen, within } from '@testing-library/react';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { Dialog } from '@ui-kit/Dialog';
import { Input } from '@ui-kit/FormControl';
import { Popover } from '@ui-kit/Popover';
import { describe, expect, it, vi } from 'vitest';

import { Field } from './Field';
import type { ControlSize, FieldProps } from './types';

const field = (props: Partial<FieldProps> = {}) =>
	render(
		<Field id="cast-delay" label="Cast Delay" className="number-picker-root" {...props}>
			<input id="cast-delay" />
		</Field>,
	);

const root = () => screen.getByTestId('input-root');

describe('Field', () => {
	it('marks the layout and the disabled state on the root', () => {
		field({ layout: 'inline', disabled: true });
		expect(root().getAttribute('data-layout')).toBe('inline');
		expect(root().hasAttribute('data-disabled')).toBe(true);
	});

	it('marks a split layout on the root', () => {
		field({ layout: 'split' });
		expect(root().getAttribute('data-layout')).toBe('split');
	});

	it('omits the state attributes when neither applies', () => {
		field();
		expect(root().hasAttribute('data-layout')).toBe(false);
		expect(root().hasAttribute('data-disabled')).toBe(false);
		expect(root().hasAttribute('role')).toBe(false);
		expect(root().hasAttribute('aria-labelledby')).toBe(false);
	});

	it('forwards role and aria-labelledby onto the root, leaving its own attributes alone', () => {
		field({ role: 'group', 'aria-labelledby': 'buffs-label', layout: 'inline' });
		expect(root().getAttribute('role')).toBe('group');
		expect(root().getAttribute('aria-labelledby')).toBe('buffs-label');
		expect(root().getAttribute('data-layout')).toBe('inline');
	});

	it('names a group by its own label when no aria-labelledby is given', () => {
		render(
			<Field label="Elixirs" role="group">
				<button type="button">Flask</button>
			</Field>,
		);

		const group = screen.getByRole('group', { name: 'Elixirs' });
		expect(group).toBe(root());
		expect(group.getAttribute('aria-labelledby')).toBe(within(group).getByTestId('form-label').id);
	});

	it('keeps an explicit aria-labelledby on a group, and gives a field that is not a group no label reference', () => {
		const { unmount } = field({ role: 'group', 'aria-labelledby': 'buffs-label' });
		expect(root().getAttribute('aria-labelledby')).toBe('buffs-label');
		unmount();

		field();
		expect(root().hasAttribute('aria-labelledby')).toBe(false);
	});

	it('gives a group no aria-labelledby without a label', () => {
		field({ role: 'group', label: undefined });
		expect(root().hasAttribute('aria-labelledby')).toBe(false);
	});

	it('hides a hidden field and keeps its children mounted', () => {
		const { rerender } = field({ hidden: true });
		expect(root().hidden).toBe(true);
		expect(root().querySelector('input')).not.toBeNull();

		rerender(
			<Field id="cast-delay" label="Cast Delay">
				<input id="cast-delay" />
			</Field>,
		);
		expect(root().hidden).toBe(false);
	});

	it('takes its test id from testId', () => {
		field({ testId: 'number-picker-root' });
		expect(screen.getByTestId('number-picker-root')).toBeTruthy();
		expect(screen.queryByTestId('input-root')).toBeNull();
	});

	it('renders no label at all without one', () => {
		field({ label: undefined });
		expect(within(root()).queryByTestId('form-label')).toBeNull();
	});

	it('puts a string description in .input-description, before the input', () => {
		field({ description: 'Delay before the first cast' });
		expect([...root().children].map(el => el.tagName)).toEqual(['LABEL', 'DIV', 'INPUT']);
		expect(within(root()).getByTestId('input-description').textContent).toBe('Delay before the first cast');
	});

	it('renders leading before the label', () => {
		field({ leading: <span data-testid="leading" /> });
		expect(root().firstElementChild).toBe(screen.getByTestId('leading'));
	});

	it('anchors a labelTooltip to the label when the field has no id', async () => {
		render(<Field label="Cast Delay" labelTooltip="Seconds of human reaction time" />);
		const label = screen.getByTestId('form-label');
		expect(label.getAttribute('data-tooltip-id')).toBeTruthy();

		fireEvent.mouseEnter(label);
		expect(await screen.findByText('Seconds of human reaction time')).toBeTruthy();
	});

	// `labelTooltip` is typed `string | ReactElement`, so the guard is only reachable from
	// an untyped caller — which is what the cast stands in for.
	it('warns instead of silently dropping a tooltip it cannot render', () => {
		const warn = vi.spyOn(console, 'warn').mockImplementation(() => {});
		field({ labelTooltip: (() => 'from a function') as unknown as string });
		expect(warn).toHaveBeenCalledOnce();
		expect(warn.mock.calls[0][0]).toMatch(/^Field "cast-delay": tooltip is neither/);
		expect(within(root()).getByTestId('form-label').getAttribute('data-tooltip-id')).toBeNull();
	});

	// Base UI merges a forwarded ref with its own, so this pins that the merge honours a cleanup the
	// ref callback returns instead of discarding it — without which anything a caller attaches to the
	// root would be attached again on every re-render and never released.
	it('forwards a ref to the root, once, and runs the cleanup it returns on unmount', () => {
		const attached: Array<Element> = [];
		const detach = vi.fn();
		const mount = (node: HTMLDivElement | null) => {
			if (!node) return;
			attached.push(node);
			return detach;
		};
		const fieldWith = (label: string) => <Field ref={mount} id="x" label={label} className="talents-picker-root" />;

		const { rerender, unmount } = render(fieldWith('One'));
		expect(attached).toEqual([root()]);

		rerender(fieldWith('Two'));
		expect(attached).toHaveLength(1);
		expect(detach).not.toHaveBeenCalled();

		unmount();
		expect(detach).toHaveBeenCalledOnce();
	});

	// Every markup string goes through LocaleHtml, which hands Field a React element rather than a
	// string.
	it('renders an element labelTooltip as markup instead of literal tags', async () => {
		field({
			labelTooltip: (
				<span>
					Used with the <b>Item Swap</b> action.
				</span>
			),
		});
		fireEvent.mouseEnter(screen.getByText('Cast Delay'));

		const tooltip = await screen.findByText('Item Swap');
		expect(tooltip.tagName).toBe('B');
		expect(document.body.textContent).not.toContain('<b>');
	});

	it('wraps the control with a clear button when clear.show is set', () => {
		const onClear = vi.fn();
		field({ clear: { show: true, label: 'Clear', testId: 'clear-btn', onClear } });
		const button = screen.getByTestId('clear-btn');
		expect(button.getAttribute('aria-label')).toBe('Clear');
		expect(button.parentElement!.querySelector('input')).toBeTruthy();

		fireEvent.click(button);
		expect(onClear).toHaveBeenCalledOnce();
	});

	it('keeps the wrapper and hides the button when clear.show is false', () => {
		const { rerender } = field({ clear: { show: false, testId: 'clear-btn', onClear: () => {} } });
		const input = root().querySelector('input');
		expect(screen.queryByTestId('clear-btn')).toBeNull();
		expect(input).not.toBeNull();
		expect(input!.parentElement).not.toBe(root());

		rerender(
			<Field id="cast-delay" label="Cast Delay" className="number-picker-root" clear={{ show: true, testId: 'clear-btn', onClear: () => {} }}>
				<input id="cast-delay" />
			</Field>,
		);
		expect(root().querySelector('input')).toBe(input);
		expect(screen.getByTestId('clear-btn')).toBeTruthy();
	});

	it('renders the after slot last', () => {
		field({ after: <span data-testid="after" /> });
		expect(root().lastElementChild).toBe(screen.getByTestId('after'));
	});
});

describe('control size', () => {
	const controls = (size?: ControlSize) => (
		<>
			<Input data-testid="input" size={size} />
			<Button size={size}>Run</Button>
			<BooleanPicker testId="checkbox" size={size} value={false} onChange={() => {}} />
			<ChoiceField testId="select" ariaLabel="Pick" size={size} options={[{ value: 1, label: 'One' }]} value={1} onChange={() => {}} />
		</>
	);
	const sizes = () => [
		screen.getByTestId('input').getAttribute('data-size'),
		screen.getByRole('button', { name: 'Run' }).getAttribute('data-size'),
		screen.getByTestId('checkbox').getAttribute('data-size'),
		screen.getByTestId('select').getAttribute('data-size'),
	];

	it.each([
		[undefined, 'sm'],
		['xs', 'xs'],
		['sm', 'sm'],
		['md', 'md'],
	] as const)('gives the input, button, checkbox and select the same size for size %s (%s)', (size, expected) => {
		render(controls(size));
		expect(sizes()).toEqual([expected, expected, expected, expected]);
	});

	it('sizes every control inside a sized field as the field, unless the control sets its own', () => {
		const { unmount } = render(<Field size="xs">{controls()}</Field>);
		expect(sizes()).toEqual(['xs', 'xs', 'xs', 'xs']);
		unmount();

		render(<Field size="xs">{controls('md')}</Field>);
		expect(sizes()).toEqual(['md', 'md', 'md', 'md']);
	});

	it('gives a button in a dialog opened from a sized field the default size', async () => {
		render(
			<Field size="xs">
				<Dialog open onOpenChange={() => {}}>
					<Button>Inside</Button>
				</Dialog>
			</Field>,
		);
		expect((await screen.findByRole('button', { name: 'Inside' })).getAttribute('data-size')).toBe('sm');
	});

	it('gives a button in a popover opened from a sized field the default size', async () => {
		render(
			<Field size="xs">
				<Popover trigger="Open" open>
					<Button>Inside</Button>
				</Popover>
			</Field>,
		);
		expect((await screen.findByRole('button', { name: 'Inside' })).getAttribute('data-size')).toBe('sm');
	});
});
