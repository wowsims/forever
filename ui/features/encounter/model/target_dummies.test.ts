import { HealingSimSettings, HealingSimSettings_Assignment as HealingAssignment } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import type { Sim } from '@sim/sim';
import { createSimStore, patchKeyed, PLAYER_FIELDS, seedKeyed, zeroVersions } from '@sim/state/sim_store';
import { beforeEach, describe, expect, it } from 'vitest';

import { watchTargetDummies } from './target_dummies';

// The rule reads four things off the player and writes one on the raid. A real `Player` needs a
// registered spec, a `Database` and a `Sim`; what it needs *here* is a store the subscription can
// select from, which is a slice with a version tuple in it.
const STORE_KEY = 0;

let store: ReturnType<typeof createSimStore>;
let dummies: number;
let sim: Sim;

const seedPlayerSlice = () => seedKeyed(store, 'players', STORE_KEY, { v: zeroVersions(PLAYER_FIELDS) } as never);

/** Anything the store notices as a player change; the rule does not care which field. */
const changePlayer = () => patchKeyed(store, 'players', STORE_KEY, {}, ['talentsString']);

const makePlayer = (overrides: Partial<Record<string, unknown>> = {}) =>
	({
		storeKey: STORE_KEY,
		sim,
		canEnableTargetDummies: () => true,
		shouldEnableTargetDummies: () => false,
		getPlayerSpec: () => ({ isHealingSpec: false }),
		...overrides,
	}) as unknown as Player<any>;

beforeEach(() => {
	store = createSimStore();
	dummies = 3;
	sim = {
		store,
		raid: {
			getTargetDummies: () => dummies,
			setTargetDummies: (value: number) => (dummies = value),
			getHealingSettings: () => store.getState().raid.healingSettings,
		},
	} as unknown as Sim;
	Object.assign(sim.raid, { sim });
	seedPlayerSlice();
});

describe('watchTargetDummies', () => {
	it.each([
		[HealingAssignment.TankHealing, 5, 4],
		[HealingAssignment.TankHealing, 10, 1],
		[HealingAssignment.TankHealing, 20, 2],
		[HealingAssignment.RaidHealing, 10, 9],
		[HealingAssignment.RaidHealing, 20, 19],
	] as const)('derives %i / %i targets without mounting settings', (assignment, groupSize, expected) => {
		store.setState(state => ({ raid: { ...state.raid, healingSettings: HealingSimSettings.create({ assignment, groupSize }) } }));
		watchTargetDummies(makePlayer({ getPlayerSpec: () => ({ isHealingSpec: true }) }), sim);
		expect(dummies).toBe(expected);
	});

	it('updates a healer count when an assignment is imported after initialization', () => {
		watchTargetDummies(makePlayer({ getPlayerSpec: () => ({ isHealingSpec: true }) }), sim);
		store.setState(state => ({
			raid: { ...state.raid, healingSettings: HealingSimSettings.create({ assignment: HealingAssignment.RaidHealing, groupSize: 20 }) },
		}));
		expect(dummies).toBe(19);
		store.setState(state => ({
			raid: { ...state.raid, healingSettings: HealingSimSettings.create({ assignment: HealingAssignment.TankHealing, groupSize: 20 }) },
		}));
		expect(dummies).toBe(2);
	});
	// The one that matters. A count restored from localStorage or pasted in from another spec's
	// link is stale on arrival; nothing changes the player afterwards, so only the apply at arm
	// time zeroes it, and every sim until then ships the extra allies.
	it('zeroes a stale count when it is armed', () => {
		watchTargetDummies(makePlayer(), sim);
		expect(dummies).toBe(0);
	});

	it('leaves a stale count alone for a spec that cannot enable them at all, so the initial apply stays behind the guard', () => {
		watchTargetDummies(makePlayer({ canEnableTargetDummies: () => false }), sim);
		expect(dummies).toBe(3);
	});

	it('zeroes the count once the player can no longer enable them', () => {
		watchTargetDummies(makePlayer(), sim);
		changePlayer();
		expect(dummies).toBe(0);
	});

	it('leaves a count the player can still enable', () => {
		watchTargetDummies(makePlayer({ shouldEnableTargetDummies: () => true }), sim);
		changePlayer();
		expect(dummies).toBe(3);
	});

	it('never arms for a spec that cannot enable them at all', () => {
		watchTargetDummies(makePlayer({ canEnableTargetDummies: () => false }), sim);
		changePlayer();
		expect(dummies).toBe(3);
	});
});
