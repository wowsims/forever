import { act, fireEvent, render, screen } from '@testing-library/react';
import { useState } from 'react';
import { describe, expect, it, vi } from 'vitest';

import { CodeTextArea } from './CodeTextArea';

// The tests load no locale, so the position reads as its key unless the values are written in here.
vi.mock('@i18n/config', () => ({
	default: { t: (_key: string, { line, column, count }: Record<string, number>) => `Ln ${line}, Col ${column} · ${count} lines` },
}));

const TEXT = '{\n  "player": {\n    "name": "Player"\n  }\n}';

const position = (line: number, column: number, count: number) => `Ln ${line}, Col ${column} · ${count} lines`;

const Editable = ({ initial = TEXT, onChange }: { initial?: string; onChange?: (value: string) => void }) => {
	const [value, setValue] = useState(initial);
	return (
		<CodeTextArea
			value={value}
			testId="code"
			onChange={next => {
				setValue(next);
				onChange?.(next);
			}}
		/>
	);
};

const textarea = () => screen.getByTestId<HTMLTextAreaElement>('code');
const frame = () => textarea().parentElement!;
const currentNumber = (container: HTMLElement) => container.querySelector('[data-current]')!.textContent;
const gutterNumbers = (text: string) => screen.getByText(text, { collapseWhitespace: false });

const moveCursor = (start: number, end = start) =>
	act(() => {
		textarea().setSelectionRange(start, end);
		document.dispatchEvent(new Event('selectionchange'));
	});

describe('CodeTextArea', () => {
	it('hands testId to the textarea and turns wrapping and spell checking off', () => {
		render(<CodeTextArea value={TEXT} testId="code" readOnly />);
		expect(textarea().getAttribute('wrap')).toBe('off');
		expect(textarea().getAttribute('spellcheck')).toBe('false');
		expect(textarea().value).toBe(TEXT);
	});

	it('numbers every line of a value set from outside, and shows the start until it is first focused', () => {
		const { rerender, container } = render(<CodeTextArea value="one" testId="code" />);
		expect(screen.getByText(position(1, 1, 1))).toBeTruthy();
		rerender(<CodeTextArea value={TEXT} testId="code" />);
		expect(screen.getByText(position(1, 1, 5))).toBeTruthy();
		expect(gutterNumbers('1\n2\n3\n4\n5')).toBeTruthy();
		expect(currentNumber(container)).toBe('1');
	});

	it('hands typed text to onChange and follows the cursor through it', () => {
		const onChange = vi.fn();
		const { container } = render(<Editable onChange={onChange} />);
		act(() => textarea().focus());
		const at = TEXT.indexOf('"name"');
		const typed = `${TEXT.slice(0, at)}ab\n${TEXT.slice(at)}`;
		// fireEvent copies the selection keys onto the textarea after the value, which leaves the caret mid-text.
		fireEvent.change(textarea(), { target: { value: typed, selectionStart: at + 3, selectionEnd: at + 3 } });
		expect(onChange).toHaveBeenCalledWith(typed);
		expect(screen.getByText(position(4, 1, 6))).toBeTruthy();
		expect(currentNumber(container)).toBe('4');
	});

	it('follows a click', () => {
		render(<Editable />);
		act(() => textarea().focus());
		textarea().setSelectionRange(5, 5);
		fireEvent.mouseUp(textarea());
		expect(screen.getByText(position(2, 4, 5))).toBeTruthy();
	});

	it('follows arrow keys and selections while focused, at the end the selection moves', () => {
		render(<Editable />);
		act(() => textarea().focus());
		moveCursor(TEXT.indexOf('"name"'));
		expect(screen.getByText(position(3, 5, 5))).toBeTruthy();

		act(() => {
			textarea().setSelectionRange(2, TEXT.indexOf('}'), 'backward');
			document.dispatchEvent(new Event('selectionchange'));
		});
		expect(screen.getByText(position(2, 1, 5))).toBeTruthy();
	});

	it("tints the cursor's line and draws its number bright", () => {
		const { container } = render(<Editable />);
		act(() => textarea().focus());
		moveCursor(TEXT.indexOf('"name"'));
		expect(frame().style.getPropertyValue('--ui-code-line')).toBe('2');
		expect(currentNumber(container)).toBe('3');
	});

	it('keeps the cursor where it was left once blurred, through a value set from outside', () => {
		const { rerender } = render(<CodeTextArea value={TEXT} testId="code" />);
		act(() => textarea().focus());
		moveCursor(5);
		act(() => textarea().blur());
		moveCursor(TEXT.length);
		expect(screen.getByText(position(2, 4, 5))).toBeTruthy();
		rerender(<CodeTextArea value={`${TEXT}\n`} testId="code" />);
		expect(screen.getByText(position(2, 4, 6))).toBeTruthy();
	});

	it('moves the numbers and the tint with the scrolled text', () => {
		render(<Editable />);
		textarea().scrollTop = 40;
		fireEvent.scroll(textarea());
		expect(frame().style.getPropertyValue('--ui-code-scroll-top')).toBe('40px');
	});
});
