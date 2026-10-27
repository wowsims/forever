import { EncounterPicker, SavedEncounter } from '@features/encounter';
import { SelectorModal } from '@features/gear/components/SelectorModal';
import { OpenSelectorModalContext, useSelectorModalState } from '@features/gear/hooks/useSelectorModal';
import { inputPicked } from '@features/inputs/components/ConnectedIconEnumPicker';
import { BuffSection, ConsumesFilters, ConsumesPicker, CustomSection, OtherSettings, PlayerSettings, RaidBuffs, SavedSettings } from '@features/settings';
import * as BuffDebuffInputs from '@features/settings/model/buffs_debuffs';
import { applyOwnerClassLabels, relevantStatOptions, type RenderableStatOptions, statOptionScope } from '@features/settings/model/stat_options';
import { StatFilterPopover, useStatFilter } from '@features/stat-filters';
import i18n from '@i18n/config';
import { PresetConfigurationCategory } from '@sim/constants/preset_categories';
import { usePlayer, useSimHost, useSpecConfig } from '@sim/context/SimHostContext';
import { useSimReady } from '@sim/hooks/useSimReady';
import type { Player } from '@sim/player/player';
import { statSetsOf } from '@sim/proto/stats';
import { memoLast } from '@sim/utils/misc';
import { ContentBlock } from '@ui-kit/ContentBlock';
import { TabPanelColumns } from '@ui-kit/TabPanelColumns';
import { useMemo, useState } from 'react';
import { useStore } from 'zustand';

import { PresetConfigurationPicker } from '../PresetConfigurationPicker';

const SETTINGS_PRESETS = [PresetConfigurationCategory.Encounter, PresetConfigurationCategory.Settings];
const GEAR_PLANNER_PRESETS = [PresetConfigurationCategory.Settings];
const BUFF_OPTIONS = [...BuffDebuffInputs.BUFFS_CONFIG, ...BuffDebuffInputs.PARTY_BUFFS_CONFIG];
const DEBUFF_OPTIONS = [...BuffDebuffInputs.DEBUFFS_CONFIG, ...BuffDebuffInputs.DEBUFFS_MISC_CONFIG];

const anyOffered = (list: ReadonlyArray<RenderableStatOptions>, player: Player<any>) =>
	list.some(option => !option.config.keptBy || inputPicked(option.config, player));

// Whether any option is offered, worked out again only when the raid slice, the player's slice or the list changes. An
// option's value is read from those two slices, so a write to any other slice returns the cached flag.
const useAnyOffered = (list: ReadonlyArray<RenderableStatOptions>, player: Player<any>) => {
	const [offered] = useState(() => memoLast<boolean>());
	return useStore(player.sim.store, state => offered([state.raid, state.players[player.storeKey], list], () => anyOffered(list, player)));
};

// The buff, party buff and debuff sections, apart from the rest of the tab so a filter write re-renders only them.
const BuffSections = ({ gearPlanner }: { gearPlanner: boolean }) => {
	const config = useSpecConfig();
	const player = usePlayer();

	const buffScope = useMemo(() => statOptionScope(BUFF_OPTIONS, config), [config]);
	const debuffScope = useMemo(() => statOptionScope(DEBUFF_OPTIONS, config), [config]);
	const buffFilter = useStatFilter('buffs', buffScope);
	const debuffFilter = useStatFilter('debuffs', debuffScope);

	const options = useMemo(() => {
		const buffStats = statSetsOf(buffFilter.shown).stats;
		const debuffStats = statSetsOf(debuffFilter.shown).stats;
		return {
			buffs: applyOwnerClassLabels(relevantStatOptions(BuffDebuffInputs.BUFFS_CONFIG, config, buffStats), player),
			partyBuffs: applyOwnerClassLabels(relevantStatOptions(BuffDebuffInputs.PARTY_BUFFS_CONFIG, config, buffStats), player),
			debuffs: applyOwnerClassLabels(relevantStatOptions(BuffDebuffInputs.DEBUFFS_CONFIG, config, debuffStats), player),
			debuffsMisc: applyOwnerClassLabels(relevantStatOptions(BuffDebuffInputs.DEBUFFS_MISC_CONFIG, config, debuffStats), player),
		};
	}, [config, player, buffFilter.shown, debuffFilter.shown]);

	// A kept option counts as offered only while it is picked, so each flag is re-read when the raid or the player
	// changes and the tab re-renders only when one flips.
	const offersBuffs = useAnyOffered(options.buffs, player);
	const offersPartyBuffs = useAnyOffered(options.partyBuffs, player);
	const offersDebuffs = useAnyOffered(options.debuffs, player);

	return (
		<>
			<BuffSection
				section="raid_buffs"
				empty={!offersBuffs}
				filter={<StatFilterPopover {...buffFilter} ariaLabel={i18n.t('stat_filters.triggers.buffs')} testId="buffs-filter" />}>
				<RaidBuffs options={options.buffs} miscOptions={[]} />
			</BuffSection>
			{offersPartyBuffs && (
				<BuffSection section="party_buffs" empty={false}>
					<RaidBuffs options={options.partyBuffs} miscOptions={[]} />
				</BuffSection>
			)}
			{!gearPlanner && !config.hideDebuffs && (
				<BuffSection
					section="debuffs"
					empty={!offersDebuffs}
					filter={<StatFilterPopover {...debuffFilter} ariaLabel={i18n.t('stat_filters.triggers.debuffs')} testId="debuffs-filter" />}>
					<RaidBuffs options={options.debuffs} miscOptions={options.debuffsMisc} />
				</BuffSection>
			)}
		</>
	);
};

export const SettingsTabBody = () => {
	const host = useSimHost();
	const config = useSpecConfig();
	const ready = useSimReady();
	const SettingsComponentUnderPlayer = config.settingsComponentUnderPlayer;
	const healingInputs = config.healingInputs?.inputs ?? [];

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
							{!gearPlanner && SettingsComponentUnderPlayer && <SettingsComponentUnderPlayer />}
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
								<ConsumesPicker encounterConsumes={true} />
							</ContentBlock>
							{hasOtherSettings && (
								<ContentBlock
									rootDataAttributes={{ 'data-block': 'other-settings' }}
									config={{ header: { title: i18n.t('settings_tab.other.title') } }}>
									<OtherSettings inputs={config.otherInputs.inputs} itemSlots={itemSwapSlots} />
								</ContentBlock>
							)}
							{!gearPlanner && healingInputs.length > 0 && (
								<ContentBlock
									rootDataAttributes={{ 'data-block': 'healing-inputs' }}
									config={{ header: { title: i18n.t('settings_tab.healing_inputs.title') } }}>
									<OtherSettings inputs={healingInputs} itemSlots={[]} />
								</ContentBlock>
							)}
						</>
					)}
				</TabPanelColumns.Col>
				<TabPanelColumns.Col testId="buff-settings">{ready && <BuffSections gearPlanner={gearPlanner} />}</TabPanelColumns.Col>
			</TabPanelColumns.Left>
			<TabPanelColumns.Right testId="settings-presets">
				<PresetConfigurationPicker categories={gearPlanner ? GEAR_PLANNER_PRESETS : SETTINGS_PRESETS} />
				{!gearPlanner && <SavedEncounter />}
				<SavedSettings />
			</TabPanelColumns.Right>
			<SelectorModal state={selector} id="item-swap-selector-modal" rail={false} />
		</OpenSelectorModalContext>
	);
};
