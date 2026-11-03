import { APLRotation } from '@generated/proto/apl';
import { makePlayer } from '@testing/apl/player';
import { describe, expect, it, vi } from 'vitest';

import { valueKindOptions } from './kind_options';
import { valueKinds } from './value_kinds';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

// The last column is whether the value is offered before the fight too; one that reads the clock is not.
const DESIGN_VALUES = [
	['targetIsActive', 'target_is_active', ['targetUnit'], true],
	['numActiveAdds', 'num_active_adds', [], true],
	['timeUntilTargetActive', 'time_until_target_active', ['targetUnit'], false],
	['timeUntilTargetDies', 'time_until_target_dies', ['targetUnit'], false],
	['timeUntilNextMove', 'time_until_next_move', [], false],
	['nextMoveDuration', 'next_move_duration', [], true],
	['ampIsActive', 'amp_is_active', ['ampType'], true],
	['timeUntilAmp', 'time_until_amp', ['ampType'], false],
	['bossAbilityIsActive', 'boss_ability_is_active', ['targetUnit', 'abilityName'], true],
	['timeUntilBossAbility', 'time_until_boss_ability', ['targetUnit', 'abilityName'], false],
	['bossAbilityCastTimeRemaining', 'boss_ability_cast_time_remaining', ['targetUnit', 'abilityName'], false],
	['bossAbilityIsKnown', 'boss_ability_is_known', ['targetUnit', 'abilityName'], true],
] as const;

const BOSS_ABILITY_VALUES = ['bossAbilityIsActive', 'timeUntilBossAbility', 'bossAbilityCastTimeRemaining', 'bossAbilityIsKnown'] as const;

describe('encounter design values', () => {
	it.each(DESIGN_VALUES)('offers %s under Encounter during the fight, and before it only when it reads no clock', (kind, key, fields, prepull) => {
		const player = makePlayer(APLRotation.create()) as any;
		const option = (isPrepull: boolean) => valueKindOptions(player, isPrepull, false).find(entry => entry.value === kind);
		expect(option(false)).toMatchObject({ label: `rotation_tab.apl.values.${key}.label`, submenu: ['encounter'] });
		if (prepull) {
			expect(option(true)).toMatchObject({ label: `rotation_tab.apl.values.${key}.label`, submenu: ['encounter'] });
		} else {
			expect(option(true)).toBeUndefined();
		}
		expect(valueKinds[kind].fields.map(field => ('field' in field ? field.field : field.type))).toEqual(fields);
	});

	it('gives the execute phase an optional target', () => {
		expect(valueKinds.isExecutePhase.fields).toContainEqual({ type: 'unit', field: 'targetUnit', unitSet: 'targets' });
	});

	it.each(BOSS_ABILITY_VALUES)('gives %s a unit whose empty choice is any target', kind => {
		expect(valueKinds[kind].fields[0]).toEqual({ type: 'unit', field: 'targetUnit', unitSet: 'any_target' });
	});
});
