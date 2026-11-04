import { describe, expect, it } from 'vitest';

import type { Player } from '../player/player';
import type { Sim } from '../sim';
import { createSimStore } from '../state/sim_store';
import { Raid } from './raid';

const makeRaid = () => new Raid({ store: createSimStore() } as unknown as Sim);
const fakePlayer = (storeKey: number) => ({ storeKey, getParty: () => null, setParty: () => {} }) as unknown as Player<any>;

describe('Raid.getTargetDummySlots', () => {
	it('takes the first empty slots of the active parties, in raid order', () => {
		const raid = makeRaid();
		raid.setPlayer(0, fakePlayer(1));
		raid.setPlayer(2, fakePlayer(2));
		raid.setNumActiveParties(2);
		raid.setTargetDummies(4);
		expect(raid.getTargetDummySlots()).toEqual([1, 3, 4, 5]);

		raid.setNumActiveParties(1);
		expect(raid.getTargetDummySlots()).toEqual([1, 3, 4]);
	});

	it('places none without dummies and 24 at most', () => {
		const raid = makeRaid();
		expect(raid.getTargetDummySlots()).toEqual([]);

		raid.setTargetDummies(30);
		expect(raid.getTargetDummySlots()).toEqual([...Array(24).keys()]);
	});
});
