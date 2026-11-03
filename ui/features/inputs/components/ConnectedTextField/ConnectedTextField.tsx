import type { ControlSize } from '@ui-kit/Field';
import { NumberField, NumberListField, StringField } from '@ui-kit/TextField';
import clsx from 'clsx';

import { useConnectedField } from '../../hooks/useConnectedField';
import type { NumberFieldConfig, NumberListFieldConfig, StringFieldConfig } from './types';

export interface ConnectedNumberFieldProps<ModObject> {
	modObject: ModObject;
	config: NumberFieldConfig<ModObject>;
	inputClassName?: string;
	size?: ControlSize;
	testId?: string;
}

export const ConnectedNumberField = <ModObject,>({ modObject, config, inputClassName, size, testId }: ConnectedNumberFieldProps<ModObject>) => {
	const { value, setValue, hidden, disabled, frameProps } = useConnectedField(modObject, config);
	if (hidden) return null;
	return (
		<NumberField
			{...frameProps}
			testId={testId}
			value={value}
			onChange={setValue}
			disabled={disabled}
			float={config.float}
			positive={config.positive}
			showZeroes={config.showZeroes}
			maxDecimalDigits={config.maxDecimalDigits}
			inputClassName={clsx(config.inputClassName, inputClassName)}
			size={size}
		/>
	);
};

export interface ConnectedStringFieldProps<M> {
	modObject: M;
	config: StringFieldConfig<M>;
	testId?: string;
}

export const ConnectedStringField = <M,>({ modObject, config, testId }: ConnectedStringFieldProps<M>) => {
	const { value, setValue, hidden, disabled, frameProps } = useConnectedField(modObject, config);
	if (hidden) return null;
	return <StringField {...frameProps} testId={testId} value={value} onChange={setValue} disabled={disabled} />;
};

export interface ConnectedNumberListFieldProps<M> {
	modObject: M;
	config: NumberListFieldConfig<M>;
	testId?: string;
}

export const ConnectedNumberListField = <M,>({ modObject, config, testId }: ConnectedNumberListFieldProps<M>) => {
	const { value, setValue, hidden, disabled, frameProps } = useConnectedField(modObject, config);
	if (hidden) return null;
	return <NumberListField {...frameProps} testId={testId} value={value} onChange={setValue} disabled={disabled} placeholder={config.placeholder} />;
};
