import type { NumberFieldProps, NumberListFieldProps } from '@ui-kit/TextField';

import type { InputConfig } from '../../input';

export type NumberFieldConfig<ModObject> = InputConfig<ModObject, number> &
	Pick<NumberFieldProps, 'float' | 'maxDecimalDigits' | 'positive' | 'showZeroes' | 'inputClassName'> & { id: string };

export type StringFieldConfig<M> = InputConfig<M, string> & { id: string };

export type NumberListFieldConfig<M> = InputConfig<M, Array<number>> & Pick<NumberListFieldProps, 'placeholder'> & { id: string };
