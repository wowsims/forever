import { memoryEnv } from '@testing/sim/memory_env';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { readTourOutcome } from '../../model/tour_steps';
import { OnboardingOutlet, OnboardingProvider, type TourRun, useOnboardingRestart } from './OnboardingProvider';

const RestartButton = () => {
	const restart = useOnboardingRestart();
	return (
		<button type="button" data-testid="restart" onClick={restart}>
			Tour
		</button>
	);
};

const Tour = ({ run }: { run: TourRun }) => (
	<div data-testid="running-tour" data-trigger={run.trigger}>
		<button type="button" data-testid="end" onClick={() => run.end('done')}>
			End
		</button>
	</div>
);

// The restart sits outside the part of the page the tour is drawn in, as the options dialog does.
const page = (ready: boolean, storage: ReturnType<typeof memoryEnv>['storage']) => (
	<OnboardingProvider storage={storage} storageKey="tour">
		<RestartButton />
		<main>
			<OnboardingOutlet ready={ready} renderTour={run => <Tour run={run} />} />
		</main>
	</OnboardingProvider>
);

const mount = (ready: boolean, storage = memoryEnv().storage) => ({ storage, view: render(page(ready, storage)) });

describe('OnboardingProvider', () => {
	beforeEach(() => vi.useFakeTimers());
	afterEach(() => vi.useRealTimers());

	it('opens once the page is ready and has settled, and remembers how it ended', () => {
		const { storage, view } = mount(false);
		act(() => vi.advanceTimersByTime(5000));
		expect(screen.queryByTestId('running-tour')).toBeNull();
		view.rerender(page(true, storage));
		act(() => vi.advanceTimersByTime(100));
		expect(screen.queryByTestId('running-tour')).toBeNull();
		act(() => vi.advanceTimersByTime(1000));
		expect(screen.getByTestId('running-tour').getAttribute('data-trigger')).toBe('auto');
		fireEvent.click(screen.getByTestId('end'));
		expect(screen.queryByTestId('running-tour')).toBeNull();
		expect(readTourOutcome(storage, 'tour')).toBe('done');
	});

	it('does not open by itself once a tour has ended, but opens when asked and gives the focus back after', () => {
		const storage = memoryEnv().storage;
		storage.setItem('tour', 'skipped');
		mount(true, storage);
		act(() => vi.advanceTimersByTime(5000));
		expect(screen.queryByTestId('running-tour')).toBeNull();
		const restart = screen.getByTestId('restart');
		restart.focus();
		fireEvent.click(restart);
		expect(screen.getByTestId('running-tour').getAttribute('data-trigger')).toBe('restart');
		screen.getByTestId('end').focus();
		fireEvent.click(screen.getByTestId('end'));
		expect(document.activeElement).toBe(restart);
	});
});
