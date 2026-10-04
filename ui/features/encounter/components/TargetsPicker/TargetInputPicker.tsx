import { ConnectedBooleanPicker } from '@features/inputs/components/ConnectedBooleanPicker';
import { ConnectedChoiceField } from '@features/inputs/components/ConnectedChoiceField';
import { ConnectedNumberField } from '@features/inputs/components/ConnectedTextField';
import type { TargetInput } from '@generated/proto/common';
import type { Encounter } from '@sim/raid/encounter';
import { Field } from '@ui-kit/Field';
import { useMemo } from 'react';

import { targetInputConfig } from './utils/configs';

export interface TargetInputPickerProps {
	encounter: Encounter;
	targetIndex: number;
	inputIndex: number;
	input: TargetInput | undefined;
}

export const TargetInputPicker = ({ encounter, targetIndex, inputIndex, input }: TargetInputPickerProps) => {
	const picker = useMemo(
		() => (input ? targetInputConfig(encounter, targetIndex, inputIndex, input) : { kind: 'none' as const }),
		[encounter, targetIndex, inputIndex, input],
	);

	return (
		<Field testId="target-input-picker-root">
			{picker.kind === 'number' && <ConnectedNumberField modObject={null} config={picker.config} />}
			{picker.kind === 'boolean' && <ConnectedBooleanPicker modObject={null} config={picker.config} />}
			{picker.kind === 'enum' && <ConnectedChoiceField modObject={null} config={picker.config} />}
		</Field>
	);
};
