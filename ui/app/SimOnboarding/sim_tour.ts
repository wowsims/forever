import { BULK_TAB_ID, RESULTS_TAB_ID } from '@features/hooks/useActivateTab';
import type { TourStep } from '@features/onboarding';
import { DETAILED_RESULTS_TABS, tabButtonId } from '@features/results/components/DetailedResults/utils';

/** What the spec page offers, so a step with nothing to show is left out of the tour and its count. */
export interface SimTourContext {
	/** A gear planner has no Simulate button, encounter, rotation, results or batch sim. */
	canSimulate: boolean;
	/** A finished run to point at. */
	hasResult: boolean;
}

const tabButton = (id: string) => `[data-testid="sim-tabs"] [data-testid="${id}"]`;
const testId = (id: string) => `[data-testid="${id}"]`;
const canSimulate = (context: SimTourContext) => context.canSimulate;

// One card per tab of the detailed results, opening it; a tab the last run has nothing for is not shown, and is passed over.
const RESULTS_TAB_STEPS: ReadonlyArray<TourStep<SimTourContext>> = DETAILED_RESULTS_TABS.map(tab => ({
	id: `results_${tab.labelKey.split('.').pop()}`,
	tab: RESULTS_TAB_ID,
	press: `#${tabButtonId(tab.id)}`,
	target: `#${tabButtonId(tab.id)}`,
	optional: true,
	applies: canSimulate,
}));

/** The spec page's tour, in order. Each step's words are `onboarding.steps.<id>.title` and `.body`. */
export const SIM_TOUR_STEPS: ReadonlyArray<TourStep<SimTourContext>> = [
	{ id: 'welcome' },
	{ id: 'gear_tab', tab: 'gear-tab', target: tabButton('gear-tab') },
	{ id: 'import', target: testId('import-link') },
	{ id: 'gear_list', tab: 'gear-tab', target: testId('gear-picker-root') },
	{ id: 'gear_picker', tab: 'gear-tab', target: `${testId('gear-picker-root')} ${testId('item-picker-root')}` },
	{ id: 'gear_presets', tab: 'gear-tab', target: testId('gear-presets') },
	{ id: 'enchants', target: testId('suggest-reforges-settings-group'), optional: true },
	{ id: 'settings_tab', tab: 'settings-tab', target: tabButton('settings-tab') },
	{ id: 'encounter', tab: 'settings-tab', target: '[data-block="encounter-settings"]', applies: canSimulate },
	{ id: 'player', tab: 'settings-tab', target: '[data-block="player-settings"]' },
	{ id: 'other', tab: 'settings-tab', target: '[data-block="other-settings"]', optional: true },
	{ id: 'consumables', tab: 'settings-tab', target: '[data-block="consumes-settings"]' },
	{ id: 'buffs', tab: 'settings-tab', target: testId('buff-settings') },
	{ id: 'settings_presets', tab: 'settings-tab', target: testId('settings-presets') },
	{ id: 'talents_tab', tab: 'talents-tab', target: tabButton('talents-tab') },
	{ id: 'rotation_tab', tab: 'rotation-tab', target: tabButton('rotation-tab'), applies: canSimulate },
	{ id: 'rotation_auto', tab: 'rotation-tab', target: testId('rotation-auto'), optional: true, applies: canSimulate },
	{ id: 'rotation_type', tab: 'rotation-tab', target: testId('rotation-type-container'), applies: canSimulate },
	{ id: 'rotation_presets', tab: 'rotation-tab', target: testId('rotation-presets'), applies: canSimulate },
	{ id: 'simulate', target: testId('dps-action'), waitFor: 'simulation', applies: canSimulate },
	{ id: 'result', target: testId('results-sim'), applies: context => context.canSimulate && context.hasResult },
	{ id: 'results_tab', tab: RESULTS_TAB_ID, target: tabButton(RESULTS_TAB_ID), applies: canSimulate },
	...RESULTS_TAB_STEPS,
	{ id: 'bulk_tab', tab: BULK_TAB_ID, target: tabButton(BULK_TAB_ID), applies: canSimulate },
	{ id: 'stat_weights', target: testId('ep-weights-action') },
	{ id: 'finish' },
];
