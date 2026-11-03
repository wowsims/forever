import { APLRotation } from '@generated/proto/apl';
import { makePlayer } from '@testing/apl/player';
import { describe, expect, it, vi } from 'vitest';

import { valueKindOptions } from './kind_options';
import { valueKinds } from './value_kinds';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

const DESIGN_VALUES = [
	['targetIsActive', 'target_is_active', ['targetUnit']],
	['numActiveAdds', 'num_active_adds', []],
	['timeUntilTargetActive', 'time_until_target_active', ['targetUnit']],
	['timeUntilTargetDies', 'time_until_target_dies', ['targetUnit']],
	['timeUntilNextMove', 'time_until_next_move', []],
	['nextMoveDuration', 'next_move_duration', []],
	['ampIsActive', 'amp_is_active', ['ampType']],
	['timeUntilAmp', 'time_until_amp', ['ampType']],
	['bossAbilityIsActive', 'boss_ability_is_active', ['targetUnit', 'abilityName']],
	['timeUntilBossAbility', 'time_until_boss_ability', ['targetUnit', 'abilityName']],
	['bossAbilityCastTimeRemaining', 'boss_ability_cast_time_remaining', ['targetUnit', 'abilityName']],
	['bossAbilityIsKnown', 'boss_ability_is_known', ['targetUnit', 'abilityName']],
] as const;

const BOSS_ABILITY_VALUES = ['bossAbilityIsActive', 'timeUntilBossAbility', 'bossAbilityCastTimeRemaining', 'bossAbilityIsKnown'] as const;

describe('encounter design values', () => {
	it.each(DESIGN_VALUES)('offers %s under Encounter, before and during the fight', (kind, key, fields) => {
		const player = makePlayer(APLRotation.create()) as any;
		for (const isPrepull of [false, true]) {
			const option = valueKindOptions(player, isPrepull, false).find(entry => entry.value === kind);
			expect(option).toMatchObject({ label: `rotation_tab.apl.values.${key}.label`, submenu: ['encounter'] });
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
