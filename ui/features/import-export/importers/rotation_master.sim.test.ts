// The same promise as rotation_master.test.ts, against a real Sim and Player: after a Rotation Master import the
// settings the simulator serializes are the ones it had, whatever they were, except the rotation.
import { IndividualSimSettings } from '@generated/proto/ui';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { encodeRmExport, type RmValue } from '../rotation_master';
import { applyRotationMasterImport, checkRotationMasterImport } from './rotation_master';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

beforeAll(() => serveDatabase());

const exportString = (value: RmValue): string => {
	const out = encodeRmExport(value);
	if (!out.ok) throw new Error(out.reason);
	return out.text;
};
const RAGE = exportString({
	rotation: [
		{ id: 'a', type: 'spell', action: 1680, ranked: false, conditions: { type: 'SPELL_AVAIL', spell: 1680 } },
		{ id: 'b', type: 'spell', action: 355, ranked: false, conditions: { type: 'NOT', value: { type: 'COMBAT', unit: 'player' } } },
	],
	cooldowns: [],
});
const withoutRotation = (settings: IndividualSimSettings): object => {
	const json = IndividualSimSettings.toJson(settings) as { player?: { rotation?: unknown } };
	if (json.player) delete json.player.rotation;
	return json;
};

describe('importing a Rotation Master rotation into a real player', () => {
	it('changes the rotation and nothing the simulator serializes', async () => {
		const { host, player } = await makeIndividualSim();
		// state the JSON importer would have reset, had the file carried none of it
		player.setName('Not the default name');
		player.setReactionTime(333);
		player.setInFrontOfTarget(true);
		const before = host.toProto();

		const checked = checkRotationMasterImport(host, RAGE);
		expect(checked.ok).toBe(true);
		if (!checked.ok) return;
		applyRotationMasterImport(host, checked.apl);
		const after = host.toProto();

		expect(withoutRotation(after)).toEqual(withoutRotation(before));
		expect(after.player?.rotation?.priorityList).toHaveLength(1);
		expect(after.player?.rotation?.prepullActions).toHaveLength(1);
		expect(player.getName()).toBe('Not the default name');
	});
});
