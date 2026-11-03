import { fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { useField } from './hooks/useField';
import { Field } from './index';

/** Stands in for a picker's swatch: registers itself on `controlRef` the way an icon picker's own trigger does, and carries `field.labelId` back out so a test can compare it against the rendered label's id. */
const ControlProbe = ({ onClick }: { onClick: () => void }) => {
	const field = useField();
	return (
		<button
			type="button"
			data-testid="control"
			aria-labelledby={field.labelId}
			ref={node => {
				field.controlRef.current = node;
			}}
			onClick={onClick}
		/>
	);
};

describe('Field parts', () => {
	it('points the label at the root id', () => {
		render(
			<Field.Root id="cast-delay">
				<Field.Label>Cast Delay</Field.Label>
			</Field.Root>,
		);
		const label = screen.getByTestId('form-label');
		expect(label.getAttribute('for')).toBe('cast-delay');
		expect(label.getAttribute('title')).toBe('Cast Delay');
	});

	it('renders a span caption that points at nothing when asked for one', () => {
		render(
			<Field.Root id="elixirs">
				<Field.Label as="span">Elixirs</Field.Label>
			</Field.Root>,
		);
		const caption = screen.getByTestId('form-label');
		expect(caption.tagName).toBe('SPAN');
		expect(caption.hasAttribute('for')).toBe(false);
		expect(caption.getAttribute('title')).toBe('Elixirs');
	});

	it('names an icon field label as a span carrying the field labelId, not a for', () => {
		render(
			<Field.Root layout="icon">
				<Field.Label>Buff</Field.Label>
				<ControlProbe onClick={() => {}} />
			</Field.Root>,
		);
		const label = screen.getByTestId('form-label');
		expect(label.tagName).toBe('SPAN');
		expect(label.hasAttribute('for')).toBe(false);
		expect(label.id).toBeTruthy();
		expect(label.hasAttribute('data-activates')).toBe(true);
		expect(screen.getByTestId('control').getAttribute('aria-labelledby')).toBe(label.id);
	});

	it('clicks the element on controlRef when an icon field label is clicked', () => {
		const onClick = vi.fn();
		render(
			<Field.Root layout="icon">
				<Field.Label>Buff</Field.Label>
				<ControlProbe onClick={onClick} />
			</Field.Root>,
		);
		fireEvent.click(screen.getByTestId('form-label'));
		expect(onClick).toHaveBeenCalledOnce();
	});

	it('keeps the label htmlFor for a non-icon layout', () => {
		render(
			<Field.Root id="stat-weight" layout="inline">
				<Field.Label>Stat Weight</Field.Label>
			</Field.Root>,
		);
		const label = screen.getByTestId('form-label');
		expect(label.tagName).toBe('LABEL');
		expect(label.getAttribute('for')).toBe('stat-weight');
		expect(label.hasAttribute('data-activates')).toBe(false);
	});

	it('captions a group root with a span by default, and with a label when asked, both carrying the field labelId', () => {
		const { rerender } = render(
			<Field.Root id="elixirs" role="group">
				<Field.Label>Elixirs</Field.Label>
				<ControlProbe onClick={() => {}} />
			</Field.Root>,
		);
		expect(screen.getByTestId('form-label').tagName).toBe('SPAN');
		expect(screen.getByTestId('input-root').getElementsByTagName('label')).toHaveLength(0);
		expect(screen.getByTestId('form-label').id).toBe(screen.getByTestId('control').getAttribute('aria-labelledby'));

		rerender(
			<Field.Root id="elixirs" role="group">
				<Field.Label as="label">Elixirs</Field.Label>
				<ControlProbe onClick={() => {}} />
			</Field.Root>,
		);
		expect(screen.getByTestId('form-label').tagName).toBe('LABEL');
		expect(screen.getByTestId('form-label').id).toBe(screen.getByTestId('control').getAttribute('aria-labelledby'));
	});

	it('gives a native control the root id, so the label names it', () => {
		render(
			<Field.Root id="cast-delay">
				<Field.Label>Cast Delay</Field.Label>
				<Field.Control>
					<input />
				</Field.Control>
			</Field.Root>,
		);
		const input = screen.getByRole('textbox', { name: 'Cast Delay' });
		expect(input.id).toBe('cast-delay');
		expect(screen.queryByTestId('field-control')).toBeNull();
	});

	it('puts testId on a native control', () => {
		render(
			<Field.Root id="cast-delay">
				<Field.Control testId="cast-delay-input">
					<input />
				</Field.Control>
			</Field.Root>,
		);
		expect(screen.getByTestId('cast-delay-input').tagName).toBe('INPUT');
	});

	it('describes a native control by the description', () => {
		render(
			<Field.Root id="cast-delay">
				<Field.Label>Cast Delay</Field.Label>
				<Field.Description>Delay before the first cast</Field.Description>
				<Field.Control>
					<input />
				</Field.Control>
			</Field.Root>,
		);
		const description = screen.getByTestId('input-description');
		expect(description.tagName).toBe('DIV');
		expect(description.id).toBeTruthy();
		expect(screen.getByRole('textbox').getAttribute('aria-describedby')).toBe(description.id);
	});

	it('wraps a control that is not native, describing the wrapper and leaving the id to the control', () => {
		render(
			<Field.Root id="rotation-type">
				<Field.Description>How the rotation is chosen</Field.Description>
				<Field.Control testId="rotation-type-control">
					<button type="button">Simple</button>
				</Field.Control>
			</Field.Root>,
		);
		const wrapper = screen.getByTestId('rotation-type-control');
		expect(wrapper.tagName).toBe('DIV');
		expect(wrapper.hasAttribute('id')).toBe(false);
		expect(wrapper.getAttribute('aria-describedby')).toBe(screen.getByTestId('input-description').id);
		expect(within(wrapper).getByRole('button', { name: 'Simple' })).toBeTruthy();
	});

	it('leaves a wrapped control undescribed without a description', () => {
		render(
			<Field.Root>
				<Field.Control>
					<button type="button">Simple</button>
				</Field.Control>
			</Field.Root>,
		);
		expect(screen.getByTestId('field-control').hasAttribute('aria-describedby')).toBe(false);
	});

	it('disables a native control through the root', () => {
		render(
			<Field.Root id="cast-delay" disabled>
				<Field.Control>
					<input />
				</Field.Control>
			</Field.Root>,
		);
		expect(screen.getByTestId('input-root').hasAttribute('data-disabled')).toBe(true);
		expect((screen.getByRole('textbox') as HTMLInputElement).disabled).toBe(true);
	});

	it('marks layout and size on the root', () => {
		const { rerender } = render(<Field.Root layout="split" size="xs" />);
		const root = screen.getByTestId('input-root');
		expect(root.getAttribute('data-layout')).toBe('split');
		expect(root.getAttribute('data-size')).toBe('xs');

		rerender(<Field.Root />);
		expect(root.hasAttribute('data-layout')).toBe(false);
		expect(root.hasAttribute('data-size')).toBe(false);
	});

	it('gives a nested root the size of its parent unless it sets its own', () => {
		render(
			<Field.Root size="xs" testId="outer">
				<Field.Root testId="inherits" />
				<Field.Root size="md" testId="own" />
			</Field.Root>,
		);
		expect(['outer', 'inherits', 'own'].map(id => screen.getByTestId(id).getAttribute('data-size'))).toEqual(['xs', 'xs', 'md']);
	});

	it('renders leading and after where they are composed, as children of the root', () => {
		render(
			<Field.Root id="cast-delay">
				<Field.Leading>
					<span />
				</Field.Leading>
				<Field.Label>Cast Delay</Field.Label>
				<Field.Control>
					<input />
				</Field.Control>
				<Field.After>s</Field.After>
			</Field.Root>,
		);
		const root = screen.getByTestId('input-root');
		expect([...root.children].map(el => el.getAttribute('data-testid') ?? el.tagName)).toEqual(['field-leading', 'form-label', 'INPUT', 'field-after']);
		expect(screen.getByTestId('field-after').textContent).toBe('s');
	});

	it('renders the clear button beside its control and calls onClear', () => {
		const onClear = vi.fn();
		render(
			<Field.Root>
				<Field.Clear show label="Clear" testId="clear-btn" onClear={onClear}>
					<input />
				</Field.Clear>
			</Field.Root>,
		);
		const button = screen.getByRole('button', { name: 'Clear' });
		expect(button).toBe(screen.getByTestId('clear-btn'));

		fireEvent.click(button);
		expect(onClear).toHaveBeenCalledOnce();
	});

	it('forwards testId on every part', () => {
		render(
			<Field.Root id="x" testId="root">
				<Field.Leading testId="leading" />
				<Field.Label testId="label">X</Field.Label>
				<Field.Description testId="description">About X</Field.Description>
				<Field.Control testId="control">
					<button type="button">X</button>
				</Field.Control>
				<Field.Clear show testId="clear" onClear={() => {}}>
					<input />
				</Field.Clear>
				<Field.After testId="after" />
			</Field.Root>,
		);
		for (const testId of ['root', 'leading', 'label', 'description', 'control', 'clear', 'after']) {
			expect(screen.getByTestId(testId)).toBeTruthy();
		}
	});

	it('shows a label tooltip on hover', async () => {
		render(
			<Field.Root>
				<Field.Label tooltip="Seconds of <b>human</b> reaction time">Cast Delay</Field.Label>
			</Field.Root>,
		);
		const label = screen.getByTestId('form-label');
		expect(label.getAttribute('data-tooltip-id')).toBeTruthy();

		fireEvent.mouseEnter(label);
		expect((await screen.findByText('human')).tagName).toBe('B');
		expect((await screen.findByRole('tooltip')).className).toContain('react-tooltip__place-top');
	});

	it('opens a label tooltip on the side it is given', async () => {
		render(
			<Field.Root>
				<Field.Label tooltip="Same as Patchwerk" tooltipPlace="right">
					Patchwerk copy
				</Field.Label>
			</Field.Root>,
		);

		fireEvent.mouseEnter(screen.getByTestId('form-label'));
		expect((await screen.findByRole('tooltip')).className).toContain('react-tooltip__place-right');
	});

	it('refuses a part outside a root', () => {
		const error = vi.spyOn(console, 'error').mockImplementation(() => {});
		expect(() => render(<Field.Label>Cast Delay</Field.Label>)).toThrow('A Field part must be rendered inside a Field.Root.');
		error.mockRestore();
	});
});
