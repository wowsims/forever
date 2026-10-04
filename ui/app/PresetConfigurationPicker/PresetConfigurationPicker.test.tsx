import { renderWithHost } from '@testing/react';
import { fireEvent, render } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { PresetBuild } from '../preset_utils';
import type { PresetConfigurationPickerViewProps } from './PresetConfigurationPicker';

const applyBuild = vi.fn();
const activeNames = new Set<string>();

vi.mock('@features/settings/model/apply_build', () => ({ applyBuild: (...args: unknown[]) => applyBuild(...args) }));
vi.mock('../preset_build_state', () => ({
	buildCategories: () => ['Gear'],
	isBuildActive: (build: { name: string }) => activeNames.has(build.name),
}));
vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));
vi.mock('@sim/state/subscriptions', async () => (await import('@testing/sim/subscriptions')).mockSubscriptions());
vi.mock('@sim/hooks/useSimReady', () => ({ useSimReady: () => ready }));

let ready = true;
const { PresetConfigurationPicker, PresetConfigurationPickerView } = await import('./PresetConfigurationPicker');

const setup = (builds: Array<Record<string, unknown>>, categories = ['gear']) =>
	renderWithHost(<PresetConfigurationPicker categories={categories as never} />, {
		individualConfig: { presets: { builds } },
		getStorageKey: (part: string) => `spec${part}`,
	});

const chips = (container: HTMLElement) => [...container.querySelectorAll<HTMLElement>('[data-testid="saved-data-set-chip"]')];

beforeEach(() => {
	applyBuild.mockClear();
	activeNames.clear();
	window.localStorage.clear();
	ready = true;
});

const phaseTabs = (container: HTMLElement) => [...container.querySelectorAll<HTMLElement>('[data-testid="preset-group-phase-tabs"] button')];

describe('PresetConfigurationPicker', () => {
	// The picker is mounted in four places with different category lists, and each shows only the
	// builds touching its own categories.
	it('keeps only the builds that carry one of its categories', () => {
		const { container } = setup(
			[
				{ name: 'P1', gear: {} },
				{ name: 'P2', talents: {} },
			],
			['gear'],
		);

		expect(chips(container).map(chip => chip.textContent)).toEqual(['P1']);
	});

	it('marks the active build, and only that one', () => {
		activeNames.add('P2');
		const { container } = setup([
			{ name: 'P1', gear: {} },
			{ name: 'P2', gear: {} },
		]);

		expect(chips(container).map(chip => chip.hasAttribute('data-active'))).toEqual([false, true]);
	});

	it('applies the build the name was clicked on', () => {
		const { container } = setup([
			{ name: 'P1', gear: {} },
			{ name: 'P2', gear: {} },
		]);

		fireEvent.click(chips(container)[1].querySelector('[data-testid="saved-data-set-name"]')!);

		expect(applyBuild).toHaveBeenCalledTimes(1);
		expect((applyBuild.mock.calls[0][0] as { name: string }).name).toBe('P2');
	});

	// Before the sim is ready, the block exists but is empty.
	it('renders the block but no chips before the sim is ready', () => {
		ready = false;
		const { container } = setup([{ name: 'P1', gear: {} }]);

		expect(container.querySelector('[data-testid="content-block"]')).not.toBeNull();
		expect(chips(container)).toHaveLength(0);
	});
});

// A spec whose builds carry a phase gets the tab bar, and only the chosen phase's chips.
describe('PresetConfigurationPicker, grouped by phase', () => {
	const phased = [
		{ name: 'P1a', gear: {}, phase: 1, group: 'Beast Mastery' },
		{ name: 'P1b', gear: {}, phase: 1, group: 'Survival' },
		{ name: 'P3', gear: {}, phase: 3 },
	];

	it('leaves an unphased spec on the flat picker', () => {
		const { container } = setup([{ name: 'P1', gear: {} }]);

		expect(phaseTabs(container)).toHaveLength(0);
		expect(chips(container)).toHaveLength(1);
	});

	it('shows one tab per phase and only the current phase\u2019s chips', () => {
		const { container } = setup(phased);

		expect(phaseTabs(container).map(tab => tab.textContent)).toEqual(['common.phase_names.1', 'common.phase_names.3']);
		// CURRENT_PHASE is 1, and it is among the builds' phases, so that tab opens selected.
		expect(phaseTabs(container).map(tab => tab.hasAttribute('data-active'))).toEqual([true, false]);
		expect(chips(container).map(chip => chip.textContent)).toEqual(['P1a', 'P1b']);
	});

	it('switches the visible chips when another phase tab is clicked', () => {
		const { container } = setup(phased);

		fireEvent.click(phaseTabs(container)[1]);

		expect(chips(container).map(chip => chip.textContent)).toEqual(['P3']);
	});

	// A default page load leaves `__presetFilters__` absent.
	it('writes the filter key only once a phase is chosen', () => {
		const { container } = setup(phased);

		expect(window.localStorage.getItem('spec__presetFilters__')).toBeNull();

		fireEvent.click(phaseTabs(container)[1]);

		expect(JSON.parse(window.localStorage.getItem('spec__presetFilters__')!)).toEqual({ phase: 3 });
	});

	// Applying a build writes its phase to the bar, so the chip that was just clicked stays on
	// screen; no tab is clicked first, so the write is the build's alone.
	it('follows the applied build to its phase', () => {
		const { container } = setup(phased);

		expect(window.localStorage.getItem('spec__presetFilters__')).toBeNull();

		fireEvent.click(chips(container)[0].querySelector('[data-testid="saved-data-set-name"]')!);

		expect(applyBuild).toHaveBeenCalledTimes(1);
		expect(JSON.parse(window.localStorage.getItem('spec__presetFilters__')!)).toEqual({ phase: 1 });
	});
});

const props = (overrides: Partial<PresetConfigurationPickerViewProps> = {}): PresetConfigurationPickerViewProps => ({
	builds: [{ name: 'P1' }, { name: 'P2' }] as Array<PresetBuild>,
	active: [false, true],
	ready: true,
	phases: [],
	phase: 1,
	onSelectPhase: vi.fn(),
	onApply: vi.fn(),
	...overrides,
});

describe('PresetConfigurationPickerView', () => {
	it('renders nothing without builds', () => {
		const { container } = render(<PresetConfigurationPickerView {...props({ builds: [] })} />);

		expect(container.firstChild).toBeNull();
	});

	it('marks the active builds and applies the one clicked', () => {
		const handlers = props();
		const { container } = render(<PresetConfigurationPickerView {...handlers} />);

		expect(chips(container).map(chip => chip.hasAttribute('data-active'))).toEqual([false, true]);
		fireEvent.click(chips(container)[0].querySelector('[data-testid="saved-data-set-name"]')!);
		expect(handlers.onApply).toHaveBeenCalledWith(handlers.builds[0]);
	});

	it('keeps the block but shows no chips before the sim is ready', () => {
		const { container } = render(<PresetConfigurationPickerView {...props({ ready: false, active: null })} />);

		expect(container.querySelector('[data-testid="content-block"]')).not.toBeNull();
		expect(chips(container)).toHaveLength(0);
	});

	it('groups phased builds under phase tabs and shows the chosen phase only', () => {
		const handlers = props({
			builds: [
				{ name: 'P1a', phase: 1 },
				{ name: 'P3', phase: 3 },
			] as Array<PresetBuild>,
			active: [false, false],
			phases: [1, 3],
			phase: 1,
		});
		const { container } = render(<PresetConfigurationPickerView {...handlers} />);

		expect(chips(container).map(chip => chip.textContent)).toEqual(['P1a']);
		fireEvent.click(phaseTabs(container)[1]);
		expect(handlers.onSelectPhase).toHaveBeenCalledWith(3);
	});

	it('labels the groups only when there is more than one', () => {
		const builds = [
			{ name: 'P1a', phase: 1, group: 'Beast Mastery' },
			{ name: 'P1b', phase: 1, group: 'Survival' },
			{ name: 'P3', phase: 3 },
		] as Array<PresetBuild>;
		const { container } = render(<PresetConfigurationPickerView {...props({ builds, active: [false, false, false], phases: [1, 3], phase: 1 })} />);

		expect([...container.querySelectorAll('[data-testid="preset-group-label"]')].map(label => label.textContent)).toEqual(['Beast Mastery', 'Survival']);
	});
});
