import { HealingSimSettings, HealingSimSettings_Assignment as HealingAssignment } from '@generated/proto/ui';
import { describe, expect, it, vi } from 'vitest';

import { HealingAssignmentInput, HealingGroupSize } from './healing_inputs';

const fakePlayer = (assignment: HealingAssignment, groupSize: number) => {
	let settings = HealingSimSettings.create({ assignment, groupSize });
	const setHealingSettings = vi.fn((next: HealingSimSettings) => {
		settings = next;
	});
	const player = { getRaid: () => ({ getHealingSettings: () => settings, setHealingSettings }) } as any;
	return { player, setHealingSettings };
};

describe('HealingGroupSize', () => {
	it('changes the raid size without replacing the assignment', () => {
		const { player, setHealingSettings } = fakePlayer(HealingAssignment.TankHealing, 10);
		HealingGroupSize.setValue(player, 20);
		expect(setHealingSettings).toHaveBeenCalledWith({ assignment: HealingAssignment.TankHealing, groupSize: 20 });
		expect(HealingGroupSize.getValue(player)).toBe(20);
	});
});

describe('HealingAssignmentInput', () => {
	it('changes the assignment without replacing the raid size', () => {
		const { player, setHealingSettings } = fakePlayer(HealingAssignment.TankHealing, 20);
		HealingAssignmentInput.setValue(player, HealingAssignment.RaidHealing);
		expect(setHealingSettings).toHaveBeenCalledWith({ assignment: HealingAssignment.RaidHealing, groupSize: 20 });
		expect(HealingAssignmentInput.getValue(player)).toBe(HealingAssignment.RaidHealing);
	});

	it('hides the assignment for a dungeon group', () => {
		expect(HealingAssignmentInput.showWhen(fakePlayer(HealingAssignment.RaidHealing, 5).player)).toBe(false);
		expect(HealingAssignmentInput.showWhen(fakePlayer(HealingAssignment.RaidHealing, 10).player)).toBe(true);
	});
});
