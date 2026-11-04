import { type ListItemsConfig, useListItems } from '@features/inputs/hooks/useListItems';
import { keyFor } from '@features/inputs/list_ids';
import { TargetInput } from '@generated/proto/common';
import type { Encounter } from '@sim/raid/encounter';
import { subscribeEncounterField } from '@sim/state/subscriptions';
import { Tree } from '@ui-kit/Tree';
import { useMemo } from 'react';

import { TargetInputPicker } from './TargetInputPicker';

export interface TargetInputsPickerProps {
	encounter: Encounter;
	targetIndex: number;
}

/**
 * The inputs a target preset declares, one picker per input. The preset fixes the set, so the list
 * offers no add, remove, copy or drag, and renders nothing while the target declares no input.
 */
export const TargetInputsPicker = ({ encounter, targetIndex }: TargetInputsPickerProps) => {
	const config = useMemo(
		(): ListItemsConfig<Encounter, TargetInput> => ({
			storeSubscribe: (subject: Encounter) => subscribeEncounterField(subject, 'targets'),
			getValue: (subject: Encounter) => subject.getTargets()[targetIndex].targetInputs.slice(),
			setValue: (subject: Encounter, newValue: Array<TargetInput>) =>
				subject.modifyTarget(targetIndex, target => {
					target.targetInputs = newValue;
				}),
			newItem: () => TargetInput.create(),
		}),
		[targetIndex],
	);
	const { items, itemConfig, hidden, disabled } = useListItems(encounter, config);
	if (hidden || !items.length) return null;

	return (
		// A group of its own: inside a target row the list shares the targets list's drag, and no target may drop in here.
		<Tree.Root group="target-inputs" density="compact" disabled={disabled} testId="target-inputs-picker">
			<Tree.Items>
				{items.map((input, inputIndex) => {
					const key = String(keyFor(input, inputIndex));
					return (
						<Tree.Item key={key} id={key} testId="target-inputs-picker-item">
							<Tree.Item.Body>
								<TargetInputPicker
									encounter={encounter}
									targetIndex={targetIndex}
									inputIndex={inputIndex}
									input={itemConfig(input, inputIndex).getValue(encounter)}
								/>
							</Tree.Item.Body>
						</Tree.Item>
					);
				})}
			</Tree.Items>
		</Tree.Root>
	);
};
