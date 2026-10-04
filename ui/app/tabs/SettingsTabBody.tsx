import { EncounterPicker, SavedEncounter } from '@features/encounter';
import { SelectorModal } from '@features/gear/components/SelectorModal';
import { OpenSelectorModalContext, useSelectorModalState } from '@features/gear/hooks/useSelectorModal';
import { BuffSection, ConsumesFilters, ConsumesPicker, CustomSection, OtherSettings, PlayerSettings, RaidBuffs, SavedSettings } from '@features/settings';
import * as BuffDebuffInputs from '@features/settings/model/buffs_debuffs';
import { applyOwnerClassLabels, relevantStatOptions } from '@features/settings/model/stat_options';
import i18n from '@i18n/config';
import { PresetConfigurationCategory } from '@sim/constants/preset_categories';
import { usePlayer, useSimHost, useSpecConfig } from '@sim/context/SimHostContext';
import { useSimReady } from '@sim/hooks/useSimReady';
import { ContentBlock } from '@ui-kit/ContentBlock';
import { TabPanelColumns } from '@ui-kit/TabPanelColumns';
import { useMemo } from 'react';

import { PresetConfigurationPicker } from '../PresetConfigurationPicker';

const SETTINGS_PRESETS = [PresetConfigurationCategory.Encounter, PresetConfigurationCategory.Settings];
const GEAR_PLANNER_PRESETS = [PresetConfigurationCategory.Settings];

export const SettingsTabBody = () => {
	const host = useSimHost();
	const config = useSpecConfig();
	const player = usePlayer();
	const ready = useSimReady();

	const options = useMemo(
		() => ({
			buffs: applyOwnerClassLabels(relevantStatOptions(BuffDebuffInputs.BUFFS_CONFIG, host), player),
			partyBuffs: applyOwnerClassLabels(relevantStatOptions(BuffDebuffInputs.PARTY_BUFFS_CONFIG, host), player),
			debuffs: applyOwnerClassLabels(relevantStatOptions(BuffDebuffInputs.DEBUFFS_CONFIG, host), player),
			debuffsMisc: applyOwnerClassLabels(relevantStatOptions(BuffDebuffInputs.DEBUFFS_MISC_CONFIG, host), player),
		}),
		[host, player],
	);

	const itemSwapSlots = config.itemSwapSlots || [];
	const hasOtherSettings = config.otherInputs.inputs.length > 0 || itemSwapSlots.length > 0;
	const selector = useSelectorModalState();
	// A gear planner never runs an encounter, so only the sections that change the character's stats are shown:
	// no encounter, debuffs or saved encounters, and no potions or explosives.
	const gearPlanner = host.simDisabled;

	return (
		<OpenSelectorModalContext value={selector.openTab}>
			<TabPanelColumns.Left variant="settings-columns">
				<TabPanelColumns.Col>
					{ready && (
						<>
							{!gearPlanner && (
								<ContentBlock
									rootDataAttributes={{ 'data-block': 'encounter-settings' }}
									config={{ header: { title: i18n.t('settings_tab.encounter.title') } }}>
									<EncounterPicker showExecuteProportion={config.encounterPicker.showExecuteProportion} />
								</ContentBlock>
							)}
							<ContentBlock
								rootDataAttributes={{ 'data-block': 'player-settings' }}
								config={{ header: { title: i18n.t('settings_tab.player.title') } }}>
								<PlayerSettings iconInputs={config.playerIconInputs} inputs={config.playerInputs?.inputs ?? []} />
							</ContentBlock>
						</>
					)}
				</TabPanelColumns.Col>
				<TabPanelColumns.Col>
					{ready && (
						<>
							{config.sections?.map(section => (
								<CustomSection key={section.id} section={section} />
							))}
							<ContentBlock
								rootDataAttributes={{ 'data-block': 'consumes-settings' }}
								config={{ header: { title: i18n.t('settings_tab.consumables.title') } }}
								headerChildren={<ConsumesFilters />}>
								<ConsumesPicker encounterConsumes={!gearPlanner} />
							</ContentBlock>
							{hasOtherSettings && (
								<ContentBlock
									rootDataAttributes={{ 'data-block': 'other-settings' }}
									config={{ header: { title: i18n.t('settings_tab.other.title') } }}>
									<OtherSettings inputs={config.otherInputs.inputs} itemSlots={itemSwapSlots} />
								</ContentBlock>
							)}
						</>
					)}
				</TabPanelColumns.Col>
				<TabPanelColumns.Col>
					{ready && (
						<>
							<BuffSection section="raid_buffs" empty={options.buffs.length === 0}>
								<RaidBuffs options={options.buffs} miscOptions={[]} />
							</BuffSection>
							{options.partyBuffs.length > 0 && (
								<BuffSection section="party_buffs" empty={false}>
									<RaidBuffs options={options.partyBuffs} miscOptions={[]} />
								</BuffSection>
							)}
							{!gearPlanner && (
								<BuffSection section="debuffs" empty={options.debuffs.length === 0}>
									<RaidBuffs options={options.debuffs} miscOptions={options.debuffsMisc} />
								</BuffSection>
							)}
						</>
					)}
				</TabPanelColumns.Col>
			</TabPanelColumns.Left>
			<TabPanelColumns.Right>
				<PresetConfigurationPicker categories={gearPlanner ? GEAR_PLANNER_PRESETS : SETTINGS_PRESETS} />
				{!gearPlanner && <SavedEncounter />}
				<SavedSettings />
			</TabPanelColumns.Right>
			<SelectorModal state={selector} id="item-swap-selector-modal" rail={false} />
		</OpenSelectorModalContext>
	);
};
