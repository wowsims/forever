import { renderWithHost } from '@testing/react';
import { act, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const trackPageView = vi.fn();
const runState = { isRunning: false, isAborting: false, abort: vi.fn() };

vi.mock('../../../../tracking/utils', () => ({ trackEvent: vi.fn(), trackPageView: (...args: unknown[]) => trackPageView(...args) }));
vi.mock('@sim/hooks/useSimRun', () => ({ useSimRun: () => runState }));
vi.mock('./ReforgeSettingsPanel', () => ({ ReforgeSettingsPanel: () => <div data-testid="settings-body" /> }));

const { ReforgePanel } = await import('./ReforgePanel');

const model = {
	settings: {},
	softCapsConfigWithLimits: [],
	previousGear: null,
} as any;

const renderPanel = () => renderWithHost(<ReforgePanel model={model} />, { sim: { runs: { isRunning: () => false } } });

beforeEach(() => {
	trackPageView.mockClear();
	runState.isRunning = false;
});

describe('ReforgePanel', () => {
	it('renders the run button and the settings trigger as one sidebar action pair', () => {
		const { container } = renderPanel();
		const [run, settings] = [...container.querySelectorAll('button')];

		expect(run.hasAttribute('data-sidebar-action')).toBe(true);
		expect(settings.hasAttribute('data-sidebar-action')).toBe(true);
		expect(settings.querySelector('.fa-gear')).not.toBeNull();
	});

	it.each([false, true])('disables the run button and shows its spinner only while a gear optimization run is in flight (%s)', isRunning => {
		runState.isRunning = isRunning;
		const { container } = renderPanel();
		const run = container.querySelector('button')!;

		expect(run.disabled).toBe(isRunning);
		expect(run.hasAttribute('aria-busy')).toBe(isRunning);
		expect(!!run.querySelector('[data-testid="loader"]')).toBe(isRunning);
	});

	it('builds no popover body and tracks no page view until the trigger is used', () => {
		renderPanel();

		expect(document.querySelector('[data-testid="settings-body"]')).toBeNull();
		expect(trackPageView).not.toHaveBeenCalled();
	});

	it('mounts no progress dialog until a run starts', () => {
		renderPanel();

		expect(document.querySelector('[data-testid="progress-tracker-dialog"]')).toBeNull();
	});

	it('builds the body and tracks the page view when the trigger opens the popover', async () => {
		const { container } = renderPanel();
		const settings = [...container.querySelectorAll('button')][1];

		await act(async () => void fireEvent.click(settings));

		expect(document.querySelector('[data-testid="settings-body"]')).not.toBeNull();
		expect(trackPageView).toHaveBeenCalledWith('Reforge Settings', 'reforge-settings');
	});
});
