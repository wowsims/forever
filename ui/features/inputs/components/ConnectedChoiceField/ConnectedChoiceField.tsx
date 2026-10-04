import { type ChoiceEntry, ChoiceField } from '@ui-kit/ChoiceField';
import type { EnumValueConfig } from '@ui-kit/MultiComboBox';
import { useMemo } from 'react';

import { useConnectedField } from '../../hooks/useConnectedField';
import type { ConnectedChoiceFieldProps } from './types';

const NO_VALUES: ReadonlyArray<EnumValueConfig> = [];

export const ConnectedChoiceField = <M, T, V = T>({ modObject, config, options, className, testId, ...presentation }: ConnectedChoiceFieldProps<M, T, V>) => {
	const { value, setValue, hidden, disabled, frameProps } = useConnectedField<M, T, V>(modObject, config);
	// A config that lists `values` is the number-valued `enum` kind, so its value is its `V`.
	const values = config.values ?? NO_VALUES;
	const mapped = useMemo(() => values.map(entry => ({ value: entry.value as V, label: entry.name, tooltip: entry.tooltip })), [values]);
	if (hidden) return null;
	const choices: ReadonlyArray<ChoiceEntry<V>> = options ?? mapped;
	return (
		<ChoiceField<V>
			{...frameProps}
			{...presentation}
			className={[frameProps.className, className]}
			testId={testId}
			value={value}
			onChange={setValue}
			disabled={disabled}
			options={choices}
		/>
	);
};
