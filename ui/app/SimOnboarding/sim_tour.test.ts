// @vitest-environment node

import { applicableSteps } from '@features/onboarding/model/tour_steps';
import { describe, expect, it } from 'vitest';

import en from '../../../assets/locales/en/translation.json';
import { SIM_TOUR_STEPS } from './sim_tour';

describe('SIM_TOUR_STEPS', () => {
	it('has a title and a body for every step', () => {
		const steps = en.onboarding.steps as Record<string, { title?: string; body?: string }>;
		for (const step of SIM_TOUR_STEPS) {
			expect(steps[step.id]?.title, step.id).toBeTruthy();
			expect(steps[step.id]?.body, step.id).toBeTruthy();
		}
		expect(Object.keys(steps).sort()).toEqual(SIM_TOUR_STEPS.map(step => step.id).sort());
	});

	it('opens and closes with a centred card, and tours a gear planner without the sim', () => {
		expect(SIM_TOUR_STEPS[0].target).toBeUndefined();
		expect(SIM_TOUR_STEPS[SIM_TOUR_STEPS.length - 1].target).toBeUndefined();
		const planner = applicableSteps(SIM_TOUR_STEPS, { canSimulate: false, hasResult: false }).map(step => step.id);
		expect(planner).not.toContain('simulate');
		expect(planner).not.toContain('rotation_tab');
		expect(planner).not.toContain('encounter');
		expect(planner).toContain('gear_list');
	});
});
