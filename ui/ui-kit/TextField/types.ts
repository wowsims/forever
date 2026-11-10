import type { ControlledProps, FieldFrameProps } from '@ui-kit/Field';

/** How a value becomes the field's text and back. */
export interface TextCodec<V> {
	format(value: V): string;
	parse(text: string): V;
}

/**
 * When an edit reaches `onChange`: `blur` on the native `change` event (blur after an edit, or
 * Enter), `input` on every keystroke, `{ debounceMs }` once typing pauses that long.
 */
export type TextFieldCommit = 'blur' | 'input' | { debounceMs: number };

/**
 * A field renders a `Field` root only for a label, description, layout or clear button; without
 * any of them it is the bare `<input>`. `className` and `testId` land on the outermost element
 * either way, and on a bare field `inputTestId` wins over `testId`.
 */
export interface TextFieldProps<V> extends FieldFrameProps, ControlledProps<V> {
	codec: TextCodec<V>;
	commit?: TextFieldCommit;
	/** Sizes the input to its text, at least 3 characters wide. */
	autosize?: boolean;
	/** A clear button inside the input while it holds text; it commits empty text. */
	clear?: { label: string };
	placeholder?: string;
	autoComplete?: 'on' | 'off';
	inputClassName?: string;
	inputTestId?: string;
}

export interface NumberCodecOptions {
	/** Formats with 2 to `maxDecimalDigits` fraction digits and reads the text back as a float. */
	float?: boolean;
	maxDecimalDigits?: number;
	/** Commits the absolute value of what was typed. */
	positive?: boolean;
	/** `false` shows zero as an empty field. */
	showZeroes?: boolean;
}

export type NumberFieldProps = Omit<TextFieldProps<number>, 'codec' | 'commit' | 'autosize'> & NumberCodecOptions;

export type StringFieldProps = Omit<TextFieldProps<string>, 'codec' | 'commit' | 'autosize'>;

export type NumberListFieldProps = Omit<TextFieldProps<Array<number>>, 'codec' | 'commit' | 'autosize'>;

export type SearchFieldProps = Omit<TextFieldProps<string>, 'codec' | 'commit' | 'autosize'> & {
	/** Commits once typing pauses this long; without it every keystroke commits. */
	debounceMs?: number;
};
