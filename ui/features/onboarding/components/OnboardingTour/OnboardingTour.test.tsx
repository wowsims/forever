import i18n from '@i18n/config';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import type { SimulationState, TourStep } from '../../model/tour_steps';
import { OnboardingTour, type TourEnd } from './OnboardingTour';

// happy-dom lays nothing out, and the tour passes over an element with an empty box.
const page = () => {
	document.body.insertAdjacentHTML('beforeend', '<div id="page"><button id="gear">Gear</button><button id="simulate">Simulate</button></div>');
	for (const element of document.querySelectorAll('#page button')) element.getBoundingClientRect = () => new DOMRect(40, 40, 120, 30);
};

const STEPS: ReadonlyArray<TourStep> = [
	{ id: 'welcome' },
	{ id: 'gear', target: '#gear', tab: 'gear-tab' },
	{ id: 'missing', target: '#not-on-this-page' },
	{ id: 'simulate', target: '#simulate', waitFor: 'simulation' },
	{ id: 'finish' },
];

const text = (step: TourStep) => ({ title: `title ${step.id}`, body: `body ${step.id}` });

const tour = (props: { simulation?: SimulationState; prepare?: (step: TourStep) => void } = {}) => {
	const onEnd = vi.fn<(end: TourEnd<TourStep>) => void>();
	const view = render(<OnboardingTour steps={STEPS} text={text} onEnd={onEnd} {...props} />);
	return { onEnd, view };
};

const stepId = () => screen.getByTestId('tour').getAttribute('data-step');
// In short steps, so the effects a step's state change starts run before the clock moves on.
const settle = (ms = 200) => {
	for (let elapsed = 0; elapsed < ms; elapsed += 50) act(() => vi.advanceTimersByTime(50));
};
const key = (name: string) => fireEvent.keyDown(screen.getByTestId('tour-card'), { key: name });

describe('OnboardingTour', () => {
	beforeEach(() => {
		vi.useFakeTimers();
		page();
	});
	afterEach(() => {
		vi.useRealTimers();
		document.getElementById('page')?.remove();
	});

	it('rings the part of the page each step explains, opening what it needs first', () => {
		const prepare = vi.fn();
		tour({ prepare });
		expect(stepId()).toBe('welcome');
		expect(screen.queryByTestId('tour-ring')).toBeNull();
		fireEvent.click(screen.getByTestId('tour-card-next'));
		settle();
		expect(stepId()).toBe('gear');
		expect(prepare).toHaveBeenLastCalledWith(STEPS[1]);
		expect(screen.getByTestId('tour-ring')).toBeTruthy();
		expect(screen.getByTestId('tour-card-title').textContent).toBe('title gear');
	});

	it('passes over a step whose part the page does not show, in the direction of travel', () => {
		tour();
		key('ArrowRight');
		settle();
		key('ArrowRight');
		settle(2000);
		expect(stepId()).toBe('simulate');
		key('ArrowLeft');
		settle(2000);
		expect(stepId()).toBe('gear');
	});

	it('presses what a step needs open first, and soon passes over a part the page may not have', () => {
		const pressed = vi.fn();
		document.getElementById('gear')!.addEventListener('click', pressed);
		const steps: ReadonlyArray<TourStep> = [
			{ id: 'welcome' },
			{ id: 'opened', press: '#gear', target: '#simulate' },
			{ id: 'maybe', target: '#not-on-this-page', optional: true },
			{ id: 'finish' },
		];
		render(<OnboardingTour steps={steps} text={text} onEnd={vi.fn()} />);
		key('ArrowRight');
		settle();
		expect(stepId()).toBe('opened');
		expect(pressed).toHaveBeenCalledOnce();
		key('ArrowRight');
		settle(600);
		expect(stepId()).toBe('finish');
	});

	it('passes over at once an optional step on the open tab whose part the page lacks', () => {
		const steps: ReadonlyArray<TourStep> = [
			{ id: 'welcome' },
			{ id: 'gear', tab: 'results', target: '#gear' },
			{ id: 'absent', tab: 'results', target: '#not-on-this-page', optional: true },
			{ id: 'simulate', tab: 'results', target: '#simulate' },
			{ id: 'finish' },
		];
		render(<OnboardingTour steps={steps} text={text} onEnd={vi.fn()} />);
		key('ArrowRight');
		settle();
		key('ArrowRight');
		expect(stepId()).toBe('simulate');
		settle();
		expect(screen.getByTestId('tour-card-progress').textContent).toContain(i18n.t('onboarding.progress', { position: 2, total: 2 }));
	});

	it('waits on the Simulate step until a run started there finishes, letting the page through meanwhile', () => {
		const { view, onEnd } = tour({ simulation: 'idle' });
		const rerender = (simulation: SimulationState) => view.rerender(<OnboardingTour steps={STEPS} text={text} onEnd={onEnd} simulation={simulation} />);
		for (let i = 0; i < 2; i++) {
			key('ArrowRight');
			settle(2000);
		}
		expect(stepId()).toBe('simulate');
		expect(screen.getByTestId('tour').hasAttribute('data-waiting')).toBe(true);
		// The arrow key does not pass a waiting step; a result that was there before the run does not either.
		key('ArrowRight');
		rerender('done');
		expect(stepId()).toBe('simulate');
		rerender('running');
		rerender('done');
		expect(stepId()).toBe('finish');
		// Back on the step, the run it already counted does not move it on again.
		settle();
		key('ArrowLeft');
		settle(2000);
		expect(stepId()).toBe('simulate');
		rerender('done');
		expect(stepId()).toBe('simulate');
	});

	it('goes back from a waiting step on the left arrow pressed on the page, but not in a control that uses the arrows', () => {
		tour({ simulation: 'idle' });
		for (let i = 0; i < 2; i++) {
			key('ArrowRight');
			settle(2000);
		}
		expect(stepId()).toBe('simulate');

		const field = document.body.appendChild(document.createElement('input'));
		fireEvent.keyDown(field, { key: 'ArrowLeft' });
		expect(stepId()).toBe('simulate');
		field.remove();

		fireEvent.keyDown(document.body, { key: 'ArrowRight' });
		expect(stepId()).toBe('simulate');
		fireEvent.keyDown(document.body, { key: 'ArrowLeft' });
		settle(2000);
		expect(stepId()).toBe('gear');
	});

	it('stays where it was when the step it is on drops out of the tour', () => {
		const { view, onEnd } = tour();
		key('ArrowRight');
		settle();
		expect(stepId()).toBe('gear');
		view.rerender(<OnboardingTour steps={STEPS.filter(step => step.id !== 'gear')} text={text} onEnd={onEnd} />);
		expect(stepId()).toBe('missing');
	});

	it('reports how and where it ended', () => {
		const { onEnd } = tour();
		key('ArrowRight');
		settle();
		key('Escape');
		expect(onEnd).toHaveBeenCalledWith({ outcome: 'skipped', step: STEPS[1], index: 1, total: STEPS.length });
	});

	it('finishes as done from the last card', () => {
		const { onEnd } = tour({ simulation: 'idle' });
		for (let i = 0; i < 2; i++) {
			key('ArrowRight');
			settle(2000);
		}
		fireEvent.click(screen.getByTestId('tour-card-next'));
		settle();
		expect(stepId()).toBe('finish');
		fireEvent.click(screen.getByTestId('tour-card-next'));
		expect(onEnd).toHaveBeenCalledWith(expect.objectContaining({ outcome: 'done' }));
	});
});
