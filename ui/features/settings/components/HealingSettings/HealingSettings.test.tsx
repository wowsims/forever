import { HealingSimSettings_Assignment as HealingAssignment } from '@generated/proto/ui';
import { render, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createStore } from 'zustand/vanilla';

import { HealingSettings } from './HealingSettings';

const mocked = vi.hoisted(() => ({ sim: null as any }));

vi.mock('@sim/context/SimHostContext', () => ({ useSim: () => mocked.sim }));

type StoreState = {
	raid: {
		healingSettings: { assignment: HealingAssignment; groupSize: number };
		targetDummies: number;
	};
};

const setUp = (assignment: HealingAssignment, groupSize: number) => {
	const store = createStore<StoreState>(() => ({ raid: { healingSettings: { assignment, groupSize }, targetDummies: 99 } }));
	const setTargetDummies = vi.fn((targetDummies: number) => store.setState(state => ({ raid: { ...state.raid, targetDummies } })));
	const setHealingSettings = vi.fn((healingSettings: StoreState['raid']['healingSettings']) =>
		store.setState(state => ({ raid: { ...state.raid, healingSettings } })),
	);
	mocked.sim = { store, raid: { setTargetDummies, setHealingSettings } };
	return { setTargetDummies };
};

describe('HealingSettings target count', () => {
	beforeEach(() => {
		mocked.sim = null;
	});

	it.each([
		[HealingAssignment.TankHealing, 5, 4],
		[HealingAssignment.TankHealing, 10, 1],
		[HealingAssignment.TankHealing, 20, 2],
		[HealingAssignment.RaidHealing, 10, 9],
		[HealingAssignment.RaidHealing, 20, 19],
	] as const)('sets target count for assignment %i and group size %i to %i', async (assignment, groupSize, expected) => {
		const { setTargetDummies } = setUp(assignment, groupSize);
		render(<HealingSettings />);

		await waitFor(() => expect(setTargetDummies).toHaveBeenCalledWith(expected));
	});
});
