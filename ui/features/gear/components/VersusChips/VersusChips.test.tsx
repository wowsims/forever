import { PlayerStats, VersusEffect, VersusEffectKind } from '@generated/proto/api';
import { AreaStats, AreaType, CreatureFamily, MobType, MobTypeStats, ScalingItemProperties, Stat, Target } from '@generated/proto/common';
import { UIItem } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { createSimStore, patchSlice, PLAYER_FIELDS, type PlayerSlice, seedKeyed, zeroVersions } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { act, fireEvent, screen } from '@testing-library/react';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';

import english from '../../../../../assets/locales/en/translation.json';
import { VersusChips } from './VersusChips';

const KEY = 0;

const ITEM = UIItem.create({
	id: 1,
	mobTypes: [MobType.MobTypeUndead],
	creatureFamilies: [CreatureFamily.CreatureFamilyWolf],
	areaTypes: [AreaType.AreaTypeSnowy],
	scalingOptions: {
		0: ScalingItemProperties.create({
			mobTypeStats: [MobTypeStats.create({ mobTypes: [MobType.MobTypeUndead], stats: { [Stat.StatAttackPower]: 81 } })],
			areaStats: [AreaStats.create({ areaType: AreaType.AreaTypeSnowy, stats: { [Stat.StatAttackPower]: 28 } })],
		}),
	},
});

const target = (mobType: MobType, creatureFamily = CreatureFamily.CreatureFamilyUnknown) => Target.create({ mobType, creatureFamily });

const setup = (item: UIItem, targets: Array<Target>, areaTypes: Array<AreaType> = [], versusEffects: Array<VersusEffect> = []) => {
	const store = createSimStore();
	patchSlice(store, 'encounter', { targets, areaTypes });
	seedKeyed(store, 'players', KEY, { currentStats: PlayerStats.create({ versusEffects }), v: zeroVersions(PLAYER_FIELDS) } as unknown as PlayerSlice);
	renderWithHost(<VersusChips item={item} />, { player: { storeKey: KEY, sim: { store } }, sim: { store } });
	return {
		setEncounter: (nextTargets: Array<Target>, nextAreaTypes: Array<AreaType>) =>
			act(() => patchSlice(store, 'encounter', { targets: nextTargets, areaTypes: nextAreaTypes })),
	};
};

const chips = () =>
	screen.queryAllByTestId('gear-versus-chip').map(chip => ({
		text: chip.textContent,
		state: chip.dataset.state,
		description: document.getElementById(chip.getAttribute('aria-describedby')!)!.textContent,
	}));

// The test setup loads no locale, so the English strings are added to show the names and reasons the chips read.
describe('VersusChips', () => {
	beforeAll(() => {
		i18n.addResourceBundle('en', 'translation', english, true, true);
	});
	afterAll(() => {
		i18n.removeResourceBundle('en', 'translation');
	});

	it('marks each mob type, family and area the item names as on, partial or off, with its value and why', () => {
		setup(ITEM, [target(MobType.MobTypeUndead), target(MobType.MobTypeMechanical)]);

		expect(chips()).toEqual([
			{ text: 'Undead 1/2', state: 'partial', description: '+81 Attack Power. Active on 1 of 2 targets' },
			{ text: 'Wolf', state: 'off', description: 'Inactive: targets are Undead (Target 1), Mechanical (Target 2)' },
			{ text: 'Snowy', state: 'off', description: '+28 Attack Power. Inactive: the encounter is not in a Snowy area' },
		]);
	});

	it('follows the encounter as its targets and areas change', () => {
		const { setEncounter } = setup(ITEM, [target(MobType.MobTypeBeast)]);
		expect(chips().map(({ state, description }) => [state, description])).toEqual([
			['off', '+81 Attack Power. Inactive: target is Beast'],
			['off', 'Inactive: target is Beast'],
			['off', '+28 Attack Power. Inactive: the encounter is not in a Snowy area'],
		]);

		setEncounter([target(MobType.MobTypeUndead), target(MobType.MobTypeUndead)], [AreaType.AreaTypeSnowy]);
		expect(chips().map(({ state, description }) => [state, description])).toEqual([
			['on', '+81 Attack Power. Active on all 2 targets'],
			['off', 'Inactive: targets are Undead (Target 1), Undead (Target 2)'],
			['on', '+28 Attack Power. Active: the encounter is in a Snowy area'],
		]);

		setEncounter([], []);
		expect(chips()[0].description).toBe('+81 Attack Power. Inactive: the encounter has no targets');
	});

	it("shows the sim's factor for the worn item in the tooltip, and no other item's", async () => {
		const effect = (itemId: number, multiplier: number) =>
			VersusEffect.create({ itemId, creatureFamilies: [CreatureFamily.CreatureFamilyWolf], multiplier, kind: VersusEffectKind.VersusEffectKindDamage });
		setup(ITEM, [target(MobType.MobTypeBeast, CreatureFamily.CreatureFamilyWolf)], [], [effect(1, 1.5), effect(2, 3)]);

		const wolf = screen.getAllByTestId('gear-versus-chip')[1];
		expect(wolf.dataset.state).toBe('on');
		fireEvent.mouseEnter(wolf);
		expect((await screen.findByRole('tooltip')).textContent).toBe('Against Wolf×1.5 damageActive: target is Wolf');
	});

	it('shows the crit chance factor and the proc-only note of the worn item', async () => {
		const wolf = (kind: VersusEffectKind, multiplier: number) =>
			VersusEffect.create({ itemId: 1, creatureFamilies: [CreatureFamily.CreatureFamilyWolf], multiplier, kind });
		setup(
			ITEM,
			[target(MobType.MobTypeBeast, CreatureFamily.CreatureFamilyWolf)],
			[],
			[wolf(VersusEffectKind.VersusEffectKindCritChance, 2), wolf(VersusEffectKind.VersusEffectKindProcOnly, 0)],
		);

		fireEvent.mouseEnter(screen.getAllByTestId('gear-versus-chip')[1]);
		expect((await screen.findByRole('tooltip')).textContent).toBe('Against Wolf×2 crit chanceProcs only against theseActive: target is Wolf');
	});

	it('shows nothing for an item that names no kind of target or area', () => {
		setup(UIItem.create({ id: 1 }), [target(MobType.MobTypeUndead)], [AreaType.AreaTypeSnowy]);

		expect(chips()).toEqual([]);
	});
});
