import { Field, hasFieldFrame } from '@ui-kit/Field';
import { Input } from '@ui-kit/FormControl';
import { useCommitChange } from '@ui-kit/hooks/useCommitChange';
import { useDraft } from '@ui-kit/hooks/useDraft';
import clsx from 'clsx';
import { useCallback, useEffect, useEffectEvent, useMemo, useState } from 'react';

import { numberCodec, numberListCodec, stringCodec } from './codecs';
import type { NumberFieldProps, NumberListFieldProps, SearchFieldProps, StringFieldProps, TextFieldProps } from './types';

const TextField = <V,>({
	codec,
	value,
	onChange,
	disabled,
	commit = 'blur',
	autosize = false,
	clear,
	placeholder,
	autoComplete,
	inputClassName,
	inputSize,
	inputTestId,
	invalid,
	className,
	testId,
	...frameProps
}: TextFieldProps<V>) => {
	const [input, setInput] = useState<HTMLInputElement | null>(null);
	const attachInput = useCallback((element: HTMLElement | null) => setInput(element instanceof HTMLInputElement ? element : null), []);

	const display = codec.format(value);
	const { text, setDraft, clear: dropDraft } = useDraft(display);

	const report = (entered: string) => onChange(codec.parse(entered));

	useCommitChange(commit === 'blur' ? input : null, committed => {
		report(committed.value);
		dropDraft();
	});

	const debounceMs = typeof commit === 'object' ? commit.debounceMs : undefined;
	const reportDebounced = useEffectEvent(report);
	useEffect(() => {
		if (debounceMs === undefined || text === display) return;
		const timer = setTimeout(() => reportDebounced(text), debounceMs);
		return () => clearTimeout(timer);
	}, [debounceMs, text, display]);

	const rooted = hasFieldFrame(frameProps, clear);

	const control = (
		<Input
			type="text"
			ref={attachInput}
			id={frameProps.id}
			className={clsx(!rooted && ['ui-text-field', className], inputClassName)}
			inputSize={inputSize}
			data-testid={inputTestId ?? (rooted ? undefined : testId)}
			disabled={disabled}
			aria-invalid={invalid || undefined}
			placeholder={placeholder}
			autoComplete={autoComplete}
			value={text}
			size={autosize ? Math.max(3, text.length) : undefined}
			onChange={event => {
				// React also reports the native `change` event here, which is the blur/Enter commit above; only a keystroke is a draft.
				if (event.nativeEvent.type !== 'input') return;
				setDraft(event.currentTarget.value);
				if (commit === 'input') report(event.currentTarget.value);
			}}
		/>
	);

	if (!rooted) return control;

	return (
		<Field
			{...frameProps}
			className={['ui-text-field', className]}
			testId={testId}
			disabled={disabled}
			clear={
				clear && {
					show: text.length > 0,
					label: clear.label,
					onClear: () => {
						report('');
						dropDraft();
					},
				}
			}>
			{control}
		</Field>
	);
};

export const NumberField = ({ float, positive, showZeroes, maxDecimalDigits, testId, ...props }: NumberFieldProps) => {
	const codec = useMemo(() => numberCodec({ float, positive, showZeroes, maxDecimalDigits }), [float, positive, showZeroes, maxDecimalDigits]);
	return <TextField {...props} codec={codec} autosize testId={testId ?? 'number-picker-root'} />;
};

export const StringField = ({ testId, ...props }: StringFieldProps) => (
	<TextField {...props} codec={stringCodec} autosize testId={testId ?? 'adaptive-string-picker-root'} />
);

export const NumberListField = ({ inputTestId, ...props }: NumberListFieldProps) => (
	<TextField {...props} codec={numberListCodec} inputTestId={inputTestId ?? 'number-list-picker-input'} />
);

export const SearchField = ({ debounceMs, ...props }: SearchFieldProps) => (
	<TextField {...props} codec={stringCodec} commit={debounceMs ? { debounceMs } : 'input'} />
);
