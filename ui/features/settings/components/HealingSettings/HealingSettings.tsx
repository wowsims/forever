import { HealingSimSettings, HealingSimSettings_Assignment as HealingAssignment } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { useSim } from '@sim/context/SimHostContext';
import { batch } from '@sim/state/batch';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { ContentBlock } from '@ui-kit/ContentBlock';
import { useEffect } from 'react';
import { useStore } from 'zustand';

const GROUP_SIZES = [5, 10, 20] as const;

const targetDummyCountFor = (assignment: HealingAssignment, raidSize: number): number =>
	raidSize === 5 ? 4 : assignment === HealingAssignment.TankHealing ? (raidSize === 20 ? 2 : 1) : raidSize - 1;

export const HealingSettings = () => {
	const sim = useSim();
	const healingSettings = useStore(sim.store, state => state.raid.healingSettings);
	const targetDummies = useStore(sim.store, state => state.raid.targetDummies);
	const assignment = healingSettings.assignment ?? HealingAssignment.TankHealing;
	const raidSize = healingSettings.groupSize === 5 ? 5 : healingSettings.groupSize === 20 ? 20 : 10;

	const applyHealingSettings = (settings: HealingSimSettings) => {
		const nextAssignment = settings.assignment ?? HealingAssignment.TankHealing;
		const nextRaidSize = settings.groupSize === 5 ? 5 : settings.groupSize === 20 ? 20 : 10;
		batch(() => {
			sim.raid.setHealingSettings(settings);
			sim.raid.setTargetDummies(targetDummyCountFor(nextAssignment, nextRaidSize));
		});
	};

	useEffect(() => {
		const count = targetDummyCountFor(assignment, raidSize);
		if (targetDummies !== count) sim.raid.setTargetDummies(count);
	}, [assignment, raidSize, sim, targetDummies]);

	const setAssignment = (value: HealingAssignment) => applyHealingSettings({ ...healingSettings, assignment: value });
	const setRaidSize = (value: number) => applyHealingSettings({ ...healingSettings, groupSize: value });
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
