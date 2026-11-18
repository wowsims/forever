import { Tabs } from '@base-ui/react/tabs';
import { GEAR_TAB_ID, ROTATION_TAB_ID, SETTINGS_TAB_ID, TALENTS_TAB_ID } from '@features/hooks/useActivateTab';
import i18n from '@i18n/config';
import { PresetConfigurationCategory } from '@sim/constants/preset_categories';
import { useSimHost } from '@sim/context/SimHostContext';
import { Dialog } from '@ui-kit/Dialog';
import { TabNav } from '@ui-kit/TabNav';
import type { ComponentType } from 'react';

import { GearPresets } from './GearPresets';
import { PresetConfigurationPicker } from './PresetConfigurationPicker';
import { useActiveSimTab } from './SimTabs';
import { RotationSidebar } from './tabs/RotationTabBody';
import { SettingsPresets } from './tabs/SettingsTabBody';
import { TalentPresets } from './tabs/TalentsTabBody';

const ALL_CATEGORIES = Object.values(PresetConfigurationCategory);

const BuildPresets = () => <PresetConfigurationPicker categories={ALL_CATEGORIES} />;

interface PresetSection {
	id: string;
	/** The sim tab whose presets column this is, which opens the dialog on it. */
	tab?: string;
	title: string;
	body: ComponentType;
}

const presetSections = (simDisabled: boolean): ReadonlyArray<PresetSection> => [
	{ id: 'presets-builds', title: i18n.t('sidebar.buttons.presets.builds'), body: BuildPresets },
	{ id: 'presets-gear', tab: GEAR_TAB_ID, title: i18n.t('gear_tab.title'), body: GearPresets },
	{ id: 'presets-settings', tab: SETTINGS_TAB_ID, title: i18n.t('settings_tab.title'), body: SettingsPresets },
	{ id: 'presets-talents', tab: TALENTS_TAB_ID, title: i18n.t('talents_tab.title'), body: TalentPresets },
	...(simDisabled ? [] : [{ id: 'presets-rotation', tab: ROTATION_TAB_ID, title: i18n.t('rotation_tab.title'), body: RotationSidebar }]),
];

// Mounted with each opening of the dialog, so it opens on the open sim tab's presets, and on the builds from any other.
const PresetSections = () => {
	const sections = presetSections(useSimHost().simDisabled);
	const activeTab = useActiveSimTab();
	return (
		<Tabs.Root className="contents" defaultValue={sections.find(section => section.tab === activeTab)?.id ?? sections[0].id}>
			<TabNav tabs={sections.map(section => ({ id: section.id, label: section.title }))} />
			{sections.map(section => (
				<Tabs.Panel key={section.id} value={section.id} id={section.id} className="flex flex-col gap-section">
					<section.body />
				</Tabs.Panel>
			))}
		</Tabs.Root>
	);
};

export interface SimPresetsDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

/** Every presets column the tabs keep at their foot, in one dialog the narrow screen's dock opens. */
export const SimPresetsDialog = ({ open, onOpenChange }: SimPresetsDialogProps) => (
	<Dialog open={open} onOpenChange={onOpenChange} testId="sim-presets-dialog" title={i18n.t('sidebar.buttons.presets.title')}>
		<PresetSections />
	</Dialog>
);
