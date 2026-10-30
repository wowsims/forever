import i18n from '@i18n/config';
import { cssVars } from '@ui-kit/utils/css';
import clsx, { type ClassValue } from 'clsx';
import type { ComponentPropsWithoutRef } from 'react';
import { useCallback, useLayoutEffect, useMemo, useRef, useState } from 'react';

type OwnProps = 'className' | 'value' | 'defaultValue' | 'onChange' | 'wrap' | 'onScroll' | 'onFocus' | 'onSelect';

export interface CodeTextAreaProps extends Omit<ComponentPropsWithoutRef<'textarea'>, OwnProps> {
	value: string;
	onChange?: (value: string) => void;
	/** Sizes the frame; the text fills it. */
	className?: ClassValue;
	testId?: string;
}

const countLines = (text: string) => {
	let lines = 1;
	for (let i = text.indexOf('\n'); i !== -1; i = text.indexOf('\n', i + 1)) lines++;
	return lines;
};

/**
 * A monospace text box for data files: line numbers beside the text, no wrapping, the cursor's line tinted and its
 * position in the bottom-right corner.
 */
export const CodeTextArea = ({ value, onChange, className, testId, ...props }: CodeTextAreaProps) => {
	const frame = useRef<HTMLDivElement>(null);
	const input = useRef<HTMLTextAreaElement>(null);

	// The cursor offset, read only while the box has focus: setting a textarea's value moves its selection to the end,
	// so an export would open on its last line, and a value written while the box is blurred would move the cursor from
	// where it was left.
	const [cursor, setCursor] = useState(0);
	const readCursor = useCallback(() => {
		const element = input.current;
		if (!element) return;
		setCursor(element.selectionDirection === 'backward' ? element.selectionStart : element.selectionEnd);
	}, []);

	// The scroll offset and the scrollbar sizes go straight to CSS variables, so scrolling renders nothing.
	const place = useCallback(() => {
		const element = input.current;
		const style = frame.current?.style;
		if (!element || !style) return;
		style.setProperty('--ui-code-scroll-top', `${element.scrollTop}px`);
		style.setProperty('--ui-code-scrollbar-x', `${element.offsetHeight - element.clientHeight}px`);
		style.setProperty('--ui-code-scrollbar-y', `${element.offsetWidth - element.clientWidth}px`);
	}, []);

	// A value written from outside fires no input event, so the scrollbars are read again after each one, and the cursor
	// too while the box has focus.
	useLayoutEffect(() => {
		if (document.activeElement === input.current) readCursor();
		place();
	}, [value, readCursor, place]);

	useLayoutEffect(() => {
		const observer = new ResizeObserver(place);
		observer.observe(input.current!);
		return () => observer.disconnect();
	}, [place]);

	const lineCount = useMemo(() => countLines(value), [value]);
	const { line, column } = useMemo(() => {
		const offset = Math.min(cursor, value.length);
		const before = value.slice(0, offset);
		return { line: countLines(before), column: offset - before.lastIndexOf('\n') };
	}, [value, cursor]);
	// The numbers follow the line count only, and the cursor's number is drawn over its dim one, so a cursor move leaves a
	// long file's gutter text alone.
	const numbers = useMemo(() => Array.from({ length: lineCount }, (_, i) => i + 1).join('\n'), [lineCount]);

	return (
		<div ref={frame} className={clsx('ui-code-text-area', className)} style={cssVars({ '--ui-code-line': String(line - 1) })}>
			<div className="ui-code-text-area-gutter" aria-hidden>
				<div className="ui-code-text-area-numbers">{numbers}</div>
				<div data-current="">{line}</div>
			</div>
			{/* After the gutter, so the tint lies over the cursor's number too. */}
			<div className="ui-code-text-area-line" />
			<textarea
				spellCheck={false}
				{...props}
				ref={input}
				value={value}
				wrap="off"
				data-testid={testId}
				className="ui-code-text-area-input"
				onChange={event => {
					onChange?.(event.currentTarget.value);
					readCursor();
				}}
				onFocus={readCursor}
				onSelect={readCursor}
				onScroll={place}
			/>
			<div className="ui-code-text-area-position" aria-hidden>
				{i18n.t('common.code_text_area.position', { line, column, count: lineCount })}
			</div>
		</div>
	);
};
