import { Encounter } from '@sim/raid/encounter';
import { renderWithHost } from '@testing/react';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { EncounterPicker, EncounterPickerView, type EncounterPickerViewProps } from './EncounterPicker';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

const renderView = (props: Partial<EncounterPickerViewProps> = {}) =>
	render(
		<EncounterPickerView
			duration={<span>duration</span>}
			addDebuffs={<span>add debuffs</span>}
			zone={<span>zone</span>}
			areaTypes={<span>area types</span>}
			targetInputs={<span>target inputs</span>}
			onAdvancedClick={() => {}}
			{...props}
		/>,
	);

describe('EncounterPickerView', () => {
	it('lays out the fields in order, and leaves out the ones not given', () => {
		const { getByTestId } = renderView({ execute: <span>execute</span>, allies: <span>allies</span> });
		const root = getByTestId('encounter-picker-root');
		expect([...root.children].map(child => child.textContent)).toEqual([
			'duration',
			'execute',
			'add debuffs',
			'zone',
			'area types',
			'allies',
			'target inputs',
			'settings_tab.encounter.advanced',
		]);
	});

	it('reports the Advanced press', () => {
		const onAdvancedClick = vi.fn();
		const { getByTestId } = renderView({ onAdvancedClick });
		fireEvent.click(getByTestId('advanced-button'));
		expect(onAdvancedClick).toHaveBeenCalledTimes(1);
	});
});

describe('EncounterPicker', () => {
	beforeAll(() => serveDatabase());

	it('shows Apply debuffs only with adds, ticked while the encounter holds no add debuffs, and the debuff delay and jitter only then', async () => {
		const { sim, player } = await makeIndividualSim();
		renderWithHost(<EncounterPicker showExecuteProportion={false} />, { sim, player });
		const apply = () => screen.getByRole('checkbox');
		const timingShown = () => [document.getElementById('encounter-add-debuff-delay'), document.getElementById('encounter-add-debuff-jitter')].map(Boolean);

		expect(sim.encounter.getTargets()).toHaveLength(1);
		expect(screen.queryByRole('checkbox')).toBeNull();
		expect(timingShown()).toEqual([false, false]);

		act(() => sim.encounter.setTargets([Encounter.defaultTargetProto(), Encounter.defaultTargetProto()]));

		expect(apply().getAttribute('aria-checked')).toBe('true');
		expect(timingShown()).toEqual([true, true]);

		act(() => void fireEvent.click(apply()));

		expect(sim.encounter.toProto().holdAddDebuffs).toBe(true);
		expect(apply().getAttribute('aria-checked')).toBe('false');
		expect(timingShown()).toEqual([false, false]);

		act(() => void fireEvent.click(apply()));

		expect(sim.encounter.toProto().holdAddDebuffs).toBe(false);
		expect(timingShown()).toEqual([true, true]);
	});
});
