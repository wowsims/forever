import type { MultiComboBoxProps } from '@ui-kit/MultiComboBox';

import type { InputConfig } from '../../input';

export type MultiComboBoxConfig<M> = InputConfig<M, Array<number>> &
	Pick<MultiComboBoxProps, 'values' | 'placeholder' | 'emptyText' | 'removeLabel' | 'openLabel'> & { id: string };
