import { act, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { HandleProgressStepList, ProgressStepList, type ProgressStepListHandle } from './ProgressStepList';

const LABELS = ['Current', 'Patchwerk', 'Council'];

const rows = () => screen.getAllByTestId('progress-step').map(row => [row.textContent, row.getAttribute('data-state')]);
const spinners = () => screen.queryAllByTestId('progress-step-spinner');

describe('ProgressStepList', () => {
	it.each([
		[0, ['running', 'pending', 'pending']],
		[1, ['done', 'running', 'pending']],
		[2, ['done', 'done', 'running']],
	])('marks the steps before step %i done and the ones after it pending', (running, states) => {
		render(<ProgressStepList labels={LABELS} running={running} />);

		expect(rows()).toEqual(LABELS.map((label, index) => [label, states[index]]));
		expect(spinners()).toHaveLength(1);
		const current = screen.getAllByTestId('progress-step')[running];
		expect(current.contains(spinners()[0])).toBe(true);
		expect(current.getAttribute('aria-current')).toBe('step');
		expect(screen.getAllByTestId('progress-step').filter(row => row.hasAttribute('aria-current'))).toHaveLength(1);
	});

	it('lists every step from the start, the first one running, and moves on when told', () => {
		const handle = { current: null as ProgressStepListHandle | null };
		render(<HandleProgressStepList ref={handle} labels={LABELS} />);

		expect(rows().map(([, state]) => state)).toEqual(['running', 'pending', 'pending']);

		act(() => handle.current!.setRunning(2));
		expect(rows().map(([, state]) => state)).toEqual(['done', 'done', 'running']);
	});
});
