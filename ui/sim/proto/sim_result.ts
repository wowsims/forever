import {
	ActionMetrics as ActionMetricsProto,
	AuraMetrics as AuraMetricsProto,
	DesignAdd,
	DistributionMetrics as DistributionMetricsProto,
	EncounterMetrics as EncounterMetricsProto,
	HitKind,
	Party as PartyProto,
	PartyMetrics as PartyMetricsProto,
	Player as PlayerProto,
	Raid as RaidProto,
	RaidMetrics as RaidMetricsProto,
	RaidSimRequest,
	RaidSimResult,
	ResourceMetrics as ResourceMetricsProto,
	TargetedActionMetrics as TargetedActionMetricsProto,
	UnitMetrics as UnitMetricsProto,
} from '@generated/proto/api';
import { Class, Encounter as EncounterProto, EncounterDesign, SpellSchool, Target as TargetProto } from '@generated/proto/common';
import { ResourceType } from '@generated/proto/spell';
import { SimRun } from '@generated/proto/ui';
import i18n from '@i18n/config';

import { CacheHandler } from '../cache/cache_handler';
import { PlayerClasses } from '../player/classes';
import { PlayerSpec } from '../player/player_spec';
import { PlayerSpecs } from '../player/specs';
import { bucket } from '../utils/collections';
import { percentOf, sum } from '../utils/math';
import { ActionId, defaultTargetIcon } from './action_id';
import {
	AuraUptimeLog,
	buildAuraUptimes,
	buildCastLogs,
	buildDpsLogs,
	buildResourceGroups,
	buildThreatGroups,
	CastLog,
	CombatLog,
	DamageLog,
	DpsLog,
	Entity,
	isDamage,
	isMajorCooldown,
	MajorCooldownLog,
	parseAll,
	populateActiveAuras,
	ResourceGroupLog,
	ThreatLogGroup,
} from './combat_log';
import { getPlayerSpecFromPlayer } from './utils';

// One pass over the raid's logs, bucketed by the unit that produced them. This replaces a
// full-array filter per player, per pet and per target: on a 25-man raid that was ~60 scans
// of the same half-million-entry array, and every pet added two more scans of its owner's
// slice on top.
class UnitLogIndex {
	// Players and targets are keyed by index alone, which is safe because raid index and
	// target index are unique per unit - Entity.equals() also compares name, but only
	// because it cannot assume that. Pets share their owner's raid index, so they are the
	// one case that needs the name, and the one case that allocates a key.
	private readonly players = new Map<number, Array<CombatLog>>();
	private readonly targets = new Map<number, Array<CombatLog>>();
	private readonly pets = new Map<number, Map<string, Array<CombatLog>>>();

	constructor(logs: Array<CombatLog>) {
		for (const log of logs) {
			const source = log.source;
			if (!source) continue;
			if (source.isPet) {
				let byName = this.pets.get(source.index);
				if (!byName) {
					byName = new Map();
					this.pets.set(source.index, byName);
				}
				UnitLogIndex.push(byName, source.name, log);
			} else {
				UnitLogIndex.push(source.isTarget ? this.targets : this.players, source.index, log);
			}
		}
	}

	private static push<K>(bucket: Map<K, Array<CombatLog>>, key: K, log: CombatLog) {
		const existing = bucket.get(key);
		if (existing) {
			existing.push(log);
		} else {
			bucket.set(key, [log]);
		}
	}

	player(raidIndex: number): Array<CombatLog> {
		return this.players.get(raidIndex) ?? [];
	}

	pet(raidIndex: number, name: string): Array<CombatLog> {
		return this.pets.get(raidIndex)?.get(name) ?? [];
	}

	target(index: number): Array<CombatLog> {
		return this.targets.get(index) ?? [];
	}
}

const simResultsCache = new CacheHandler<SimResult>({
	keysToKeep: 2,
});

export interface SimResultFilter {
	// Raid index of the player to display, or null for all players.
	player?: number | null;

	// Target index of the target to display, or null for all targets.
	target?: number | null;
}

class SimResultData {
	readonly request: RaidSimRequest;
	readonly result: RaidSimResult;

	constructor(request: RaidSimRequest, result: RaidSimResult) {
		this.request = request;
		this.result = result;
	}

	get iterations() {
		return this.request.simOptions?.iterations || 1;
	}

	get duration() {
		return this.result.avgIterationDuration || 1;
	}

	get firstIterationDuration() {
		return this.result.firstIterationDuration || 1;
	}
}

// Holds all the data from a simulation call, and provides helper functions
// for parsing it.
export class SimResult {
	readonly id: string;
	readonly request: RaidSimRequest;
	readonly result: RaidSimResult;

	readonly raidMetrics: RaidMetrics;
	readonly encounterMetrics: EncounterMetrics;
	readonly logs: Array<CombatLog>;

	private players: Array<UnitMetrics>;
	private units: Array<UnitMetrics>;

	private constructor(request: RaidSimRequest, result: RaidSimResult, raidMetrics: RaidMetrics, encounterMetrics: EncounterMetrics, logs: Array<CombatLog>) {
		this.id = request.requestId;
		this.request = request;
		this.result = result;
		this.raidMetrics = raidMetrics;
		this.encounterMetrics = encounterMetrics;
		this.logs = logs;

		this.players = raidMetrics.parties.map(party => party.players).flat();
		this.units = this.players.concat(encounterMetrics.targets);
	}

	get iterations() {
		return this.request.simOptions?.iterations || 1;
	}

	get duration() {
		return this.result.avgIterationDuration || 1;
	}

	get firstIterationDuration() {
		return this.result.firstIterationDuration || 1;
	}

	getPlayers(filter?: SimResultFilter): Array<UnitMetrics> {
		if (filter?.player || filter?.player === 0) {
			const player = this.getUnitWithIndex(filter.player);
			return player ? [player] : [];
		} else {
			return this.raidMetrics.parties.map(party => party.players).flat();
		}
	}

	getRaidIndexedPlayers(filter?: SimResultFilter): Array<UnitMetrics> {
		if (filter?.player || filter?.player === 0) {
			const player = this.getPlayerWithRaidIndex(filter.player);
			return player ? [player] : [];
		} else {
			return this.raidMetrics.parties.map(party => party.players).flat();
		}
	}

	// Returns the first player, regardless of which party / raid slot its in.
	getFirstPlayer(): UnitMetrics | null {
		return this.getPlayers()[0] || null;
	}

	getPlayerWithIndex(unitIndex: number): UnitMetrics | null {
		return this.players.find(player => player.unitIndex == unitIndex) || null;
	}
	getPlayerWithRaidIndex(raidIndex: number): UnitMetrics | null {
		return this.players.find(player => player.index == raidIndex) || null;
	}

	getTargets(filter?: SimResultFilter): Array<UnitMetrics> {
		if (filter?.target || filter?.target === 0) {
			const target = this.getUnitWithIndex(filter.target);
			return target ? [target] : [];
		} else {
			return this.encounterMetrics.targets.slice();
		}
	}

	getTargetWithEncounterIndex(index: number): UnitMetrics | null {
		return this.getTargets().find(target => target.index == index) || null;
	}
	getUnitWithIndex(unitIndex: number): UnitMetrics | null {
		return this.units.find(unit => unit.unitIndex == unitIndex) || null;
	}

	getDamageMetrics(filter: SimResultFilter): DistributionMetricsProto {
		if (filter.player || filter.player === 0) {
			return this.getPlayerWithIndex(filter.player)?.dps || DistributionMetricsProto.create();
		}

		return this.raidMetrics.dps;
	}

	// The damage a unit's own side deals it, such as a sapper's hit on its thrower or a hit on a pet: every
	// action of that side, pets included, cut down to its damage on the unit. It counts as damage taken.
	getOwnSideDamageOn(unit: UnitMetrics): Array<ActionMetrics> {
		const side = unit.isTarget ? this.encounterMetrics.targets : this.players.flatMap(player => [player, ...player.pets]);
		return side.flatMap(source => source.getDamageActionsOn(unit.unitIndex));
	}

	getRaidIndexedActionMetrics(filter?: SimResultFilter): Array<ActionMetrics> {
		return ActionMetrics.joinById(
			this.getRaidIndexedPlayers(filter)
				.map(player => player.getPlayerAndPetActions().map(action => action.forTarget(filter)))
				.flat(),
		);
	}

	getDebuffMetrics(filter?: SimResultFilter): Array<AuraMetrics> {
		return AuraMetrics.joinById(
			this.getTargets(filter)
				.map(target => target.auras)
				.flat(),
		).filter(aura => aura.uptimePercent != 0);
	}

	toProto(): SimRun {
		return SimRun.create({
			request: this.request,
			result: this.result,
		});
	}

	static async fromProto(proto: SimRun): Promise<SimResult> {
		return SimResult.makeNew(proto.request || RaidSimRequest.create(), proto.result || RaidSimResult.create());
	}

	static async makeNew(request: RaidSimRequest, result: RaidSimResult): Promise<SimResult> {
		const id = request.requestId;

		const cachedResult = simResultsCache.get(id);
		if (cachedResult) return cachedResult;

		const resultData = new SimResultData(request, result);
		const logs = await parseAll(result);
		const logIndex = new UnitLogIndex(logs);
		const raidPromise = RaidMetrics.makeNew(resultData, request.raid!, result.raidMetrics!, logIndex);
		const encounterPromise = EncounterMetrics.makeNew(resultData, request.encounter!, result.encounterMetrics!, logIndex);

		const raidMetrics = await raidPromise;
		const encounterMetrics = await encounterPromise;

		const simResult = new SimResult(request, result, raidMetrics, encounterMetrics, logs);
		simResultsCache.set(id, simResult);

		return simResult;
	}
}

export class RaidMetrics {
	private readonly metrics: RaidMetricsProto;

	readonly dps: DistributionMetricsProto;
	readonly hps: DistributionMetricsProto;
	readonly parties: Array<PartyMetrics>;

	private constructor(metrics: RaidMetricsProto, parties: Array<PartyMetrics>) {
		this.metrics = metrics;
		this.dps = this.metrics.dps!;
		this.hps = this.metrics.hps!;
		this.parties = parties;
	}

	static async makeNew(resultData: SimResultData, raid: RaidProto, metrics: RaidMetricsProto, logIndex: UnitLogIndex): Promise<RaidMetrics> {
		const numParties = Math.min(raid.parties.length, metrics.parties.length);

		const parties = await Promise.all(
			[...new Array(numParties).keys()].map(i => PartyMetrics.makeNew(resultData, raid.parties[i], metrics.parties[i], i, logIndex)),
		);

		return new RaidMetrics(metrics, parties);
	}
}

export class PartyMetrics {
	readonly players: Array<UnitMetrics>;

	private constructor(players: Array<UnitMetrics>) {
		this.players = players;
	}

	static async makeNew(
		resultData: SimResultData,
		party: PartyProto,
		metrics: PartyMetricsProto,
		partyIndex: number,
		logIndex: UnitLogIndex,
	): Promise<PartyMetrics> {
		const numPlayers = Math.min(party.players.length, metrics.players.length);
		const players = await Promise.all(
			[...new Array(numPlayers).keys()]
				.filter(i => party.players[i].class != Class.ClassUnknown)
				.map(i => UnitMetrics.makeNewPlayer(resultData, party.players[i], metrics.players[i], partyIndex * 5 + i, false, logIndex)),
		);

		return new PartyMetrics(players);
	}
}

export class UnitMetrics {
	// If this Unit is a pet, player is the owner. If it's a target, player is null.
	private readonly player: PlayerProto | null;
	private readonly target: TargetProto | null;
	private readonly metrics: UnitMetricsProto;

	readonly index: number;
	readonly unitIndex: number;
	readonly name: string;
	// For a target an add wave of the encounter design spawns, the wave and the add it spawns as; null otherwise.
	readonly designAddName: string | null;
	readonly spec: PlayerSpec<any> | null;
	readonly petActionId: ActionId | null;
	readonly iconUrl: string;
	readonly classColor: string;
	readonly dps: DistributionMetricsProto;
	readonly hps: DistributionMetricsProto;
	readonly tps: DistributionMetricsProto;
	readonly dtps: DistributionMetricsProto;
	readonly tmi: DistributionMetricsProto;
	readonly tto: DistributionMetricsProto;
	readonly actions: Array<ActionMetrics>;
	readonly auras: Array<AuraMetrics>;
	readonly resources: Array<ResourceMetrics>;
	readonly pets: Array<UnitMetrics>;
	private readonly iterations: number;
	private readonly duration: number;

	readonly logs: Array<CombatLog>;

	private readonly firstIterationDuration: number;
	private memoDamageDealtLogs?: Array<DamageLog>;
	private memoGroupedResourceLogs?: Record<ResourceType, Array<ResourceGroupLog>>;
	private memoDpsLogs?: Array<DpsLog>;
	private memoAuraUptimeLogs?: Array<AuraUptimeLog>;
	private memoMajorCooldownLogs?: Array<MajorCooldownLog>;
	private memoCastLogs?: Array<CastLog>;
	private memoThreatLogs?: Array<ThreatLogGroup>;
	private memoMajorCooldownAuraUptimeLogs?: Array<AuraUptimeLog>;

	get damageDealtLogs(): Array<DamageLog> {
		return (this.memoDamageDealtLogs ??= this.logs.filter(isDamage).filter(log => log.target?.isTarget !== this.isTarget));
	}

	get dpsLogs(): Array<DpsLog> {
		if (!this.memoDpsLogs) {
			this.memoDpsLogs = buildDpsLogs(this.damageDealtLogs);
			populateActiveAuras(this.memoDpsLogs, this.auraUptimeLogs);
		}
		return this.memoDpsLogs;
	}

	get castLogs(): Array<CastLog> {
		return (this.memoCastLogs ??= buildCastLogs(this.logs));
	}

	get threatLogs(): Array<ThreatLogGroup> {
		return (this.memoThreatLogs ??= buildThreatGroups(this.logs.filter(log => !isDamage(log) || log.target?.isTarget !== this.isTarget)));
	}

	get auraUptimeLogs(): Array<AuraUptimeLog> {
		return (this.memoAuraUptimeLogs ??= buildAuraUptimes(
			this.logs,
			new Entity(this.name, '', this.index, this.target != null, this.isPet),
			this.firstIterationDuration,
		));
	}

	get majorCooldownLogs(): Array<MajorCooldownLog> {
		return (this.memoMajorCooldownLogs ??= this.logs.filter(isMajorCooldown));
	}

	get groupedResourceLogs(): Record<ResourceType, Array<ResourceGroupLog>> {
		if (!this.memoGroupedResourceLogs) {
			this.memoGroupedResourceLogs = buildResourceGroups(this.logs);
			populateActiveAuras(this.memoGroupedResourceLogs[ResourceType.ResourceTypeMana], this.auraUptimeLogs);
		}
		return this.memoGroupedResourceLogs;
	}

	// Aura uptime logs, filtered to include only auras that correspond to a
	// major cooldown.
	get majorCooldownAuraUptimeLogs(): Array<AuraUptimeLog> {
		if (!this.memoMajorCooldownAuraUptimeLogs) {
			const mcdIds = new Set(this.majorCooldownLogs.map(mcdLog => mcdLog.actionId!.equalityKey()));
			this.memoMajorCooldownAuraUptimeLogs = this.auraUptimeLogs.filter(auraLog => mcdIds.has(auraLog.actionId!.equalityKey()));
		}
		return this.memoMajorCooldownAuraUptimeLogs;
	}

	private constructor(
		player: PlayerProto | null,
		target: TargetProto | null,
		petActionId: ActionId | null,
		metrics: UnitMetricsProto,
		index: number,
		actions: Array<ActionMetrics>,
		auras: Array<AuraMetrics>,
		resources: Array<ResourceMetrics>,
		pets: Array<UnitMetrics>,
		logs: Array<CombatLog>,
		resultData: SimResultData,
		designAddName: string | null = null,
	) {
		this.player = player;
		this.target = target;
		this.metrics = metrics;

		this.index = index;
		this.unitIndex = metrics.unitIndex;
		this.name = metrics.name;
		this.designAddName = designAddName;
		this.spec = this.player ? getPlayerSpecFromPlayer(this.player) : null;
		this.petActionId = petActionId;
		this.iconUrl = this.isPlayer ? (this.spec?.getIcon('medium') ?? '') : this.isTarget ? defaultTargetIcon : '';
		this.classColor = this.isTarget ? '' : (PlayerClasses.getCssScheme(PlayerSpecs.getPlayerClass(this.spec as PlayerSpec<any>)) ?? '');
		this.dps = this.metrics.dps!;
		this.hps = this.metrics.hps!;
		this.tps = this.metrics.threat!;
		this.dtps = this.metrics.dtps!;
		this.tmi = this.metrics.tmi!;
		this.tto = this.metrics.tto!;
		this.actions = actions;
		this.auras = auras;
		this.resources = resources;
		this.pets = pets;
		this.logs = logs;
		this.iterations = resultData.iterations;
		this.duration = resultData.duration;
		this.firstIterationDuration = resultData.firstIterationDuration;
	}

	get label() {
		if (this.target == null) {
			return `${this.name} (#${this.index + 1})`;
		} else {
			return `${this.name} (${this.designAddName ?? this.target.name})`;
		}
	}

	get isPlayer() {
		return this.player != null;
	}

	get isTarget() {
		return this.target != null;
	}

	get isPet() {
		return this.petActionId != null;
	}

	// The actions with button presses by the rotation on the filtered target, the ones APM counts.
	apmActions(filter?: SimResultFilter): Array<ActionMetrics> {
		return this.actions.map(action => action.forTarget(filter)).filter(action => action.manualCasts != 0);
	}

	// Actions per minute: the rotation's button presses on the filtered target, pets left out.
	actionsPerMinute(filter?: SimResultFilter): number {
		return sum(this.apmActions(filter).map(action => action.manualCastsPerMinute));
	}

	// Returns the unit index of the target of this unit, as selected by the filter.
	getTargetIndex(filter?: SimResultFilter): number | null {
		if (!filter) {
			return null;
		}

		const index = this.isPlayer ? filter.target : filter.player;
		if (index == null || index == -1) {
			return null;
		}

		return index;
	}

	get inFrontOfTarget(): boolean {
		if (this.isTarget) {
			return true;
		} else if (this.isPlayer) {
			return this.player!.inFrontOfTarget;
		} else {
			return false; // TODO pets
		}
	}

	get chanceOfDeath(): DistributionMetricsProto {
		const p = this.metrics.chanceOfDeath;
		const err = Math.sqrt((p * (1 - p)) / this.iterations);

		return DistributionMetricsProto.create({
			avg: p * 100,
			stdev: err * 100,
		});
	}

	get maxThreat() {
		return this.threatLogs[this.threatLogs.length - 1]?.threatAfter || 0;
	}

	get secondsOomAvg() {
		return this.metrics.secondsOomAvg;
	}

	get totalDamage() {
		return this.dps.avg * this.duration;
	}

	get totalDamageTaken() {
		return this.dtps.avg * this.duration;
	}

	getPlayerAndPetActions(): Array<ActionMetrics> {
		return this.actions.concat(this.pets.map(pet => pet.getPlayerAndPetActions()).flat());
	}

	private getActionsForDisplay(): Array<ActionMetrics> {
		return this.actions.filter(e => e.hitAttempts != 0 || e.tps != 0 || e.dps != 0);
	}

	getMeleeActions(): Array<ActionMetrics> {
		return this.getActionsForDisplay().filter(e => e.isMeleeAction);
	}

	getSpellActions(): Array<ActionMetrics> {
		return this.getActionsForDisplay().filter(e => !e.isMeleeAction);
	}

	getDamageActions(): Array<ActionMetrics> {
		return this.getActionsForDisplay()
			.filter(e => e.hps === 0)
			.map(e => e.againstOpponents())
			.filter(e => e.dps !== 0);
	}

	getDamageActionsOn(unitIndex: number): Array<ActionMetrics> {
		return this.getActionsForDisplay()
			.filter(e => e.hps === 0 && e.targets.some(target => target.data.unitIndex === unitIndex))
			.map(e => e.forUnit(unitIndex))
			.filter(e => e.damage > 0);
	}

	getThreatActions(): Array<ActionMetrics> {
		return this.getActionsForDisplay()
			.filter(e => e.hps === 0)
			.map(e => e.againstOpponents())
			.filter(e => e.tps !== 0);
	}

	getHealingActions(): Array<ActionMetrics> {
		return this.getActionsForDisplay();
	}

	getResourceMetrics(resourceType: ResourceType): Array<ResourceMetrics> {
		return this.resources.filter(resource => resource.type == resourceType);
	}

	static async makeNewPlayer(
		resultData: SimResultData,
		player: PlayerProto,
		metrics: UnitMetricsProto,
		raidIndex: number,
		isPet: boolean,
		logIndex: UnitLogIndex,
	): Promise<UnitMetrics> {
		const playerLogs = isPet ? logIndex.pet(raidIndex, metrics.name) : logIndex.player(raidIndex);

		const actionsPromise = Promise.all(metrics.actions.map(actionMetrics => ActionMetrics.makeNew(null, resultData, actionMetrics, raidIndex)));
		const aurasPromise = Promise.all(metrics.auras.map(auraMetrics => AuraMetrics.makeNew(null, resultData, auraMetrics, raidIndex)));
		const resourcesPromise = Promise.all(metrics.resources.map(resourceMetrics => ResourceMetrics.makeNew(null, resultData, resourceMetrics, raidIndex)));
		const petsPromise = Promise.all(metrics.pets.map(petMetrics => UnitMetrics.makeNewPlayer(resultData, player, petMetrics, raidIndex, true, logIndex)));

		let petIdPromise: Promise<ActionId | null> = Promise.resolve(null);
		if (isPet) {
			petIdPromise = ActionId.fromPetName(metrics.name).fill(raidIndex);
		}

		const actions = await actionsPromise;
		const auras = await aurasPromise;
		const resources = await resourcesPromise;
		const pets = await petsPromise;
		const petActionId = await petIdPromise;

		const playerMetrics = new UnitMetrics(player, null, petActionId, metrics, raidIndex, actions, auras, resources, pets, playerLogs, resultData);
		actions.forEach(action => {
			action.unit = playerMetrics;
			action.resources = resources.filter(resourceMetrics => resourceMetrics.actionId.equals(action.actionId));
		});
		auras.forEach(aura => (aura.unit = playerMetrics));
		resources.forEach(resource => (resource.unit = playerMetrics));
		return playerMetrics;
	}

	static async makeNewTarget(
		resultData: SimResultData,
		target: TargetProto,
		metrics: UnitMetricsProto,
		index: number,
		logIndex: UnitLogIndex,
		designAddName: string | null = null,
	): Promise<UnitMetrics> {
		const targetLogs = logIndex.target(index);

		const actionsPromise = Promise.all(metrics.actions.map(actionMetrics => ActionMetrics.makeNew(null, resultData, actionMetrics, index)));
		const aurasPromise = Promise.all(metrics.auras.map(auraMetrics => AuraMetrics.makeNew(null, resultData, auraMetrics)));

		const actions = await actionsPromise;
		const auras = await aurasPromise;

		const targetMetrics = new UnitMetrics(null, target, null, metrics, index, actions, auras, [], [], targetLogs, resultData, designAddName);
		actions.forEach(action => (action.unit = targetMetrics));
		auras.forEach(aura => (aura.unit = targetMetrics));
		return targetMetrics;
	}
}

// The add's template and its name as the encounter designer shows them: the wave's name, else "Wave N", then the
// add's, else "Add N".
const designAddTarget = (design: EncounterDesign | undefined, add: DesignAdd): { template: TargetProto; name: string } => {
	const wave = design?.addWaves[add.waveIndex];
	const template = wave?.adds[add.addIndex] ?? TargetProto.create();
	const waveName =
		wave?.name ||
		i18n.t('settings_tab.encounter.designer.event_name', {
			event: i18n.t('settings_tab.encounter.designer.events.wave'),
			number: add.waveIndex + 1,
		});
	const addName = template.name || i18n.t('settings_tab.encounter.designer.adds.add_title', { number: add.addIndex + 1 });
	return { template, name: `${waveName} · ${addName}` };
};

export class EncounterMetrics {
	private readonly encounter: EncounterProto;
	private readonly metrics: EncounterMetricsProto;

	readonly targets: Array<UnitMetrics>;

	private constructor(encounter: EncounterProto, metrics: EncounterMetricsProto, targets: Array<UnitMetrics>) {
		this.encounter = encounter;
		this.metrics = metrics;
		this.targets = targets;
	}

	static async makeNew(
		resultData: SimResultData,
		encounter: EncounterProto,
		metrics: EncounterMetricsProto,
		logIndex: UnitLogIndex,
	): Promise<EncounterMetrics> {
		// The sim lists the encounter's own targets, then every add slot of the design's waves. A target past the
		// encounter's own without a design add is the stand-in of an encounter without targets.
		const namesSeen = new Map<string, number>();
		const targets = await Promise.all(
			metrics.targets.map((targetMetrics, i) => {
				if (i < encounter.targets.length) return UnitMetrics.makeNewTarget(resultData, encounter.targets[i], targetMetrics, i, logIndex);
				if (!targetMetrics.designAdd) return null;
				const { template, name } = designAddTarget(encounter.design, targetMetrics.designAdd);
				const seen = (namesSeen.get(name) ?? 0) + 1;
				namesSeen.set(name, seen);
				const uniqueName = seen > 1 ? i18n.t('settings_tab.encounter.designer.occurrence_name', { name, number: seen }) : name;
				return UnitMetrics.makeNewTarget(resultData, template, targetMetrics, i, logIndex, uniqueName);
			}),
		);

		return new EncounterMetrics(
			encounter,
			metrics,
			targets.filter(target => target != null),
		);
	}

	get durationSeconds() {
		return this.encounter.duration;
	}
}

export class AuraMetrics {
	unit: UnitMetrics | null;
	readonly actionId: ActionId;
	readonly name: string;
	readonly iconUrl: string;
	private readonly resultData: SimResultData;
	private readonly iterations: number;
	private readonly duration: number;
	private readonly data: AuraMetricsProto;

	private constructor(unit: UnitMetrics | null, actionId: ActionId, data: AuraMetricsProto, resultData: SimResultData) {
		this.unit = unit;
		this.actionId = actionId;
		this.name = actionId.name;
		this.iconUrl = actionId.iconUrl;
		this.data = data;
		this.resultData = resultData;
		this.iterations = resultData.iterations;
		this.duration = resultData.duration;
	}

	get uptimePercent() {
		return percentOf(this.data.uptimeSecondsAvg, this.duration);
	}

	get averageProcs() {
		return this.data.procsAvg;
	}

	get ppm() {
		return this.data.procsAvg / (this.duration / 60);
	}

	static async makeNew(unit: UnitMetrics | null, resultData: SimResultData, auraMetrics: AuraMetricsProto, playerIndex?: number): Promise<AuraMetrics> {
		const actionId = await ActionId.fromProto(auraMetrics.id!).fill(playerIndex);
		return new AuraMetrics(unit, actionId, auraMetrics, resultData);
	}

	// Merges an array of metrics into a single metrics.
	static merge(auras: Array<AuraMetrics>, { removeTag, actionIdOverride }: { removeTag?: boolean; actionIdOverride?: ActionId } = {}): AuraMetrics {
		const firstAura = auras[0];
		const unit = auras.every(aura => aura.unit == firstAura.unit) ? firstAura.unit : null;
		let actionId = actionIdOverride || firstAura.actionId;
		if (removeTag) {
			actionId = actionId.withoutTag();
		}
		return new AuraMetrics(
			unit,
			actionId,
			AuraMetricsProto.create({
				uptimeSecondsAvg: Math.max(...auras.map(a => a.data.uptimeSecondsAvg)),
			}),
			firstAura.resultData,
		);
	}

	// Groups similar metrics, i.e. metrics with the same item/spell/other ID but
	// different tags, and returns them as separate arrays.
	static groupById(auras: Array<AuraMetrics>, useTag?: boolean): Array<Array<AuraMetrics>> {
		if (useTag) {
			return Object.values(bucket(auras, aura => aura.actionId.toString()));
		} else {
			return Object.values(bucket(auras, aura => aura.actionId.toStringIgnoringTag()));
		}
	}

	// Merges aura metrics that have the same name/ID, adding their stats together.
	static joinById(auras: Array<AuraMetrics>, useTag?: boolean): Array<AuraMetrics> {
		return AuraMetrics.groupById(auras, useTag).map(aurasToJoin => AuraMetrics.merge(aurasToJoin));
	}
}
export class ResourceMetrics {
	unit: UnitMetrics | null;
	readonly actionId: ActionId;
	readonly name: string;
	readonly iconUrl: string;
	readonly type: ResourceType;
	private readonly resultData: SimResultData;
	private readonly iterations: number;
	private readonly duration: number;
	private readonly data: ResourceMetricsProto;

	private constructor(unit: UnitMetrics | null, actionId: ActionId, data: ResourceMetricsProto, resultData: SimResultData) {
		this.unit = unit;
		this.actionId = actionId;
		this.name = actionId.name;
		this.iconUrl = actionId.iconUrl;
		this.type = data.type;
		this.resultData = resultData;
		this.iterations = resultData.iterations;
		this.duration = resultData.duration;
		this.data = data;
	}

	get events() {
		return this.data.events / this.iterations;
	}

	get gain() {
		return this.data.gain / this.iterations;
	}

	get gainPerSecond() {
		return this.data.gain / this.iterations / this.duration;
	}

	get avgGain() {
		return this.data.gain / this.data.events;
	}

	get wastedGain() {
		return (this.data.gain - this.data.actualGain) / this.iterations;
	}

	static async makeNew(
		unit: UnitMetrics | null,
		resultData: SimResultData,
		resourceMetrics: ResourceMetricsProto,
		playerIndex?: number,
	): Promise<ResourceMetrics> {
		const actionId = await ActionId.fromProto(resourceMetrics.id!).fill(playerIndex);
		return new ResourceMetrics(unit, actionId, resourceMetrics, resultData);
	}

	// Merges an array of metrics into a single metrics.
	static merge(
		resources: Array<ResourceMetrics>,
		{ removeTag, actionIdOverride }: { removeTag?: boolean; actionIdOverride?: ActionId } = {},
	): ResourceMetrics {
		const firstResource = resources[0];
		const unit = resources.every(resource => resource.unit == firstResource.unit) ? firstResource.unit : null;
		let actionId = actionIdOverride || firstResource.actionId;
		if (removeTag) {
			actionId = actionId.withoutTag();
		}
		return new ResourceMetrics(
			unit,
			actionId,
			ResourceMetricsProto.create({
				events: sum(resources.map(a => a.data.events)),
				gain: sum(resources.map(a => a.data.gain)),
				actualGain: sum(resources.map(a => a.data.actualGain)),
			}),
			firstResource.resultData,
		);
	}

	// Groups similar metrics, i.e. metrics with the same item/spell/other ID but
	// different tags, and returns them as separate arrays.
	static groupById(resources: Array<ResourceMetrics>, useTag?: boolean): Array<Array<ResourceMetrics>> {
		if (useTag) {
			return Object.values(bucket(resources, resource => resource.actionId.toString()));
		} else {
			return Object.values(bucket(resources, resource => resource.actionId.toStringIgnoringTag()));
		}
	}
}

// Manages the metrics for a single unit action (e.g. Lightning Bolt).
export class ActionMetrics {
	unit: UnitMetrics | null;
	readonly actionId: ActionId;
	readonly name: string;
	readonly iconUrl: string;
	readonly spellSchool: SpellSchool | null;
	readonly targets: Array<TargetedActionMetrics>;
	private readonly resultData: SimResultData;
	private readonly iterations: number;
	private readonly duration: number;
	private readonly data: ActionMetricsProto;
	private readonly combinedMetrics: TargetedActionMetrics;
	resources: Array<ResourceMetrics>;

	private constructor(unit: UnitMetrics | null, actionId: ActionId, data: ActionMetricsProto, resultData: SimResultData) {
		this.unit = unit;
		this.actionId = actionId;
		this.name = actionId.name;
		this.iconUrl = actionId.iconUrl;
		this.resultData = resultData;
		this.iterations = resultData.iterations;
		this.duration = resultData.duration;
		this.data = data;
		this.spellSchool = data.spellSchool;
		this.targets = data.targets.map(
			tam =>
				new TargetedActionMetrics(tam, {
					iterations: this.iterations,
					duration: this.duration,
				}),
		);
		this.combinedMetrics = TargetedActionMetrics.merge(this.targets);
		this.resources = [];
	}

	get isMeleeAction() {
		return this.data.isMelee;
	}

	get isPassiveAction() {
		return this.data.isPassive;
	}

	get totalDamagePercent() {
		const totalAvgDps = this.resultData.result.raidMetrics?.dps?.avg;
		if (!totalAvgDps) return undefined;

		return percentOf(this.avgDamage, totalAvgDps * this.duration);
	}

	get damage() {
		return this.combinedMetrics.damage;
	}

	get avgDamage() {
		return this.combinedMetrics.avgDamage;
	}

	get avgHitDamage() {
		return (
			this.avgDamage -
			this.avgTickDamage -
			this.avgCritDamage +
			this.avgCritTickDamage -
			this.avgGlanceDamage -
			this.avgBlockDamage -
			this.avgBlockedCritDamage -
			this.avgCrushDamage
		);
	}

	get resistedDamage() {
		return this.combinedMetrics.resistedDamage;
	}

	get avgResistedDamage() {
		return this.combinedMetrics.avgResistedDamage;
	}

	get critDamage() {
		return this.combinedMetrics.critDamage;
	}

	get avgCritDamage() {
		return this.combinedMetrics.avgCritDamage;
	}

	get resistedCritDamage() {
		return this.combinedMetrics.resistedCritDamage;
	}

	get avgResistedCritDamage() {
		return this.combinedMetrics.avgResistedCritDamage;
	}

	get tickDamage() {
		return this.combinedMetrics.tickDamage;
	}

	get avgTickDamage() {
		return this.combinedMetrics.avgTickDamage;
	}

	get resistedTickDamage() {
		return this.combinedMetrics.resistedTickDamage;
	}

	get avgResistedTickDamage() {
		return this.combinedMetrics.avgResistedTickDamage;
	}

	get critTickDamage() {
		return this.combinedMetrics.critTickDamage;
	}

	get avgCritTickDamage() {
		return this.combinedMetrics.avgCritTickDamage;
	}
	get resistedCritTickDamage() {
		return this.combinedMetrics.resistedCritTickDamage;
	}

	get avgResistedCritTickDamage() {
		return this.combinedMetrics.avgResistedCritTickDamage;
	}

	get glanceDamage() {
		return this.combinedMetrics.glanceDamage;
	}

	get avgGlanceDamage() {
		return this.combinedMetrics.avgGlanceDamage;
	}

	get blockDamage() {
		return this.combinedMetrics.blockDamage;
	}

	get crushDamage() {
		return this.combinedMetrics.crushDamage;
	}

	get avgCrushDamage() {
		return this.combinedMetrics.avgCrushDamage;
	}

	get avgBlockDamage() {
		return this.combinedMetrics.avgBlockDamage;
	}

	get blockedCritDamage() {
		return this.combinedMetrics.blockedCritDamage;
	}

	get avgBlockedCritDamage() {
		return this.combinedMetrics.avgBlockedCritDamage;
	}

	get dps() {
		return this.combinedMetrics.dps;
	}

	get threat() {
		return this.combinedMetrics.threat;
	}

	get avgThreat() {
		return this.combinedMetrics.avgThreat;
	}

	get totalThreatPercent() {
		const totalAvgTps = this.resultData.result.raidMetrics?.parties[0].players[0].threat?.avg;
		if (!totalAvgTps) return undefined;

		return percentOf(this.avgThreat, totalAvgTps * this.duration);
	}

	get tps() {
		return this.combinedMetrics.tps;
	}

	get totalHealingPercent() {
		const totalAvgHps = this.resultData.result.raidMetrics?.hps?.avg;
		if (!totalAvgHps) return undefined;

		return percentOf(this.avgHealing, totalAvgHps * this.duration);
	}

	get healing() {
		return this.combinedMetrics.healing;
	}

	get avgHealing() {
		return this.combinedMetrics.healing / this.iterations;
	}

	get critHealing() {
		return this.combinedMetrics.critHealing;
	}

	get avgCritHealing() {
		return this.combinedMetrics.critHealing / this.iterations;
	}

	get hps() {
		return this.combinedMetrics.hps;
	}

	get casts() {
		if (this.isPassiveAction) return 0;
		return this.combinedMetrics.casts;
	}

	get castsPerMinute() {
		if (this.isPassiveAction) return 0;
		return this.combinedMetrics.castsPerMinute;
	}

	// Presses by the rotation, which a passive action can have too (Judgement is cast by hand and flagged passive).
	get manualCasts() {
		return this.combinedMetrics.manualCasts;
	}

	get manualCastsPerMinute() {
		return this.combinedMetrics.manualCastsPerMinute;
	}

	get avgCastTimeMs() {
		if (this.isPassiveAction) return 0;
		return this.combinedMetrics.avgCastTimeMs;
	}

	get hpm() {
		const totalHealing = this.combinedMetrics.hps * this.duration;
		const manaMetrics = this.resources.find(r => r.type == ResourceType.ResourceTypeMana);
		if (manaMetrics) {
			return totalHealing / -manaMetrics.gain;
		}

		return 0;
	}

	get damageThroughput() {
		if (this.unit?.isPet && !this.actionId.spellId) return 0;
		return this.combinedMetrics.damageThroughput;
	}

	get threatThroughput() {
		if (this.unit?.isPet && !this.actionId.spellId) return 0;
		return this.combinedMetrics.threatThroughput;
	}

	get healingThroughput() {
		return this.combinedMetrics.healingThroughput;
	}

	get shielding() {
		return this.combinedMetrics.shielding;
	}

	get avgCast() {
		if (this.isPassiveAction) return 0;
		return this.combinedMetrics.avgCast;
	}

	get avgCastHit() {
		if (!this.combinedMetrics.avgCast) return 0;
		return this.combinedMetrics.avgCast - this.avgCastTick;
	}

	get avgCastTick() {
		return this.combinedMetrics.avgCastTick;
	}

	get avgCastHealing() {
		if (this.isPassiveAction) return 0;
		return this.combinedMetrics.avgCastHealing;
	}

	get avgCastThreat() {
		if (this.isPassiveAction) return 0;
		return this.combinedMetrics.avgCastThreat;
	}

	get landedHits() {
		return this.combinedMetrics.landedHits;
	}

	get landedTicks() {
		return this.combinedMetrics.landedTicks;
	}

	get landedRolls() {
		return this.combinedMetrics.landedRolls;
	}

	get hitAttempts() {
		return this.combinedMetrics.hitAttempts;
	}

	get avgHit() {
		return this.combinedMetrics.avgHit;
	}

	get avgTick() {
		return this.combinedMetrics.avgTick;
	}

	get avgHitHealing() {
		return this.combinedMetrics.avgHitHealing;
	}

	get avgHitThreat() {
		return this.combinedMetrics.avgHitThreat;
	}

	get totalMisses() {
		return this.misses + this.dodges + this.parries;
	}

	get totalMissesPercent() {
		return this.missPercent + this.dodgePercent + this.parryPercent;
	}

	get misses() {
		return this.combinedMetrics.misses;
	}

	get missPercent() {
		return this.combinedMetrics.missPercent;
	}

	get dodges() {
		return this.combinedMetrics.dodges;
	}

	get dodgePercent() {
		return this.combinedMetrics.dodgePercent;
	}

	get parries() {
		return this.combinedMetrics.parries;
	}

	get parryPercent() {
		return this.combinedMetrics.parryPercent;
	}

	get hits() {
		return this.combinedMetrics.hits;
	}

	get resistedHits() {
		return this.combinedMetrics.resistedHits;
	}

	get hitPercent() {
		return this.combinedMetrics.hitPercent;
	}

	get resistedHitPercent() {
		return this.combinedMetrics.resistedHitPercent;
	}

	get ticks() {
		return this.combinedMetrics.ticks;
	}

	get resistedTicks() {
		return this.combinedMetrics.resistedTicks;
	}

	get resistedTickPercent() {
		return this.combinedMetrics.resistedTickPercent;
	}

	get critTicks() {
		return this.combinedMetrics.critTicks;
	}

	get critTickPercent() {
		return this.combinedMetrics.critTickPercent;
	}

	get resistedCritTicks() {
		return this.combinedMetrics.resistedCritTicks;
	}

	get resistedCritTickPercent() {
		return this.combinedMetrics.resistedCritTickPercent;
	}

	get crushes() {
		return this.combinedMetrics.crushes;
	}

	get crushPercent() {
		return this.combinedMetrics.crushPercent;
	}

	get blocks() {
		return this.combinedMetrics.blocks;
	}

	get blockPercent() {
		return this.combinedMetrics.blockPercent;
	}

	get blockedCrits() {
		return this.combinedMetrics.blockedCrits;
	}

	get blockedCritPercent() {
		return this.combinedMetrics.blockedCritPercent;
	}

	get glances() {
		return this.combinedMetrics.glances;
	}

	get glancePercent() {
		return this.combinedMetrics.glancePercent;
	}

	get crits() {
		return this.combinedMetrics.crits;
	}

	get critPercent() {
		return this.combinedMetrics.critPercent;
	}

	get resistedCrits() {
		return this.combinedMetrics.resistedCrits;
	}

	get resistedCritPercent() {
		return this.combinedMetrics.resistedCritPercent;
	}

	get healingPercent() {
		return this.combinedMetrics.healingPercent;
	}

	get healingCritPercent() {
		return this.combinedMetrics.healingCritPercent;
	}

	range(kind: HitKind) {
		return this.combinedMetrics.range(kind);
	}

	private perIteration(range: DistributionMetricsProto) {
		return (range.avg * rangeCount(range)) / this.iterations;
	}

	// Each value against the action's own amount: its damage, or its healing when it deals none.
	get customMetrics() {
		const actionTotal = this.avgDamage || this.avgHealing;
		return this.combinedMetrics.customMetrics.flatMap(({ key, range }) => {
			if (!range) return [];
			const value = this.perIteration(range);
			return [{ key, value, percentage: percentOf(value, actionTotal), average: range.avg, min: range.min, max: range.max }];
		});
	}

	private bounds(kind: HitKind) {
		const range = this.range(kind);
		return range ? { min: range.min, max: range.max } : {};
	}

	get healingDone() {
		const entry = (kind: HitKind) => {
			const range = this.range(kind);
			const value = range ? this.perIteration(range) : 0;
			return { value, percentage: percentOf(value, this.avgHealing), average: range?.avg ?? 0, ...this.bounds(kind) };
		};
		return {
			heal: entry(HitKind.HitKindHeal),
			critHeal: entry(HitKind.HitKindCritHeal),
			tick: entry(HitKind.HitKindHealTick),
			critTick: entry(HitKind.HitKindCritHealTick),
			shield: entry(HitKind.HitKindShield),
		};
	}

	get damageDone() {
		const normalHitAvgDamage = Number(
			(
				this.avgDamage -
				this.avgResistedDamage +
				this.avgResistedTickDamage +
				this.avgResistedCritDamage -
				this.avgResistedCritTickDamage -
				this.avgCritDamage +
				this.avgCritTickDamage -
				this.avgTickDamage -
				this.avgGlanceDamage -
				this.avgBlockDamage -
				this.avgBlockedCritDamage -
				this.avgCrushDamage
			).toFixed(8),
		);
		const normalResistedHitAvgDamage = Number(
			(this.avgResistedDamage - this.avgResistedCritDamage - this.avgResistedTickDamage + this.avgResistedCritTickDamage).toFixed(8),
		);
		const critHitAvgDamage = Number((this.avgCritDamage - this.avgResistedCritDamage - this.avgCritTickDamage + this.avgResistedCritTickDamage).toFixed(8));
		const resistedCritHitAvgDamage = Number((this.avgResistedCritDamage - this.avgResistedCritTickDamage).toFixed(8));

		const normalTickAvgDamage = Number(
			(this.avgTickDamage - this.avgCritTickDamage - this.avgResistedTickDamage + this.avgResistedCritTickDamage).toFixed(8),
		);
		const normalResistedTickAvgDamage = Number((this.avgResistedTickDamage - this.avgResistedCritTickDamage).toFixed(8));
		const normalCritTickAvgDamage = Number((this.avgCritTickDamage - this.avgResistedCritTickDamage).toFixed(8));

		return {
			hit: {
				value: normalHitAvgDamage,
				percentage: percentOf(normalHitAvgDamage, this.avgDamage),
				average: normalHitAvgDamage / (this.hits - this.resistedHits),
				...this.bounds(HitKind.HitKindHit),
			},
			resistedHit: {
				value: normalResistedHitAvgDamage,
				percentage: percentOf(normalResistedHitAvgDamage, this.avgDamage),
				average: normalResistedHitAvgDamage / this.resistedHits,
				...this.bounds(HitKind.HitKindResistedHit),
			},
			critHit: {
				value: critHitAvgDamage,
				percentage: percentOf(critHitAvgDamage, this.avgDamage),
				average: critHitAvgDamage / (this.crits - this.resistedCrits),
				...this.bounds(HitKind.HitKindCrit),
			},
			resistedCritHit: {
				value: resistedCritHitAvgDamage,
				percentage: percentOf(resistedCritHitAvgDamage, this.avgDamage),
				average: resistedCritHitAvgDamage / this.resistedCrits,
				...this.bounds(HitKind.HitKindResistedCrit),
			},
			tick: {
				value: normalTickAvgDamage,
				percentage: percentOf(normalTickAvgDamage, this.avgDamage),
				average: normalTickAvgDamage / (this.ticks - this.resistedTicks),
				...this.bounds(HitKind.HitKindTick),
			},
			resistedTick: {
				value: normalResistedTickAvgDamage,
				percentage: percentOf(normalResistedTickAvgDamage, this.avgDamage),
				average: normalResistedTickAvgDamage / this.resistedTicks,
				...this.bounds(HitKind.HitKindResistedTick),
			},
			critTick: {
				value: normalCritTickAvgDamage,
				percentage: percentOf(normalCritTickAvgDamage, this.avgDamage),
				average: normalCritTickAvgDamage / (this.critTicks - this.resistedCritTicks),
				...this.bounds(HitKind.HitKindCritTick),
			},
			resistedCritTick: {
				value: this.avgResistedCritTickDamage,
				percentage: percentOf(this.avgResistedCritTickDamage, this.avgDamage),
				average: this.avgResistedCritTickDamage / this.resistedCritTicks,
				...this.bounds(HitKind.HitKindResistedCritTick),
			},
			glance: {
				value: this.avgGlanceDamage,
				percentage: percentOf(this.avgGlanceDamage, this.avgDamage),
				average: this.avgGlanceDamage / this.glances,
				...this.bounds(HitKind.HitKindGlance),
			},
			block: {
				value: this.avgBlockDamage,
				percentage: percentOf(this.avgBlockDamage, this.avgDamage),
				average: this.avgBlockDamage / this.blocks,
				...this.bounds(HitKind.HitKindBlock),
			},
			blockedCrit: {
				value: this.avgBlockedCritDamage,
				percentage: percentOf(this.avgBlockedCritDamage, this.avgDamage),
				average: this.avgBlockedCritDamage / this.blockedCrits,
				...this.bounds(HitKind.HitKindBlockedCrit),
			},
			crush: {
				value: this.avgCrushDamage,
				percentage: percentOf(this.avgCrushDamage, this.avgDamage),
				average: this.avgCrushDamage / this.crushes,
				...this.bounds(HitKind.HitKindCrush),
			},
		};
	}

	// The action without the units on its own side when it damages or threatens any of them. Damage and
	// threat it puts on its own side, such as a sapper's hit on its thrower, are taken, not done.
	againstOpponents(): ActionMetrics {
		const enemyIndices = new Set(this.resultData.result.encounterMetrics!.targets.map(target => target.unitIndex));
		const onEnemySide = this.unit?.isTarget ?? false;
		const isOpponent = (target: TargetedActionMetricsProto) => enemyIndices.has(target.unitIndex) !== onEnemySide;
		if (this.data.targets.every(target => isOpponent(target) || (target.damage === 0 && target.threat === 0))) {
			return this;
		}

		return this.withTargets(this.data.targets.filter(isOpponent));
	}

	forTarget(filter?: SimResultFilter): ActionMetrics {
		const unitIndex = this.unit!.getTargetIndex(filter);
		return unitIndex == null ? this : this.forUnit(unitIndex);
	}

	forUnit(unitIndex: number): ActionMetrics {
		const target = this.data.targets.find(target => target.unitIndex == unitIndex);
		if (!target) {
			throw new Error('Could not find target with unitIndex ' + unitIndex);
		}
		return this.withTargets([target]);
	}

	private withTargets(targets: Array<TargetedActionMetricsProto>): ActionMetrics {
		return new ActionMetrics(this.unit, this.actionId, ActionMetricsProto.create({ ...this.data, targets }), this.resultData);
	}

	static async makeNew(unit: UnitMetrics | null, resultData: SimResultData, actionMetrics: ActionMetricsProto, playerIndex?: number): Promise<ActionMetrics> {
		const actionId = await ActionId.fromProto(actionMetrics.id!).fill(playerIndex);
		return new ActionMetrics(unit, actionId, actionMetrics, resultData);
	}

	// Merges an array of metrics into a single metric.
	static merge(actions: Array<ActionMetrics>, { removeTag, actionIdOverride }: { removeTag?: boolean; actionIdOverride?: ActionId } = {}): ActionMetrics {
		const firstAction = actions[0];
		const unit = firstAction.unit;
		let actionId = actionIdOverride || firstAction.actionId;
		if (removeTag) {
			actionId = actionId.withoutTag();
		}

		const maxTargets = Math.max(...actions.map(action => action.targets.length));
		const mergedTargets = [...Array(maxTargets).keys()].map(i => TargetedActionMetrics.merge(actions.map(action => action.targets[i])));
		const isAllPassiveSpells = actions.every(action => action.isPassiveAction);

		return new ActionMetrics(
			unit,
			actionId,
			ActionMetricsProto.create({
				isMelee: firstAction.isMeleeAction,
				isPassive: isAllPassiveSpells,
				targets: mergedTargets.map(t => t.data),
				spellSchool: firstAction.spellSchool || undefined,
			}),
			firstAction.resultData,
		);
	}

	// Groups similar metrics, i.e. metrics with the same item/spell/other ID but
	// different tags, and returns them as separate arrays.
	static groupById(actions: Array<ActionMetrics>, useTag?: boolean): Array<Array<ActionMetrics>> {
		if (useTag) {
			return Object.values(bucket(actions, action => action.actionId.toString()));
		} else {
			return Object.values(bucket(actions, action => action.actionId.toStringIgnoringTag()));
		}
	}

	// Merges action metrics that have the same name/ID, adding their stats together.
	static joinById(actions: Array<ActionMetrics>, useTag?: boolean): Array<ActionMetrics> {
		return ActionMetrics.groupById(actions, useTag).map(actionsToJoin => ActionMetrics.merge(actionsToJoin));
	}
}

type TargetedActionMetricsOptions = {
	iterations: number;
	duration: number;
};

const rangeCount = (range: DistributionMetricsProto | undefined) => range?.aggregatorData?.n ?? 0;

// Joins per-value ranges into one with Chan's pairwise update of the count, mean and sum of squared
// deviations; a range carries only its count, so the bounds skip a side that landed no values.
const mergeRanges = (ranges: Array<DistributionMetricsProto | undefined>): DistributionMetricsProto | undefined => {
	const filled = ranges.filter((range): range is DistributionMetricsProto => rangeCount(range) > 0);
	if (!filled.length) return undefined;

	let n = 0;
	let avg = 0;
	let m2 = 0;
	for (const range of filled) {
		const rn = rangeCount(range);
		const delta = range.avg - avg;
		const total = n + rn;
		m2 += range.stdev * range.stdev * rn + (delta * delta * n * rn) / total;
		avg += (delta * rn) / total;
		n = total;
	}
	return DistributionMetricsProto.create({
		avg,
		stdev: Math.sqrt(m2 / n),
		min: Math.min(...filled.map(range => range.min)),
		max: Math.max(...filled.map(range => range.max)),
		aggregatorData: { n, sumSq: 0 },
	});
};

const mergeRangeMaps = (maps: Array<TargetedActionMetricsProto['ranges']>): TargetedActionMetricsProto['ranges'] => {
	if (maps.length === 1) return maps[0];
	const merged: TargetedActionMetricsProto['ranges'] = {};
	for (const kind of new Set(maps.flatMap(map => Object.keys(map).map(Number)))) {
		const range = mergeRanges(maps.map(map => map[kind]));
		if (range) merged[kind] = range;
	}
	return merged;
};

const mergeCustomMetrics = (lists: Array<TargetedActionMetricsProto['customMetrics']>): TargetedActionMetricsProto['customMetrics'] => {
	if (lists.length === 1) return lists[0];
	const keys = [...new Set(lists.flatMap(list => list.map(metric => metric.key)))];
	return keys.flatMap(key => {
		const range = mergeRanges(lists.map(list => list.find(metric => metric.key == key)?.range));
		return range ? [{ key, range }] : [];
	});
};

// Manages the metrics for a single action applied to a specific target.
class TargetedActionMetrics {
	private readonly iterations: number;
	private readonly duration: number;
	readonly data: TargetedActionMetricsProto;

	readonly landedHitsRaw: number;
	readonly landedTicksRaw: number;
	readonly hitAttempts: number;
	readonly landedRollsRaw: number;
	private readonly rollAttempts: number;

	constructor(data: TargetedActionMetricsProto, { iterations, duration }: TargetedActionMetricsOptions) {
		this.iterations = iterations;
		this.duration = duration;
		// A range or custom metric that holds no values is dropped here, so no read has to check for one.
		this.data = {
			...data,
			ranges: Object.fromEntries(Object.entries(data.ranges).filter(([, range]) => rangeCount(range) > 0)),
			customMetrics: data.customMetrics.filter(metric => rangeCount(metric.range) > 0),
		};

		this.landedHitsRaw = this.data.hits + this.data.crits + this.data.blocks + this.data.blockedCrits + this.data.glances + this.data.crushes;
		this.landedTicksRaw = this.data.ticks + this.data.critTicks;

		// A DoT counts no hit when its cast lands, so a spell that lands only ticks takes its casts,
		// which include the misses, as its hit attempts.
		this.hitAttempts =
			this.landedHitsRaw == 0 && this.landedTicksRaw > 0 ? this.data.casts : this.data.misses + this.data.dodges + this.data.parries + this.landedHitsRaw;

		// Ticks that roll their own hit are attempts of their own, and a tick miss is the only sign that they
		// roll: a hit-capped row of them reads 0% over its casts as it would over its ticks.
		const ticksRoll = this.data.tickMisses > 0;
		this.landedRollsRaw = this.landedHitsRaw + (ticksRoll ? this.landedTicksRaw : 0);
		this.rollAttempts = ticksRoll ? this.landedRollsRaw + this.data.misses + this.data.tickMisses + this.data.dodges + this.data.parries : this.hitAttempts;
	}

	get damage() {
		return this.data.damage;
	}

	get avgDamage() {
		return this.data.damage / this.iterations;
	}

	get resistedDamage() {
		return this.data.resistedDamage;
	}

	get avgResistedDamage() {
		return this.data.resistedDamage / this.iterations;
	}

	get critDamage() {
		return this.data.critDamage;
	}

	get avgCritDamage() {
		return this.data.critDamage / this.iterations;
	}

	get resistedCritDamage() {
		return this.data.resistedCritDamage;
	}

	get avgResistedCritDamage() {
		return this.data.resistedCritDamage / this.iterations;
	}

	get tickDamage() {
		return this.data.tickDamage;
	}

	get avgTickDamage() {
		return this.data.tickDamage / this.iterations;
	}

	get resistedTickDamage() {
		return this.data.resistedTickDamage;
	}

	get avgResistedTickDamage() {
		return this.data.resistedTickDamage / this.iterations;
	}

	get critTickDamage() {
		return this.data.critTickDamage;
	}

	get avgCritTickDamage() {
		return this.data.critTickDamage / this.iterations;
	}

	get resistedCritTickDamage() {
		return this.data.resistedCritTickDamage;
	}

	get avgResistedCritTickDamage() {
		return this.data.resistedCritTickDamage / this.iterations;
	}

	get glanceDamage() {
		return this.data.glanceDamage;
	}

	get avgGlanceDamage() {
		return this.data.glanceDamage / this.iterations;
	}

	get crushDamage() {
		return this.data.crushDamage;
	}

	get avgCrushDamage() {
		return this.data.crushDamage / this.iterations;
	}

	get blockDamage() {
		return this.data.blockDamage;
	}

	get avgBlockDamage() {
		return this.data.blockDamage / this.iterations;
	}

	get blockedCritDamage() {
		return this.data.blockedCritDamage;
	}

	get avgBlockedCritDamage() {
		return this.data.blockedCritDamage / this.iterations;
	}

	get dps() {
		return this.data.damage / this.iterations / this.duration;
	}

	get threat() {
		return this.data.threat;
	}

	get avgThreat() {
		return this.data.threat / this.iterations;
	}

	get tps() {
		return this.data.threat / this.iterations / this.duration;
	}

	get healing() {
		return this.data.healing + this.data.shielding;
	}

	get avgHealing() {
		return (this.data.healing + this.data.shielding) / this.iterations;
	}

	get critHealing() {
		return this.data.critHealing;
	}

	get avgCritHealing() {
		return this.data.critHealing / this.iterations;
	}

	get shielding() {
		return this.data.shielding;
	}

	get hps() {
		return (this.data.healing + this.data.shielding) / this.iterations / this.duration;
	}

	get casts() {
		return this.data.casts / this.iterations;
	}

	get castsPerMinute() {
		return this.casts / (this.duration / 60);
	}

	get manualCasts() {
		return this.data.manualCasts / this.iterations;
	}

	get manualCastsPerMinute() {
		return this.manualCasts / (this.duration / 60);
	}

	get avgCastTimeMs() {
		return this.data.castTimeMs / this.iterations / this.casts;
	}

	get damageThroughput() {
		if (this.avgCastTimeMs) {
			return this.avgCast / (this.avgCastTimeMs / 1000);
		} else {
			return 0;
		}
	}

	get threatThroughput() {
		if (this.avgCastTimeMs) {
			return Math.max(this.avgHitThreat, this.avgCastThreat) / (this.avgCastTimeMs / 1000);
		} else {
			return 0;
		}
	}

	get healingThroughput() {
		if (this.avgCastTimeMs) {
			return this.hps / (this.avgCastTimeMs / 1000);
		} else {
			return 0;
		}
	}

	get timeSpentCastingMs() {
		return this.data.castTimeMs / this.iterations;
	}

	get avgCast() {
		if (!this.casts) return 0;
		return this.data.damage / this.iterations / (this.casts || 1);
	}

	get avgCastTick() {
		return this.data.tickDamage / this.iterations / (this.casts || 1);
	}

	get avgCastHealing() {
		return (this.data.healing + this.data.shielding) / this.iterations / (this.casts || 1);
	}

	get avgCastThreat() {
		return this.data.threat / this.iterations / (this.casts || 1);
	}

	get landedHits() {
		return this.landedHitsRaw / this.iterations;
	}

	get landedTicks() {
		return this.landedTicksRaw / this.iterations;
	}

	get landedRolls() {
		return this.landedRollsRaw / this.iterations;
	}

	get avgHit() {
		const lhr = this.landedHitsRaw;
		return lhr == 0 ? 0 : (this.data.damage - this.data.tickDamage) / lhr;
	}

	get avgTick() {
		const ltr = this.landedTicksRaw;
		return ltr == 0 ? 0 : this.data.tickDamage / ltr;
	}

	get avgHitHealing() {
		return (this.data.healing + this.data.shielding) / this.iterations / this.landedHits;
	}

	// The per-value range of an outcome, absent when the action landed none.
	range(kind: HitKind): DistributionMetricsProto | undefined {
		return this.data.ranges[kind];
	}

	get customMetrics() {
		return this.data.customMetrics;
	}

	get avgHitThreat() {
		const lhr = this.landedHitsRaw;
		return lhr == 0 ? 0 : this.data.threat / lhr;
	}

	get totalMisses() {
		return this.misses + this.dodges + this.parries;
	}

	get totalMissesPercent() {
		return this.missPercent + this.dodgePercent + this.parryPercent;
	}

	get misses() {
		return (this.data.misses + this.data.tickMisses) / this.iterations;
	}

	get missPercent() {
		return percentOf(this.data.misses + this.data.tickMisses, this.rollAttempts);
	}

	get dodges() {
		return this.data.dodges / this.iterations;
	}

	get dodgePercent() {
		return percentOf(this.data.dodges, this.rollAttempts);
	}

	get parries() {
		return this.data.parries / this.iterations;
	}

	get parryPercent() {
		return percentOf(this.data.parries, this.rollAttempts);
	}

	get hits() {
		return this.data.hits / this.iterations;
	}

	get hitPercent() {
		return percentOf(this.data.hits, this.hitAttempts);
	}

	get resistedHits() {
		return this.data.resistedHits / this.iterations;
	}

	get resistedHitPercent() {
		return percentOf(this.data.resistedHits, this.hitAttempts);
	}

	get ticks() {
		return this.data.ticks / this.iterations;
	}

	get resistedTicks() {
		return this.data.resistedTicks / this.iterations;
	}

	get resistedTickPercent() {
		return percentOf(this.data.resistedTicks, this.data.ticks + this.data.critTicks);
	}

	get critTicks() {
		return this.data.critTicks / this.iterations;
	}

	get critTickPercent() {
		return percentOf(this.data.critTicks, this.data.ticks + this.data.critTicks);
	}

	get resistedCritTicks() {
		return this.data.resistedCritTicks / this.iterations;
	}

	get resistedCritTickPercent() {
		return percentOf(this.data.resistedCritTicks, this.data.ticks + this.data.critTicks);
	}

	get blocks() {
		return this.data.blocks / this.iterations;
	}

	get crushes() {
		return this.data.crushes / this.iterations;
	}

	get crushPercent() {
		return percentOf(this.data.crushes, this.hitAttempts);
	}

	get blockPercent() {
		return percentOf(this.data.blocks, this.hitAttempts);
	}

	get blockedCrits() {
		return this.data.blockedCrits / this.iterations;
	}

	get blockedCritPercent() {
		return percentOf(this.data.blockedCrits, this.hitAttempts);
	}

	get glances() {
		return this.data.glances / this.iterations;
	}

	get glancePercent() {
		return percentOf(this.data.glances, this.hitAttempts);
	}

	get crits() {
		return this.data.crits / this.iterations;
	}

	get critPercent() {
		return percentOf(this.data.crits, this.hitAttempts);
	}

	get resistedCrits() {
		return this.data.resistedCrits / this.iterations;
	}

	get resistedCritPercent() {
		return percentOf(this.data.resistedCrits, this.hitAttempts);
	}

	get healingPercent() {
		return percentOf(this.healing - this.critHealing, this.healing);
	}

	get healingCritPercent() {
		return percentOf(this.data.critHealing, this.healing);
	}

	// Merges an array of metrics into a single metric.
	static merge(actions: Array<TargetedActionMetrics>): TargetedActionMetrics {
		const { iterations = 1, duration = 1 } = actions[0];

		return new TargetedActionMetrics(
			TargetedActionMetricsProto.create({
				casts: sum(actions.map(a => a.data.casts)),
				manualCasts: sum(actions.map(a => a.data.manualCasts)),
				hits: sum(actions.map(a => a.data.hits)),
				resistedHits: sum(actions.map(a => a.data.resistedHits)),
				crits: sum(actions.map(a => a.data.crits)),
				resistedCrits: sum(actions.map(a => a.data.resistedCrits)),
				ticks: sum(actions.map(a => a.data.ticks)),
				resistedTicks: sum(actions.map(a => a.data.resistedTicks)),
				critTicks: sum(actions.map(a => a.data.critTicks)),
				resistedCritTicks: sum(actions.map(a => a.data.resistedCritTicks)),
				misses: sum(actions.map(a => a.data.misses)),
				tickMisses: sum(actions.map(a => a.data.tickMisses)),
				dodges: sum(actions.map(a => a.data.dodges)),
				parries: sum(actions.map(a => a.data.parries)),
				blocks: sum(actions.map(a => a.data.blocks)),
				crushes: sum(actions.map(a => a.data.crushes)),
				blockedCrits: sum(actions.map(a => a.data.blockedCrits)),
				glances: sum(actions.map(a => a.data.glances)),
				damage: sum(actions.map(a => a.data.damage)),
				resistedDamage: sum(actions.map(a => a.data.resistedDamage)),
				critDamage: sum(actions.map(a => a.data.critDamage)),
				resistedCritDamage: sum(actions.map(a => a.data.resistedCritDamage)),
				tickDamage: sum(actions.map(a => a.data.tickDamage)),
				resistedTickDamage: sum(actions.map(a => a.data.resistedTickDamage)),
				critTickDamage: sum(actions.map(a => a.data.critTickDamage)),
				resistedCritTickDamage: sum(actions.map(a => a.data.resistedCritTickDamage)),
				glanceDamage: sum(actions.map(a => a.data.glanceDamage)),
				blockDamage: sum(actions.map(a => a.data.blockDamage)),
				blockedCritDamage: sum(actions.map(a => a.data.blockedCritDamage)),
				crushDamage: sum(actions.map(a => a.data.crushDamage)),
				threat: sum(actions.map(a => a.data.threat)),
				healing: sum(actions.map(a => a.data.healing)),
				critHealing: sum(actions.map(a => a.data.critHealing)),
				shielding: sum(actions.map(a => a.data.shielding)),
				castTimeMs: sum(actions.map(a => a.data.castTimeMs)),
				ranges: mergeRangeMaps(actions.map(a => a.data.ranges)),
				customMetrics: mergeCustomMetrics(actions.map(a => a.data.customMetrics)),
			}),
			{
				iterations,
				duration,
			},
		);
	}
}
