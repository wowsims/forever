import { act, fireEvent, render, screen } from '@testing-library/react';
import { type Ref, useImperativeHandle, useState } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { NumberField, NumberListField, SearchField, StringField } from './TextField';
import type { SearchFieldProps, StringFieldProps } from './types';

interface HarnessHandle<V> {
	set: (next: V) => void;
}

const commits = vi.fn<(next: unknown) => void>();

type HarnessProps = Partial<Omit<StringFieldProps, 'value' | 'onChange'>> & {
	initial?: string;
	/** What the parent keeps of a committed value; `null` keeps nothing. */
	keep?: (next: string) => string | null;
	handle?: Ref<HarnessHandle<string>>;
};

// The parent owns the value; `handle.set` is a change landing from outside the field.
const Harness = ({ initial = '', keep = next => next, handle, ...props }: HarnessProps) => {
	const [value, setValue] = useState(initial);
	useImperativeHandle(handle, () => ({ set: setValue }), []);
	return (
		<StringField
			id="rotation-name"
			{...props}
			value={value}
			onChange={next => {
				commits(next);
				const kept = keep(next);
				if (kept !== null) setValue(kept);
			}}
		/>
	);
};

const SearchHarness = ({
	initial = '',
	handle,
	...props
}: Partial<Omit<SearchFieldProps, 'value' | 'onChange'>> & { initial?: string; handle?: Ref<HarnessHandle<string>> }) => {
	const [value, setValue] = useState(initial);
	useImperativeHandle(handle, () => ({ set: setValue }), []);
	return (
		<SearchField
			placeholder="Search"
			inputTestId="search-input"
			{...props}
			value={value}
			onChange={next => {
				commits(next);
				setValue(next);
			}}
		/>
	);
};

const input = () => screen.getByRole('textbox') as HTMLInputElement;
const type = (text: string) => fireEvent.input(input(), { target: { value: text } });
const handleRef = <V,>() => ({ current: null as HarnessHandle<V> | null });

beforeEach(() => {
	commits.mockClear();
});

describe('StringField', () => {
	it.each([
		[undefined, 'sm'],
		['xs', 'xs'],
		['md', 'md'],
	] as const)('sizes the input as %s by default or as asked (%s)', (size, expected) => {
		render(<Harness size={size} />);

		expect(input().getAttribute('data-size')).toBe(expected);
	});

	describe('shell', () => {
		it('renders the bare input when there is no label, description, layout or clear button', () => {
			const { container } = render(<Harness className="w-2/5" testId="rotation-name-field" />);
			expect(container.firstElementChild).toBe(input());
			expect(input().type).toBe('text');
			expect(input().id).toBe('rotation-name');
			expect(screen.getByTestId('rotation-name-field')).toBe(input());
		});

		it('puts the input test id on a bare input over the field test id', () => {
			render(<Harness testId="rotation-name-field" inputTestId="rotation-name-input" />);
			expect(screen.getByTestId('rotation-name-input')).toBe(input());
			expect(screen.queryByTestId('rotation-name-field')).toBeNull();
		});

		it('renders a field root for a label, with the description', () => {
			const { container } = render(
				<Harness label="Rotation Name" description="Shown in the list" testId="rotation-name-field" inputTestId="rotation-name-input" />,
			);
			expect(container.firstElementChild).toBe(screen.getByTestId('rotation-name-field'));
			expect(screen.getByLabelText('Rotation Name')).toBe(input());
			expect(screen.getByText('Shown in the list')).toBeTruthy();
			expect(screen.getByTestId('rotation-name-input')).toBe(input());
		});

		it.each([
			['a description', { description: 'Shown in the list' }],
			['a layout', { layout: 'inline' as const }],
			['a clear button', { clear: { label: 'Clear' } }],
		])('renders a field root for %s', (_, props) => {
			const { container } = render(<Harness {...props} />);
			expect(container.firstElementChild).not.toBe(input());
			expect(container.firstElementChild!.contains(input())).toBe(true);
		});

		it.each([
			['a labelled field, marking the root', 'Rotation Name'],
			['a bare input', undefined],
		])('disables the input of %s when disabled', (_, label) => {
			render(<Harness label={label} testId="rotation-name-field" disabled />);
			expect(input().disabled).toBe(true);
			if (label) expect(screen.getByTestId('rotation-name-field').hasAttribute('data-disabled')).toBe(true);
		});

		it('passes the placeholder and autocomplete through', () => {
			render(<Harness placeholder="Name it" autoComplete="off" />);
			expect(input().placeholder).toBe('Name it');
			expect(input().getAttribute('autocomplete')).toBe('off');
		});
	});

	describe('commit on blur or Enter', () => {
		it('shows the value and reports a committed edit', () => {
			render(<Harness initial="abc" />);
			expect(input().value).toBe('abc');

			fireEvent.change(input(), { target: { value: 'xyz' } });
			expect(commits).toHaveBeenCalledExactlyOnceWith('xyz');
			expect(input().value).toBe('xyz');
		});

		// The field listens for the native `change` event, not React's onChange (which is the input
		// event) and not onBlur. Both alternatives look equivalent and are not: typing would commit per
		// keystroke, and onBlur is a delegated listener that never fires when the close path detaches
		// the node, which is what happens to the bonus-stat popover on Escape. Only half of that is
		// reachable under happy-dom; this covers that half.
		it('does not commit while the field is being typed into, only on the native change event', () => {
			render(<Harness initial="abc" />);

			type('typed');
			expect(commits).not.toHaveBeenCalled();
			expect(input().value).toBe('typed');

			fireEvent.change(input());
			expect(commits).toHaveBeenCalledExactlyOnceWith('typed');
		});

		it('reports nothing when the field is blurred without an edit', () => {
			render(<Harness initial="abc" />);
			fireEvent.focus(input());
			fireEvent.blur(input());
			expect(commits).not.toHaveBeenCalled();
			expect(input().value).toBe('abc');
		});

		it('drops the draft when the value changes underneath it', () => {
			const handle = handleRef<string>();
			render(<Harness initial="abc" handle={handle} />);
			type('typed-but-uncommitted');

			act(() => handle.current!.set('xyz'));
			expect(input().value).toBe('xyz');
		});

		it('shows the value the parent kept after a commit it clamped to the shown value', () => {
			render(<Harness initial="kept" keep={() => 'kept'} />);
			type('dropped');
			fireEvent.change(input());
			expect(commits).toHaveBeenCalledExactlyOnceWith('dropped');
			expect(input().value).toBe('kept');
		});

		it('shows the value after a commit the parent does not take', () => {
			render(<Harness initial="kept" keep={() => null} />);
			type('dropped');
			fireEvent.change(input());
			expect(input().value).toBe('kept');
		});
	});

	it('offers no clear button unless asked', () => {
		render(<Harness initial="abc" label="Rotation Name" />);
		expect(screen.queryByRole('button')).toBeNull();
	});

	describe('autosize', () => {
		it('sizes the input to its text, at least 3 wide, following the draft and the value', () => {
			const handle = handleRef<string>();
			render(<Harness handle={handle} />);
			expect(input().size).toBe(3);

			type('123456');
			expect(input().size).toBe(6);

			act(() => handle.current!.set('ab'));
			expect(input().size).toBe(3);

			act(() => handle.current!.set('a much longer value'));
			expect(input().size).toBe('a much longer value'.length);
		});
	});

	it('reports the text raw, sized to it, with the adaptive-string-picker-root test id on its root', () => {
		const onChange = vi.fn();
		render(<StringField id="custom-name" label="Custom Name" value="initial-value" onChange={onChange} />);
		expect(input().size).toBe('initial-value'.length);
		expect(screen.getByTestId('adaptive-string-picker-root').contains(input())).toBe(true);

		fireEvent.change(input(), { target: { value: '  raw 123 text  ' } });
		expect(onChange).toHaveBeenCalledExactlyOnceWith('  raw 123 text  ');
	});
});

const NumberHarness = ({
	initial = 0,
	handle,
	...props
}: Partial<Parameters<typeof NumberField>[0]> & { initial?: number; handle?: Ref<HarnessHandle<number>> }) => {
	const [value, setValue] = useState(initial);
	useImperativeHandle(handle, () => ({ set: setValue }), []);
	return (
		<NumberField
			id="cast-delay"
			label="Cast Delay"
			{...props}
			value={value}
			onChange={next => {
				commits(next);
				setValue(next);
			}}
		/>
	);
};

describe('NumberField', () => {
	it('carries the number-picker-root test id on its root by default', () => {
		render(<NumberHarness />);
		expect(screen.getByTestId('number-picker-root').contains(input())).toBe(true);
	});

	it('rejects autosize, since NumberField always sets it itself', () => {
		// @ts-expect-error NumberFieldProps omits autosize, the same as StringField and the others.
		const withAutosize = <NumberField id="x" value={0} onChange={() => {}} autosize />;
		expect(withAutosize).toBeTruthy();
	});

	it('shows a whole number with String(value) and sizes to it', () => {
		render(<NumberHarness initial={12345} />);
		expect(input().value).toBe('12345');
		expect(input().size).toBe(5);
	});

	it.each([
		[1234.5, undefined, '1234.50'],
		[3.14159, 4, '3.1416'],
	])('formats the float %d ungrouped, at maxDecimalDigits %s (2 by default)', (initial, maxDecimalDigits, shown) => {
		render(<NumberHarness initial={initial} float maxDecimalDigits={maxDecimalDigits} />);
		expect(input().value).toBe(shown);
	});

	it('shows zero as an empty field when showZeroes is false, and as 0 otherwise', () => {
		const { unmount } = render(<NumberHarness initial={0} showZeroes={false} />);
		expect(input().value).toBe('');
		unmount();

		render(<NumberHarness initial={0} />);
		expect(input().value).toBe('0');
	});

	it('reads an integer back with parseInt, falling back to 0', () => {
		render(<NumberHarness initial={5} />);
		fireEvent.change(input(), { target: { value: '12abc' } });
		expect(commits).toHaveBeenLastCalledWith(12);
		fireEvent.change(input(), { target: { value: 'abc' } });
		expect(commits).toHaveBeenLastCalledWith(0);
	});

	it('reads a float back as a float, falling back to 0 on garbage', () => {
		render(<NumberHarness float />);
		type('12.5');
		fireEvent.change(input());
		expect(commits).toHaveBeenLastCalledWith(12.5);

		fireEvent.change(input(), { target: { value: 'abc' } });
		expect(commits).toHaveBeenLastCalledWith(0);
	});

	it('shows the parsed value after a commit that leaves the value where it was', () => {
		render(<NumberHarness initial={12} />);
		fireEvent.change(input(), { target: { value: '12abc' } });
		expect(commits).toHaveBeenLastCalledWith(12);
		expect(input().value).toBe('12');
	});

	it.each([
		['an integer', {}, '-8', 8, '8'],
		['a float', { float: true }, '-12.5', 12.5, '12.50'],
		['a float past the thousands', { float: true }, '-1234.5', 1234.5, '1234.50'],
	])('commits the absolute value of %s when positive', (_, options, typed, committed, shown) => {
		render(<NumberHarness positive {...options} />);
		fireEvent.change(input(), { target: { value: typed } });
		expect(commits).toHaveBeenLastCalledWith(committed);
		expect(input().value).toBe(shown);
	});

	it('shows a value the parent changes underneath it, dropping a draft', () => {
		const handle = handleRef<number>();
		render(<NumberHarness initial={1} handle={handle} />);
		type('123');
		act(() => handle.current!.set(9));
		expect(input().value).toBe('9');
	});
});

const ListHarness = ({
	initial = [],
	keep = next => next,
	handle,
}: {
	initial?: Array<number>;
	keep?: (next: Array<number>) => Array<number> | null;
	handle?: Ref<HarnessHandle<Array<number>>>;
}) => {
	const [value, setValue] = useState(initial);
	useImperativeHandle(handle, () => ({ set: setValue }), []);
	return (
		<NumberListField
			id="thresholds"
			placeholder="20, 40, ..."
			value={value}
			onChange={next => {
				commits(next);
				const kept = keep(next);
				if (kept) setValue(kept);
			}}
		/>
	);
};

describe('NumberListField', () => {
	it('shows the list comma-joined, with the number-list-picker-input test id and the placeholder', () => {
		render(<ListHarness initial={[1, 2, 3]} />);
		expect(input().value).toBe('1,2,3');
		expect(screen.getByTestId('number-list-picker-input')).toBe(input());
		expect(input().placeholder).toBe('20, 40, ...');
	});

	it.each([
		['', []],
		['1,2,x,3', [1, 2, 3]],
		['1.5,2.25', [1.5, 2.25]],
	])('reads %j back as %j', (typed, parsed) => {
		render(<ListHarness initial={[9]} />);
		fireEvent.change(input(), { target: { value: typed } });
		expect(commits).toHaveBeenCalledExactlyOnceWith(parsed);
	});

	// A re-render that leaves the value equal would otherwise stomp the field mid-edit, e.g. while
	// typing the trailing comma in '1,2,'.
	it('keeps an edit that still parses to the current value across a same-value notification', () => {
		const handle = handleRef<Array<number>>();
		render(<ListHarness initial={[1, 2]} handle={handle} />);
		type('1,2,');
		act(() => handle.current!.set([1, 2]));
		expect(input().value).toBe('1,2,');
	});

	it.each([
		['1,', [1], '1'],
		['1, 2,', [1, 2], '1,2'],
	])('normalises %j to the committed value on commit', (typed, parsed, shown) => {
		render(<ListHarness initial={[5]} />);
		type(typed);
		fireEvent.change(input());
		expect(commits).toHaveBeenCalledExactlyOnceWith(parsed);
		expect(input().value).toBe(shown);
	});

	it('normalises on commit even when the parent keeps nothing', () => {
		render(<ListHarness initial={[1]} keep={() => null} />);
		type('1,');
		fireEvent.change(input());
		expect(input().value).toBe('1');
	});

	it('shows what the parent kept when it rewrites the committed value', () => {
		render(<ListHarness keep={next => [...next].sort((a, b) => a - b)} />);
		type('3, 1,');
		fireEvent.change(input());
		expect(input().value).toBe('1,3');
	});
});

describe('SearchField', () => {
	it('is a bare input that reports every keystroke', () => {
		const { container } = render(<SearchHarness />);
		expect(container.firstElementChild).toBe(screen.getByTestId('search-input'));
		type('a');
		expect(commits).toHaveBeenLastCalledWith('a');
		type('ab');
		expect(commits).toHaveBeenLastCalledWith('ab');
		expect(input().value).toBe('ab');
	});

	it('shows a value the parent changes underneath it', () => {
		const handle = handleRef<string>();
		render(<SearchHarness initial="abc" handle={handle} />);
		act(() => handle.current!.set('xyz'));
		expect(input().value).toBe('xyz');
	});

	it('shows the clear button once there is text, and clearing commits empty text', () => {
		render(<SearchHarness clear={{ label: 'Clear' }} />);
		expect(screen.queryByRole('button', { name: 'Clear' })).toBeNull();

		type('abc');
		fireEvent.click(screen.getByRole('button', { name: 'Clear' }));
		expect(commits).toHaveBeenLastCalledWith('');
		expect(input().value).toBe('');
	});

	it('leaves the size alone', () => {
		render(<SearchHarness initial="a much longer value" />);
		expect(input().hasAttribute('size')).toBe(false);
	});

	describe('debounced', () => {
		beforeEach(() => vi.useFakeTimers());
		afterEach(() => vi.useRealTimers());

		it('keeps the field responsive and reports once typing pauses for debounceMs', () => {
			render(<SearchHarness debounceMs={150} />);
			type('Ashkandi');
			expect(input().value).toBe('Ashkandi');
			expect(commits).not.toHaveBeenCalled();

			act(() => vi.advanceTimersByTime(149));
			expect(commits).not.toHaveBeenCalled();

			act(() => vi.advanceTimersByTime(1));
			expect(commits).toHaveBeenCalledExactlyOnceWith('Ashkandi');
		});

		it('restarts the window on every keystroke', () => {
			render(<SearchHarness debounceMs={150} />);
			type('a');
			act(() => vi.advanceTimersByTime(100));
			type('ab');
			act(() => vi.advanceTimersByTime(100));
			expect(commits).not.toHaveBeenCalled();

			act(() => vi.advanceTimersByTime(50));
			expect(commits).toHaveBeenCalledExactlyOnceWith('ab');
		});

		it('reports nothing on mount', () => {
			render(<SearchHarness debounceMs={150} initial="abc" />);
			act(() => vi.advanceTimersByTime(1000));
			expect(commits).not.toHaveBeenCalled();
		});

		it('drops a pending commit when the value changes underneath it', () => {
			const handle = handleRef<string>();
			render(<SearchHarness debounceMs={150} handle={handle} />);
			type('typed');
			act(() => handle.current!.set('xyz'));
			act(() => vi.advanceTimersByTime(1000));
			expect(commits).not.toHaveBeenCalled();
			expect(input().value).toBe('xyz');
		});
	});
});
