import { healingGroupSize } from '@features/encounter/model/target_dummies';
import { HealingSimSettings_Assignment as HealingAssignment } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { Player } from '@sim/player/player';
import type { CustomSection } from '@sim/spec_config';

export const HealingGroupSize = {
	id: 'healing-group-size',
	type: 'enum' as const,
	label: i18n.t('settings_tab.healing_settings.group_size'),
	values: [
		{ name: i18n.t('settings_tab.healing_settings.dungeon_group'), value: 5 },
		{ name: i18n.t('settings_tab.healing_settings.ten_player'), value: 10 },
		{ name: i18n.t('settings_tab.healing_settings.twenty_player'), value: 20 },
	],
	storeField: 'raid:healingSettings' as const,
	getValue: (player: Player<any>) => healingGroupSize(player.getRaid()!.getHealingSettings()),
	setValue: (player: Player<any>, newValue: number) => {
		const raid = player.getRaid()!;
		raid.setHealingSettings({ ...raid.getHealingSettings(), groupSize: newValue });
	},
};

// A dungeon group's heal targets do not depend on the assignment, so it shows for raids only.
export const HealingAssignmentInput = {
	id: 'healing-assignment',
	type: 'enum' as const,
	label: i18n.t('settings_tab.healing_settings.assignment'),
	values: [
		{ name: i18n.t('settings_tab.healing_settings.tank_healing'), value: HealingAssignment.TankHealing },
		{ name: i18n.t('settings_tab.healing_settings.raid_healing'), value: HealingAssignment.RaidHealing },
	],
	storeField: 'raid:healingSettings' as const,
	showWhen: (player: Player<any>) => healingGroupSize(player.getRaid()!.getHealingSettings()) !== 5,
	getValue: (player: Player<any>) => player.getRaid()!.getHealingSettings().assignment,
	setValue: (player: Player<any>, newValue: HealingAssignment) => {
		const raid = player.getRaid()!;
		raid.setHealingSettings({ ...raid.getHealingSettings(), assignment: newValue });
	},
};

export const HealingSettingsSection: CustomSection<any> = {
	id: 'healing-settings',
	title: i18n.t('settings_tab.healing_settings.title'),
	inputs: [HealingGroupSize, HealingAssignmentInput],
};
