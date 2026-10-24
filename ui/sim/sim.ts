import {
	BulkCandidatesRequest,
	BulkCandidatesResult,
	BulkCombinationCountRequest,
	BulkCombinationCountResult,
	BulkEnchantSlotOptions,
	BulkGearCandidate,
	BulkSettings,
	BulkSimRequest,
	BulkSimResult,
	ComputeStatsRequest,
	type ComputeStatsResult,
	ErrorOutcome,
	ErrorOutcomeType,
	PlayerStats,
	Raid as RaidProto,
	RaidSimRequest,
	RaidSimResult,
	ReforgeOptimizeRequest,
	ReforgeOptimizeResult,
	ReforgeSettings,
	SimOptions,
	SimType,
	StatWeightsRequest,
	StatWeightsResult,
	StatWeightsTarget,
	StatWeightsTargetMetric,
} from '@generated/proto/api';
import {
	ArmorType,
	ConsumableType,
	Encounter as EncounterProto,
	EquipmentSpec,
	Faction,
	Profession,
	PseudoStat,
	RangedWeaponType,
	Stat,
	UnitReference,
	UnitReference_Type as UnitType,
	WeaponType,
} from '@generated/proto/common';
import {
	DatabaseFilters,
	RaidFilterOption,
	SimlParseRequest,
	SimlParseResult,
	SimlPrintRequest,
	SimlPrintResult,
	SimSettings as SimSettingsProto,
	SourceFilterOption,
} from '@generated/proto/ui';
import i18n from '@i18n/config';
import { getLang } from '@i18n/locale_service';
import { SimRequest } from '@worker/types';

import { hasTouch } from '../shared/pointer';
import { makeBulkGearDatabase, makeBulkItemDatabaseFromSpecs } from './bulk/gear_database';
import { dedupeGearSets, throwIfAborted } from './bulk/utils';
import { CURRENT_PHASE, LOCAL_STORAGE_PREFIX } from './constants/other';
import { Player, UnitMetadata } from './player/player';
import { addPartyBuffItems } from './proto/buff_items';
import { Database } from './proto/database';
import { Gear } from './proto/gear';
import { extendPlayerProtoWithMissingEffects } from './proto/proto_migration';
import { SimResult } from './proto/sim_result';
import { StatCap, Stats } from './proto/stats';
import { Encounter } from './raid/encounter';
import { Raid } from './raid/raid';
import { SimRuns } from './sim_runs';
import { RequestTypes, SimSignalManager, type SimSignals } from './sim_signal_manager';
import { batch } from './state/batch';
import type { Env } from './state/env';
import { Emitter } from './state/events';
import { makeReforgeConfigRequestFields } from './state/reforge_request';
import { createSimStore, patchSlice, SimSettingsSlice, UISlice } from './state/sim_store';
import { subscribeStatsInputs, subscribeUiField } from './state/subscriptions';
import { getEnumValues } from './utils/collections';
import { environmentOf } from './utils/env';
import { hashString, noop } from './utils/misc';
import { runConcurrentBulkSim, runConcurrentSim, runConcurrentStatWeights } from './wasm';
import { generateRequestId, WorkerPool, WorkerProgressCallback } from './workers/worker_pool';

export const WASM_CONCURRENCY_STORAGE_KEY = `${LOCAL_STORAGE_PREFIX}_wasmconcurrency`;
// Without the spec prefix, so every spec page shares the one setting.
export const SHOW_POWER_TOOLS_STORAGE_KEY = `${LOCAL_STORAGE_PREFIX}_showpowertools`;
export type ReforgeOptimizeConfig = {
	gear: Gear;
	preCapEPWeights: Stats;
	undershootCaps: Stats;
	settings: ReforgeSettings;
	softCaps: StatCap[];
	enchantOptions: BulkEnchantSlotOptions[];
	debug?: boolean;
};

interface SimProps {
	// The type of sim. Default `SimType.SimTypeIndividual`
	type?: SimType;
	// Browser adapter used by the domain layer (storage / location / hardware /
	// page hide). Page entries pass `browserEnv`; test harnesses an in-memory one.
	env: Env;
}

export type RunSimOptions = {
	silent?: boolean; // If true, don't emit the simResultEmitter event.
	debug?: boolean;
	singleIteration?: boolean; // If true, only run a single iteration (for testing purposes).
	iterations?: number;
	// Replaces the request player's equipment and database, rather than merging into them.
	gear?: Gear;
	onProgress?: WorkerProgressCallback;
	raw?: boolean;
};

export type StatWeightsOptions = {
	iterations?: number; // Unset uses the sim's iteration setting; a target ignores it and sizes its first pass itself.
	targetEpCi?: number; // The 90% ± of each targeted EP to reach by adding iterations; unset runs the iterations once.
	maxIterationMultiplier?: number;
	targetMetrics?: StatWeightsTargetMetric[]; // The metrics the target applies to; unset means DPS only.
	hpsReferenceStat?: Stat; // The stat HPS EP divides by; unset means the EP reference stat.
};

// Core Sim module which deals only with api types, no UI-related stuff.
export class Sim {
	private readonly workerPool: WorkerPool;

	readonly type: SimType;
	// Browser adapter for the domain layer; see state/env.ts.
	readonly env: Env;
	readonly raid: Raid;
	readonly encounter: Encounter;

	private db_: Database | null = null;

	// The Zustand store backing this sim's state (see state/sim_store.ts).
	// Declared before the facades that write it.
	readonly store = createSimStore();

	private set(patch: Partial<SimSettingsSlice>) {
		patchSlice(this.store, 'sim', patch);
	}

	readonly crashEmitter = new Emitter<SimError>();

	// Fires when a raid sim API call completes.
	readonly simResultEmitter = new Emitter<SimResult>();

	private readonly _initPromise: Promise<any>;
	isNative: boolean | undefined = undefined;

	// These callbacks are needed so we can apply BuffBot modifications automatically before sending requests.
	private modifyRaidProto: (raidProto: RaidProto) => void = noop;

	readonly signalManager: SimSignalManager;
	readonly runs: SimRuns;

	constructor({ type, env }: SimProps) {
		this.type = type ?? SimType.SimTypeIndividual;
		this.env = env;

		this.workerPool = new WorkerPool(1);
		subscribeUiField(
			this,
			'wasmConcurrency',
		)(async () => {
			// Prevent using worker concurrency when not running wasm. Local sim has native threading.
			if (await this.workerPool.isWasm()) {
				const nWorker = Math.max(1, Math.min(this.getWasmConcurrency(), this.env.hardwareConcurrency));
				this.workerPool.setNumWorkers(nWorker);
			}
		});

		let wasmConcurrencySetting = parseInt(this.env.storage.getItem(WASM_CONCURRENCY_STORAGE_KEY) ?? 'NaN');
		if (isNaN(wasmConcurrencySetting)) {
			wasmConcurrencySetting = 0;
			// Set a default worker count if env supports multiple threads. Should not be too high as to be safe for all situations.
			// TODO: Set based on browser/engine? E.g. Firefox has significant RAM and CPU usage per worker while Chrome can run many without a downside.
			if (this.env.hardwareConcurrency > 1) {
				wasmConcurrencySetting = Math.min(4, Math.floor(this.env.hardwareConcurrency / 2));
			}
		}
		this.setWasmConcurrency(wasmConcurrencySetting);
		this.setShowPowerTools(this.env.storage.getItem(SHOW_POWER_TOOLS_STORAGE_KEY) === 'true');

		this.signalManager = new SimSignalManager();
		this.runs = new SimRuns(this.store, this.signalManager);

		this._initPromise = Database.get().then(async db => {
			this.db_ = db;
			await this.resolveIsNative();
		});

		this.raid = new Raid(this);
		this.encounter = new Encounter(this);

		// Stats recompute on any raid/encounter change — one selector, so a batch
		// touching both recomputes once. Skipped while the initial settings load
		// is applying, so the load's many writes cost one recompute, which
		// applyLoadedSettings runs when it finishes.
		subscribeStatsInputs(this)(() => {
			if (this.applyingLoadedSettings) return;
			this.updateCharacterStats();
		});

		// Initial language write: initialization, not a change.
		this.store.setState(s => ({ ui: { ...s.ui, language: getLang() } }));
	}

	// True only while loadIndividualSettings is applying stored settings; see the
	// stats subscription in the constructor. Subscribers mint their own event ids,
	// so the load cannot be recognised by its EventID.
	private applyingLoadedSettings = false;

	// Runs `apply` with the initial-settings-load flag set. Anything that does not
	// go through here keeps the normal recompute-on-change behaviour.
	applyLoadedSettings<T>(apply: () => T): T {
		this.applyingLoadedSettings = true;
		try {
			return apply();
		} finally {
			this.applyingLoadedSettings = false;
			// Not redundant: every recompute was suppressed for the whole load, and currentStats is never persisted.
			this.updateCharacterStats();
		}
	}

	waitForInit(): Promise<void> {
		return this._initPromise;
	}

	private async resolveIsNative() {
		try {
			this.isNative = !(await this.isWasm());
		} catch {
			// Probe failed - fall back to the hostname heuristic (a local host runs the
			// native sim, an external one runs wasm).
			this.isNative = environmentOf(this.env.location.hostname) === 'local';
		}
	}

	/**
	 * Check if workers are running wasm.
	 * @returns true if workers are running wasm.
	 */
	isWasm() {
		return this.workerPool.isWasm();
	}

	/**
	 * Whether the current environment should use wasm/worker concurrency methods.
	 * @returns true if running wasm workers and concurrency setting is active.
	 */
	async shouldUseWasmConcurrency() {
		return (await this.isWasm()) && this.getWasmConcurrency() >= 2 && this.workerPool.getNumWorkers() >= 2;
	}

	get db(): Database {
		return this.db_!;
	}

	setModifyRaidProto(newModFn: (raidProto: RaidProto) => void) {
		this.modifyRaidProto = newModFn;
	}

	getModifiedRaidProto(): RaidProto {
		const raidProto = this.raid.toProto(false, true);
		this.modifyRaidProto(raidProto);

		raidProto.parties.forEach(party => {
			party.players.forEach(player => {
				if (!player.equipment) {
					return;
				}

				const isEnchanter = [player.profession1, player.profession2].includes(Profession.Enchanting);

				// Remove ring enchants if not an enchanter.
				if (!isEnchanter) {
					player.equipment = this.db.lookupEquipmentSpec(player.equipment).withoutEnchanting().asSpec();
				}

				extendPlayerProtoWithMissingEffects(player, this.db);
			});
		});

		return raidProto;
	}

	makeRaidSimRequest(options: RunSimOptions = {}): RaidSimRequest {
		const raid = this.getModifiedRaidProto();
		const encounter = this.encounter.toProto();

		// TODO: remove any replenishment from sim request here? probably makes more sense to do it inside the sim to protect against accidents
		return RaidSimRequest.create({
			requestId: generateRequestId(SimRequest.raidSimAsync),
			type: this.type,
			raid: raid,
			encounter: encounter,
			simOptions: SimOptions.create({
				iterations: options.singleIteration ? 1 : (options.iterations ?? this.getIterations()),
				randomSeed: BigInt(this.nextRngSeed()),
				debugFirstIteration: true,
				debug: options.debug ?? false,
			}),
		});
	}

	private requireRunnableSetup() {
		if (this.raid.isEmpty()) {
			throw new Error(i18n.t('sim.run_errors.raid_empty'));
		} else if (this.encounter.getTargets().length < 1) {
			throw new Error(i18n.t('sim.run_errors.no_targets'));
		}
	}

	// Runs `run` registered as `type`, so an abort of that type or of every type reaches its signals.
	private async runRegistered<T>(type: RequestTypes, run: (signals: SimSignals) => Promise<T>): Promise<T> {
		this.requireRunnableSetup();

		const signals = this.signalManager.registerRunning(type);
		try {
			await this.waitForInit();
			return await run(signals);
		} catch (error) {
			if (error instanceof SimError) throw error;
			console.error(error);
			throw new Error(i18n.t('sim.run_errors.sim_failed'));
		} finally {
			this.signalManager.unregisterRunning(signals);
		}
	}

	// An aborted run comes back with its error set; any other error throws.
	private async simulate(request: RaidSimRequest, onProgress: WorkerProgressCallback, signals: SimSignals): Promise<RaidSimResult> {
		// Only use worker based concurrency when running wasm (local sim has native threading),
		// and never for a lone iteration, which has nothing to split.
		const result =
			request.simOptions!.iterations > 1 && (await this.shouldUseWasmConcurrency())
				? await runConcurrentSim(request, this.workerPool, onProgress, signals)
				: await this.workerPool.raidSimAsync(request, onProgress, signals);
		if (result.error && result.error.type == ErrorOutcomeType.ErrorOutcomeError) throw new SimError(result.error.message);
		return result;
	}

	runSim(options: RunSimOptions & { raw: true }): Promise<[RaidSimRequest, RaidSimResult] | ErrorOutcome>;
	runSim(options?: RunSimOptions & { raw?: false }): Promise<SimResult | ErrorOutcome>;
	runSim(options: RunSimOptions = {}): Promise<SimResult | ErrorOutcome | [RaidSimRequest, RaidSimResult]> {
		return this.runRegistered(RequestTypes.IndividualSim, async (signals): Promise<SimResult | ErrorOutcome | [RaidSimRequest, RaidSimResult]> => {
			const request = this.makeRaidSimRequest(options);
			const player = request.raid!.parties[0].players[0];

			if (options.gear) {
				const gear = Sim.prepareGear(options.gear, [player.profession1, player.profession2].includes(Profession.Enchanting));
				player.database = gear.toDatabase(this.db);
				addPartyBuffItems(player.database, this.db, request.raid!.parties[0].buffs);
				player.equipment = gear.asSpec();
				if (player.consumables) player.consumables = gear.adjustImbues(player.consumables);
			}

			const result = await this.simulate(request, options.onProgress ?? noop, signals);
			if (result.error) return result.error;

			if (options.raw) return [request, result];

			const simResult = await SimResult.makeNew(request, result);
			if (!options.silent) {
				this.simResultEmitter.emit(simResult);
			}
			return simResult;
		});
	}

	// Runs a request built by makeRaidSimRequest as `type`, emitting no result. An aborted run comes back with its error set.
	runRaidSimRequest(request: RaidSimRequest, type: RequestTypes, onProgress: WorkerProgressCallback): Promise<RaidSimResult> {
		return this.runRegistered(type, signals => this.simulate(request, onProgress, signals));
	}

	// Normalizes gear for a sim/bulk/reforge request
	private static prepareGear(gear: Gear, isEnchanter: boolean): Gear {
		// Remove ring enchants if not an enchanter.
		if (!isEnchanter) {
			gear = gear.withoutEnchanting();
		}
		return gear;
	}

	async runBulkSim(
		onProgress: WorkerProgressCallback,
		bulkSettings: BulkSettings,
		abortSignal?: AbortSignal,
		encounterOverride?: EncounterProto,
	): Promise<BulkSimResult | ErrorOutcome> {
		this.requireRunnableSetup();

		const signals = this.signalManager.registerRunning(RequestTypes.BulkSim);
		try {
			await this.waitForInit();

			const requestId = generateRequestId(SimRequest.bulkSimAsync);
			const baseRequest = this.makeRaidSimRequest();
			if (encounterOverride) {
				baseRequest.encounter = EncounterProto.clone(encounterOverride);
			}
			baseRequest.requestId = requestId;
			baseRequest.simOptions!.debugFirstIteration = false;
			baseRequest.simOptions!.debug = false;

			const player = baseRequest.raid!.parties[0].players[0];
			const isEnchanter = [player.profession1, player.profession2].includes(Profession.Enchanting);

			const baselineGear = Sim.prepareGear(this.raid.getActivePlayers()[0].getGear(), isEnchanter);
			if (!this.getFixedRngSeed()) {
				// Derive the seed from the run's content instead of Math.random(): the same
				// setup then reproduces bit-identical results (the whole pipeline is
				// deterministic given a seed), while any change to the setup draws a fresh
				// sample. An explicit fixed RNG seed still takes precedence above. The two
				// parts are hashed individually (they are already JSON) and the digests
				// combined, avoiding a second full serialization pass.
				const contentHash = hashString(
					hashString(EquipmentSpec.toJsonString(baselineGear.asSpec())) + hashString(BulkSettings.toJsonString(bulkSettings)),
				);
				const contentSeed = Number(BigInt('0x' + contentHash.slice(0, 8)));
				baseRequest.simOptions!.randomSeed = BigInt(contentSeed);
				// makeRaidSimRequest already drew and recorded a random seed; overwrite the
				// record too so getLastUsedRngSeed() reflects the seed actually used.
				this.setLastUsedRngSeed(contentSeed);
			}
			const useWasmBulkSim = await this.isWasm();
			// The native sim generates the candidates itself when the request carries none.
			let candidates: BulkGearCandidate[] = [];
			if (useWasmBulkSim) {
				const bulkCandidatesResult = await this.getBulkCandidates(bulkSettings);
				if (bulkCandidatesResult.error) {
					throw new SimError(bulkCandidatesResult.error.message || i18n.t('sim.run_errors.bulk_candidates_failed'));
				}
				const gearSets = bulkCandidatesResult.candidates.flatMap(candidate => (candidate.gear ? [candidate.gear] : []));
				candidates = dedupeGearSets(gearSets, [baselineGear.asSpec()]).map((gear, index) => ({ index, gear }));
			}
			throwIfAborted(abortSignal);
			const bulkGearDatabase = makeBulkItemDatabaseFromSpecs(this.db, baselineGear, bulkSettings);
			player.database = player.database ? Database.mergeSimDatabases(player.database, bulkGearDatabase) : bulkGearDatabase;
			player.equipment = baselineGear.asSpec();
			baseRequest.raid!.parties[0].players[0] = player;
			throwIfAborted(abortSignal);
			const bulkRequest = BulkSimRequest.create({
				requestId,
				baseRequest,
				candidates,
				topResults: 5,
				bulkSettings,
			});
			// Only use worker based concurrency when running wasm. Local sim has native threading.
			const result = useWasmBulkSim
				? await runConcurrentBulkSim(bulkRequest, this.workerPool, onProgress, signals)
				: await this.workerPool.bulkSimAsync(bulkRequest, onProgress, signals);
			if (result.error) {
				if (result.error.type != ErrorOutcomeType.ErrorOutcomeError) return result.error;
				throw new SimError(result.error.message);
			}

			return result;
		} catch (error) {
			if (error instanceof SimError) throw error;
			console.error(error);
			throw new Error(i18n.t('sim.run_errors.bulk_sim_failed'));
		} finally {
			this.signalManager.unregisterRunning(signals);
		}
	}

	private makeBulkBaseRequest(bulkSettings: BulkSettings): RaidSimRequest {
		const baseRequest = this.makeRaidSimRequest();
		const player = baseRequest.raid!.parties[0].players[0];
		const isEnchanter = [player.profession1, player.profession2].includes(Profession.Enchanting);
		const baselineGear = Sim.prepareGear(this.raid.getActivePlayers()[0].getGear(), isEnchanter);
		const bulkGearDatabase = makeBulkItemDatabaseFromSpecs(this.db, baselineGear, bulkSettings);
		player.database = player.database ? Database.mergeSimDatabases(player.database, bulkGearDatabase) : bulkGearDatabase;
		player.equipment = baselineGear.asSpec();
		baseRequest.raid!.parties[0].players[0] = player;
		return baseRequest;
	}

	async getBulkCombinationCount(bulkSettings: BulkSettings): Promise<BulkCombinationCountResult> {
		this.requireRunnableSetup();

		await this.waitForInit();
		const baseRequest = this.makeBulkBaseRequest(bulkSettings);
		const request = BulkCombinationCountRequest.create({
			baseRequest,
			bulkSettings,
		});
		return await this.workerPool.bulkCombinationCount(request);
	}

	async getBulkCandidates(bulkSettings: BulkSettings): Promise<BulkCandidatesResult> {
		this.requireRunnableSetup();

		await this.waitForInit();
		const baseRequest = this.makeBulkBaseRequest(bulkSettings);
		const request = BulkCandidatesRequest.create({
			baseRequest,
			bulkSettings,
		});
		return await this.workerPool.bulkCandidates(request);
	}

	// SIML (sim/siml/SPEC.md), read and written by the sim.
	simlParse(request: SimlParseRequest): Promise<SimlParseResult> {
		return this.workerPool.simlParse(request);
	}

	simlPrint(request: SimlPrintRequest): Promise<SimlPrintResult> {
		return this.workerPool.simlPrint(request);
	}

	// Stats for a hypothetical gear set, without touching the player's own current stats or
	// metadata. Used by the gear optimizer.
	async getCharacterStatsForGear(gear: Gear): Promise<PlayerStats> {
		await this.waitForInit();

		const raidProto = this.getModifiedRaidProto();
		const player = raidProto.parties[0].players[0];
		player.database = gear.toDatabase(this.db);
		addPartyBuffItems(player.database, this.db, raidProto.parties[0].buffs);
		player.equipment = gear.asSpec();
		extendPlayerProtoWithMissingEffects(player, this.db);

		const result = await this.computeRaidStats(raidProto);
		if (result.errorResult != '') {
			this.crashEmitter.emit(new SimError(result.errorResult));
		}

		return result.raidStats!.parties[0].players[0];
	}

	// The stats of `raidProto`, a raid from getModifiedRaidProto() the caller has edited, without touching the players' own
	// current stats or metadata. An error comes back in `errorResult` for the caller to raise.
	async computeRaidStats(raidProto: RaidProto): Promise<ComputeStatsResult> {
		await this.waitForInit();
		return this.workerPool.computeStats(
			ComputeStatsRequest.create({
				raid: raidProto,
				encounter: this.encounter.toProto(),
			}),
		);
	}

	// This should be invoked internally whenever stats might have changed.
	private characterStatsVersion = 0;

	async updateCharacterStats() {
		const version = ++this.characterStatsVersion;
		await this.waitForInit();
		// Capture the current players so we avoid issues if something changes while
		// request is in-flight.

		const players = this.raid.getPlayers();
		const req = ComputeStatsRequest.create({
			raid: this.getModifiedRaidProto(),
			encounter: this.encounter.toProto(),
		});
		const result = await this.workerPool.computeStats(req);
		// Two recomputes can be in flight at once and they settle in any order, so a reply that has
		// already been superseded must not write: it would overwrite newer stats with older ones.
		if (version !== this.characterStatsVersion) return;
		if (result.errorResult != '') {
			this.crashEmitter.emit(new SimError(result.errorResult));
			return;
		}

		batch(async () => {
			const playerUpdatePromises = result
				.raidStats!.parties.map((partyStats, partyIndex) =>
					partyStats.players.map((playerStats, playerIndex) => {
						const player = players[partyIndex * 5 + playerIndex];
						if (player) {
							player.setCurrentStats(playerStats);
							return player.updateMetadata();
						} else {
							return null;
						}
					}),
				)
				.flat()
				.filter(p => p != null) as Array<Promise<boolean>>;

			const targetUpdatePromise = this.encounter.targetsMetadata.update(result.encounterStats!.targets.map(t => t.metadata!));

			const anyUpdates = await Promise.all(playerUpdatePromises.concat([targetUpdatePromise]));
			if (anyUpdates.some(v => v)) {
				this.store.setState(s => ({ sim: { ...s.sim, metadataVersion: s.sim.metadataVersion + 1 } }));
			}
		});
	}

	// Runs the gear optimizer for Player 0 without triggering any metadata updates or touching the UI.
	async reforgeOptimize(config: ReforgeOptimizeConfig): Promise<ReforgeOptimizeResult> {
		const signals = this.signalManager.registerRunning(RequestTypes.ReforgeOptimize);
		try {
			await this.waitForInit();
			const raid = this.getModifiedRaidProto();
			const player = raid.parties[0].players[0];
			// The WASM sim has no enchant table of its own, so the enchants the optimizer may pick ship with the gear.
			player.database = makeBulkGearDatabase(
				this.db,
				[config.gear],
				[],
				config.enchantOptions.flatMap(options => options.effectIds),
			);
			addPartyBuffItems(player.database, this.db, raid.parties[0].buffs);
			player.equipment = config.gear.asSpec();
			raid.parties[0].players[0] = player;

			const request = ReforgeOptimizeRequest.create({
				requestId: generateRequestId(SimRequest.reforgeOptimizeAsync),
				raid,
				...makeReforgeConfigRequestFields(config),
				debug: config.debug ?? false,
			});
			const result = await this.workerPool.reforgeOptimizeAsync(request, signals);
			if (result.error) {
				throw new SimError(result.error.message);
			}
			return result;
		} finally {
			this.signalManager.unregisterRunning(signals);
		}
	}

	async statWeights(
		player: Player<any>,
		epStats: Array<Stat>,
		epPseudoStats: Array<PseudoStat>,
		epReferenceStat: Stat,
		onProgress: WorkerProgressCallback,
		options: StatWeightsOptions = {},
	): Promise<StatWeightsResult> {
		this.requireRunnableSetup();

		await this.waitForInit();

		if (player.getParty() == null) {
			console.warn('Trying to get stat weights without a party!');
			return StatWeightsResult.create();
		} else {
			const tanks = this.raid
				.getTanks()
				.map(tank => tank.index)
				.includes(player.getRaidIndex())
				? [UnitReference.create({ type: UnitType.Player, index: 0 })]
				: [];
			// `toDatabase()` fills items and enchants only, so without this the request carries
			// no consumables or spell effects. lib.wasm is built without the `with_db` tag, so a
			// worker that has not already handled a full sim request cannot resolve them and computes
			// the stat's sub-sims with no flask, elixir, food or potion.
			const playerProto = player.toProto(false, true);
			extendPlayerProtoWithMissingEffects(playerProto, this.db);

			const request = StatWeightsRequest.create({
				player: playerProto,
				raidBuffs: this.raid.getBuffs(),
				partyBuffs: player.getParty()!.getBuffs(),
				debuffs: this.raid.getDebuffs(),
				encounter: this.encounter.toProto(),
				simOptions: SimOptions.create({
					iterations: options.iterations ?? this.getIterations(),
					randomSeed: BigInt(this.nextRngSeed()),
					debug: false,
				}),
				tanks: tanks,

				statsToWeigh: epStats,
				pseudoStatsToWeigh: epPseudoStats,
				epReferenceStat: epReferenceStat,
				hpsReferenceStat: options.hpsReferenceStat ?? epReferenceStat,
				target:
					options.targetEpCi === undefined
						? undefined
						: StatWeightsTarget.create({
								epCi: options.targetEpCi,
								maxIterationMultiplier: options.maxIterationMultiplier,
								metrics: options.targetMetrics,
							}),
			});

			const signals = this.signalManager.registerRunning(RequestTypes.StatWeights);
			try {
				let result: StatWeightsResult;
				// Only use worker based concurrency when running wasm.
				if (await this.shouldUseWasmConcurrency()) {
					result = await runConcurrentStatWeights(request, this.workerPool, onProgress, signals);
				} else {
					result = await this.workerPool.statWeightsAsync(request, onProgress, signals);
				}
				if (result.error) {
					if (result.error.type != ErrorOutcomeType.ErrorOutcomeError) return result;
					throw new SimError(result.error.message);
				}
				return result;
			} catch (error) {
				if (error instanceof SimError) throw error;
				console.error(error);
				throw new Error(i18n.t('sim.run_errors.stat_weights_failed'));
			} finally {
				this.signalManager.unregisterRunning(signals);
			}
		}
	}

	getUnitMetadata(ref: UnitReference | undefined, contextPlayer: Player<any> | null, defaultRef: UnitReference): UnitMetadata | undefined {
		if (!ref || ref.type == UnitType.Unknown) {
			return this.getUnitMetadata(defaultRef, contextPlayer, defaultRef);
		} else if (ref.type == UnitType.Player) {
			return this.raid.getPlayerFromUnitReference(ref)?.getMetadata();
		} else if (ref.type == UnitType.Target) {
			return this.encounter.targetsMetadata.asList()[ref.index];
		} else if (ref.type == UnitType.Pet) {
			const owner = this.raid.getPlayerFromUnitReference(ref.owner, contextPlayer);
			if (owner) {
				return owner.getPetMetadatas().asList()[ref.index];
			} else {
				return undefined;
			}
		} else if (ref.type == UnitType.Self) {
			return contextPlayer?.getMetadata();
		} else if (ref.type == UnitType.CurrentTarget) {
			return this.encounter.targetsMetadata.asList()[0];
		}
		return undefined;
	}

	getPhase(): number {
		return this.store.getState().sim.phase;
	}
	setPhase(newPhase: number) {
		if (newPhase != this.getPhase()) {
			this.set({ phase: newPhase });
		}
	}

	getFaction(): Faction {
		return this.store.getState().sim.faction;
	}
	setFaction(newFaction: Faction) {
		if (newFaction != this.getFaction() && !!newFaction) {
			this.set({ faction: newFaction });
		}
	}

	getFixedRngSeed(): number {
		return this.store.getState().sim.fixedRngSeed;
	}
	setFixedRngSeed(newFixedRngSeed: number) {
		if (newFixedRngSeed != this.getFixedRngSeed()) {
			this.set({ fixedRngSeed: newFixedRngSeed });
		}
	}

	static MAX_RNG_SEED = Math.pow(2, 32) - 1;
	// Never 0: the sim replaces a 0 seed with the clock, so a seed shared by several requests would not hold.
	private nextRngSeed(): number {
		let rngSeed = 0;
		if (this.getFixedRngSeed()) {
			rngSeed = this.getFixedRngSeed();
		} else {
			rngSeed = Math.floor(Math.random() * Sim.MAX_RNG_SEED) + 1;
		}

		this.setLastUsedRngSeed(rngSeed);
		return rngSeed;
	}
	// Counts as a change on every call, so
	// subscribers watch the version counter, not the value.
	private setLastUsedRngSeed(rngSeed: number) {
		this.store.setState(s => ({ sim: { ...s.sim, lastUsedRngSeed: rngSeed, lastUsedRngSeedVersion: s.sim.lastUsedRngSeedVersion + 1 } }));
	}
	getLastUsedRngSeed(): number {
		return this.store.getState().sim.lastUsedRngSeed;
	}

	getFilters(): DatabaseFilters {
		// Make a defensive copy
		return DatabaseFilters.clone(this.store.getState().sim.filters);
	}
	// The consumable pickers' minimum level by ConsumableType, where the filters set none.
	static readonly DEFAULT_MIN_CONSUMABLE_LVL: Partial<Record<ConsumableType, number>> = {
		[ConsumableType.ConsumableTypePotion]: 40,
		[ConsumableType.ConsumableTypeFlask]: 0,
		[ConsumableType.ConsumableTypeElixir]: 40,
		[ConsumableType.ConsumableTypeScroll]: 40,
		[ConsumableType.ConsumableTypeExplosive]: 30,
		[ConsumableType.ConsumableTypeImbue]: 40,
		[ConsumableType.ConsumableTypeWeaponImbue]: 40,
	};
	getMinConsumableLvl(type: ConsumableType): number {
		return this.store.getState().sim.filters.minConsumableLvl[type] ?? Sim.DEFAULT_MIN_CONSUMABLE_LVL[type] ?? 0;
	}
	setFilters(newFilters: DatabaseFilters) {
		if (DatabaseFilters.equals(newFilters, this.store.getState().sim.filters)) {
			return;
		}

		// Make a defensive copy; replace-on-write, so subscribers can use
		// reference equality.
		const filters = DatabaseFilters.clone(newFilters);
		this.set({ filters });
	}

	private get ui(): UISlice {
		return this.store.getState().ui;
	}
	private setUi(patch: Partial<UISlice>) {
		patchSlice(this.store, 'ui', patch);
	}

	getShowDamageMetrics(): boolean {
		return this.ui.showDamageMetrics;
	}
	setShowDamageMetrics(newShowDamageMetrics: boolean) {
		if (newShowDamageMetrics != this.ui.showDamageMetrics) this.setUi({ showDamageMetrics: newShowDamageMetrics });
	}

	getShowThreatMetrics(): boolean {
		return this.ui.showThreatMetrics;
	}
	setShowThreatMetrics(newShowThreatMetrics: boolean) {
		if (newShowThreatMetrics != this.ui.showThreatMetrics) this.setUi({ showThreatMetrics: newShowThreatMetrics });
	}

	// A plain field read with no "threat metrics on, and the player is a tank" term: `toProto`
	// serializes whatever this returns, so a derived form would write showHealingMetrics: true into
	// every tank spec's saved settings.
	getShowHealingMetrics(): boolean {
		return this.ui.showHealingMetrics;
	}
	setShowHealingMetrics(newShowHealingMetrics: boolean) {
		if (newShowHealingMetrics != this.ui.showHealingMetrics) this.setUi({ showHealingMetrics: newShowHealingMetrics });
	}

	getShowExperimental(): boolean {
		return this.ui.showExperimental;
	}
	setShowExperimental(newShowExperimental: boolean) {
		if (newShowExperimental != this.ui.showExperimental) this.setUi({ showExperimental: newShowExperimental });
	}

	getShowPowerTools(): boolean {
		return this.ui.showPowerTools;
	}
	setShowPowerTools(newShowPowerTools: boolean) {
		if (newShowPowerTools != this.ui.showPowerTools) {
			this.env.storage.setItem(SHOW_POWER_TOOLS_STORAGE_KEY, newShowPowerTools.toString());
			this.setUi({ showPowerTools: newShowPowerTools });
		}
	}

	getWasmConcurrency(): number {
		return this.ui.wasmConcurrency;
	}
	setWasmConcurrency(newWasmConcurrency: number) {
		if (newWasmConcurrency != this.ui.wasmConcurrency) {
			this.env.storage.setItem(WASM_CONCURRENCY_STORAGE_KEY, newWasmConcurrency.toString());
			this.setUi({ wasmConcurrency: newWasmConcurrency });
		}
	}

	getShowQuickSwap(): boolean {
		return !hasTouch() && this.ui.showQuickSwap;
	}
	setShowQuickSwap(newShowQuickSwap: boolean) {
		if (newShowQuickSwap != this.ui.showQuickSwap) this.setUi({ showQuickSwap: newShowQuickSwap });
	}

	getShowEPValues(): boolean {
		return this.ui.showEPValues;
	}
	setShowEPValues(newShowEPValues: boolean) {
		if (newShowEPValues != this.ui.showEPValues) this.setUi({ showEPValues: newShowEPValues });
	}

	getLanguage(): string {
		return this.ui.language;
	}
	setLanguage(newLanguage: string) {
		newLanguage = newLanguage || 'en';
		if (newLanguage != this.ui.language) this.setUi({ language: newLanguage });
	}

	getIterations(): number {
		return this.store.getState().sim.iterations;
	}
	setIterations(newIterations: number) {
		if (newIterations != this.getIterations()) {
			this.set({ iterations: newIterations });
		}
	}

	static readonly ALL_ARMOR_TYPES = (getEnumValues(ArmorType) as Array<ArmorType>).filter(v => v != 0);
	static readonly ALL_WEAPON_TYPES = (getEnumValues(WeaponType) as Array<WeaponType>).filter(v => v != 0);
	static readonly ALL_RANGED_WEAPON_TYPES = (getEnumValues(RangedWeaponType) as Array<RangedWeaponType>).filter(v => v != 0);
	static readonly ALL_SOURCES = (getEnumValues(SourceFilterOption) as Array<SourceFilterOption>).filter(v => v != 0);
	static readonly ALL_RAIDS = (getEnumValues(RaidFilterOption) as Array<RaidFilterOption>).filter(v => v != 0);
	static readonly DEFAULT_ITERATIONS = 12500;

	toProto(): SimSettingsProto {
		const filters = this.getFilters();
		if (filters.armorTypes.length == Sim.ALL_ARMOR_TYPES.length) {
			filters.armorTypes = [];
		}
		if (filters.weaponTypes.length == Sim.ALL_WEAPON_TYPES.length) {
			filters.weaponTypes = [];
		}
		if (filters.rangedWeaponTypes.length == Sim.ALL_RANGED_WEAPON_TYPES.length) {
			filters.rangedWeaponTypes = [];
		}
		if (filters.sources.length == Sim.ALL_SOURCES.length) {
			filters.sources = [];
		}
		if (filters.raids.length == Sim.ALL_RAIDS.length) {
			filters.raids = [];
		}

		return SimSettingsProto.create({
			iterations: this.getIterations(),
			phase: this.getPhase(),
			fixedRngSeed: BigInt(this.getFixedRngSeed()),
			showDamageMetrics: this.getShowDamageMetrics(),
			showThreatMetrics: this.getShowThreatMetrics(),
			showHealingMetrics: this.getShowHealingMetrics(),
			showExperimental: this.getShowExperimental(),
			showQuickSwap: this.getShowQuickSwap(),
			showEpValues: this.getShowEPValues(),
			language: this.getLanguage(),
			faction: this.getFaction(),
			filters: filters,
		});
	}

	fromProto(proto: SimSettingsProto) {
		batch(() => {
			this.setIterations(proto.iterations || 12500);
			this.setPhase(proto.phase || CURRENT_PHASE);
			this.setFixedRngSeed(Number(proto.fixedRngSeed));
			this.setShowDamageMetrics(proto.showDamageMetrics);
			this.setShowThreatMetrics(proto.showThreatMetrics);
			this.setShowHealingMetrics(proto.showHealingMetrics);
			this.setShowExperimental(proto.showExperimental);
			this.setShowQuickSwap(proto.showQuickSwap);
			this.setShowEPValues(proto.showEpValues);
			this.setLanguage(proto.language);
			this.setFaction(proto.faction || Faction.Alliance);

			const filters = proto.filters || this.defaultFilters();
			if (filters.armorTypes.length == 0) {
				if (this.type == SimType.SimTypeIndividual) {
					// There is no armor specialization, so a class can wear every armor type it
					// is proficient in -- a druid leather and cloth, a shaman mail down to cloth.
					filters.armorTypes = this.raid.getActivePlayers()[0].getPlayerClass().armorTypes.slice();
				} else {
					filters.armorTypes = Sim.ALL_ARMOR_TYPES.slice();
				}
			}
			if (filters.weaponTypes.length == 0) {
				filters.weaponTypes = Sim.ALL_WEAPON_TYPES.slice();
			}
			if (filters.rangedWeaponTypes.length == 0) {
				filters.rangedWeaponTypes = Sim.ALL_RANGED_WEAPON_TYPES.slice();
			}
			if (filters.sources.length == 0) {
				filters.sources = Sim.ALL_SOURCES.slice();
			}
			// A binary link keeps raid values the enum reserves.
			filters.raids = filters.raids.filter(raid => Sim.ALL_RAIDS.some(known => known === raid));
			if (filters.raids.length == 0) {
				filters.raids = Sim.ALL_RAIDS.slice();
			}
			this.setFilters(filters);
		});
	}

	applyDefaults(isTankSim: boolean, isHealingSim: boolean) {
		this.fromProto(
			SimSettingsProto.create({
				iterations: Sim.DEFAULT_ITERATIONS,
				phase: CURRENT_PHASE,
				faction: Faction.Alliance,
				showDamageMetrics: !isHealingSim,
				showThreatMetrics: isTankSim,
				showHealingMetrics: isHealingSim,
				showQuickSwap: true,
				language: this.getLanguage(), // Don't change language.
				filters: this.defaultFilters(),
				showEpValues: false,
				useSoftCapBreakpoints: true,
			}),
		);
	}

	defaultFilters(): DatabaseFilters {
		const { favoriteItems = [], favoriteRandomSuffixes = [], favoriteReforges = [], favoriteEnchants = [], statFilters } = this.getFilters();
		return DatabaseFilters.create({
			oneHandedWeapons: true,
			twoHandedWeapons: true,
			favoriteItems,
			favoriteEnchants,
			favoriteRandomSuffixes,
			favoriteReforges,
			statFilters,
		});
	}
}

export class SimError extends Error {
	readonly errorStr: string;

	constructor(errorStr: string) {
		super(errorStr);
		this.errorStr = errorStr;
	}
}
