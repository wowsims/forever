// How many target dummies a page starts with: what its defaults.siml says, or the count of a spec that keeps its
// defaults in TypeScript.
import { APLRotation } from '@generated/proto/apl';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, EquipmentSpec, Spec } from '@generated/proto/common';
import { IndividualSimSettings, SavedTalents } from '@generated/proto/ui';
import { Stats } from '@sim/proto/stats';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { applyIndividualDefaults, type DefaultsHost } from './apply_defaults';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

// The page of `spec` once its defaults are applied: `settings` stands for a defaults.siml, and none for TS defaults.
const dummiesAfterDefaults = async (spec: Spec, settings?: IndividualSimSettings) => {
	const defaults = {
		gear: EquipmentSpec.create(),
		consumables: ConsumesSpec.create(),
		talents: SavedTalents.create(),
		specOptions: {},
		individualBuffs: IndividualBuffs.create(),
		partyBuffs: PartyBuffs.create(),
		raidBuffs: RaidBuffs.create(),
		debuffs: Debuffs.create(),
		epWeights: new Stats(),
		settings,
	};
	const config = { autoRotation: () => APLRotation.create(), epStats: [], presets: { epWeights: [] }, defaults };
	const { sim, host } = await makeIndividualSim({ spec, config });
	const stub = { applyDefaults: () => {} };
	applyIndividualDefaults(Object.assign(host, { statWeightActionSettings: stub, statSweepSettings: stub }) as unknown as DefaultsHost<any>);
	return sim.raid.getTargetDummies();
};

beforeAll(() => serveDatabase());

describe('defaults from a defaults.siml', () => {
	it('bring the dummies its raid_buffs: section writes, with no encounter: section', async () => {
		expect(await dummiesAfterDefaults(Spec.SpecHunter, IndividualSimSettings.create({ targetDummies: 3 }))).toBe(3);
	});

	it('bring none when the file writes none', async () => {
		expect(await dummiesAfterDefaults(Spec.SpecHunter, IndividualSimSettings.create())).toBe(0);
	});
});

describe('defaults in TypeScript', () => {
	it('bring none to a spec that does damage', async () => {
		expect(await dummiesAfterDefaults(Spec.SpecDpsWarrior)).toBe(0);
	});

	it('bring nine to a healing spec', async () => {
		expect(await dummiesAfterDefaults(Spec.SpecHolyPaladin)).toBe(9);
	});
});
