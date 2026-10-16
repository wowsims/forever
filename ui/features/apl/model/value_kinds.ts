import {
	APLValue,
	APLValueActionGroupUsed,
	APLValueActiveItemSwapSet,
	APLValueAllItemStatProcsActive,
	APLValueAnd,
	APLValueAnyItemStatProcsActive,
	APLValueAnyItemStatProcsAvailable,
	APLValueAnyStatBuffCooldownsActive,
	APLValueAnyStatBuffCooldownsMinDuration,
	APLValueAuraDuration,
	APLValueAuraICDIsReady,
	APLValueAuraInternalCooldown,
	APLValueAuraIsActive,
	APLValueAuraIsInactive,
	APLValueAuraIsKnown,
	APLValueAuraNumStacks,
	APLValueAuraRemainingTime,
	APLValueAuraShouldRefresh,
	APLValueAutoSwingTime,
	APLValueAutoTimeSinceLast,
	APLValueAutoTimeToNext,
	APLValueBossCurrentTarget,
	APLValueBossSpellIsCasting,
	APLValueBossSpellTimeToReady,
	APLValueChannelClipDelay,
	APLValueCompare,
	APLValueConst,
	APLValueCurrentComboPoints,
	APLValueCurrentEnergy,
	APLValueCurrentGenericResource,
	APLValueCurrentHealth,
	APLValueCurrentHealthPercent,
	APLValueCurrentMana,
	APLValueCurrentManaPercent,
	APLValueCurrentManaRegen,
	APLValueCurrentRage,
	APLValueCurrentTime,
	APLValueCurrentTimePercent,
	APLValueDotBaseDuration,
	APLValueDotIsActive,
	APLValueDotIsActiveOnAllTargets,
	APLValueDotLowestRemainingTime,
	APLValueDotPercentIncrease,
	APLValueDotRemainingTime,
	APLValueDotTickFrequency,
	APLValueDotTimeToNextTick,
	APLValueDruidBleedIsActive,
	APLValueEncounterInArea,
	APLValueEnergyTimeToTarget,
	APLValueFrontOfTarget,
	APLValueGCDIsReady,
	APLValueGCDTimeToReady,
	APLValueInputDelay,
	APLValueIsExecutePhase,
	APLValueItemProcsMaxRemainingICD,
	APLValueItemProcsMinRemainingTime,
	APLValueMath,
	APLValueMax,
	APLValueMaxComboPoints,
	APLValueMaxEnergy,
	APLValueMaxHealth,
	APLValueMaxMana,
	APLValueMaxRage,
	APLValueMin,
	APLValueMultipleCdUsages,
	APLValueNot,
	APLValueNumberTargets,
	APLValueNumEquippedStatProcItems,
	APLValueNumStatBuffCooldowns,
	APLValueOr,
	APLValueProtectionPaladinDamageTakenLastGlobal,
	APLValueRemainingTime,
	APLValueRemainingTimePercent,
	APLValueSelectedConjured,
	APLValueSelectedPotion,
	APLValueSequenceIsComplete,
	APLValueSequenceIsReady,
	APLValueSequenceTimeToReady,
	APLValueSpellCanCast,
	APLValueSpellCastTime,
	APLValueSpellChanneledTicks,
	APLValueSpellCPM,
	APLValueSpellCurrentCost,
	APLValueSpellFullCooldown,
	APLValueSpellGCDHastedDuration,
	APLValueSpellInFlight,
	APLValueSpellIsChanneling,
	APLValueSpellIsKnown,
	APLValueSpellIsReady,
	APLValueSpellNumCharges,
	APLValueSpellTimeToCharge,
	APLValueSpellTimeToReady,
	APLValueSpellTravelTime,
	APLValueTargetIsMobType,
	APLValueTimeToNextEnergyTick,
	APLValueTotemRemainingTime,
	APLValueUnitDistance,
	APLValueUnitIsMoving,
	APLValueWarlockActivePet,
	APLValueWarlockAssignedCurseIsActive,
	APLValueWarlockIsAssignedCurse,
} from '@generated/proto/apl';
import { Class, Spec } from '@generated/proto/common';
import i18n from '@i18n/config';
import { Player } from '@sim/player/player';
import { itemSwapEnabledSpecs } from '@sim/spec_config';

import {
	actionIdFieldConfig,
	areaTypeFieldConfig,
	autoSwingTypeFieldConfig,
	autoTypeFieldConfig,
	booleanFieldConfig,
	comparisonOperatorFieldConfig,
	curseTypeFieldConfig,
	executePhaseThresholdFieldConfig,
	groupNameFieldConfig,
	itemSwapSetFieldConfig,
	makeUseDotBaseValueCheckbox,
	mathOperatorFieldConfig,
	minIcdInput,
	mobTypeFieldConfig,
	placeholderNameFieldConfig,
	reactionTimeCheckbox,
	statTypeFieldConfig,
	stringFieldConfig,
	summonTypeFieldConfig,
	totemTypeFieldConfig,
	unitFieldConfig,
	valueFieldConfig,
	ValueFieldDescriptor,
	valueListFieldConfig,
	variableNameFieldConfig,
} from './field_descriptors';

type APLValue_Value = APLValue['value'];
export type APLValueKind = APLValue_Value['oneofKind'];
export type ValidAPLValueKind = NonNullable<APLValueKind>;

export type APLValueImplStruct<F extends APLValueKind> = Extract<APLValue_Value, { oneofKind: F }>;

// Get the implementation type for a specific kind using infer
type APLValueImplFor<F extends ValidAPLValueKind> = APLValueImplStruct<F> extends { [K in F]: infer T } ? T : never;

// Map all valid kinds to their implementation types
export type APLValueImplMap = {
	[K in ValidAPLValueKind]: APLValueImplFor<K>;
};

export type APLValueImplType = APLValueImplMap[ValidAPLValueKind] | undefined;

// The DOM-free half of an APL value kind. `components/ValuePicker` renders `fields`
// through `FieldGroup`.
export type ValueKindModel<T> = {
	label: string;
	submenu?: Array<string>;
	shortDescription: string;
	fullDescription?: string;
	newValue: () => T;
	includeIf?: (player: Player<any>, isPrepull: boolean, isGroup: boolean) => boolean;
	dynamicStringResolver?: (value: string, player: Player<any>) => string;
	fields: Array<ValueFieldDescriptor>;
};

function inputBuilder<T extends APLValueImplType>(config: ValueKindModel<T>): ValueKindModel<T> {
	return config;
}

const includeAutoValue = (player: Player<any>, isPrepull: boolean) => {
	const spec = player.getSpec();
	return !isPrepull && spec !== Spec.SpecBalanceDruid && spec !== Spec.SpecElementalShaman;
};

export const valueKinds: { [f in ValidAPLValueKind]: ValueKindModel<APLValueImplMap[f]> } = {
	// Operators
	const: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.const.label'),
		shortDescription: i18n.t('rotation_tab.apl.values.const.tooltip'),
		fullDescription: i18n.t('rotation_tab.apl.values.const.full_description'),
		newValue: APLValueConst.create,
		fields: [stringFieldConfig('val')],
	}),
	cmp: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.compare.label'),
		submenu: ['logic'],
		shortDescription: i18n.t('rotation_tab.apl.values.compare.tooltip'),
		newValue: APLValueCompare.create,
		fields: [valueFieldConfig('lhs'), comparisonOperatorFieldConfig('op'), valueFieldConfig('rhs')],
	}),
	math: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.math.label'),
		submenu: ['logic'],
		shortDescription: i18n.t('rotation_tab.apl.values.math.tooltip'),
		newValue: APLValueMath.create,
		fields: [valueFieldConfig('lhs'), mathOperatorFieldConfig('op'), valueFieldConfig('rhs')],
	}),
	max: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.max.label'),
		submenu: ['logic'],
		shortDescription: i18n.t('rotation_tab.apl.values.max.tooltip'),
		newValue: APLValueMax.create,
		fields: [valueListFieldConfig('vals')],
	}),
	min: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.min.label'),
		submenu: ['logic'],
		shortDescription: i18n.t('rotation_tab.apl.values.min.tooltip'),
		newValue: APLValueMin.create,
		fields: [valueListFieldConfig('vals')],
	}),
	and: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.all_of.label'),
		submenu: ['logic'],
		shortDescription: i18n.t('rotation_tab.apl.values.all_of.tooltip'),
		newValue: APLValueAnd.create,
		fields: [valueListFieldConfig('vals')],
	}),
	or: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.any_of.label'),
		submenu: ['logic'],
		shortDescription: i18n.t('rotation_tab.apl.values.any_of.tooltip'),
		newValue: APLValueOr.create,
		fields: [valueListFieldConfig('vals')],
	}),
	not: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.not.label'),
		submenu: ['logic'],
		shortDescription: i18n.t('rotation_tab.apl.values.not.tooltip'),
		newValue: APLValueNot.create,
		fields: [valueFieldConfig('val')],
	}),

	// Encounter
	currentTime: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.current_time.label'),
		submenu: ['encounter'],
		shortDescription: i18n.t('rotation_tab.apl.values.current_time.tooltip'),
		newValue: APLValueCurrentTime.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [],
	}),
	currentTimePercent: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.current_time_percent.label'),
		submenu: ['encounter'],
		shortDescription: i18n.t('rotation_tab.apl.values.current_time_percent.tooltip'),
		newValue: APLValueCurrentTimePercent.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [],
	}),
	remainingTime: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.remaining_time.label'),
		submenu: ['encounter'],
		shortDescription: i18n.t('rotation_tab.apl.values.remaining_time.tooltip'),
		newValue: APLValueRemainingTime.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [],
	}),
	remainingTimePercent: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.remaining_time_percent.label'),
		submenu: ['encounter'],
		shortDescription: i18n.t('rotation_tab.apl.values.remaining_time_percent.tooltip'),
		newValue: APLValueRemainingTimePercent.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [],
	}),
	isExecutePhase: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.is_execute_phase.label'),
		submenu: ['encounter'],
		shortDescription: i18n.t('rotation_tab.apl.values.is_execute_phase.tooltip'),
		newValue: APLValueIsExecutePhase.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [executePhaseThresholdFieldConfig('threshold')],
	}),
	numberTargets: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.num_targets.label'),
		submenu: ['encounter'],
		shortDescription: i18n.t('rotation_tab.apl.values.num_targets.tooltip'),
		newValue: APLValueNumberTargets.create,
		fields: [],
	}),
	frontOfTarget: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.in_front_of_target.label'),
		submenu: ['encounter'],
		shortDescription: i18n.t('rotation_tab.apl.values.in_front_of_target.tooltip'),
		newValue: APLValueFrontOfTarget.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [],
	}),
	encounterInArea: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.encounter_in_area.label'),
		submenu: ['encounter'],
		shortDescription: i18n.t('rotation_tab.apl.values.encounter_in_area.tooltip'),
		newValue: APLValueEncounterInArea.create,
		fields: [areaTypeFieldConfig('areaType')],
	}),

	// Boss
	bossSpellIsCasting: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.spell_is_casting.label'),
		submenu: ['boss'],
		shortDescription: i18n.t('rotation_tab.apl.values.spell_is_casting.tooltip'),
		newValue: APLValueBossSpellIsCasting.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [unitFieldConfig('targetUnit', 'targets'), actionIdFieldConfig('spellId', 'non_instant_spells', 'targetUnit', 'currentTarget')],
	}),
	bossSpellTimeToReady: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.spell_time_to_ready.label'),
		submenu: ['boss'],
		shortDescription: i18n.t('rotation_tab.apl.values.spell_time_to_ready.tooltip'),
		newValue: APLValueBossSpellTimeToReady.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [unitFieldConfig('targetUnit', 'targets'), actionIdFieldConfig('spellId', 'spells', 'targetUnit', 'currentTarget')],
	}),
	bossCurrentTarget: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.boss_current_target.label'),
		submenu: ['boss'],
		shortDescription: i18n.t('rotation_tab.apl.values.boss_current_target.tooltip'),
		newValue: APLValueBossCurrentTarget.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [unitFieldConfig('targetUnit', 'targets')],
	}),
	targetIsMobType: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.target_is_mob_type.label'),
		submenu: ['boss'],
		shortDescription: i18n.t('rotation_tab.apl.values.target_is_mob_type.tooltip'),
		newValue: APLValueTargetIsMobType.create,
		fields: [unitFieldConfig('targetUnit', 'targets'), mobTypeFieldConfig('mobType')],
	}),

	// Unit
	unitIsMoving: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.unit_is_moving.label'),
		submenu: ['unit'],
		shortDescription: i18n.t('rotation_tab.apl.values.unit_is_moving.tooltip'),
		newValue: APLValueUnitIsMoving.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [unitFieldConfig('sourceUnit', 'aura_sources')],
	}),
	unitDistance: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.distance_to_unit.label'),
		submenu: ['unit'],
		shortDescription: i18n.t('rotation_tab.apl.values.distance_to_unit.tooltip'),
		newValue: APLValueUnitDistance.create,
		fields: [unitFieldConfig('sourceUnit', 'aura_sources')],
	}),

	// Resources
	currentHealth: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.current_health.label'),
		submenu: ['resources', 'health'],
		shortDescription: i18n.t('rotation_tab.apl.values.current_health.tooltip'),
		newValue: APLValueCurrentHealth.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [unitFieldConfig('sourceUnit', 'aura_sources')],
	}),
	currentHealthPercent: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.current_health_percent.label'),
		submenu: ['resources', 'health'],
		shortDescription: i18n.t('rotation_tab.apl.values.current_health_percent.tooltip'),
		newValue: APLValueCurrentHealthPercent.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [unitFieldConfig('sourceUnit', 'aura_sources')],
	}),
	maxHealth: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.max_health.label'),
		submenu: ['resources', 'health'],
		shortDescription: i18n.t('rotation_tab.apl.values.max_health.tooltip'),
		newValue: APLValueMaxHealth.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [],
	}),
	currentMana: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.current_mana.label'),
		submenu: ['resources', 'mana'],
		shortDescription: i18n.t('rotation_tab.apl.values.current_mana.tooltip'),
		newValue: APLValueCurrentMana.create,
		includeIf(player: Player<any>, isPrepull: boolean) {
			const clss = player.getClass();
			return !isPrepull && clss !== Class.ClassRogue && clss !== Class.ClassWarrior;
		},
		fields: [],
	}),
	currentManaPercent: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.current_mana_percent.label'),
		submenu: ['resources', 'mana'],
		shortDescription: i18n.t('rotation_tab.apl.values.current_mana_percent.tooltip'),
		newValue: APLValueCurrentManaPercent.create,
		includeIf(player: Player<any>, isPrepull: boolean) {
			const clss = player.getClass();
			return !isPrepull && clss !== Class.ClassRogue && clss !== Class.ClassWarrior;
		},
		fields: [],
	}),
	maxMana: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.max_mana.label'),
		submenu: ['resources', 'mana'],
		shortDescription: i18n.t('rotation_tab.apl.values.max_mana.tooltip'),
		newValue: APLValueMaxMana.create,
		includeIf(player: Player<any>, isPrepull: boolean) {
			const clss = player.getClass();
			return !isPrepull && clss !== Class.ClassRogue && clss !== Class.ClassWarrior;
		},
		fields: [],
	}),
	currentManaRegen: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.current_mana_regen.label'),
		submenu: ['resources', 'mana'],
		shortDescription: i18n.t('rotation_tab.apl.values.current_mana_regen.tooltip'),
		newValue: APLValueCurrentManaRegen.create,
		includeIf(player: Player<any>, isPrepull: boolean) {
			const clss = player.getClass();
			return !isPrepull && clss !== Class.ClassRogue && clss !== Class.ClassWarrior;
		},
		fields: [],
	}),
	currentRage: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.current_rage.label'),
		submenu: ['resources', 'rage'],
		shortDescription: i18n.t('rotation_tab.apl.values.current_rage.tooltip'),
		newValue: APLValueCurrentRage.create,
		includeIf(player: Player<any>, isPrepull: boolean) {
			const clss = player.getClass();
			const spec = player.getSpec();
			return !isPrepull && (spec === Spec.SpecFeralCatDruid || spec === Spec.SpecFeralBearDruid || clss === Class.ClassWarrior);
		},
		fields: [],
	}),
	maxRage: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.max_rage.label'),
		submenu: ['resources', 'rage'],
		shortDescription: i18n.t('rotation_tab.apl.values.max_rage.tooltip'),
		newValue: APLValueMaxRage.create,
		includeIf(player: Player<any>, isPrepull: boolean) {
			const clss = player.getClass();
			const spec = player.getSpec();
			return !isPrepull && (spec === Spec.SpecFeralCatDruid || spec === Spec.SpecFeralBearDruid || clss === Class.ClassWarrior);
		},
		fields: [],
	}),
	currentEnergy: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.current_energy.label'),
		submenu: ['resources', 'energy'],
		shortDescription: i18n.t('rotation_tab.apl.values.current_energy.tooltip'),
		newValue: APLValueCurrentEnergy.create,
		includeIf(player: Player<any>, isPrepull: boolean) {
			const clss = player.getClass();
			const spec = player.getSpec();
			return !isPrepull && (spec === Spec.SpecFeralCatDruid || spec === Spec.SpecFeralBearDruid || clss === Class.ClassRogue);
		},
		fields: [],
	}),
	maxEnergy: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.max_energy.label'),
		submenu: ['resources', 'energy'],
		shortDescription: i18n.t('rotation_tab.apl.values.max_energy.tooltip'),
		newValue: APLValueMaxEnergy.create,
		includeIf(player: Player<any>, isPrepull: boolean) {
			const clss = player.getClass();
			const spec = player.getSpec();
			return !isPrepull && (spec === Spec.SpecFeralCatDruid || spec === Spec.SpecFeralBearDruid || clss === Class.ClassRogue);
		},
		fields: [],
	}),
	timeToNextEnergyTick: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.time_to_next_energy_tick.label'),
		submenu: ['resources', 'energy'],
		shortDescription: i18n.t('rotation_tab.apl.values.time_to_next_energy_tick.tooltip'),
		newValue: APLValueTimeToNextEnergyTick.create,
		includeIf(player: Player<any>, isPrepull: boolean) {
			const clss = player.getClass();
			const spec = player.getSpec();
			return !isPrepull && (spec === Spec.SpecFeralCatDruid || spec === Spec.SpecFeralBearDruid || clss === Class.ClassRogue);
		},
		fields: [],
	}),
	energyTimeToTarget: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.estimated_time_to_target_energy.label'),
		submenu: ['resources', 'energy'],
		shortDescription: i18n.t('rotation_tab.apl.values.estimated_time_to_target_energy.tooltip'),
		newValue: APLValueEnergyTimeToTarget.create,
		includeIf(player: Player<any>, isPrepull: boolean) {
			const clss = player.getClass();
			const spec = player.getSpec();
			return !isPrepull && (spec === Spec.SpecFeralCatDruid || spec === Spec.SpecFeralBearDruid || clss === Class.ClassRogue);
		},
		fields: [valueFieldConfig('targetEnergy')],
	}),
	currentComboPoints: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.current_combo_points.label'),
		submenu: ['resources', 'combo_points'],
		shortDescription: i18n.t('rotation_tab.apl.values.current_combo_points.tooltip'),
		newValue: APLValueCurrentComboPoints.create,
		includeIf(player: Player<any>, isPrepull: boolean) {
			const clss = player.getClass();
			const spec = player.getSpec();
			return !isPrepull && (spec === Spec.SpecFeralCatDruid || spec === Spec.SpecFeralBearDruid || clss === Class.ClassRogue);
		},
		fields: [],
	}),
	maxComboPoints: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.max_combo_points.label'),
		submenu: ['resources', 'combo_points'],
		shortDescription: i18n.t('rotation_tab.apl.values.max_combo_points.tooltip'),
		newValue: APLValueMaxComboPoints.create,
		includeIf(player: Player<any>, isPrepull: boolean) {
			const clss = player.getClass();
			const spec = player.getSpec();
			return !isPrepull && (spec === Spec.SpecFeralCatDruid || spec === Spec.SpecFeralBearDruid || clss === Class.ClassRogue);
		},
		fields: [],
	}),
	currentGenericResource: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.generic_resource.label'),
		submenu: ['resources'],
		shortDescription: i18n.t('rotation_tab.apl.values.generic_resource.tooltip'),
		newValue: APLValueCurrentGenericResource.create,
		fields: [],
		includeIf: (_player: Player<any>, isPrepull: boolean) => !isPrepull,
		dynamicStringResolver: (_value: string, _player: Player<any>) => '',
	}),

	// GCD
	gcdIsReady: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.gcd_is_ready.label'),
		submenu: ['gcd'],
		shortDescription: i18n.t('rotation_tab.apl.values.gcd_is_ready.tooltip'),
		newValue: APLValueGCDIsReady.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [],
	}),
	gcdTimeToReady: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.gcd_time_to_ready.label'),
		submenu: ['gcd'],
		shortDescription: i18n.t('rotation_tab.apl.values.gcd_time_to_ready.tooltip'),
		newValue: APLValueGCDTimeToReady.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [],
	}),

	// Auto attacks
	autoSwingTime: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.auto_swing_time.label'),
		submenu: ['auto'],
		shortDescription: i18n.t('rotation_tab.apl.values.auto_swing_time.tooltip'),
		newValue: APLValueAutoSwingTime.create,
		includeIf: includeAutoValue,
		fields: [autoSwingTypeFieldConfig('autoType')],
	}),
	autoTimeToNext: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.time_to_next_auto.label'),
		submenu: ['auto'],
		shortDescription: i18n.t('rotation_tab.apl.values.time_to_next_auto.tooltip'),
		newValue: APLValueAutoTimeToNext.create,
		includeIf: includeAutoValue,
		fields: [autoTypeFieldConfig('autoType')],
	}),
	autoTimeSinceLast: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.time_since_last_auto.label'),
		submenu: ['auto'],
		shortDescription: i18n.t('rotation_tab.apl.values.time_since_last_auto.tooltip'),
		newValue: APLValueAutoTimeSinceLast.create,
		includeIf: includeAutoValue,
		fields: [autoTypeFieldConfig('autoType')],
	}),

	// Spells
	spellIsKnown: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.spell_known.label'),
		submenu: ['spell'],
		shortDescription: i18n.t('rotation_tab.apl.values.spell_known.tooltip'),
		newValue: APLValueSpellIsKnown.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [actionIdFieldConfig('spellId', 'castable_spells', '')],
	}),
	spellCurrentCost: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.current_cost.label'),
		submenu: ['spell'],
		shortDescription: i18n.t('rotation_tab.apl.values.current_cost.tooltip'),
		newValue: APLValueSpellCurrentCost.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [actionIdFieldConfig('spellId', 'castable_spells', '')],
	}),
	spellCanCast: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.can_cast.label'),
		submenu: ['spell'],
		shortDescription: i18n.t('rotation_tab.apl.values.can_cast.tooltip'),
		fullDescription: i18n.t('rotation_tab.apl.values.can_cast.full_description'),
		newValue: APLValueSpellCanCast.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [actionIdFieldConfig('spellId', 'castable_spells', '')],
	}),
	spellIsReady: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.is_ready.label'),
		submenu: ['spell'],
		shortDescription: i18n.t('rotation_tab.apl.values.is_ready.tooltip'),
		newValue: APLValueSpellIsReady.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [actionIdFieldConfig('spellId', 'castable_spells', '')],
	}),
	spellTimeToReady: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.time_to_ready.label'),
		submenu: ['spell'],
		shortDescription: i18n.t('rotation_tab.apl.values.time_to_ready.tooltip'),
		newValue: APLValueSpellTimeToReady.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [actionIdFieldConfig('spellId', 'castable_spells', '')],
	}),
	spellCastTime: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.cast_time.label'),
		submenu: ['spell'],
		shortDescription: i18n.t('rotation_tab.apl.values.cast_time.tooltip'),
		newValue: APLValueSpellCastTime.create,
		fields: [actionIdFieldConfig('spellId', 'castable_spells', '')],
	}),
	spellTravelTime: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.travel_time.label'),
		submenu: ['spell'],
		shortDescription: i18n.t('rotation_tab.apl.values.travel_time.tooltip'),
		newValue: APLValueSpellTravelTime.create,
		fields: [actionIdFieldConfig('spellId', 'castable_spells', '')],
	}),
	spellCpm: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.cpm.label'),
		submenu: ['spell'],
		shortDescription: i18n.t('rotation_tab.apl.values.cpm.tooltip'),
		newValue: APLValueSpellCPM.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [actionIdFieldConfig('spellId', 'castable_spells', '')],
	}),
	spellIsChanneling: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.is_channeling.label'),
		submenu: ['spell'],
		shortDescription: i18n.t('rotation_tab.apl.values.is_channeling.tooltip'),
		newValue: APLValueSpellIsChanneling.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [actionIdFieldConfig('spellId', 'channel_spells', '')],
	}),
	spellChanneledTicks: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.channeled_ticks.label'),
		submenu: ['spell'],
		shortDescription: i18n.t('rotation_tab.apl.values.channeled_ticks.tooltip'),
		newValue: APLValueSpellChanneledTicks.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [actionIdFieldConfig('spellId', 'channel_spells', '')],
	}),
	spellNumCharges: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.number_of_charges.label'),
		submenu: ['spell'],
		shortDescription: i18n.t('rotation_tab.apl.values.number_of_charges.tooltip'),
		newValue: APLValueSpellNumCharges.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [actionIdFieldConfig('spellId', 'castable_spells', '')],
	}),
	spellTimeToCharge: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.time_to_next_charge.label'),
		submenu: ['spell'],
		shortDescription: i18n.t('rotation_tab.apl.values.time_to_next_charge.tooltip'),
		newValue: APLValueSpellTimeToCharge.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [actionIdFieldConfig('spellId', 'castable_spells', '')],
	}),
	spellGcdHastedDuration: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.gcd_hasted_duration.label'),
		submenu: ['spell'],
		shortDescription: i18n.t('rotation_tab.apl.values.gcd_hasted_duration.tooltip'),
		newValue: APLValueSpellGCDHastedDuration.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [actionIdFieldConfig('spellId', 'castable_spells', '')],
	}),
	spellFullCooldown: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.full_cooldown.label'),
		submenu: ['spell'],
		shortDescription: i18n.t('rotation_tab.apl.values.full_cooldown.tooltip'),
		newValue: APLValueSpellFullCooldown.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [actionIdFieldConfig('spellId', 'castable_spells', '')],
	}),
	channelClipDelay: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.channel_clip_delay.label'),
		submenu: ['spell'],
		shortDescription: i18n.t('rotation_tab.apl.values.channel_clip_delay.tooltip'),
		newValue: APLValueChannelClipDelay.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [],
	}),
	inputDelay: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.input_delay.label'),
		submenu: ['spell'],
		shortDescription: i18n.t('rotation_tab.apl.values.input_delay.tooltip'),
		newValue: APLValueInputDelay.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [],
	}),
	spellInFlight: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.spell_in_flight.label'),
		submenu: ['spell'],
		shortDescription: i18n.t('rotation_tab.apl.values.spell_in_flight.tooltip'),
		newValue: APLValueSpellInFlight.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [actionIdFieldConfig('spellId', 'spells_with_travelTime', '')],
	}),

	// Auras
	auraIsKnown: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.aura_known.label'),
		submenu: ['aura'],
		shortDescription: i18n.t('rotation_tab.apl.values.aura_known.tooltip'),
		newValue: APLValueAuraIsKnown.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [unitFieldConfig('sourceUnit', 'aura_sources'), actionIdFieldConfig('auraId', 'auras', 'sourceUnit')],
	}),
	auraIsActive: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.aura_active.label'),
		submenu: ['aura'],
		shortDescription: i18n.t('rotation_tab.apl.values.aura_active.tooltip'),
		newValue: () => APLValueAuraIsActive.create({ includeReactionTime: true }),
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [unitFieldConfig('sourceUnit', 'aura_sources'), actionIdFieldConfig('auraId', 'auras', 'sourceUnit'), reactionTimeCheckbox()],
	}),
	auraIsInactive: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.aura_inactive.label'),
		submenu: ['aura'],
		shortDescription: i18n.t('rotation_tab.apl.values.aura_inactive.tooltip'),
		newValue: () => APLValueAuraIsInactive.create({ includeReactionTime: true }),
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [unitFieldConfig('sourceUnit', 'aura_sources'), actionIdFieldConfig('auraId', 'auras', 'sourceUnit'), reactionTimeCheckbox()],
	}),
	auraRemainingTime: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.aura_remaining_time.label'),
		submenu: ['aura'],
		shortDescription: i18n.t('rotation_tab.apl.values.aura_remaining_time.tooltip'),
		newValue: APLValueAuraRemainingTime.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [unitFieldConfig('sourceUnit', 'aura_sources'), actionIdFieldConfig('auraId', 'auras', 'sourceUnit')],
	}),
	auraNumStacks: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.aura_num_stacks.label'),
		submenu: ['aura'],
		shortDescription: i18n.t('rotation_tab.apl.values.aura_num_stacks.tooltip'),
		newValue: () => APLValueAuraNumStacks.create({ includeReactionTime: true }),
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [unitFieldConfig('sourceUnit', 'aura_sources'), actionIdFieldConfig('auraId', 'stackable_auras', 'sourceUnit'), reactionTimeCheckbox()],
	}),
	auraInternalCooldown: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.aura_remaining_icd.label'),
		submenu: ['aura'],
		shortDescription: i18n.t('rotation_tab.apl.values.aura_remaining_icd.tooltip'),
		newValue: APLValueAuraInternalCooldown.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [unitFieldConfig('sourceUnit', 'aura_sources'), actionIdFieldConfig('auraId', 'icd_auras', 'sourceUnit')],
	}),
	auraIcdIsReady: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.aura_icd_is_ready.label'),
		submenu: ['aura'],
		shortDescription: i18n.t('rotation_tab.apl.values.aura_icd_is_ready.tooltip'),
		newValue: () => APLValueAuraICDIsReady.create({ includeReactionTime: true }),
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [unitFieldConfig('sourceUnit', 'aura_sources'), actionIdFieldConfig('auraId', 'icd_auras', 'sourceUnit'), reactionTimeCheckbox()],
	}),
	auraShouldRefresh: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.aura_should_refresh.label'),
		submenu: ['aura'],
		shortDescription: i18n.t('rotation_tab.apl.values.aura_should_refresh.tooltip'),
		fullDescription: i18n.t('rotation_tab.apl.values.aura_should_refresh.full_description'),
		newValue: () =>
			APLValueAuraShouldRefresh.create({
				maxOverlap: {
					value: {
						oneofKind: 'const',
						const: {
							val: '0ms',
						},
					},
				},
			}),
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [
			unitFieldConfig('sourceUnit', 'aura_sources_targets_first'),
			actionIdFieldConfig('auraId', 'exclusive_effect_auras', 'sourceUnit', 'currentTarget'),
			valueFieldConfig('maxOverlap', {
				label: i18n.t('rotation_tab.apl.values.overlap.label'),
				labelTooltip: i18n.t('rotation_tab.apl.values.overlap.tooltip'),
			}),
		],
	}),
	auraDuration: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.aura_duration.label'),
		submenu: ['aura'],
		shortDescription: i18n.t('rotation_tab.apl.values.aura_duration.tooltip'),
		fullDescription: i18n.t('rotation_tab.apl.values.aura_duration.full_description'),
		newValue: APLValueAuraDuration.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [unitFieldConfig('sourceUnit', 'aura_sources'), actionIdFieldConfig('auraId', 'auras', 'sourceUnit')],
	}),

	// Aura Sets
	allItemStatProcsActive: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.all_item_stat_procs_active.label'),
		submenu: ['aura_sets'],
		shortDescription: i18n.t('rotation_tab.apl.values.all_item_stat_procs_active.tooltip'),
		fullDescription: i18n.t('rotation_tab.apl.values.all_item_stat_procs_active.full_description'),
		newValue: () =>
			APLValueAllItemStatProcsActive.create({
				statType1: -1,
				statType2: -1,
				statType3: -1,
			}),
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [statTypeFieldConfig('statType1'), statTypeFieldConfig('statType2'), statTypeFieldConfig('statType3'), minIcdInput],
	}),
	anyItemStatProcsActive: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.any_item_stat_procs_active.label'),
		submenu: ['aura_sets'],
		shortDescription: i18n.t('rotation_tab.apl.values.any_item_stat_procs_active.tooltip'),
		fullDescription: i18n.t('rotation_tab.apl.values.any_item_stat_procs_active.full_description'),
		newValue: () =>
			APLValueAnyItemStatProcsActive.create({
				statType1: -1,
				statType2: -1,
				statType3: -1,
			}),
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [statTypeFieldConfig('statType1'), statTypeFieldConfig('statType2'), statTypeFieldConfig('statType3'), minIcdInput],
	}),
	anyItemStatProcsAvailable: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.any_item_stat_procs_available.label'),
		submenu: ['aura_sets'],
		shortDescription: i18n.t('rotation_tab.apl.values.any_item_stat_procs_available.tooltip'),
		newValue: () =>
			APLValueAnyItemStatProcsAvailable.create({
				statType1: -1,
				statType2: -1,
				statType3: -1,
			}),
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [statTypeFieldConfig('statType1'), statTypeFieldConfig('statType2'), statTypeFieldConfig('statType3'), minIcdInput],
	}),
	itemProcsMinRemainingTime: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.item_procs_min_remaining_time.label'),
		submenu: ['aura_sets'],
		shortDescription: i18n.t('rotation_tab.apl.values.item_procs_min_remaining_time.tooltip'),
		newValue: () =>
			APLValueItemProcsMinRemainingTime.create({
				statType1: -1,
				statType2: -1,
				statType3: -1,
			}),
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [statTypeFieldConfig('statType1'), statTypeFieldConfig('statType2'), statTypeFieldConfig('statType3'), minIcdInput],
	}),
	itemProcsMaxRemainingIcd: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.item_procs_max_remaining_icd.label'),
		submenu: ['aura_sets'],
		shortDescription: i18n.t('rotation_tab.apl.values.item_procs_max_remaining_icd.tooltip'),
		newValue: () =>
			APLValueItemProcsMaxRemainingICD.create({
				statType1: -1,
				statType2: -1,
				statType3: -1,
			}),
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [statTypeFieldConfig('statType1'), statTypeFieldConfig('statType2'), statTypeFieldConfig('statType3'), minIcdInput],
	}),
	numEquippedStatProcItems: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.num_equipped_stat_proc_items.label'),
		submenu: ['aura_sets'],
		shortDescription: i18n.t('rotation_tab.apl.values.num_equipped_stat_proc_items.tooltip'),
		newValue: () =>
			APLValueNumEquippedStatProcItems.create({
				statType1: -1,
				statType2: -1,
				statType3: -1,
			}),
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [statTypeFieldConfig('statType1'), statTypeFieldConfig('statType2'), statTypeFieldConfig('statType3'), minIcdInput],
	}),
	numStatBuffCooldowns: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.num_stat_buff_cooldowns.label'),
		submenu: ['aura_sets'],
		shortDescription: i18n.t('rotation_tab.apl.values.num_stat_buff_cooldowns.tooltip'),
		fullDescription: i18n.t('rotation_tab.apl.values.num_stat_buff_cooldowns.full_description'),
		newValue: () =>
			APLValueNumStatBuffCooldowns.create({
				statType1: -1,
				statType2: -1,
				statType3: -1,
			}),
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [statTypeFieldConfig('statType1'), statTypeFieldConfig('statType2'), statTypeFieldConfig('statType3')],
	}),
	anyStatBuffCooldownsActive: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.any_stat_buff_cooldowns_active.label'),
		submenu: ['aura_sets'],
		shortDescription: i18n.t('rotation_tab.apl.values.any_stat_buff_cooldowns_active.tooltip'),
		fullDescription: i18n.t('rotation_tab.apl.values.any_stat_buff_cooldowns_active.full_description'),
		newValue: () =>
			APLValueAnyStatBuffCooldownsActive.create({
				statType1: -1,
				statType2: -1,
				statType3: -1,
			}),
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [statTypeFieldConfig('statType1'), statTypeFieldConfig('statType2'), statTypeFieldConfig('statType3')],
	}),
	anyStatBuffCooldownsMinDuration: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.any_stat_buff_cooldowns_min_duration.label'),
		submenu: ['aura_sets'],
		shortDescription: i18n.t('rotation_tab.apl.values.any_stat_buff_cooldowns_min_duration.tooltip'),
		fullDescription: i18n.t('rotation_tab.apl.values.any_stat_buff_cooldowns_min_duration.full_description'),
		newValue: () =>
			APLValueAnyStatBuffCooldownsMinDuration.create({
				statType1: -1,
				statType2: -1,
				statType3: -1,
			}),
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [statTypeFieldConfig('statType1'), statTypeFieldConfig('statType2'), statTypeFieldConfig('statType3')],
	}),

	// DoT
	dotIsActive: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.dot_is_active.label'),
		submenu: ['dot'],
		shortDescription: i18n.t('rotation_tab.apl.values.dot_is_active.tooltip'),
		newValue: APLValueDotIsActive.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [unitFieldConfig('targetUnit', 'targets'), actionIdFieldConfig('spellId', 'dot_spells', '')],
	}),
	dotIsActiveOnAllTargets: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.dot_is_active_on_all_targets.label'),
		submenu: ['dot'],
		shortDescription: i18n.t('rotation_tab.apl.values.dot_is_active_on_all_targets.tooltip'),
		newValue: APLValueDotIsActiveOnAllTargets.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [actionIdFieldConfig('spellId', 'dot_spells')],
	}),
	dotRemainingTime: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.dot_remaining_time.label'),
		submenu: ['dot'],
		shortDescription: i18n.t('rotation_tab.apl.values.dot_remaining_time.tooltip'),
		newValue: APLValueDotRemainingTime.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [unitFieldConfig('targetUnit', 'targets'), actionIdFieldConfig('spellId', 'dot_spells', '')],
	}),
	dotLowestRemainingTime: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.dot_lowest_remaining_time.label'),
		submenu: ['dot'],
		shortDescription: i18n.t('rotation_tab.apl.values.dot_lowest_remaining_time.tooltip'),
		newValue: APLValueDotLowestRemainingTime.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [actionIdFieldConfig('spellId', 'dot_spells', '')],
	}),
	dotTickFrequency: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.dot_tick_frequency.label'),
		submenu: ['dot'],
		shortDescription: i18n.t('rotation_tab.apl.values.dot_tick_frequency.tooltip'),
		newValue: APLValueDotTickFrequency.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [unitFieldConfig('targetUnit', 'targets'), actionIdFieldConfig('spellId', 'dot_spells', '')],
	}),
	dotTimeToNextTick: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.dot_time_to_next_tick.label'),
		submenu: ['dot'],
		shortDescription: i18n.t('rotation_tab.apl.values.dot_time_to_next_tick.tooltip'),
		newValue: APLValueDotTimeToNextTick.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [unitFieldConfig('targetUnit', 'targets'), actionIdFieldConfig('spellId', 'dot_spells', '')],
	}),
	dotBaseDuration: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.dot_base_duration.label'),
		submenu: ['dot'],
		shortDescription: i18n.t('rotation_tab.apl.values.dot_base_duration.tooltip'),
		newValue: APLValueDotBaseDuration.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [actionIdFieldConfig('spellId', 'dot_spells', '')],
	}),
	dotPercentIncrease: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.dot_percent_increase.label'),
		submenu: ['dot'],
		shortDescription: i18n.t('rotation_tab.apl.values.dot_percent_increase.tooltip'),
		newValue: APLValueDotPercentIncrease.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [unitFieldConfig('targetUnit', 'targets'), actionIdFieldConfig('spellId', 'expected_dot_spells', ''), makeUseDotBaseValueCheckbox()],
	}),
	dotCritPercentIncrease: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.dot_crit_percent_increase.label'),
		submenu: ['dot'],
		shortDescription: i18n.t('rotation_tab.apl.values.dot_crit_percent_increase.tooltip'),
		newValue: APLValueDotPercentIncrease.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [unitFieldConfig('targetUnit', 'targets'), actionIdFieldConfig('spellId', 'expected_dot_spells', ''), makeUseDotBaseValueCheckbox()],
	}),
	dotTickRatePercentIncrease: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.dot_tick_rate_percent_increase.label'),
		submenu: ['dot'],
		shortDescription: i18n.t('rotation_tab.apl.values.dot_tick_rate_percent_increase.tooltip'),
		newValue: APLValueDotPercentIncrease.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [unitFieldConfig('targetUnit', 'targets'), actionIdFieldConfig('spellId', 'expected_dot_spells', ''), makeUseDotBaseValueCheckbox()],
	}),
	sequenceIsComplete: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.sequence_is_complete.label'),
		submenu: ['sequence'],
		shortDescription: i18n.t('rotation_tab.apl.values.sequence_is_complete.tooltip'),
		newValue: APLValueSequenceIsComplete.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [stringFieldConfig('sequenceName')],
	}),
	sequenceIsReady: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.sequence_is_ready.label'),
		submenu: ['sequence'],
		shortDescription: i18n.t('rotation_tab.apl.values.sequence_is_ready.tooltip'),
		newValue: APLValueSequenceIsReady.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [stringFieldConfig('sequenceName')],
	}),
	sequenceTimeToReady: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.sequence_time_to_ready.label'),
		submenu: ['sequence'],
		shortDescription: i18n.t('rotation_tab.apl.values.sequence_time_to_ready.tooltip'),
		newValue: APLValueSequenceTimeToReady.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [stringFieldConfig('sequenceName')],
	}),

	// Class/spec specific values
	totemRemainingTime: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.totem_remaining_time.label'),
		submenu: ['shaman'],
		shortDescription: i18n.t('rotation_tab.apl.values.totem_remaining_time.tooltip'),
		newValue: APLValueTotemRemainingTime.create,
		includeIf: (player: Player<any>, isPrepull: boolean) => !isPrepull && player.getClass() == Class.ClassShaman,
		fields: [totemTypeFieldConfig('totemType'), reactionTimeCheckbox()],
	}),
	warlockAssignedCurseIsActive: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.warlock_assigned_curse_is_active.label'),
		submenu: ['warlock'],
		shortDescription: i18n.t('rotation_tab.apl.values.warlock_assigned_curse_is_active.tooltip'),
		newValue: APLValueWarlockAssignedCurseIsActive.create,
		includeIf: (player: Player<any>, isPrepull: boolean) => !isPrepull && player.getSpec() == Spec.SpecWarlock,
		fields: [unitFieldConfig('targetUnit', 'targets')],
	}),
	warlockIsAssignedCurse: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.warlock_is_assigned_curse.label'),
		submenu: ['warlock'],
		shortDescription: i18n.t('rotation_tab.apl.values.warlock_is_assigned_curse.tooltip'),
		newValue: APLValueWarlockIsAssignedCurse.create,
		includeIf: (player: Player<any>, isPrepull: boolean) => !isPrepull && player.getSpec() == Spec.SpecWarlock,
		fields: [curseTypeFieldConfig('curseType')],
	}),
	warlockActivePet: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.warlock_active_pet.label'),
		submenu: ['warlock'],
		shortDescription: i18n.t('rotation_tab.apl.values.warlock_active_pet.tooltip'),
		newValue: APLValueWarlockActivePet.create,
		includeIf: (player: Player<any>, isPrepull: boolean) => !isPrepull && player.getSpec() == Spec.SpecWarlock,
		fields: [summonTypeFieldConfig('summon')],
	}),
	protectionPaladinDamageTakenLastGlobal: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.protection_paladin_damage_taken_last_global.label'),
		submenu: ['tank'],
		shortDescription: i18n.t('rotation_tab.apl.values.protection_paladin_damage_taken_last_global.tooltip'),
		newValue: APLValueProtectionPaladinDamageTakenLastGlobal.create,
		includeIf: (player: Player<any>, isPrepull: boolean) => !isPrepull && player.getSpec() === Spec.SpecProtectionPaladin,
		fields: [],
	}),
	druidBleedIsActive: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.druid_bleed_is_active.label'),
		submenu: ['feral_druid'],
		shortDescription: i18n.t('rotation_tab.apl.values.druid_bleed_is_active.tooltip'),
		newValue: APLValueDruidBleedIsActive.create,
		includeIf(player: Player<any>, isPrepull: boolean) {
			const spec = player.getSpec();
			return !isPrepull && (spec === Spec.SpecFeralCatDruid || spec === Spec.SpecFeralBearDruid);
		},
		fields: [unitFieldConfig('targetUnit', 'targets')],
	}),
	variableRef: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.variable_ref.label'),
		submenu: ['variables'],
		shortDescription: i18n.t('rotation_tab.apl.values.variable_ref.tooltip'),
		newValue: () => ({ name: '' }),
		fields: [variableNameFieldConfig('name')],
	}),
	variablePlaceholder: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.variable_placeholder.label'),
		submenu: ['variables'],
		shortDescription: i18n.t('rotation_tab.apl.values.variable_placeholder.tooltip'),
		fullDescription: i18n.t('rotation_tab.apl.values.variable_placeholder.full_description'),
		includeIf: (_player: Player<any>, isPrepull: boolean, isGroup: boolean) => !isPrepull && isGroup, // Only show in groups, not prepull or priority list
		newValue: () => ({ name: '' }),
		fields: [
			placeholderNameFieldConfig('name', {
				labelTooltip: i18n.t('rotation_tab.apl.values.variable_placeholder.name_tooltip'),
			}),
		],
	}),
	actionGroupUsed: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.action_group_used.label'),
		submenu: ['variables'],
		shortDescription: i18n.t('rotation_tab.apl.values.action_group_used.tooltip'),
		newValue: APLValueActionGroupUsed.create,
		fields: [groupNameFieldConfig('name')],
	}),
	activeItemSwapSet: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.active_item_swap_set.label'),
		submenu: ['misc'],
		shortDescription: i18n.t('rotation_tab.apl.values.active_item_swap_set.tooltip'),
		includeIf: (player: Player<any>, isPrepull: boolean) => !isPrepull && itemSwapEnabledSpecs.includes(player.getSpec()),
		newValue: APLValueActiveItemSwapSet.create,
		fields: [itemSwapSetFieldConfig('swapSet')],
	}),
	selectedPotion: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.selected_potion.label'),
		submenu: ['misc'],
		shortDescription: i18n.t('rotation_tab.apl.values.selected_potion.tooltip'),
		newValue: APLValueSelectedPotion.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [actionIdFieldConfig('potionId', 'potions')],
	}),
	selectedConjured: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.selected_conjured.label'),
		submenu: ['misc'],
		shortDescription: i18n.t('rotation_tab.apl.values.selected_conjured.tooltip'),
		newValue: APLValueSelectedConjured.create,
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [actionIdFieldConfig('conjuredId', 'conjured_items')],
	}),
	multipleCdUsages: inputBuilder({
		label: i18n.t('rotation_tab.apl.values.multiple_cd_usages.label'),
		submenu: ['misc'],
		shortDescription: i18n.t('rotation_tab.apl.values.multiple_cd_usages.tooltip'),
		newValue: () =>
			APLValueMultipleCdUsages.create({
				offset: {
					value: {
						oneofKind: 'const',
						const: {
							val: '0ms',
						},
					},
				},
			}),
		includeIf: (_: Player<any>, isPrepull: boolean) => !isPrepull,
		fields: [
			actionIdFieldConfig('baseSpellId', 'castable_spells', '', undefined, {
				label: i18n.t('rotation_tab.apl.values.multiple_cd_usages.base_spell'),
				labelTooltip: i18n.t('rotation_tab.apl.values.multiple_cd_usages.base_spell_tooltip'),
			}),
			actionIdFieldConfig('targetSpellId', 'castable_spells', '', undefined, {
				label: i18n.t('rotation_tab.apl.values.multiple_cd_usages.target_spell'),
				labelTooltip: i18n.t('rotation_tab.apl.values.multiple_cd_usages.target_spell_tooltip'),
			}),
			valueFieldConfig('offset', {
				label: i18n.t('rotation_tab.apl.values.offset.label'),
				labelTooltip: i18n.t('rotation_tab.apl.values.offset.tooltip'),
			}),
			booleanFieldConfig('alignCdEnd', i18n.t('rotation_tab.apl.values.multiple_cd_usages.align_cd_end'), {
				labelTooltip: i18n.t('rotation_tab.apl.values.multiple_cd_usages.align_cd_end_tooltip'),
			}),
		],
	}),
};
