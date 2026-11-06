import { InputType, TargetInput } from '@generated/proto/common';
import type { Encounter } from '@sim/raid/encounter';

export const isTargetInputShown = (inputs: ReadonlyArray<TargetInput>, input: TargetInput): boolean =>
	!input.shownBy || inputs.some(other => other.label === input.shownBy && other.inputType === InputType.Bool && other.boolValue);

export const repairTargetInputs = (encounter: Encounter): void => {
	const presets = encounter.sim.db.getAllPresetTargets();
	const preset = presets.find(entry => encounter.primaryTarget.id === entry.target?.id);
	const targetInputs = preset?.target?.targetInputs || [];
	const current = encounter.primaryTarget.targetInputs;
	if (current.length === targetInputs.length && current.every((input, index) => input.label === targetInputs[index].label)) return;
	encounter.modifyTarget(0, target => {
		target.targetInputs = targetInputs.map(input => TargetInput.clone(input));
	});
};
