import { Tabs } from '@base-ui/react/tabs';
import { BulkBuildsColumns } from '@features/bulk/components/BulkBuildsColumns';
import { BulkCompareButton } from '@features/bulk/components/BulkCompareButton';
import { BulkEncounterResults } from '@features/bulk/components/BulkEncounterResults';
import { BulkItemSearch } from '@features/bulk/components/BulkItemSearch';
import { BulkPickerGroups } from '@features/bulk/components/BulkPickerGroups';
import { BulkProgressDialog } from '@features/bulk/components/BulkProgress';
import { BulkResults } from '@features/bulk/components/BulkResults';
import { BulkResultsCsvButton } from '@features/bulk/components/BulkResultsCsvButton';
import { BulkSettings } from '@features/bulk/components/BulkSettings';
import { BulkBuildsProvider } from '@features/bulk/context/BulkBuildsContext';
import { useBulkMode } from '@features/bulk/hooks/useBulkMode';
import { useBulkResults } from '@features/bulk/hooks/useBulkResults';
import { useBulkResultSelection } from '@features/bulk/hooks/useBulkResultSelection';
import { useBulkState } from '@features/bulk/hooks/useBulkState';
import { addBulkItems, clearBulkItems } from '@features/bulk/model/gear/items';
import { SelectorModal } from '@features/gear/components/SelectorModal';
import { OpenSelectorModalContext, useSelectorModalState } from '@features/gear/hooks/useSelectorModal';
import { useActivateTab } from '@features/hooks/useActivateTab';
import { BulkGearImporterDialog } from '@features/import-export';
import { ItemSpec } from '@generated/proto/common';
import i18n from '@i18n/config';
import { REPO_RELEASES_URL } from '@sim/constants/other';
import { useSimHost } from '@sim/context/SimHostContext';
import { useSimReady } from '@sim/hooks/useSimReady';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { TabNav, TabPanel, TabPanels } from '@ui-kit/TabNav';
import { TabPanelColumns } from '@ui-kit/TabPanelColumns';
import { useEffect, useRef, useState } from 'react';

const PANES = [
	{ id: 'bulkSetupTab', labelKey: 'bulk_tab.tabs.setup' },
	{ id: 'bulkResultsTab', labelKey: 'bulk_tab.tabs.results' },
] as const;

type BulkPaneId = (typeof PANES)[number]['id'];

export interface BulkTabActionsViewProps {
	showNativeDownload: boolean;
	onImportBags: () => void;
	onImportFavorites: () => void;
	onClear: () => void;
}

export const BulkTabActionsView = ({ showNativeDownload, onImportBags, onImportFavorites, onClear }: BulkTabActionsViewProps) => (
	<>
		{showNativeDownload && (
			<p className="mb-0">
				<Button href={REPO_RELEASES_URL} target="_blank" variant="unstyled">
					<Icon name="gauge-high" className="mr-1" />
					{i18n.t('bulk_tab.download_native')}
				</Button>
			</p>
		)}
		<div className="flex grid-flow-col gap-2" data-testid="bulk-gear-actions">
			<Button variant="secondary" onClick={onImportBags}>
				<Icon name="download" style="base" /> {i18n.t('bulk_tab.actions.import_bags')}
			</Button>
			<Button variant="secondary" onClick={onImportFavorites}>
				<Icon name="download" style="base" /> {i18n.t('bulk_tab.actions.import_favorites')}
			</Button>
			<Button variant="danger" className="ml-auto" onClick={onClear}>
				<Icon name="xmark" />
				{i18n.t('bulk_tab.actions.clear_items')}
			</Button>
		</div>
	</>
);

export const BulkTabBody = () => {
	const host = useSimHost();
	const ready = useSimReady();
	const activateTab = useActivateTab();
	const results = useBulkResults();
	const isRunning = useBulkState(slice => slice.isRunning);
	const { builds } = useBulkMode();

	const [activeId, setActiveId] = useState<BulkPaneId>('bulkSetupTab');
	const [importOpen, setImportOpen] = useState(false);
	const [selectedEncounter, selectEncounter] = useBulkResultSelection();
	const selector = useSelectorModalState();
	const resultsMode = useRef(builds);
	// Starting a run clears its mode's results and drops back to setup; finishing one opens the results. Switching modes
	// swaps the results shown and keeps the pane.
	useEffect(() => {
		if (resultsMode.current !== builds) {
			resultsMode.current = builds;
			return;
		}
		setActiveId(results ? 'bulkResultsTab' : 'bulkSetupTab');
	}, [results, builds]);

	return (
		<OpenSelectorModalContext value={selector.openTab}>
			<Tabs.Root className="w-full pt-2" data-testid="bulk-tab-tabs" value={activeId} onValueChange={next => setActiveId(next as BulkPaneId)}>
				<TabNav tabs={PANES.map(pane => ({ id: pane.id, label: i18n.t(pane.labelKey) }))} />
				<TabPanels>
					<TabPanel value="bulkSetupTab">
						<TabPanelColumns.Root fullWidth>
							<BulkBuildsProvider>
								<TabPanelColumns.Left variant="auto-columns">
									{builds ? (
										<BulkBuildsColumns />
									) : (
										<>
											<BulkTabActionsView
												showNativeDownload={ready && host.sim.isNative === false}
												onImportBags={() => setImportOpen(true)}
												onImportFavorites={() =>
													addBulkItems(
														host.player,
														host.sim.getFilters().favoriteItems.map(itemID => ItemSpec.create({ id: itemID })),
													)
												}
												onClear={() => clearBulkItems(host.player)}
											/>
											<BulkItemSearch ready={ready} />
											<BulkPickerGroups />
										</>
									)}
								</TabPanelColumns.Left>
								<BulkSettings />
							</BulkBuildsProvider>
						</TabPanelColumns.Root>
					</TabPanel>
					<TabPanel value="bulkResultsTab">
						<TabPanelColumns.Root fullWidth>
							<TabPanelColumns.Left variant="stacked">
								<BulkResults onActivateTab={activateTab} selected={selectedEncounter} />
							</TabPanelColumns.Left>
							{results && results.length > 0 && (
								<BulkEncounterResults
									selected={selectedEncounter}
									onSelect={selectEncounter}
									footer={
										<div className="flex flex-wrap gap-2">
											<BulkResultsCsvButton />
											{results.length > 1 && <BulkCompareButton />}
										</div>
									}
								/>
							)}
						</TabPanelColumns.Root>
					</TabPanel>
				</TabPanels>
			</Tabs.Root>
			{importOpen && <BulkGearImporterDialog open onOpenChange={setImportOpen} />}
			<SelectorModal state={selector} id="bulk-selector-modal" rail={false} />
			{isRunning && <BulkProgressDialog />}
		</OpenSelectorModalContext>
	);
};
