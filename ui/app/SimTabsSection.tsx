import { BULK_TAB_ID, RESULTS_TAB_ID, ROTATION_TAB_ID } from '@features/hooks/useActivateTab';
import { DetailedResults } from '@features/results/components/DetailedResults';
import i18n from '@i18n/config';
import { SimTabPane } from '@ui-kit/SimTabPane';
import { type ReactNode, useMemo } from 'react';

import type { SimHostObject } from './individual_sim_ui';
import { SimTabDef } from './SimTabDef';
import { SimTabs } from './SimTabs';
import { BulkTabBody } from './tabs/BulkTabBody';
import { GearTabBody } from './tabs/GearTabBody';
import { RotationTabPane } from './tabs/RotationTabPane';
import { SettingsTabBody } from './tabs/SettingsTabBody';
import { TalentsTabBody } from './tabs/TalentsTabBody';

export interface SimTabsSectionProps {
	host: SimHostObject<any> | null;
	children: ReactNode;
}

// The pane list is built once per host, so `SimTabsPanels` keeps the same elements (and every
// `keepMounted` pane's state) across a shell re-render that has nothing to do with the tabs — the
// sticky header toggling, most often.
export const SimTabsSection = ({ host, children }: SimTabsSectionProps) => {
	const defs = useMemo(
		() =>
			host && (
				<>
					<SimTabDef id="gear-tab" title={i18n.t('gear_tab.title')}>
						<SimTabPane id="gear-tab">
							<GearTabBody />
						</SimTabPane>
					</SimTabDef>
					<SimTabDef id="settings-tab" title={i18n.t('settings_tab.title')}>
						<SimTabPane
							id="settings-tab"
							className="[&_.ui-columns-left-settings]:max-xl:flex-wrap [&_.ui-columns-left-settings>.ui-columns-col]:max-xl:basis-full"
							contentClassName="max-xxl:flex-col">
							<SettingsTabBody />
						</SimTabPane>
					</SimTabDef>
					<SimTabDef id="talents-tab" title={i18n.t('talents_tab.title')}>
						<SimTabPane id="talents-tab">
							<TalentsTabBody />
						</SimTabPane>
					</SimTabDef>
					{/* The three tabs that only mean something once a sim has run. A gear planner never runs one. */}
					{!host.simDisabled && (
						<>
							<SimTabDef id={ROTATION_TAB_ID} title={i18n.t('rotation_tab.title')}>
								<RotationTabPane />
							</SimTabDef>
							{/* Not a `SimTabPane`: the doubled id and the missing content container are what the stylesheets select on. */}
							<SimTabDef id={RESULTS_TAB_ID} title={i18n.t('results_tab.title')}>
								<div id={RESULTS_TAB_ID}>
									<DetailedResults resultsManager={host.raidSimResultsManager} />
								</div>
							</SimTabDef>
							<SimTabDef id={BULK_TAB_ID} title={i18n.t('bulk_tab.title')}>
								<SimTabPane id={BULK_TAB_ID} className="p-0">
									<BulkTabBody />
								</SimTabPane>
							</SimTabDef>
						</>
					)}
				</>
			),
		[host],
	);

	// Always `SimTabs`, even with no `defs` yet: switching element types here on the render the host
	// arrives would remount `children` (the header and `main`), taking the sticky header's observer
	// and `StickyHeaderContext` down with the old nodes.
	return <SimTabs defs={defs}>{children}</SimTabs>;
};
