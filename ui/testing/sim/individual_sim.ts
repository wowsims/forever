import { readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { APLRotation } from '@generated/proto/apl';
import { Spec } from '@generated/proto/common';
import { UIDatabase } from '@generated/proto/ui';
import { Player, type PlayerConfig, registerSpecConfig } from '@sim/player/player';
import { PlayerSpecs } from '@sim/player/specs';
import { Stats } from '@sim/proto/stats';
import { Sim } from '@sim/sim';
import type { IndividualSimHost } from '@sim/sim_host';
import type { IndividualSettingsHost } from '@sim/state/persistence';
import { applyIndividualSimSettings, type IndividualSimSerializationContext, individualSimSettingsToProto } from '@sim/state/serialization';
import { vi } from 'vitest';

import { fakeHost } from './fake_host';
import { memoryEnv } from './memory_env';

// The database every Sim loads, answered from `db` in place of the network. Call it before the
// first Sim is built: the database loads once per test file.
export const serveDatabase = (db: UIDatabase = UIDatabase.create()) => {
	vi.stubGlobal('fetch', () => Promise.resolve({ json: () => Promise.resolve(UIDatabase.toJson(db)) }));
};

// serveDatabase with the committed assets/database/db.json, for a test that needs real items and consumables.
export const serveRealDatabase = () => {
	const dbPath = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../../assets/database/db.json');
	serveDatabase(UIDatabase.fromJson(JSON.parse(readFileSync(dbPath, 'utf8'))));
};

const BARE_CONFIG = {
	autoRotation: () => APLRotation.create(),
	epStats: [],
	presets: { epWeights: [] },
};

// A real Sim with one real Player, serialized the way the sim UI does it, once its database has
// loaded. `config` is the spec's; a bare one is registered when none is given. The test mocks the
// worker pool (worker_pool.ts).
export const makeIndividualSim = async ({ spec = Spec.SpecDpsWarrior, config = BARE_CONFIG as object, env = memoryEnv() } = {}) => {
	registerSpecConfig(spec, config as PlayerConfig<any>);
	const sim = new Sim({ env });
	await sim.waitForInit();
	const player = new Player(PlayerSpecs.fromProto(spec), sim);
	sim.raid.setPlayer(0, player);
	const ctx: IndividualSimSerializationContext = { sim, player, defaultEpWeights: new Stats() };
	const host: IndividualSimHost<any> & IndividualSettingsHost = Object.assign(
		fakeHost({
			sim,
			player,
			individualConfig: config,
			toProto: (categories?: Parameters<IndividualSimHost<any>['toProto']>[0]) => individualSimSettingsToProto(ctx, categories),
			fromProto: (...args: Parameters<IndividualSimHost<any>['fromProto']>) => applyIndividualSimSettings(ctx, ...args),
		}),
		{ applyDefaults: () => sim.applyDefaults(false, false) },
	);
	return { sim, player, host };
};
