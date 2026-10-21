import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { type Ref, useImperativeHandle, useState } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { BooleanPicker, type BooleanPickerProps } from './BooleanPicker';

interface HarnessHandle {
	set: (next: boolean) => void;
}

type HarnessProps = Omit<BooleanPickerProps, 'value' | 'onChange'> & {
	initial?: boolean;
	accept?: (next: boolean) => boolean;
	handle?: Ref<HarnessHandle>;
};

const changes = vi.fn<(next: boolean) => void>();

// The parent owns the value; `handle.set` is a change landing from outside the picker.
const Harness = ({ initial = false, accept = next => next, handle, ...props }: HarnessProps) => {
	const [value, setValue] = useState(initial);
	useImperativeHandle(handle, () => ({ set: setValue }), []);
	return (
		<BooleanPicker
			id="use-item-swap"
			label="Enable Item Swap"
			{...props}
			value={value}
			onChange={next => {
				changes(next);
				setValue(accept(next));
			}}
		/>
	);
};

const checkbox = () => screen.getByRole('checkbox') as HTMLButtonElement;
const isChecked = (el: HTMLElement) => el.getAttribute('aria-checked') === 'true';

beforeEach(() => {
	changes.mockClear();
});

describe('BooleanPicker', () => {
	it('shows the value and reports a click', () => {
		render(<Harness initial />);
		expect(isChecked(checkbox())).toBe(true);

		fireEvent.click(checkbox());
		expect(changes).toHaveBeenLastCalledWith(false);
		expect(isChecked(checkbox())).toBe(false);
	});

	it('shows a value the parent changes underneath it', () => {
		const handle = { current: null as HarnessHandle | null };
		render(<Harness handle={handle} />);
		act(() => handle.current!.set(true));
		expect(isChecked(checkbox())).toBe(true);
	});

	it('renders the check mark only while the box is checked', async () => {
		render(<Harness initial />);
		expect(checkbox().firstElementChild).not.toBeNull();

		fireEvent.click(checkbox());
		await waitFor(() => expect(checkbox().firstElementChild).toBeNull());
	});

	it('follows the value the parent keeps, not the click', () => {
		render(<Harness accept={() => false} />);
		fireEvent.click(checkbox());
		expect(changes).toHaveBeenLastCalledWith(true);
		expect(isChecked(checkbox())).toBe(false);
	});

	it('disables the input and marks the root when disabled', () => {
		render(<Harness disabled />);
		expect(checkbox().getAttribute('aria-disabled')).toBe('true');
		expect(checkbox().closest('[data-testid="boolean-picker-root"]')!.hasAttribute('data-disabled')).toBe(true);
	});

	it('renders a labelled field: the label targets the checkbox, and the description sits alongside it', () => {
		const { container } = render(<Harness description="Swaps mid-fight" />);
		const root = container.firstElementChild!;
		expect(root.getAttribute('data-testid')).toBe('boolean-picker-root');
		expect(screen.getByText('Enable Item Swap').getAttribute('for')).toBe(checkbox().nextElementSibling!.id);
		expect(root.querySelector('[data-testid="boolean-picker-input"]')).toBe(checkbox());
	});

	it.each([
		[true, ['LABEL', 'DIV', 'SPAN', 'INPUT']],
		[false, ['SPAN', 'INPUT', 'LABEL', 'DIV']],
	])('orders the label and the box, and marks the root, for reverse=%s', (reverse, order) => {
		const { container } = render(<Harness reverse={reverse} description="Swaps mid-fight" />);
		const root = container.firstElementChild!;
		expect(root.hasAttribute('data-reverse')).toBe(reverse);
		expect([...root.children].map(el => el.tagName)).toEqual(order);
	});

	it('takes a root testId and a layout', () => {
		const { container } = render(<Harness testId="log-runner-show-debug" layout="inline" />);
		const root = container.firstElementChild!;
		expect(root.getAttribute('data-testid')).toBe('log-runner-show-debug');
		expect(root.getAttribute('data-layout')).toBe('inline');
	});

	it('toggles the value when the label is clicked', () => {
		render(<Harness />);
		fireEvent.click(screen.getByText('Enable Item Swap'));
		expect(changes).toHaveBeenLastCalledWith(true);
	});

	it('renders the checkbox alone, with no Field root, when there is no label, description or layout', () => {
		const { container } = render(<Harness label={undefined} />);
		expect(container.firstElementChild).toBe(checkbox());
		expect(container.querySelector('[data-testid="boolean-picker-root"]')).toBeNull();
	});

	it('marks a disabled, unlabelled checkbox directly, with no root to carry it', () => {
		render(<Harness label={undefined} disabled />);
		expect(checkbox().getAttribute('aria-disabled')).toBe('true');
	});

	it('takes a testId on the bare checkbox when there is no label', () => {
		render(<Harness label={undefined} testId="sw-stat-toggle" />);
		expect(screen.getByTestId('sw-stat-toggle')).toBe(checkbox());
	});

	it('points the checkbox at the text it is described by, and at nothing without it', () => {
		const { unmount } = render(<Harness aria-describedby="swap-note" />);
		expect(checkbox().getAttribute('aria-describedby')).toBe('swap-note');
		unmount();

		render(<Harness />);
		expect(checkbox().hasAttribute('aria-describedby')).toBe(false);
	});
});
