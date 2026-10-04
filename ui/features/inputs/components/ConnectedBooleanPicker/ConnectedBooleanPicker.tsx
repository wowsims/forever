import { BooleanPicker } from '@ui-kit/BooleanPicker';

import { useConnectedField } from '../../hooks/useConnectedField';
import type { BooleanPickerConfig } from './types';

export interface ConnectedBooleanPickerProps<ModObject> {
	modObject: ModObject;
	config: BooleanPickerConfig<ModObject>;
	testId?: string;
}

export const ConnectedBooleanPicker = <ModObject,>({ modObject, config, testId }: ConnectedBooleanPickerProps<ModObject>) => {
	const { value, setValue, hidden, disabled, frameProps } = useConnectedField(modObject, config);
	if (hidden) return null;
	return <BooleanPicker {...frameProps} testId={testId} value={value} onChange={setValue} disabled={disabled} reverse={config.reverse} />;
};
