import {
	APLValueAutoAttackType,
	APLValueAutoSwingTime_SwingType as AutoSwingType,
	APLValueIsExecutePhase_ExecutePhaseThreshold as ExecutePhaseThreshold,
} from '@generated/proto/apl';
import { ActionID, AreaType, Class, Stat } from '@generated/proto/common';
import type { Player } from '@sim/player/player';
import { getEnumValues } from '@sim/utils/collections';
import { describe, expect, it, vi } from 'vitest';

import type { APLFieldDescriptor } from './field_descriptors';
import {
	actionFieldConfig,
	actionIdFieldConfig,
	actionListFieldConfig,
	areaTypeFieldConfig,
	autoSwingTypeFieldConfig,
	autoTypeFieldConfig,
	booleanFieldConfig,
	comparisonOperatorFieldConfig,
	curseTypeFieldConfig,
	damageAmpTypeFieldConfig,
	executePhaseThresholdFieldConfig,
	groupNameFieldConfig,
	groupReferenceVariablesFieldConfig,
	itemSwapSetFieldConfig,
	makeUseDotBaseValueCheckbox,
	mathOperatorFieldConfig,
	minIcdInput,
	mobTypeFieldConfig,
	numberFieldConfig,
	placeholderNameFieldConfig,
	reactionTimeCheckbox,
	statTypeFieldConfig,
	stringFieldConfig,
	summonTypeFieldConfig,
	totemTypeFieldConfig,
	unitFieldConfig,
	valueFieldConfig,
	valueListFieldConfig,
	variableNameFieldConfig,
} from './field_descriptors';
import { newKindImpl as newKindImplFor, resolveField as resolveFieldFor } from './field_specs';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

const playerOf = (clss: Class) => ({ getClass: () => clss }) as unknown as Player<any>;
const warrior = playerOf(Class.ClassWarrior);

const resolveField = (descriptor: APLFieldDescriptor) => resolveFieldFor(descriptor, warrior);
const newKindImpl = (kind: Parameters<typeof newKindImplFor>[0]) => newKindImplFor(kind, warrior);

describe('resolveField kinds', () => {
	it('resolves the fixed-enum descriptor types to kind enum', () => {
		const descriptors = [
			curseTypeFieldConfig('f'),
			summonTypeFieldConfig('f'),
			autoTypeFieldConfig('f'),
			autoSwingTypeFieldConfig('f'),
			statTypeFieldConfig('f'),
			itemSwapSetFieldConfig('f'),
			comparisonOperatorFieldConfig('f'),
			mathOperatorFieldConfig('f'),
			executePhaseThresholdFieldConfig('f'),
			totemTypeFieldConfig('f'),
			mobTypeFieldConfig('f'),
			areaTypeFieldConfig('f'),
			damageAmpTypeFieldConfig('f'),
		];
		for (const descriptor of descriptors) {
			expect(resolveField(descriptor).kind).toBe('enum');
		}
	});

	it('resolves minIcd to a number field named minIcdSeconds', () => {
		const spec = resolveField(minIcdInput);
		expect(spec.kind).toBe('number');
		expect(spec.field).toBe('minIcdSeconds');
	});

	it('resolves the two alias checkboxes to kind boolean with their own field names', () => {
		expect(resolveField(reactionTimeCheckbox())).toMatchObject({ kind: 'boolean', field: 'includeReactionTime' });
		expect(resolveField(makeUseDotBaseValueCheckbox())).toMatchObject({ kind: 'boolean', field: 'useBaseValue' });
	});

	it('resolves variableName and groupName to rotationName over their own source', () => {
		expect(resolveField(variableNameFieldConfig('f'))).toMatchObject({ kind: 'rotationName', source: 'variables' });
		expect(resolveField(groupNameFieldConfig('f'))).toMatchObject({ kind: 'rotationName', source: 'groups' });
	});

	it('resolves the remaining descriptor types to their own kind', () => {
		expect(resolveField(actionIdFieldConfig('f', 'spells')).kind).toBe('actionId');
		expect(resolveField(unitFieldConfig('f', 'targets')).kind).toBe('unit');
		expect(resolveField(booleanFieldConfig('f')).kind).toBe('boolean');
		expect(resolveField(numberFieldConfig('f', false)).kind).toBe('number');
		expect(resolveField(stringFieldConfig('f')).kind).toBe('string');
		expect(resolveField(placeholderNameFieldConfig('f')).kind).toBe('placeholderName');
		expect(resolveField(groupReferenceVariablesFieldConfig('f', 'g')).kind).toBe('groupVariables');
		expect(resolveField(valueFieldConfig('f')).kind).toBe('value');
		expect(resolveField(valueListFieldConfig('f')).kind).toBe('valueList');
		expect(resolveField(actionFieldConfig('f')).kind).toBe('action');
		expect(resolveField(actionListFieldConfig('f')).kind).toBe('actionList');
	});
});

describe('resolveField newValue defaults', () => {
	it('defaults simple kinds to their empty value', () => {
		expect(resolveField(booleanFieldConfig('f')).newValue()).toBe(false);
		expect(resolveField(numberFieldConfig('f', false)).newValue()).toBe(0);
		expect(resolveField(stringFieldConfig('f')).newValue()).toBe('');
		expect(resolveField(variableNameFieldConfig('f')).newValue()).toBe('');
		expect(resolveField(groupNameFieldConfig('f')).newValue()).toBe('');
		expect(resolveField(placeholderNameFieldConfig('f')).newValue()).toBe('');
		expect(resolveField(unitFieldConfig('f', 'targets')).newValue()).toBeUndefined();
		expect(resolveField(groupReferenceVariablesFieldConfig('f', 'g')).newValue()).toEqual([]);
		expect(resolveField(valueListFieldConfig('f')).newValue()).toEqual([]);
		expect(resolveField(actionListFieldConfig('f')).newValue()).toEqual([]);
		expect(resolveField(minIcdInput).newValue()).toBe(0);
		expect(resolveField(reactionTimeCheckbox()).newValue()).toBe(false);
		expect(resolveField(makeUseDotBaseValueCheckbox()).newValue()).toBe(false);
	});

	it('defaults actionId to a fresh empty ActionID', () => {
		expect(resolveField(actionIdFieldConfig('f', 'spells')).newValue()).toEqual(ActionID.create());
	});

	it('defaults value and action to a fresh APLValue carrying a uuid', () => {
		const value = resolveField(valueFieldConfig('f')).newValue();
		const action = resolveField(actionFieldConfig('f')).newValue();
		expect(typeof value.uuid.value).toBe('string');
		expect(value.uuid.value.length).toBeGreaterThan(0);
		expect(typeof action.uuid.value).toBe('string');
		expect(value.uuid.value).not.toBe(action.uuid.value);
	});
});

describe('resolveField label precedence', () => {
	it('lets booleanFieldConfig label win over options.label', () => {
		const descriptor = booleanFieldConfig('f', 'Explicit Label', { label: 'Options Label' });
		expect(resolveField(descriptor).label).toBe('Explicit Label');
	});

	it('reads label and labelTooltip off descriptor.options for other field types', () => {
		const unit = resolveField(unitFieldConfig('f', 'targets', { label: 'Unit Label', labelTooltip: 'Unit Tooltip' }));
		expect(unit.label).toBe('Unit Label');
		expect(unit.labelTooltip).toBe('Unit Tooltip');

		const number = resolveField(numberFieldConfig('f', false, { label: 'Number Label', labelTooltip: 'Number Tooltip' }));
		expect(number.label).toBe('Number Label');
		expect(number.labelTooltip).toBe('Number Tooltip');
	});
});

describe('resolveField actionId defaultUnitRef', () => {
	it('defaults to self when the descriptor gives no defaultUnitRef', () => {
		const spec = resolveField(actionIdFieldConfig('f', 'spells'));
		expect(spec.kind).toBe('actionId');
		if (spec.kind === 'actionId') expect(spec.defaultUnitRef).toBe('self');
	});

	it('keeps an explicit defaultUnitRef', () => {
		const spec = resolveField(actionIdFieldConfig('f', 'spells', undefined, 'currentTarget'));
		if (spec.kind === 'actionId') expect(spec.defaultUnitRef).toBe('currentTarget');
	});
});

describe('resolveField statType table', () => {
	it('leads with value -1 and then covers every Stat enum member in order', () => {
		const spec = resolveField(statTypeFieldConfig('f'));
		if (spec.kind !== 'enum') throw new Error('expected enum kind');
		expect(spec.table.options[0].value).toBe(-1);
		expect(spec.table.options.slice(1).map(option => option.value)).toEqual(getEnumValues(Stat));
	});
});

describe('resolveField auto tables by class', () => {
	const enumSpec = (descriptor: APLFieldDescriptor, player: Player<any>) => {
		const spec = resolveFieldFor(descriptor, player);
		if (spec.kind !== 'enum') throw new Error('expected enum kind');
		return spec;
	};

	it.each([Class.ClassMage, Class.ClassPriest, Class.ClassWarlock])('offers class %s only the wand', clss => {
		const autoTypes = enumSpec(autoTypeFieldConfig('f'), playerOf(clss));
		const swingTypes = enumSpec(autoSwingTypeFieldConfig('f'), playerOf(clss));
		expect(autoTypes.table.options.map(option => option.value)).toEqual([APLValueAutoAttackType.WandAuto]);
		expect(swingTypes.table.options.map(option => option.value)).toEqual([AutoSwingType.Wand]);
		expect(autoTypes.newValue()).toBe(APLValueAutoAttackType.WandAuto);
		expect(swingTypes.newValue()).toBe(AutoSwingType.Wand);
	});

	it.each([Class.ClassWarrior, Class.ClassRogue, Class.ClassHunter, Class.ClassPaladin, Class.ClassShaman, Class.ClassDruid])(
		'offers class %s no wand',
		clss => {
			const autoTypes = enumSpec(autoTypeFieldConfig('f'), playerOf(clss));
			const swingTypes = enumSpec(autoSwingTypeFieldConfig('f'), playerOf(clss));
			expect(autoTypes.table.options.map(option => option.value)).not.toContain(APLValueAutoAttackType.WandAuto);
			expect(swingTypes.table.options.map(option => option.value)).not.toContain(AutoSwingType.Wand);
			expect(autoTypes.newValue()).toBe(APLValueAutoAttackType.AnyAuto);
			expect(swingTypes.newValue()).toBe(AutoSwingType.MainHand);
		},
	);
});

describe('newKindImpl', () => {
	it('sets action-id fields and leaves the rest to the pickers', () => {
		const impl = newKindImpl({
			newValue: () => ({}),
			fields: [actionIdFieldConfig('spellId', 'castable_spells'), valueFieldConfig('lhs'), unitFieldConfig('target', 'targets')],
		})() as Record<string, unknown>;
		expect(impl).toEqual({ spellId: ActionID.create() });
	});

	it('keeps an action id the kind already set', () => {
		const spellId = ActionID.create({ rawId: { oneofKind: 'spellId', spellId: 1 } });
		const impl = newKindImpl({ newValue: () => ({ spellId }), fields: [actionIdFieldConfig('spellId', 'castable_spells')] })();
		expect(impl).toEqual({ spellId });
	});

	it('sets unset enum fields to the default their picker writes', () => {
		const impl = newKindImpl({
			newValue: () => ({ areaType: AreaType.AreaTypeUnknown, threshold: ExecutePhaseThreshold.Unknown }),
			fields: [areaTypeFieldConfig('areaType'), executePhaseThresholdFieldConfig('threshold')],
		})();
		expect(impl).toEqual({ areaType: AreaType.AreaTypeForestGrassland, threshold: ExecutePhaseThreshold.E20 });
	});
});
