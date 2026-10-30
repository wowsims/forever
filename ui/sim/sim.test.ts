import {
	BulkCandidatesResult,
	BulkEnchantSlotOptions,
	BulkSettings,
	BulkSimResult,
	ErrorOutcome,
	ErrorOutcomeType,
	Raid as RaidProto,
	RaidSimRequest,
	RaidSimResult,
	ReforgeOptimizeResult,
	ReforgeSettings,
} from '@generated/proto/api';
import {
	ConsumableType,
	Encounter as EncounterProto,
	EquipmentSpec,
	ItemRandomSuffix,
	ItemSlot,
	ItemSpec,
	ItemType,
	Profession,
	Stat,
	Target,
} from '@generated/proto/common';
import { UIEnchant, UIItem } from '@generated/proto/ui';
import { makeEquippedItem } from '@testing/storybook/fixtures/gear';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { EquippedItem } from './proto/equipped_item';
import { Gear } from './proto/gear';
import { Stats } from './proto/stats';
import { Sim, SimError } from './sim';
import { RequestTypes } from './sim_signal_manager';
import type { Env } from './state/env';

const mocks = vi.hoisted(() => ({
	raidSimAsync: vi.fn(),
	runConcurrentSim: vi.fn(),
	isWasm: vi.fn(() => Promise.resolve(false)),
	getNumWorkers: vi.fn(() => 1),
	computeStats: vi.fn(() => new Promise(() => {})),
	reforgeOptimizeAsync: vi.fn(),
	runConcurrentBulkSim: vi.fn(),
	bulkSimAsync: vi.fn(),
	bulkCandidates: vi.fn(),
	db: {} as Record<string, unknown>,
}));

vi.mock('./workers/worker_pool', () => ({
	generateRequestId: () => 'test-request',
	WorkerPool: class {
		setNumWorkers() {}
		getNumWorkers = mocks.getNumWorkers;
		isWasm = mocks.isWasm;
		raidSimAsync = mocks.raidSimAsync;
		computeStats = mocks.computeStats;
		reforgeOptimizeAsync = mocks.reforgeOptimizeAsync;
		bulkSimAsync = mocks.bulkSimAsync;
		bulkCandidates = mocks.bulkCandidates;
	},
}));

beforeEach(() => {
	mocks.isWasm.mockReset().mockResolvedValue(false);
	for (const key of Object.keys(mocks.db)) delete mocks.db[key];
});

vi.mock('./wasm', () => ({
	runConcurrentBulkSim: mocks.runConcurrentBulkSim,
	runConcurrentSim: mocks.runConcurrentSim,
	runConcurrentStatWeights: vi.fn(),
}));

vi.mock('./proto/database', () => ({
	Database: { get: () => Promise.resolve(mocks.db) },
}));

const makeEnv = (): Env => ({
	storage: { getItem: () => null, setItem: () => {}, removeItem: () => {} },
	location: { hash: '', search: '', href: 'http://localhost/', hostname: 'localhost', setHash: () => {} },
	hardwareConcurrency: 1,
	onPageHide: () => () => {},
});

// A stand-in for the request `makeRaidSimRequest` builds: every option the merged call acts on is
// read back off the request or the player, so stubbing it keeps the assertions on the option
// handling rather than on proto assembly.
const makeRequest = (iterations: number, profession = Profession.ProfessionUnknown) =>
	RaidSimRequest.create({
		raid: { parties: [{ players: [{ profession1: profession }] }] },
		simOptions: { iterations },
	});

const makeGear = (label: string) =>
	({
		withoutEnchanting: vi.fn(() => makeGear(`${label}-stripped`)),
		toDatabase: vi.fn(() => ({ label })),
		asSpec: vi.fn(() => ({ label })),
		adjustImbues: vi.fn((consumables: unknown) => consumables),
	}) as unknown as Gear;

const runningSignals = (sim: Sim) => (sim.signalManager as unknown as { running: Map<unknown, unknown> }).running;

const makeSim = ({ iterations = 100, profession = Profession.ProfessionUnknown, concurrent = false } = {}) => {
	const sim = new Sim({ env: makeEnv() });
	vi.spyOn(sim.raid, 'isEmpty').mockReturnValue(false);
	vi.spyOn(sim.encounter, 'getTargets').mockReturnValue([{}] as any);
	vi.spyOn(sim, 'waitForInit').mockResolvedValue();
	vi.spyOn(sim, 'shouldUseWasmConcurrency').mockResolvedValue(concurrent);
	vi.spyOn(sim, 'makeRaidSimRequest').mockImplementation(() => makeRequest(iterations, profession));
	return sim;
};

describe('Sim.runSim', () => {
	beforeEach(() => {
		mocks.raidSimAsync.mockReset().mockResolvedValue(RaidSimResult.create());
		mocks.runConcurrentSim.mockReset().mockResolvedValue(RaidSimResult.create());
	});

	it('rejects before registering anything when the raid is empty', async () => {
		const sim = makeSim();
		vi.spyOn(sim.raid, 'isEmpty').mockReturnValue(true);

		await expect(sim.runSim({ raw: true })).rejects.toThrow('sim.run_errors.raid_empty');
		expect(runningSignals(sim).size).toBe(0);
		expect(mocks.raidSimAsync).not.toHaveBeenCalled();
	});

	it('rejects when the encounter has no targets', async () => {
		const sim = makeSim();
		vi.spyOn(sim.encounter, 'getTargets').mockReturnValue([]);

		await expect(sim.runSim({ raw: true })).rejects.toThrow('sim.run_errors.no_targets');
	});

	it('substitutes the gear option onto the player, stripping ring enchants for a non-enchanter', async () => {
		const sim = makeSim();
		const gear = makeGear('gear');

		await sim.runSim({ gear, raw: true });

		expect(gear.withoutEnchanting).toHaveBeenCalled();
		const [request] = mocks.raidSimAsync.mock.calls[0];
		const player = request.raid.parties[0].players[0];
		expect(player.database).toEqual({ label: 'gear-stripped' });
		expect(player.equipment).toEqual({ label: 'gear-stripped' });
	});

	it('keeps the ring enchants when the player is an enchanter', async () => {
		const sim = makeSim({ profession: Profession.Enchanting });
		const gear = makeGear('gear');

		await sim.runSim({ gear, raw: true });

		expect(gear.withoutEnchanting).not.toHaveBeenCalled();
		const [request] = mocks.raidSimAsync.mock.calls[0];
		expect(request.raid.parties[0].players[0].database).toEqual({ label: 'gear' });
	});

	it('leaves the player untouched when no gear option is given', async () => {
		const sim = makeSim();

		await sim.runSim({ raw: true });

		const [request] = mocks.raidSimAsync.mock.calls[0];
		const player = request.raid.parties[0].players[0];
		expect(player.database).toBeUndefined();
		expect(player.equipment).toBeUndefined();
	});

	it('returns the raw request/result pair without building or emitting a SimResult', async () => {
		const sim = makeSim();
		const result = RaidSimResult.create();
		mocks.raidSimAsync.mockResolvedValue(result);
		const emitted = vi.fn();
		sim.simResultEmitter.on(emitted);

		const returned = await sim.runSim({ raw: true });

		expect(Array.isArray(returned)).toBe(true);
		const [request, raw] = returned as [RaidSimRequest, RaidSimResult];
		expect(request).toBe(mocks.raidSimAsync.mock.calls[0][0]);
		expect(raw).toBe(result);
		expect(emitted).not.toHaveBeenCalled();
	});

	it('returns an aborted outcome instead of throwing, and unregisters its signals', async () => {
		const sim = makeSim();
		const aborted = RaidSimResult.create({ error: ErrorOutcome.create({ type: ErrorOutcomeType.ErrorOutcomeAborted, message: 'aborted' }) });
		mocks.raidSimAsync.mockResolvedValue(aborted);

		const returned = await sim.runSim();

		expect(returned).toBe(aborted.error);
		expect(runningSignals(sim).size).toBe(0);
	});

	it('still throws a SimError for a real error outcome', async () => {
		const sim = makeSim();
		mocks.raidSimAsync.mockResolvedValue(
			RaidSimResult.create({ error: ErrorOutcome.create({ type: ErrorOutcomeType.ErrorOutcomeError, message: 'boom' }) }),
		);

		await expect(sim.runSim()).rejects.toThrow(SimError);
		expect(runningSignals(sim).size).toBe(0);
	});

	it('does not split a single iteration across workers even with wasm concurrency available', async () => {
		const sim = makeSim({ iterations: 1, concurrent: true });

		await sim.runSim({ singleIteration: true, raw: true });

		expect(mocks.raidSimAsync).toHaveBeenCalledTimes(1);
		expect(mocks.runConcurrentSim).not.toHaveBeenCalled();
	});

	it('splits a multi-iteration run when wasm concurrency is available', async () => {
		const sim = makeSim({ iterations: 100, concurrent: true });

		await sim.runSim({ raw: true });

		expect(mocks.runConcurrentSim).toHaveBeenCalledTimes(1);
		expect(mocks.raidSimAsync).not.toHaveBeenCalled();
	});

	it('passes a no-op progress callback through when the option is omitted', async () => {
		const sim = makeSim();

		await sim.runSim({ raw: true });

		expect(typeof mocks.raidSimAsync.mock.calls[0][1]).toBe('function');
		expect(() => mocks.raidSimAsync.mock.calls[0][1]({})).not.toThrow();
	});

	it('forwards the progress callback option to the worker', async () => {
		const sim = makeSim();
		const onProgress = vi.fn();

		await sim.runSim({ onProgress, raw: true });

		expect(mocks.raidSimAsync.mock.calls[0][1]).toBe(onProgress);
	});
});

describe('Sim.updateCharacterStats', () => {
	const statsResult = (health: number) => ({
		errorResult: '',
		raidStats: { parties: [{ players: [{ finalStats: { stats: [health] } }] }] },
		encounterStats: { targets: [] },
	});

	const makeStatsSim = () => {
		const sim = makeSim();
		vi.spyOn(sim, 'getModifiedRaidProto').mockReturnValue({} as any);
		vi.spyOn(sim.encounter, 'toProto').mockReturnValue({} as any);
		vi.spyOn(sim.encounter.targetsMetadata, 'update').mockResolvedValue(false);
		return sim;
	};

	beforeEach(() => {
		mocks.computeStats.mockReset();
	});

	// The hazard: both calls build a request from current state, and whichever reply lands last
	// writes. Without a guard the older one can land last and overwrite the newer stats.
	it('discards a reply that a newer recompute has already superseded', async () => {
		const sim = makeStatsSim();
		const player = { setCurrentStats: vi.fn(), updateMetadata: vi.fn(async () => false) };
		vi.spyOn(sim.raid, 'getPlayers').mockReturnValue([player] as any);

		let settleFirst!: (value: unknown) => void;
		mocks.computeStats
			.mockImplementationOnce(() => new Promise(resolve => (settleFirst = resolve)))
			.mockImplementationOnce(() => Promise.resolve(statsResult(200)));

		const stale = sim.updateCharacterStats();
		const fresh = sim.updateCharacterStats();
		await fresh;
		settleFirst(statsResult(100));
		await stale;

		expect(player.setCurrentStats).toHaveBeenCalledTimes(1);
		expect(player.setCurrentStats.mock.calls[0][0]).toEqual({ finalStats: { stats: [200] } });
	});

	it('writes when nothing has superseded it', async () => {
		const sim = makeStatsSim();
		const player = { setCurrentStats: vi.fn(), updateMetadata: vi.fn(async () => false) };
		vi.spyOn(sim.raid, 'getPlayers').mockReturnValue([player] as any);
		mocks.computeStats.mockResolvedValue(statsResult(300));

		await sim.updateCharacterStats();

		expect(player.setCurrentStats).toHaveBeenCalledWith({ finalStats: { stats: [300] } });
	});

	// A stale error would otherwise raise a crash toast for work whose result nobody wants.
	it('does not report an error from a superseded reply', async () => {
		const sim = makeStatsSim();
		vi.spyOn(sim.raid, 'getPlayers').mockReturnValue([] as any);
		const crashes: unknown[] = [];
		sim.crashEmitter.on(e => crashes.push(e));

		let settleFirst!: (value: unknown) => void;
		mocks.computeStats
			.mockImplementationOnce(() => new Promise(resolve => (settleFirst = resolve)))
			.mockImplementationOnce(() => Promise.resolve(statsResult(400)));

		const stale = sim.updateCharacterStats();
		const fresh = sim.updateCharacterStats();
		await fresh;
		settleFirst({ errorResult: 'boom' });
		await stale;

		expect(crashes).toEqual([]);
	});
});

describe('Sim stats for an edited player', () => {
	const hit = (value: number) => Stats.fromMap({ [Stat.StatHitRating]: value });

	const makeBonusSim = () => {
		const sim = makeSim();
		vi.spyOn(sim, 'getModifiedRaidProto').mockReturnValue(RaidProto.create({ parties: [{ players: [{ bonusStats: hit(5).toProto() }] }] }));
		vi.spyOn(sim.encounter, 'toProto').mockReturnValue({} as any);
		return sim;
	};

	beforeEach(() => {
		mocks.computeStats.mockReset();
	});

	it('reports an error result for a gear set as a crash and still returns the stats', async () => {
		const sim = makeBonusSim();
		const crashes: unknown[] = [];
		sim.crashEmitter.on(e => crashes.push(e));
		mocks.computeStats.mockResolvedValue({ errorResult: 'boom', raidStats: { parties: [{ players: [{ finalStats: hit(7).toProto() }] }] } } as any);

		const playerStats = await sim.getCharacterStatsForGear(makeGear('candidate'));

		expect(crashes).toEqual([new SimError('boom')]);
		expect(Stats.fromProto(playerStats.finalStats).getStat(Stat.StatHitRating)).toBe(7);
	});

	it('hands back the stats of the raid it is given, an error result included, and reports no crash', async () => {
		const sim = makeBonusSim();
		const crashes: unknown[] = [];
		sim.crashEmitter.on(e => crashes.push(e));
		const result = { errorResult: 'boom', raidStats: { parties: [{ players: [{ finalStats: hit(7).toProto() }] }] } };
		mocks.computeStats.mockResolvedValue(result as any);
		const raid = RaidProto.create({ parties: [{ players: [{ bonusStats: hit(15).toProto() }] }] });

		expect(await sim.computeRaidStats(raid)).toBe(result);
		const [request] = mocks.computeStats.mock.calls[0] as unknown as [{ raid: RaidProto }];
		expect(request.raid).toEqual(raid);
		expect(crashes).toEqual([]);
	});
});

describe('Sim.runRaidSimRequest', () => {
	beforeEach(() => {
		mocks.raidSimAsync.mockReset();
	});

	it('runs the request it is given and emits no result', async () => {
		const sim = makeSim();
		const emitted = vi.fn();
		sim.simResultEmitter.on(emitted);
		const result = RaidSimResult.create({ raidMetrics: { parties: [] } });
		mocks.raidSimAsync.mockResolvedValue(result);
		const request = makeRequest(2);

		expect(await sim.runRaidSimRequest(request, RequestTypes.StatSweep, vi.fn())).toBe(result);
		expect(mocks.raidSimAsync.mock.calls[0][0]).toBe(request);
		expect(emitted).not.toHaveBeenCalled();
	});

	it('is stopped by an abort of its own type and not by an abort of the plain sims', async () => {
		const sim = makeSim();
		const aborted = ErrorOutcome.create({ type: ErrorOutcomeType.ErrorOutcomeAborted });
		mocks.raidSimAsync.mockImplementation(
			(_request, _onProgress, signals) => new Promise(resolve => signals.abort.onTrigger(async () => resolve(RaidSimResult.create({ error: aborted })))),
		);
		const run = sim.runRaidSimRequest(makeRequest(2), RequestTypes.StatSweep, vi.fn());
		await vi.waitFor(() => expect(mocks.raidSimAsync).toHaveBeenCalled());

		await sim.signalManager.abortType(RequestTypes.IndividualSim | RequestTypes.BulkSim);
		let settled = false;
		void run.then(() => (settled = true));
		await Promise.resolve();
		expect(settled).toBe(false);

		await sim.signalManager.abortType(RequestTypes.StatSweep);
		expect((await run).error).toEqual(aborted);
	});

	it('throws a run that failed', async () => {
		mocks.raidSimAsync.mockResolvedValue(
			RaidSimResult.create({ error: ErrorOutcome.create({ type: ErrorOutcomeType.ErrorOutcomeError, message: 'boom' }) }),
		);

		await expect(makeSim().runRaidSimRequest(makeRequest(2), RequestTypes.StatSweep, vi.fn())).rejects.toThrow(new SimError('boom'));
	});
});

describe('Sim.makeRaidSimRequest seed', () => {
	const seedOf = (sim: Sim) => {
		vi.mocked(sim.makeRaidSimRequest).mockRestore();
		vi.spyOn(sim, 'getModifiedRaidProto').mockReturnValue(RaidProto.create());
		vi.spyOn(sim.encounter, 'toProto').mockReturnValue({} as any);
		return sim.makeRaidSimRequest().simOptions!.randomSeed;
	};

	it('never draws 0, which the sim would replace with the clock', () => {
		const random = vi.spyOn(Math, 'random').mockReturnValue(0);
		try {
			expect(seedOf(makeSim())).toBeGreaterThan(0n);
		} finally {
			random.mockRestore();
		}
	});

	it('takes the fixed seed when one is set', () => {
		const sim = makeSim();
		sim.setFixedRngSeed(42);

		expect(seedOf(sim)).toBe(42n);
	});
});

describe('Sim.reforgeOptimize', () => {
	it('ships every row of each enchant the optimizer may pick, beside the equipped ones', async () => {
		const crusader = UIEnchant.create({ effectId: 1900, itemId: 1, name: 'Enchant Weapon - Crusader', type: ItemType.ItemTypeWeapon });
		const lesserImpact = UIEnchant.create({ effectId: 1897, itemId: 2, name: 'Enchant 2H Weapon - Lesser Impact', type: ItemType.ItemTypeWeapon });
		const superiorStriking = UIEnchant.create({ effectId: 1897, itemId: 3, name: 'Enchant Weapon - Superior Striking', type: ItemType.ItemTypeWeapon });
		Object.assign(mocks.db, {
			getItemEffectRandPropPoints: () => undefined,
			getEnchantsByEffectId: (effectId: number) => [crusader, lesserImpact, superiorStriking].filter(enchant => enchant.effectId === effectId),
		});
		mocks.reforgeOptimizeAsync.mockReset().mockResolvedValue(ReforgeOptimizeResult.create());
		const sim = makeSim();
		vi.spyOn(sim, 'getModifiedRaidProto').mockReturnValue(RaidProto.create({ parties: [{ players: [{}] }] }));

		await sim.reforgeOptimize({
			gear: new Gear({ [ItemSlot.ItemSlotMainHand]: makeEquippedItem({ enchant: crusader }) }),
			preCapEPWeights: new Stats(),
			undershootCaps: new Stats(),
			settings: ReforgeSettings.create(),
			softCaps: [],
			enchantOptions: [BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotMainHand, effectIds: [1897] })],
		});

		const [request] = mocks.reforgeOptimizeAsync.mock.calls[0];
		const shipped = request.raid.parties[0].players[0].database.enchants.map((enchant: { name: string }) => enchant.name).sort();
		expect(shipped).toEqual([crusader.name, lesserImpact.name, superiorStriking.name].sort());
	});
});

describe('Sim.runBulkSim', () => {
	const equipmentWithBack = (back: ItemSpec) =>
		EquipmentSpec.create({
			items: Array.from({ length: ItemSlot.ItemSlotRanged + 1 }, (_, slot) => (slot === ItemSlot.ItemSlotBack ? back : ItemSpec.create())),
		});

	it('ships every item, suffix and enchant a candidate can carry, from the baseline and the settings alone', async () => {
		const cloakAgility = UIEnchant.create({ effectId: 849, itemId: 1, name: 'Enchant Cloak - Minor Agility', type: ItemType.ItemTypeBack });
		const cloakResistance = UIEnchant.create({ effectId: 1889, itemId: 2, name: 'Enchant Cloak - Superior Defense', type: ItemType.ItemTypeBack });
		const cloakDodge = UIEnchant.create({ effectId: 2622, itemId: 3, name: 'Enchant Cloak - Dodge', type: ItemType.ItemTypeBack });
		const baselineCloak = new EquippedItem({
			item: UIItem.create({ id: 20, type: ItemType.ItemTypeBack, scalingOptions: { 0: { ilvl: 60 } } }),
			enchant: cloakAgility,
			randomSuffix: ItemRandomSuffix.create({ id: 7 }),
		});
		const selectedCloak = new EquippedItem({
			item: UIItem.create({ id: 30, type: ItemType.ItemTypeBack, scalingOptions: { 0: { ilvl: 60 } } }),
			enchant: cloakDodge,
			randomSuffix: ItemRandomSuffix.create({ id: 8 }),
		});
		Object.assign(mocks.db, {
			lookupItemSpec: (spec: ItemSpec) => (spec.id === 30 ? selectedCloak : null),
			getItemEffectRandPropPoints: () => undefined,
			getEnchantsByEffectId: (effectId: number) => [cloakAgility, cloakResistance, cloakDodge].filter(enchant => enchant.effectId === effectId),
		});
		mocks.isWasm.mockResolvedValue(true);
		mocks.runConcurrentBulkSim.mockReset().mockResolvedValue(BulkSimResult.create());
		const sim = makeSim();
		vi.spyOn(sim.raid, 'getActivePlayers').mockReturnValue([{ getGear: () => new Gear({ [ItemSlot.ItemSlotBack]: baselineCloak }) }] as any);
		const backSpecs = [
			ItemSpec.create({ id: 20, randomSuffix: 7, enchant: 849 }),
			ItemSpec.create({ id: 20, randomSuffix: 7, enchant: 1889 }),
			ItemSpec.create({ id: 30, randomSuffix: 8, enchant: 2622 }),
			ItemSpec.create({ id: 30, randomSuffix: 8, enchant: 1889 }),
		];
		mocks.bulkCandidates
			.mockReset()
			.mockResolvedValue(BulkCandidatesResult.create({ candidates: backSpecs.map((back, index) => ({ index, gear: equipmentWithBack(back) })) }));

		await sim.runBulkSim(
			() => {},
			BulkSettings.create({
				items: [ItemSpec.create({ id: 30, randomSuffix: 8, enchant: 2622 })],
				enchants: [BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [1889] })],
			}),
		);

		const [request] = mocks.runConcurrentBulkSim.mock.calls[0];
		const database = request.baseRequest.raid.parties[0].players[0].database;
		const shipped = {
			items: new Set(database.items.map((item: { id: number }) => item.id)),
			suffixes: new Set(database.randomSuffixes.map((suffix: { id: number }) => suffix.id)),
			enchants: new Set(database.enchants.map((enchant: { effectId: number }) => enchant.effectId)),
		};
		for (const back of backSpecs) {
			expect(shipped.items).toContain(back.id);
			expect(shipped.suffixes).toContain(back.randomSuffix);
			expect(shipped.enchants).toContain(back.enchant);
		}
	});

	const makeEmptyBulkSim = () => {
		Object.assign(mocks.db, { lookupItemSpec: () => null, getItemEffectRandPropPoints: () => undefined, getEnchantsByEffectId: () => [] });
		mocks.runConcurrentBulkSim.mockReset().mockResolvedValue(BulkSimResult.create());
		mocks.bulkSimAsync.mockReset().mockResolvedValue(BulkSimResult.create());
		mocks.bulkCandidates.mockReset();
		const sim = makeSim();
		vi.spyOn(sim.raid, 'getActivePlayers').mockReturnValue([{ getGear: () => new Gear({}) }] as any);
		return sim;
	};

	it('sends the native sim no candidates, so it generates them itself', async () => {
		const sim = makeEmptyBulkSim();

		await sim.runBulkSim(() => {}, BulkSettings.create());

		expect(mocks.bulkCandidates).not.toHaveBeenCalled();
		expect(mocks.bulkSimAsync.mock.calls[0][0].candidates).toEqual([]);
	});

	it('sends an override encounter whole and leaves the live encounter as it is', async () => {
		const sim = makeEmptyBulkSim();
		const live = sim.encounter.toProto();
		const override = EncounterProto.create({
			duration: 90,
			durationVariation: 10,
			executeProportion20: 0.5,
			targets: [Target.create(), Target.create()],
		});

		await sim.runBulkSim(() => {}, BulkSettings.create(), undefined, override);

		const sent = mocks.bulkSimAsync.mock.calls[0][0].baseRequest.encounter;
		expect(sent).toEqual(override);
		expect(sent).not.toBe(override);
		expect(sim.encounter.toProto()).toEqual(live);
	});

	it('drops the WASM candidates that repeat the baseline or one another', async () => {
		mocks.isWasm.mockResolvedValue(true);
		const sim = makeEmptyBulkSim();
		const cloak = equipmentWithBack(ItemSpec.create({ id: 30 }));
		mocks.bulkCandidates.mockResolvedValue(
			BulkCandidatesResult.create({
				candidates: [new Gear({}).asSpec(), cloak, cloak].map((gear, index) => ({ index, gear })),
			}),
		);

		await sim.runBulkSim(() => {}, BulkSettings.create());

		expect(mocks.runConcurrentBulkSim.mock.calls[0][0].candidates).toEqual([{ index: 0, gear: cloak }]);
	});
});

describe('Sim.getMinConsumableLvl', () => {
	it('takes the default for a type the filters leave unset, and the filters where they set one, 0 included', () => {
		const sim = new Sim({ env: makeEnv() });
		expect(sim.getMinConsumableLvl(ConsumableType.ConsumableTypePotion)).toBe(Sim.DEFAULT_MIN_CONSUMABLE_LVL[ConsumableType.ConsumableTypePotion]);
		expect(sim.getMinConsumableLvl(ConsumableType.ConsumableTypeFood)).toBe(0);

		const filters = sim.getFilters();
		filters.minConsumableLvl[ConsumableType.ConsumableTypePotion] = 0;
		filters.minConsumableLvl[ConsumableType.ConsumableTypeFood] = 45;
		sim.setFilters(filters);
		expect(sim.getMinConsumableLvl(ConsumableType.ConsumableTypePotion)).toBe(0);
		expect(sim.getMinConsumableLvl(ConsumableType.ConsumableTypeFood)).toBe(45);
	});
});
