import { Sim } from '@sim/sim';
import { createSimStore, patchSlice } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { act, fireEvent, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { IterationsPicker } from './IterationsPicker';

vi.mock('../../../tracking/utils', () => ({ trackEvent: () => {} }));

const setup = (iterations: number) => {
	const store = createSimStore();
	patchSlice(store, 'sim', { iterations });
	const sim = {
		store,
		getIterations: () => store.getState().sim.iterations,
		setIterations: (value: number) => patchSlice(store, 'sim', { iterations: value }),
	} as unknown as Sim;
	renderWithHost(<IterationsPicker sim={sim} />, { sim });
	return { store };
};

const pressed = () => screen.getAllByTestId('iterations-preset').map(button => button.getAttribute('aria-pressed'));

describe('IterationsPicker', () => {
	it('offers the default and twice the default, pressing the one that matches the count', () => {
		setup(Sim.DEFAULT_ITERATIONS);

		expect(pressed()).toEqual(['true', 'false']);
	});

	it('presses neither preset for any other count', () => {
		setup(3000);

		expect(pressed()).toEqual(['false', 'false']);
	});

	it('sets the count from a preset and follows it', () => {
		const { store } = setup(3000);

		fireEvent.click(screen.getAllByTestId('iterations-preset')[1]);

		expect(store.getState().sim.iterations).toBe(2 * Sim.DEFAULT_ITERATIONS);
		expect(pressed()).toEqual(['false', 'true']);
	});

	it('follows a count changed elsewhere', () => {
		const { store } = setup(3000);

		act(() => patchSlice(store, 'sim', { iterations: Sim.DEFAULT_ITERATIONS }));

		expect(pressed()).toEqual(['true', 'false']);
	});
});
