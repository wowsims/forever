import { memoryEnv } from '@testing/sim/memory_env';
import { describe, expect, it } from 'vitest';

import { applicableSteps, cardKindAt, readTourOutcome, stepCount, type TourStep, writeTourOutcome } from './tour_steps';

interface Context {
	canSimulate: boolean;
}

const STEPS: ReadonlyArray<TourStep<Context>> = [
	{ id: 'welcome' },
	{ id: 'gear', target: '#gear' },
	{ id: 'simulate', target: '#simulate', applies: context => context.canSimulate },
	{ id: 'talents', target: '#talents' },
	{ id: 'finish' },
];

describe('applicableSteps', () => {
	it('leaves out the steps the page has nothing to show for', () => {
		expect(applicableSteps(STEPS, { canSimulate: true }).map(step => step.id)).toEqual(['welcome', 'gear', 'simulate', 'talents', 'finish']);
		expect(applicableSteps(STEPS, { canSimulate: false }).map(step => step.id)).toEqual(['welcome', 'gear', 'talents', 'finish']);
	});
});

describe('cardKindAt and stepCount', () => {
	it('opens with the welcome, closes with the last card, and counts only the steps between', () => {
		expect(STEPS.map((_, index) => cardKindAt(STEPS, index))).toEqual(['welcome', 'step', 'step', 'step', 'finish']);
		expect(stepCount(STEPS, 1)).toEqual({ position: 1, total: 3 });
		expect(stepCount(STEPS, 3)).toEqual({ position: 3, total: 3 });
	});

	it('counts a centred card in the middle of the tour as a step', () => {
		const steps: ReadonlyArray<TourStep> = [{ id: 'welcome' }, { id: 'note' }, { id: 'gear', target: '#gear' }, { id: 'finish' }];
		expect(cardKindAt(steps, 1)).toBe('step');
		expect(stepCount(steps, 2)).toEqual({ position: 2, total: 2 });
	});
});

describe('readTourOutcome and writeTourOutcome', () => {
	it('remembers how the tour ended and ignores anything else under the key', () => {
		const { storage } = memoryEnv();
		expect(readTourOutcome(storage, 'tour')).toBeNull();
		writeTourOutcome(storage, 'tour', 'skipped');
		expect(readTourOutcome(storage, 'tour')).toBe('skipped');
		storage.setItem('tour', 'something else');
		expect(readTourOutcome(storage, 'tour')).toBeNull();
	});

	it('counts the tour as seen when storage throws, so it does not open on every page', () => {
		const blocked = {
			getItem: () => {
				throw new Error('blocked');
			},
			setItem: () => {
				throw new Error('blocked');
			},
			removeItem: () => {},
		};
		expect(readTourOutcome(blocked, 'tour')).toBe('skipped');
		expect(() => writeTourOutcome(blocked, 'tour', 'done')).not.toThrow();
	});
});
