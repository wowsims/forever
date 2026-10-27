import { healingGroupSize } from '@features/encounter/model/target_dummies';
import { HealingSimSettings_Assignment as HealingAssignment } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { useSim } from '@sim/context/SimHostContext';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { ContentBlock } from '@ui-kit/ContentBlock';
import { useStore } from 'zustand';

const GROUP_SIZES = [5, 10, 20] as const;

export const HealingSettings = () => {
	const sim = useSim();
	const healingSettings = useStore(sim.store, state => state.raid.healingSettings);
	const assignment = healingSettings.assignment;
	const raidSize = healingGroupSize(healingSettings);
	const setAssignment = (value: HealingAssignment) => sim.raid.setHealingSettings({ ...healingSettings, assignment: value });
	const setRaidSize = (value: number) => sim.raid.setHealingSettings({ ...healingSettings, groupSize: value });
	const assignmentOptions = [
		{ value: HealingAssignment.TankHealing, label: i18n.t('settings_tab.healing_settings.tank_healing') },
		{ value: HealingAssignment.RaidHealing, label: i18n.t('settings_tab.healing_settings.raid_healing') },
	];
	const raidSizeOptions = GROUP_SIZES.map(size => ({
		value: size,
		label: i18n.t(
			size === 5
				? 'settings_tab.healing_settings.dungeon_group'
				: size === 10
					? 'settings_tab.healing_settings.ten_player'
					: 'settings_tab.healing_settings.twenty_player',
		),
	}));

	return (
		<ContentBlock rootDataAttributes={{ 'data-block': 'healing-settings' }} config={{ header: { title: i18n.t('settings_tab.healing_settings.title') } }}>
			<div className="flex flex-col gap-4">
				<ChoiceField
					label={i18n.t('settings_tab.healing_settings.group_size')}
					value={raidSize}
					options={raidSizeOptions}
					onChange={setRaidSize}
					testId="healing-group-size"
				/>
				{raidSize !== 5 && (
					<ChoiceField
						label={i18n.t('settings_tab.healing_settings.assignment')}
						value={assignment}
						options={assignmentOptions}
						onChange={setAssignment}
						testId="healing-assignment"
					/>
				)}
			</div>
		</ContentBlock>
	);
};
