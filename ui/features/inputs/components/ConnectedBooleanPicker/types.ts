import type { BooleanPickerProps } from '@ui-kit/BooleanPicker';

import type { InputConfig } from '../../input';

export type BooleanPickerConfig<ModObject> = InputConfig<ModObject, boolean> & Pick<BooleanPickerProps, 'reverse'> & { id: string };
