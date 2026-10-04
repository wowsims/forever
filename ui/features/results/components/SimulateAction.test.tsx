import { renderWithHost } from '@testing/react';
import { act, fireEvent } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { SimulateAction } from './SimulateAction';

vi.mock('../../../tracking/utils', () => ({ trackEvent: () => {} }));

const setup = () => {
	let finishRun = () => {};
	const runIndividualSim = vi.fn(() => new Promise<void>(resolve => (finishRun = resolve)));
	const abort = vi.fn(async () => {});
	const view = renderWithHost(<SimulateAction />, {
		sim: { runs: { isRunning: () => false, abort }, getIterations: () => 1000 },
		resultsViewer: { addAbortButton: () => {}, removeAbortButton: () => {}, setProgress: () => {}, hideAll: () => {} },
		runIndividualSim,
	});
	return { view, runIndividualSim, abort, finishRun: () => act(async () => finishRun()) };
};

describe('SimulateAction', () => {
	it('turns into Stop while its run is in flight, and back once the run settles', async () => {
		const { view, runIndividualSim, finishRun } = setup();

		fireEvent.click(view.getByTestId('dps-action'));

		expect(runIndividualSim).toHaveBeenCalledTimes(1);
		expect(view.queryByTestId('dps-action')).toBeNull();
		expect(view.getByTestId('sim-stop-action').textContent).toContain('sidebar.results.stop');
		expect(view.getByTestId('sim-stop-action').getAttribute('data-variant')).toBe('cancel');

		await finishRun();
		expect(view.getByTestId('dps-action')).not.toBeNull();
	});

	it('disables and relabels Stop before it asks the run to abort', async () => {
		const { view, abort } = setup();
		fireEvent.click(view.getByTestId('dps-action'));
		const seen: Array<{ label: string; disabled: boolean }> = [];
		abort.mockImplementation(async () => {
			const button = view.getByTestId('sim-stop-action') as HTMLButtonElement;
			seen.push({ label: button.textContent!.trim(), disabled: button.disabled });
		});

		await act(async () => fireEvent.click(view.getByTestId('sim-stop-action')));

		expect(seen).toEqual([{ label: 'sidebar.results.stopping', disabled: true }]);
	});
});
