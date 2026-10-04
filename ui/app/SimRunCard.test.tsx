import { fireEvent, render } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { SimHostObject } from './individual_sim_ui';
import { SimRunCard } from './SimRunCard';
import { TabActivationContext } from './tab_activation';

vi.mock('@features/results/components/IterationsPicker', () => ({ IterationsPicker: () => <div data-testid="iterations-picker" /> }));
vi.mock('@features/results/components/SimulateAction', () => ({ SimulateAction: () => <button data-testid="dps-action" /> }));
vi.mock('@features/stat-weights/components/StatWeightsAction', () => ({ StatWeightsAction: () => <button data-testid="ep-weights-action" /> }));
vi.mock('./SidebarActions', () => ({ SidebarActions: () => <div data-testid="registry-actions" /> }));
vi.mock('./NoticeNativeSim', () => ({ NoticeNativeSim: () => null }));
const simStatus = vi.hoisted(() => ({ current: 'ready' as 'loading' | 'ready' | 'error' }));
vi.mock('@sim/hooks/useSimStatus', () => ({ useSimStatus: () => ({ status: simStatus.current, error: null }) }));
vi.mock('@features/results/components/SimResultsPanel', () => ({
	SimResultsPanel: ({ onOpenBreakdown }: { onOpenBreakdown?: () => void }) => (
		<div data-testid="results-viewer">{onOpenBreakdown && <button data-testid="results-open-breakdown" onClick={onOpenBreakdown} />}</div>
	),
}));

const host = (parts: Partial<SimHostObject<any>> = {}) => ({ disabled: false, simDisabled: false, ...parts }) as SimHostObject<any>;

const renderCard = (value: SimHostObject<any> | null, activate = vi.fn()) => ({
	activate,
	...render(
		<TabActivationContext value={activate}>
			<SimRunCard host={value} sim={{} as never} />
		</TabActivationContext>,
	),
});

const ids = (element: Element) => [...element.children].map(child => child.getAttribute('data-testid'));

describe('SimRunCard', () => {
	beforeEach(() => {
		simStatus.current = 'ready';
	});

	it('covers the whole card while the sim initialises, and uncovers it once ready', () => {
		simStatus.current = 'loading';
		const { getByTestId, queryByTestId, rerender } = renderCard(host());
		expect(getByTestId('sim-run-card').getAttribute('aria-busy')).toBe('true');
		expect(queryByTestId('run-card-loading')).not.toBeNull();

		simStatus.current = 'ready';
		rerender(
			<TabActivationContext value={vi.fn()}>
				<SimRunCard host={host()} sim={{} as never} />
			</TabActivationContext>,
		);
		expect(queryByTestId('run-card-loading')).toBeNull();
	});

	it('puts stat weights beside Simulate, the spec actions under them, and the last run under the divider', () => {
		const { getByTestId } = renderCard(host());

		expect(ids(getByTestId('run-card-controls'))).toEqual(['iterations-picker', 'sim-run-card-run-row', 'sim-run-card-actions']);
		expect(ids(getByTestId('sim-run-card-run-row'))).toEqual(['dps-action', 'ep-weights-action']);
		expect(ids(getByTestId('sim-run-card-actions'))).toEqual(['registry-actions']);
		expect(ids(getByTestId('run-card-body'))).toEqual(['results-viewer']);
	});

	it('shows only the iterations picker before the host is constructed', () => {
		const { getByTestId } = renderCard(null);

		expect(ids(getByTestId('sim-run-card'))).toEqual(['run-card-controls']);
		expect(ids(getByTestId('run-card-controls'))).toEqual(['iterations-picker']);
	});

	it('keeps stat weights without Simulate, and no breakdown link, for a sim that never runs', () => {
		const { getByTestId, queryByTestId } = renderCard(host({ simDisabled: true }));

		expect(ids(getByTestId('sim-run-card-run-row'))).toEqual(['ep-weights-action']);
		expect(queryByTestId('results-open-breakdown')).toBeNull();
	});

	it('opens the Results tab from the full breakdown link', () => {
		const { getByTestId, activate } = renderCard(host());

		fireEvent.click(getByTestId('results-open-breakdown'));

		expect(activate).toHaveBeenCalledWith('detailed-results-tab-tab');
	});
});
