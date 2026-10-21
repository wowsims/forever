export interface DetailedResultsTabConfig {
	id: string;
	labelKey: string;
	className?: string;
	/** The tab's table reads the Group ranks toggle, which shows only while such a tab is open. */
	groupsRanks?: boolean;
}

export const DETAILED_RESULTS_TABS: ReadonlyArray<DetailedResultsTabConfig> = [
	{ id: 'damageTab', labelKey: 'results_tab.details.tabs.damage', groupsRanks: true },
	{ id: 'threatTab', labelKey: 'results_tab.details.tabs.threat', groupsRanks: true },
	{ id: 'healingTab', labelKey: 'results_tab.details.tabs.healing', groupsRanks: true },
	{ id: 'damageTakenTab', labelKey: 'results_tab.details.tabs.damage_taken', groupsRanks: true },
	{ id: 'buffsTab', labelKey: 'results_tab.details.tabs.buffs' },
	{ id: 'debuffsTab', labelKey: 'results_tab.details.tabs.debuffs' },
	{ id: 'castsTab', labelKey: 'results_tab.details.tabs.casts', groupsRanks: true },
	{ id: 'resourcesTab', labelKey: 'results_tab.details.tabs.resources' },
	{ id: 'timelineTab', labelKey: 'results_tab.details.tabs.timeline' },
	{ id: 'logTab', labelKey: 'results_tab.details.tabs.log' },
];

export const DEFAULT_DETAILED_RESULTS_TAB = 'damageTab';

/** Ties a pane to the tab button that controls it. */
export const tabButtonId = (tabId: string): string => `${tabId}-nav`;
