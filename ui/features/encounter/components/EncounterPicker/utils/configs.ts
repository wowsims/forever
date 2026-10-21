import type { ChoiceFieldConfig, NumberFieldConfig } from '@features/inputs/configs';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import type { Encounter } from '@sim/raid/encounter';
import type { Raid } from '@sim/raid/raid';

import { trackEvent } from '../../../../../tracking/utils';

// Every field here is off limits while the encounter runs to a health target rather than a clock.
const whileTimed = (encounter: Encounter) => () => !encounter.getUseHealth();

interface Field {
	id: string;
	key: string;
	label: string;
	category: 'duration' | 'execute';
	get: (encounter: Encounter) => number;
	set: (encounter: Encounter, value: number) => void;
}

const field = (encounter: Encounter, spec: Field): NumberFieldConfig<Encounter> => ({
	id: spec.id,
	label: i18n.t(`settings_tab.encounter.${spec.key}.label`),
	labelTooltip: i18n.t(`settings_tab.encounter.${spec.key}.tooltip`),
	storeField: 'encounter:*',
	getValue: spec.get,
	setValue: (subject, newValue) => {
		trackEvent({ action: 'settings', category: spec.category, label: spec.label, value: newValue });
		spec.set(subject, newValue);
	},
	enableWhen: whileTimed(encounter),
});

export const durationConfigs = (encounter: Encounter): Array<NumberFieldConfig<Encounter>> =>
	(
		[
			{ id: 'encounter-duration', key: 'duration', label: 'duration', category: 'duration', get: e => e.getDuration(), set: (e, v) => e.setDuration(v) },
			{
				id: 'encounter-duration-variation',
				key: 'duration_variation',
				label: 'variation',
				category: 'duration',
				get: e => e.getDurationVariation(),
				set: (e, v) => e.setDurationVariation(v),
			},
		] satisfies Array<Field>
	).map(spec => field(encounter, spec));

// Stored as a proportion, shown as a percentage.
const EXECUTE_BANDS: Array<Field> = [
	{
		id: 'encounter-execute-proportion',
		key: 'execute_duration_20',
		label: 'execute_20',
		category: 'execute',
		get: (e: Encounter) => e.getExecuteProportion20() * 100,
		set: (e: Encounter, v: number) => e.setExecuteProportion20(v / 100),
	},
	{
		id: 'encounter-execute-proportion-35',
		key: 'execute_duration_35',
		label: 'execute_35',
		category: 'execute',
		get: (e: Encounter) => e.getExecuteProportion35() * 100,
		set: (e: Encounter, v: number) => e.setExecuteProportion35(v / 100),
	},
];

export const executeConfigs = (encounter: Encounter): Array<NumberFieldConfig<Encounter>> => EXECUTE_BANDS.map(band => field(encounter, band));

export const numAlliesConfig = (player: Player<any>): NumberFieldConfig<Raid> => ({
	id: 'encounter-num-allies',
	label: i18n.t('settings_tab.encounter.num_allies.label'),
	labelTooltip: i18n.t('settings_tab.encounter.num_allies.tooltip'),
	storeField: ['raid:targetDummies', 'itemSwap'],
	getValue: (raid: Raid) => raid.getTargetDummies(),
	setValue: (raid: Raid, newValue: number) => raid.setTargetDummies(newValue),
	showWhen: () => player.shouldEnableTargetDummies(),
});

export const minBaseDamageConfig = (): NumberFieldConfig<Encounter> => ({
	id: 'encounter-min-base-damage',
	label: i18n.t('settings_tab.encounter.min_base_damage.label'),
	labelTooltip: i18n.t('settings_tab.encounter.min_base_damage.tooltip'),
	storeField: 'encounter:*',
	getValue: (encounter: Encounter) => encounter.primaryTarget.minBaseDamage,
	setValue: (encounter: Encounter, newValue: number) =>
		encounter.modifyTarget(0, target => {
			target.minBaseDamage = newValue;
		}),
});

export const presetEncounterConfig = (encounter: Encounter): ChoiceFieldConfig<Encounter> => {
	const presets = encounter.sim.db.getAllPresetEncounters();
	return {
		id: 'encounter-preset-encouter',
		label: i18n.t('settings_tab.encounter.encounter_preset.label'),
		extraClassNames: ['w-full'],
		values: [{ name: i18n.t('common.custom'), value: -1 }, ...presets.map((preset, index) => ({ name: preset.path, value: index }))],
		storeField: 'encounter:*',
		getValue: (subject: Encounter) => presets.findIndex(preset => subject.matchesPreset(preset)),
		setValue: (subject: Encounter, newValue: number) => {
			if (newValue !== -1) subject.applyPreset(presets[newValue]);
		},
	};
};

export const zoneConfig = (encounter: Encounter): ChoiceFieldConfig<Encounter> => {
	const zones = encounter.sim.db
		.getAllZones()
		.filter(zone => zone.areaTypes.length > 0)
		.sort((a, b) => a.name.localeCompare(b.name));
	return {
		id: 'encounter-zone',
		label: i18n.t('settings_tab.encounter.zone.label'),
		labelTooltip: i18n.t('settings_tab.encounter.zone.tooltip'),
		extraClassNames: ['w-full'],
		values: [{ name: i18n.t('common.custom'), value: 0 }, ...zones.map(zone => ({ name: zone.name, value: zone.id }))],
		storeField: 'encounter:*',
		getValue: (subject: Encounter) => subject.getZoneId(),
		setValue: (subject: Encounter, newValue: number) => {
			trackEvent({ action: 'settings', category: 'area', label: 'zone', value: newValue });
			subject.setZone(newValue);
		},
	};
};
