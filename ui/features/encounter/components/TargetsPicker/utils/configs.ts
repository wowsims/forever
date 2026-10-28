import type { BooleanPickerConfig, ChoiceFieldConfig, NumberFieldConfig } from '@features/inputs/configs';
import { CreatureFamily, InputType, MobType, SpellSchool, Stat, Target as TargetProto, TargetInput } from '@generated/proto/common';
import i18n from '@i18n/config';
import {
	translateCreatureFamily,
	translateMobType,
	translateSpellSchool,
	translateStat,
	translateTargetInputLabel,
	translateTargetInputTooltip,
} from '@i18n/localization';
import type { Encounter } from '@sim/raid/encounter';
import { type StoreSubscribe, subscribeEncounterField } from '@sim/state/subscriptions';
import { distinct } from '@sim/utils/collections';
import { randomUUID } from '@sim/utils/misc';

import { trackEvent, type TrackEventProps } from '../../../../../tracking/utils';

export const ALL_TARGET_STATS: Array<{ stat: Stat; tooltip: string; threatMetrics?: boolean }> = [
	{ stat: Stat.StatHealth, tooltip: '' },
	{ stat: Stat.StatArmor, tooltip: '' },
	{ stat: Stat.StatArcaneResistance, tooltip: '' },
	{ stat: Stat.StatFireResistance, tooltip: '' },
	{ stat: Stat.StatFrostResistance, tooltip: '' },
	{ stat: Stat.StatNatureResistance, tooltip: '' },
	{ stat: Stat.StatShadowResistance, tooltip: '' },
	{ stat: Stat.StatAttackPower, tooltip: '', threatMetrics: true },
	{ stat: Stat.StatBlockValue, tooltip: '', threatMetrics: true },
];

const mobTypeEnumValues = [
	MobType.MobTypeUnknown,
	MobType.MobTypeBeast,
	MobType.MobTypeDemon,
	MobType.MobTypeDragonkin,
	MobType.MobTypeElemental,
	MobType.MobTypeGiant,
	MobType.MobTypeHumanoid,
	MobType.MobTypeMechanical,
	MobType.MobTypeUndead,
].map(value => ({ name: translateMobType(value), value }));

// The families each mob type holds; Aquatic spans beasts, elementals and humanoids.
const creatureFamiliesByMobType: Partial<Record<MobType, ReadonlyArray<CreatureFamily>>> = {
	[MobType.MobTypeBeast]: [CreatureFamily.CreatureFamilyAquatic],
	[MobType.MobTypeDemon]: [CreatureFamily.CreatureFamilySatyr],
	[MobType.MobTypeElemental]: [CreatureFamily.CreatureFamilyEarthElemental, CreatureFamily.CreatureFamilyFireElemental, CreatureFamily.CreatureFamilyAquatic],
	[MobType.MobTypeGiant]: [CreatureFamily.CreatureFamilyMountainGiant],
	[MobType.MobTypeHumanoid]: [CreatureFamily.CreatureFamilyNaga, CreatureFamily.CreatureFamilyAquatic],
};

export const creatureFamiliesOf = (mobType: MobType): ReadonlyArray<CreatureFamily> => creatureFamiliesByMobType[mobType] ?? [];

const spellSchoolValues = [
	SpellSchool.SpellSchoolPhysical,
	SpellSchool.SpellSchoolArcane,
	SpellSchool.SpellSchoolFire,
	SpellSchool.SpellSchoolFrost,
	SpellSchool.SpellSchoolHoly,
	SpellSchool.SpellSchoolNature,
	SpellSchool.SpellSchoolShadow,
].map(value => ({ name: translateSpellSchool(value), value }));

const targetId = (idPrefix: string, name: string) => `${idPrefix}-picker-${name}`;

const equalTargetsIgnoreInputs = (target1: TargetProto | undefined, target2: TargetProto | undefined): boolean => {
	if (!!target1 !== !!target2) return false;
	if (!target1) return true;
	const modTarget2 = TargetProto.clone(target2!);
	modTarget2.targetInputs = target1.targetInputs;
	return TargetProto.equals(target1, modTarget2);
};

/** What the field configs read and write: one target, wherever it lives. */
export interface TargetFieldContext {
	/** Starts every field id, so two pickers on a page never share one. */
	idPrefix: string;
	getTarget: () => TargetProto;
	modifyTarget: (edit: (target: TargetProto) => void) => void;
	storeSubscribe: () => StoreSubscribe;
}

/** An encounter's target, which the preset and AI fields also need the encounter for. */
export interface EncounterTargetFieldContext extends TargetFieldContext {
	encounter: Encounter;
	targetIndex: number;
}

export const encounterTargetContext = (encounter: Encounter, targetIndex: number): EncounterTargetFieldContext => ({
	encounter,
	targetIndex,
	idPrefix: `target-${targetIndex}`,
	getTarget: () => encounter.getTarget(targetIndex) || TargetProto.create(),
	modifyTarget: edit => encounter.modifyTarget(targetIndex, edit),
	storeSubscribe: () => subscribeEncounterField(encounter, 'targets'),
});

export const npcConfig = ({ encounter, targetIndex, idPrefix, getTarget, storeSubscribe }: EncounterTargetFieldContext): ChoiceFieldConfig<null> => {
	const presetTargets = encounter.sim.db.getAllPresetTargets();
	return {
		id: targetId(idPrefix, 'npc'),
		label: i18n.t('settings_tab.encounter.npc.label'),
		labelTooltip: i18n.t('settings_tab.encounter.npc.tooltip'),
		values: [{ name: i18n.t('common.custom'), value: -1 }].concat(presetTargets.map((preset, index) => ({ name: preset.path, value: index }))),
		storeSubscribe,
		getValue: () => presetTargets.findIndex(preset => equalTargetsIgnoreInputs(getTarget(), preset.target)),
		setValue: (_: null, newValue: number) => {
			if (newValue === -1) return;
			const preset = presetTargets[newValue];
			trackEvent({ action: 'settings', category: 'targets', label: 'preset', value: preset.target?.name || preset.path });
			encounter.applyPresetTarget(preset, targetIndex);
		},
	};
};

export const aiConfig = ({ encounter, idPrefix, getTarget, storeSubscribe, modifyTarget }: EncounterTargetFieldContext): ChoiceFieldConfig<null> => {
	const presetTargets = encounter.sim.db.getAllPresetTargets();
	const byNpcId = distinct(presetTargets, (a, b) => a.target?.id === b.target?.id);
	return {
		id: targetId(idPrefix, 'ai'),
		label: i18n.t('settings_tab.encounter.ai.label'),
		labelTooltip: i18n.t('settings_tab.encounter.ai.tooltip'),
		values: [{ name: i18n.t('common.none'), value: 0 }].concat(byNpcId.map(preset => ({ name: preset.path, value: preset.target!.id }))),
		storeSubscribe,
		getValue: () => getTarget().id,
		setValue: (_: null, newValue: number) => {
			modifyTarget(target => {
				target.id = newValue;
				trackEvent({ action: 'settings', category: 'targets', label: 'ai', value: target.name });
				target.targetInputs = (presetTargets.find(preset => target.id === preset.target?.id)?.target?.targetInputs || []).map(input =>
					TargetInput.clone(input),
				);
			});
		},
	};
};

export const levelConfig = ({ idPrefix, getTarget, storeSubscribe, modifyTarget }: TargetFieldContext): ChoiceFieldConfig<null> => ({
	id: targetId(idPrefix, 'level'),
	label: i18n.t('settings_tab.encounter.level'),
	values: [63, 62, 61, 60].map(value => ({ name: String(value), value })),
	storeSubscribe,
	getValue: () => getTarget().level,
	setValue: (_: null, newValue: number) => {
		trackEvent({ action: 'settings', category: 'targets', label: 'level', value: newValue });
		modifyTarget(target => {
			target.level = newValue;
		});
	},
});

export const mobTypeConfig = ({ idPrefix, getTarget, storeSubscribe, modifyTarget }: TargetFieldContext): ChoiceFieldConfig<null> => ({
	id: targetId(idPrefix, 'mob-type'),
	label: i18n.t('settings_tab.encounter.mob_type'),
	values: mobTypeEnumValues,
	storeSubscribe,
	getValue: () => getTarget().mobType,
	setValue: (_: null, newValue: number) => {
		trackEvent({ action: 'settings', category: 'targets', label: 'mob_type', value: newValue });
		modifyTarget(target => {
			target.mobType = newValue;
			if (!creatureFamiliesOf(newValue).includes(target.creatureFamily)) {
				target.creatureFamily = CreatureFamily.CreatureFamilyUnknown;
			}
		});
	},
});

export const creatureFamilyConfig = ({ idPrefix, getTarget, storeSubscribe, modifyTarget }: TargetFieldContext, mobType: MobType): ChoiceFieldConfig<null> => ({
	id: targetId(idPrefix, 'creature-family'),
	label: i18n.t('settings_tab.encounter.creature_family.label'),
	labelTooltip: i18n.t('settings_tab.encounter.creature_family.tooltip'),
	values: [CreatureFamily.CreatureFamilyUnknown, ...creatureFamiliesOf(mobType)].map(value => ({ name: translateCreatureFamily(value), value })),
	storeSubscribe,
	getValue: () => getTarget().creatureFamily,
	setValue: (_: null, newValue: number) => {
		trackEvent({ action: 'settings', category: 'targets', label: 'creature_family', value: newValue });
		modifyTarget(target => {
			target.creatureFamily = newValue;
		});
	},
});

export const tankIndexConfig = ({ idPrefix, getTarget, storeSubscribe, modifyTarget }: TargetFieldContext): ChoiceFieldConfig<null> => ({
	id: targetId(idPrefix, 'tanked-by'),
	label: i18n.t('settings_tab.encounter.tanked_by.label'),
	labelTooltip: i18n.t('settings_tab.encounter.tanked_by.tooltip'),
	values: [
		{ name: i18n.t('common.none'), value: -1 },
		{ name: i18n.t('common.tanks.main_tank'), value: 0 },
		{ name: i18n.t('common.tanks.tank_2'), value: 1 },
		{ name: i18n.t('common.tanks.tank_3'), value: 2 },
		{ name: i18n.t('common.tanks.tank_4'), value: 3 },
	],
	storeSubscribe,
	getValue: () => getTarget().tankIndex,
	setValue: (_: null, newValue: number) => {
		trackEvent({ action: 'settings', category: 'targets', label: 'tank_index', value: newValue });
		modifyTarget(target => {
			target.tankIndex = newValue;
		});
	},
});

export const statConfig = ({ idPrefix, getTarget, storeSubscribe, modifyTarget }: TargetFieldContext, stat: Stat, tooltip: string) =>
	({
		id: `${idPrefix}-picker-stats-${stat}`,
		label: translateStat(stat),
		labelTooltip: tooltip,
		storeSubscribe,
		getValue: () => getTarget().stats[stat],
		setValue: (_: null, newValue: number) => {
			modifyTarget(target => {
				target.stats[stat] = newValue;
			});
		},
	}) satisfies NumberFieldConfig<null>;

interface NumberField {
	name: string;
	key: string;
	label: string;
	float?: boolean;
	get: (target: TargetProto) => number;
	set: (target: TargetProto, value: number) => void;
}

const NUMBER_FIELDS: Array<NumberField> = [
	{
		name: 'swing-speed',
		key: 'swing_speed',
		label: 'swing_speed',
		float: true,
		get: target => target.swingSpeed,
		set: (target, value) => {
			target.swingSpeed = value;
		},
	},
	{
		name: 'min-base-damage',
		key: 'min_base_damage',
		label: 'min_base_damage',
		get: target => target.minBaseDamage,
		set: (target, value) => {
			target.minBaseDamage = value;
		},
	},
	{
		name: 'damage-spread',
		key: 'damage_spread',
		label: 'damage_spread',
		float: true,
		get: target => target.damageSpread,
		set: (target, value) => {
			target.damageSpread = value;
		},
	},
];

export const numberConfigs = ({ idPrefix, getTarget, storeSubscribe, modifyTarget }: TargetFieldContext): Array<NumberFieldConfig<null>> =>
	NUMBER_FIELDS.map(field => ({
		id: targetId(idPrefix, field.name),
		label: i18n.t(`settings_tab.encounter.${field.key}.label`),
		labelTooltip: i18n.t(`settings_tab.encounter.${field.key}.tooltip`),
		float: field.float,
		storeSubscribe,
		getValue: () => field.get(getTarget()),
		setValue: (_: null, newValue: number) => {
			trackEvent({ action: 'settings', category: 'targets', label: field.label, value: newValue });
			modifyTarget(target => field.set(target, newValue));
		},
	}));

interface BooleanField {
	name: string;
	key: string;
	label: string;
	get: (target: TargetProto) => boolean;
	set: (target: TargetProto, value: boolean) => void;
	enabledWhen?: (target: TargetProto) => boolean;
}

const BOOLEAN_FIELDS: Array<BooleanField> = [
	{
		name: 'dual-wield',
		key: 'dual_wield',
		label: 'dual_wield',
		get: target => target.dualWield,
		set: (target, value) => {
			target.dualWield = value;
		},
	},
	{
		name: 'dw-miss-penalty',
		key: 'dual_wield_penalty',
		label: 'dual_wield_penalty',
		get: target => target.dualWieldPenalty,
		set: (target, value) => {
			target.dualWieldPenalty = value;
		},
		enabledWhen: target => target.dualWield,
	},
	{
		name: 'parry-haste',
		key: 'parry_haste',
		label: 'parry_haste',
		get: target => target.parryHaste,
		set: (target, value) => {
			target.parryHaste = value;
		},
	},
	{
		name: 'can-crush',
		key: 'can_crush',
		label: 'can_crush',
		get: target => target.canCrush,
		set: (target, value) => {
			target.canCrush = value;
		},
	},
];

export const booleanConfigs = ({ idPrefix, getTarget, storeSubscribe, modifyTarget }: TargetFieldContext): Array<BooleanPickerConfig<null>> =>
	BOOLEAN_FIELDS.map(field => ({
		id: targetId(idPrefix, field.name),
		label: i18n.t(`settings_tab.encounter.${field.key}.label`),
		labelTooltip: i18n.t(`settings_tab.encounter.${field.key}.tooltip`),
		layout: 'inline',
		reverse: true,
		storeSubscribe,
		getValue: () => field.get(getTarget()),
		setValue: (_: null, newValue: boolean) => {
			trackEvent({ action: 'settings', category: 'targets', label: field.label, value: newValue });
			modifyTarget(target => field.set(target, newValue));
		},
		...(field.enabledWhen ? { enableWhen: () => field.enabledWhen!(getTarget()) } : {}),
	}));

export const spellSchoolConfig = ({ idPrefix, getTarget, storeSubscribe, modifyTarget }: TargetFieldContext): ChoiceFieldConfig<null> => ({
	id: targetId(idPrefix, 'spell-school'),
	label: i18n.t('settings_tab.encounter.spell_school.label'),
	labelTooltip: i18n.t('settings_tab.encounter.spell_school.tooltip'),
	values: spellSchoolValues,
	storeSubscribe,
	getValue: () => getTarget().spellSchool,
	setValue: (_: null, newValue: number) => {
		trackEvent({ action: 'settings', category: 'targets', label: 'spell_school', value: newValue });
		modifyTarget(target => {
			target.spellSchool = newValue;
		});
	},
});

export const targetInputConfig = (encounter: Encounter, targetIndex: number, inputIndex: number, input: TargetInput) => {
	const getTargetInput = () => encounter.getTarget(targetIndex)!.targetInputs[inputIndex] || TargetInput.create();
	const tracking: TrackEventProps = { action: 'settings', category: 'targets', label: input.label };
	const write =
		<T>(apply: (target: TargetInput, value: T) => void) =>
		(_: null, newValue: T) => {
			trackEvent({ ...tracking, value: newValue as string | number | boolean });
			encounter.modifyTarget(targetIndex, target => {
				const draft = target.targetInputs[inputIndex];
				if (draft) apply(draft, newValue);
			});
		};
	const shared = {
		id: randomUUID(),
		label: translateTargetInputLabel(input.label),
		labelTooltip: translateTargetInputTooltip(input.label, input.tooltip),
		storeSubscribe: () => subscribeEncounterField(encounter, 'targets'),
	};

	if (input.inputType === InputType.Number) {
		return {
			kind: 'number' as const,
			config: {
				...shared,
				float: true,
				getValue: () => getTargetInput().numberValue,
				setValue: write<number>((draft, value) => {
					draft.numberValue = value;
				}),
			} satisfies NumberFieldConfig<null>,
		};
	}
	if (input.inputType === InputType.Bool) {
		return {
			kind: 'boolean' as const,
			config: {
				...shared,
				layout: 'inline',
				getValue: () => getTargetInput().boolValue,
				setValue: write<boolean>((draft, value) => {
					draft.boolValue = value;
				}),
			} satisfies BooleanPickerConfig<null>,
		};
	}
	if (input.inputType === InputType.Enum) {
		return {
			kind: 'enum' as const,
			config: {
				...shared,
				labelTooltip: undefined,
				values: input.enumOptions.map((option, index) => ({ value: index, name: option })),
				getValue: () => getTargetInput().enumValue,
				setValue: write<number>((draft, value) => {
					draft.enumValue = value;
				}),
			} satisfies ChoiceFieldConfig<null>,
		};
	}
	return { kind: 'none' as const };
};
