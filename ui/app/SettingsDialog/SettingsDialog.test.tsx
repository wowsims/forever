import { supportedLanguages } from '@i18n/locale_service';
import { createSimStore } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { SettingsDialog, SettingsDialogView, type SettingsDialogViewProps } from './SettingsDialog';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

const toasts = vi.hoisted(() => [] as Array<{ variant: string; body: unknown }>);
vi.mock('@ui-kit/Toast', async importOriginal => ({
	...(await importOriginal<typeof import('@ui-kit/Toast')>()),
	toastManager: {
		add: (options: { variant: string; body: unknown }) => {
			toasts.push(options);
			return '';
		},
		close: () => {},
	},
}));
const trackEvent = vi.hoisted(() => vi.fn());
vi.mock('../../tracking/utils', () => ({ trackEvent }));

class FakeSim {
	readonly store = createSimStore();
	wasm = Promise.resolve(true);

	isWasm() {
		return this.wasm;
	}
	setFixedRngSeed(value: number) {
		this.store.setState(state => ({ sim: { ...state.sim, fixedRngSeed: value } }));
	}
	recordLastUsedRngSeed(value: number) {
		this.store.setState(state => ({
			sim: { ...state.sim, lastUsedRngSeed: value, lastUsedRngSeedVersion: state.sim.lastUsedRngSeedVersion + 1 },
		}));
	}
	setLanguage(value: string) {
		this.store.setState(state => ({ ui: { ...state.ui, language: value } }));
	}
	getShowThreatMetrics() {
		return this.store.getState().ui.showThreatMetrics;
	}
	setShowThreatMetrics(value: boolean) {
		this.store.setState(state => ({ ui: { ...state.ui, showThreatMetrics: value } }));
	}
	setShowExperimental(value: boolean) {
		this.store.setState(state => ({ ui: { ...state.ui, showExperimental: value } }));
	}
	setShowPowerTools(value: boolean) {
		this.store.setState(state => ({ ui: { ...state.ui, showPowerTools: value } }));
	}
	setShowQuickSwap(value: boolean) {
		this.store.setState(state => ({ ui: { ...state.ui, showQuickSwap: value } }));
	}
	setWasmConcurrency(value: number) {
		this.store.setState(state => ({ ui: { ...state.ui, wasmConcurrency: value } }));
	}
}

const mount = (sim: FakeSim, applyDefaults = vi.fn()) => {
	const host = { applyDefaults, player: { getPlayerSpec: () => ({ isHealingSpec: false }) } } as never;
	const rendered = renderWithHost(<SettingsDialog open onOpenChange={() => {}} host={host} />, { sim });
	return { ...rendered, applyDefaults };
};

const byId = (id: string) => document.getElementById(id) as HTMLElement;
const concurrencyRow = () => screen.getByTestId('use-concurrency-container');

describe('SettingsDialog', () => {
	it('keeps the concurrency row mounted and hides it with the attribute when the sim is not wasm', async () => {
		const sim = new FakeSim();
		sim.wasm = Promise.resolve(false);
		mount(sim);

		expect(concurrencyRow().hidden).toBe(false);
		await waitFor(() => expect(concurrencyRow().hidden).toBe(true));
		expect(concurrencyRow().isConnected).toBe(true);
	});

	it('writes a toggle through to the sim', () => {
		const sim = new FakeSim();
		mount(sim);

		act(() => {
			fireEvent.click(byId('simui-show-threat-metrics'));
		});
		expect(sim.getShowThreatMetrics()).toBe(true);
	});

	it('follows the last used rng seed', async () => {
		const sim = new FakeSim();
		mount(sim);

		const lastUsedRngSeedText = () => screen.getByText('info.options.fixed_rng_seed.last_used').parentElement!.textContent;
		expect(lastUsedRngSeedText()).toContain('0');
		act(() => {
			sim.recordLastUsedRngSeed(1234);
		});
		await waitFor(() => expect(lastUsedRngSeedText()).toContain('1234'));
	});

	it('restores defaults, reports it and toasts', () => {
		toasts.length = 0;
		trackEvent.mockClear();
		const { applyDefaults } = mount(new FakeSim());

		act(() => {
			fireEvent.click(screen.getByRole('button', { name: 'info.options.restore_defaults.button' }));
		});

		expect(applyDefaults).toHaveBeenCalledTimes(1);
		expect(trackEvent).toHaveBeenCalledWith({ action: 'settings', category: 'restore-defaults', label: 'restore' });
		expect(toasts).toEqual([{ variant: 'success', body: 'info.options.restore_defaults.success_message' }]);
	});
});

const props = (overrides: Partial<SettingsDialogViewProps> = {}): SettingsDialogViewProps => ({
	open: true,
	onOpenChange: vi.fn(),
	fixedRngSeed: 0,
	onFixedRngSeedChange: vi.fn(),
	lastUsedRngSeed: 0,
	language: 'en',
	onLanguageChange: vi.fn(),
	showThreatMetrics: false,
	onShowThreatMetricsChange: vi.fn(),
	showDamageMetrics: false,
	onShowDamageMetricsChange: vi.fn(),
	showDamageMetricsRow: false,
	showExperimental: false,
	onShowExperimentalChange: vi.fn(),
	showPowerTools: false,
	onShowPowerToolsChange: vi.fn(),
	showQuickSwap: true,
	onShowQuickSwapChange: vi.fn(),
	wasmConcurrency: 0,
	onWasmConcurrencyChange: vi.fn(),
	showConcurrency: true,
	maxWorkers: 4,
	firefox: false,
	onApplyDefaults: vi.fn(),
	...overrides,
});

const pick = async (id: string, option: string) => {
	await userEvent.click(byId(id));
	await userEvent.click(screen.getByRole('option', { name: option }));
};
const checked = (testId: string) => within(screen.getByTestId(testId)).getByRole('checkbox').getAttribute('aria-checked');
const workerNote = () => concurrencyRow().querySelector('.ui-help-text') as HTMLElement;

describe('SettingsDialogView', () => {
	it('shows each value it is given', () => {
		render(<SettingsDialogView {...props({ fixedRngSeed: 42, lastUsedRngSeed: 1234, showThreatMetrics: true, wasmConcurrency: 3 })} />);

		expect((byId('simui-fixed-rng-seed') as HTMLInputElement).value).toBe('42');
		expect(screen.getByText('info.options.fixed_rng_seed.last_used').parentElement?.textContent).toContain('1234');
		expect(byId('simui-language-picker').textContent).toBe(supportedLanguages.en);
		expect(checked('show-threat-metrics-picker')).toBe('true');
		expect(checked('show-experimental-picker')).toBe('false');
		expect(checked('show-quick-swap-picker')).toBe('true');
		expect(byId('simui-concurrent-workers-picker').textContent).toBe('3');
	});

	it('offers Off and two up to the maximum worker count', async () => {
		render(<SettingsDialogView {...props({ maxWorkers: 4 })} />);

		await userEvent.click(byId('simui-concurrent-workers-picker'));
		expect(screen.getAllByRole('option').map(option => option.textContent)).toEqual(['info.options.use_multiple_cpu_cores.off', '2', '3', '4']);
	});

	it('reports each change through its handler', async () => {
		const handlers = props();
		render(<SettingsDialogView {...handlers} />);

		act(() => {
			fireEvent.click(byId('simui-show-threat-metrics'));
			fireEvent.click(byId('simui-show-experimental'));
			fireEvent.click(byId('simui-show-power-tools'));
			fireEvent.click(byId('simui-show-quick-swap'));
		});
		await pick('simui-concurrent-workers-picker', '2');

		expect(handlers.onShowThreatMetricsChange).toHaveBeenCalledWith(true);
		expect(handlers.onShowExperimentalChange).toHaveBeenCalledWith(true);
		expect(handlers.onShowPowerToolsChange).toHaveBeenCalledWith(true);
		expect(handlers.onShowQuickSwapChange).toHaveBeenCalledWith(false);
		expect(handlers.onWasmConcurrencyChange).toHaveBeenCalledWith(2);
	});

	it('commits the seed on change', () => {
		const handlers = props();
		render(<SettingsDialogView {...handlers} />);

		act(() => {
			fireEvent.change(byId('simui-fixed-rng-seed'), { target: { value: '99' } });
		});

		expect(handlers.onFixedRngSeedChange).toHaveBeenCalledWith(99);
	});

	it('reports the chosen language by its code', async () => {
		const handlers = props();
		render(<SettingsDialogView {...handlers} />);
		const [code, name] = Object.entries(supportedLanguages).at(-1)!;

		await pick('simui-language-picker', name);

		expect(handlers.onLanguageChange).toHaveBeenCalledWith(code);
	});

	it('selects English for a language it does not know', () => {
		render(<SettingsDialogView {...props({ language: '' })} />);

		expect(byId('simui-language-picker').textContent).toBe(supportedLanguages.en);
	});

	it('keeps the concurrency row mounted and hidden when it is not shown', () => {
		render(<SettingsDialogView {...props({ showConcurrency: false })} />);

		expect(concurrencyRow().hidden).toBe(true);
	});

	it('shows the worker note on firefox only', () => {
		const { rerender } = render(<SettingsDialogView {...props()} />);
		expect(workerNote().hidden).toBe(true);
		expect(workerNote().textContent).toBe('');

		rerender(<SettingsDialogView {...props({ firefox: true })} />);
		expect(workerNote().hidden).toBe(false);
		expect(workerNote().textContent).toContain('info.options.use_multiple_cpu_cores.firefox_note');
	});

	it('applies the defaults from the footer button', () => {
		const handlers = props();
		render(<SettingsDialogView {...handlers} />);

		act(() => {
			fireEvent.click(screen.getByRole('button', { name: 'info.options.restore_defaults.button' }));
		});

		expect(handlers.onApplyDefaults).toHaveBeenCalledTimes(1);
	});

	it('offers to restart the tour only where the page has one', () => {
		const onRestartTour = vi.fn();
		const { rerender } = render(<SettingsDialogView {...props()} />);
		expect(screen.queryByRole('button', { name: 'info.options.restart_tour' })).toBeNull();

		rerender(<SettingsDialogView {...props({ onRestartTour })} />);
		act(() => {
			fireEvent.click(screen.getByRole('button', { name: 'info.options.restart_tour' }));
		});

		expect(onRestartTour).toHaveBeenCalledTimes(1);
	});
});
