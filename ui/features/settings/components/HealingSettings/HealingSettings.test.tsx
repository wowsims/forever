import { HealingSimSettings_Assignment as HealingAssignment } from '@generated/proto/ui';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
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
	return { setHealingSettings };
};

describe('HealingSettings', () => {
	beforeEach(() => {
		mocked.sim = null;
	});

	it('changes the raid size without replacing the assignment', async () => {
		const { setHealingSettings } = setUp(HealingAssignment.TankHealing, 10);
		render(<HealingSettings />);
		await userEvent.click(within(screen.getByTestId('healing-group-size')).getByRole('combobox'));
		await userEvent.click(screen.getByRole('option', { name: 'settings_tab.healing_settings.twenty_player' }));
		expect(setHealingSettings).toHaveBeenCalledWith({ assignment: HealingAssignment.TankHealing, groupSize: 20 });
	});

	it('hides the assignment for a dungeon group', () => {
		setUp(HealingAssignment.RaidHealing, 5);
		render(<HealingSettings />);
		expect(screen.queryByTestId('healing-assignment')).toBeNull();
	});
});
