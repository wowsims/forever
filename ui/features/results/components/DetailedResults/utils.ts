export interface DetailedResultsTabConfig {
	id: string;
	labelKey: string;
	className?: string;
	/** The tab's table reads the Group ranks toggle, which shows only while such a tab is open. */
	groupsRanks?: boolean;
	/** The tab's table reads the Show APM breakdown toggle, which shows only while such a tab is open. */
	breaksDownApm?: boolean;
}

// Keep this tab disabled while it only reports spells used by the rotation: without downranking
// and spells absent from the rotation, the table is incomplete and not useful as a spell reference.
// Re-enable it once both cases are supported.
export const HEALING_SPELLS_RESULTS_ENABLED = false;

const ALL_DETAILED_RESULTS_TABS: ReadonlyArray<DetailedResultsTabConfig> = [
	{ id: 'damageTab', labelKey: 'results_tab.details.tabs.damage', groupsRanks: true },
	{ id: 'healingSpellsTab', labelKey: 'results_tab.details.tabs.healing_spells' },
	{ id: 'threatTab', labelKey: 'results_tab.details.tabs.threat', groupsRanks: true },
	{ id: 'healingTab', labelKey: 'results_tab.details.tabs.healing', groupsRanks: true },
	{ id: 'damageTakenTab', labelKey: 'results_tab.details.tabs.damage_taken', groupsRanks: true },
	{ id: 'buffsTab', labelKey: 'results_tab.details.tabs.buffs' },
	{ id: 'debuffsTab', labelKey: 'results_tab.details.tabs.debuffs' },
	{ id: 'castsTab', labelKey: 'results_tab.details.tabs.casts', groupsRanks: true, breaksDownApm: true },
	{ id: 'resourcesTab', labelKey: 'results_tab.details.tabs.resources' },
	{ id: 'timelineTab', labelKey: 'results_tab.details.tabs.timeline' },
	{ id: 'logTab', labelKey: 'results_tab.details.tabs.log' },
];

export const DETAILED_RESULTS_TABS = ALL_DETAILED_RESULTS_TABS.filter(tab => HEALING_SPELLS_RESULTS_ENABLED || tab.id !== 'healingSpellsTab');

export const DEFAULT_DETAILED_RESULTS_TAB = 'damageTab';
export const HEALING_SPELLS_RESULTS_TAB = 'healingSpellsTab';

/** Ties a pane to the tab button that controls it. */
export const tabButtonId = (tabId: string): string => `${tabId}-nav`;
