import type { ChoiceFieldConfig, NumberFieldConfig } from '@features/inputs/configs';
import type { InputConfig } from '@features/inputs/input';
import type { Encounter as EncounterProto, PresetEncounter } from '@generated/proto/common';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import type { Encounter } from '@sim/raid/encounter';
import type { Raid } from '@sim/raid/raid';
import type { ChoiceEntry } from '@ui-kit/ChoiceField';
import type { SavedDataEntry } from '@ui-kit/hooks/useSavedData';

import { trackEvent } from '../../../../../tracking/utils';

interface Field {
	id: string;
	key: string;
	label: string;
	category: 'duration' | 'execute';
	get: (encounter: Encounter) => number;
	set: (encounter: Encounter, value: number) => void;
}

const field = (spec: Field): NumberFieldConfig<Encounter> => ({
	id: spec.id,
	label: i18n.t(`settings_tab.encounter.${spec.key}.label`),
	labelTooltip: i18n.t(`settings_tab.encounter.${spec.key}.tooltip`),
	storeField: 'encounter:*',
	getValue: spec.get,
	setValue: (subject, newValue) => {
		trackEvent({ action: 'settings', category: spec.category, label: spec.label, value: newValue });
		spec.set(subject, newValue);
	},
});

export const durationConfigs = (): Array<NumberFieldConfig<Encounter>> =>
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
	).map(field);

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

export const executeConfigs = (): Array<NumberFieldConfig<Encounter>> => EXECUTE_BANDS.map(field);

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

const CUSTOM_ENCOUNTER = 'custom';
const presetChoice = (index: number) => `preset:${index}`;
const ownChoice = (name: string) => `own:${name}`;

// The choice reads the store alone, never `own`: the input caches it until the next store write, so a
// list captured here would answer one write late. An entry of My encounters is the option of its name,
// listed or not; any other named encounter (an import) is no option, so the field shows its placeholder;
// an unnamed one is the preset it matches.
export const encounterChoiceConfig = (
	encounter: Encounter,
	own: ReadonlyArray<SavedDataEntry<EncounterProto>>,
	field: Pick<InputConfig<Encounter, string>, 'extraClassNames'> & { id: string },
): InputConfig<Encounter, string> & { id: string } => {
	const presets = encounter.sim.db.getAllPresetEncounters();
	return {
		...field,
		label: i18n.t('settings_tab.encounter.encounter_preset.label'),
		storeField: 'encounter:*',
		getValue: (subject: Encounter) => {
			if (subject.getName()) return subject.isMyEncounter() ? ownChoice(subject.getName()) : '';
			const index = presets.findIndex(preset => subject.matchesPreset(preset));
			return index === -1 ? CUSTOM_ENCOUNTER : presetChoice(index);
		},
		setValue: (subject: Encounter, choice: string) => {
			const preset = presets.findIndex((_, index) => presetChoice(index) === choice);
			if (preset !== -1) {
				trackEvent({ action: 'settings', category: 'encounter', label: 'preset', value: presets[preset].path });
				return subject.applyPreset(presets[preset]);
			}
			const entry = own.find(entry => ownChoice(entry.name) === choice);
			if (entry) subject.fromProto({ ...entry.data, myEncounter: true });
		},
	};
};

export const encounterChoices = (presets: ReadonlyArray<PresetEncounter>, own: ReadonlyArray<SavedDataEntry<EncounterProto>>): Array<ChoiceEntry<string>> => [
	{ value: CUSTOM_ENCOUNTER, label: i18n.t('common.custom') },
	{
		label: i18n.t('settings_tab.encounter.encounter_preset.presets'),
		options: presets.map((preset, index) => ({ value: presetChoice(index), label: preset.path })),
	},
	...(own.length
		? [{ label: i18n.t('settings_tab.encounter.encounter_preset.own'), options: own.map(entry => ({ value: ownChoice(entry.name), label: entry.name })) }]
		: []),
];

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
