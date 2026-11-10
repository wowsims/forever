import {
	ActionMetrics,
	DistributionMetrics,
	EncounterMetrics,
	HitKind,
	Party,
	PartyMetrics,
	Player,
	Raid,
	RaidMetrics,
	RaidSimRequest,
	RaidSimResult,
	SimType,
	TargetedActionMetrics,
	UnitMetrics,
} from '@generated/proto/api';
import { ActionID, Class, Encounter, ItemSlot, OtherAction, Race, Target } from '@generated/proto/common';
import { DpsWarrior } from '@generated/proto/warrior';
import { SimResult } from '@sim/proto/sim_result';

const ITERATIONS = 100;
const DURATION = 180;
const PLAYER_UNIT = 0;
const TARGET_UNIT = 1;

const perSecond = (total: number) => total / ITERATIONS / DURATION;

// A bell over five buckets, two standard deviations either side, summing to the iteration count.
const distribution = (avg: number) => {
	const stdev = avg * 0.05;
	const counts = [5, 24, 42, 24, 5];
	return DistributionMetrics.create({
		avg,
		stdev,
		min: avg - 2 * stdev,
		max: avg + 2 * stdev,
		hist: Object.fromEntries(counts.map((count, step) => [Math.round(avg + (step - 2) * stdev), count])),
	});
};

export const valueRange = (n: number, avg: number, stdev: number, min: number, max: number) =>
	DistributionMetrics.create({ avg, stdev, min, max, aggregatorData: { n, sumSq: 0 } });

export interface SimRunOptions {
	// Multiplies every damage and threat figure, so a second run reads as a change against the first.
	dpsScale?: number;
}

let runs = 0;

/**
 * One warrior against one level 63 target: 100 iterations of 180 seconds, whose only action is the
 * main-hand auto attack. That action's id is an `OtherAction`, which needs no item database to name,
 * and there are no combat logs.
 */
export const makeSimRun = ({ dpsScale = 1 }: SimRunOptions = {}): { request: RaidSimRequest; result: RaidSimResult } => {
	const damage = 5_855_400 * dpsScale;
	const threat = damage * 0.8;

	const player = Player.create({
		name: 'Player',
		race: Race.RaceHuman,
		class: Class.ClassWarrior,
		spec: { oneofKind: 'dpsWarrior', dpsWarrior: DpsWarrior.create() },
	});
	const target = Target.create({ id: 31146, name: 'Raid Target', level: 63 });

	const mainHand = ActionMetrics.create({
		id: ActionID.create({ rawId: { oneofKind: 'otherId', otherId: OtherAction.OtherActionAttack }, tag: ItemSlot.ItemSlotMainHand }),
		isMelee: true,
		targets: [
			TargetedActionMetrics.create({
				unitIndex: TARGET_UNIT,
				casts: 7_000,
				hits: 3_800,
				crits: 1_700,
				misses: 600,
				dodges: 400,
				glances: 500,
				damage,
				critDamage: 2_628_200 * dpsScale,
				glanceDamage: 289_875 * dpsScale,
				threat,
				ranges: {
					[HitKind.HitKindHit]: valueRange(3_800, 773, 40, 650, 900),
					[HitKind.HitKindCrit]: valueRange(1_700, 1_546, 80, 1_300, 1_800),
				},
				customMetrics: [{ key: 'divine_aegis', range: valueRange(200, 410, 50, 300, 520) }],
			}),
		],
	});

	const playerMetrics = UnitMetrics.create({
		name: player.name,
		unitIndex: PLAYER_UNIT,
		dps: distribution(perSecond(damage)),
		threat: distribution(perSecond(threat)),
		dtps: DistributionMetrics.create(),
		tmi: DistributionMetrics.create(),
		hps: DistributionMetrics.create(),
		tto: DistributionMetrics.create(),
		actions: [mainHand],
	});
	const targetMetrics = UnitMetrics.create({
		name: target.name,
		unitIndex: TARGET_UNIT,
		dps: DistributionMetrics.create(),
		threat: DistributionMetrics.create(),
		dtps: DistributionMetrics.create(),
		tmi: DistributionMetrics.create(),
		hps: DistributionMetrics.create(),
		tto: DistributionMetrics.create(),
	});

	const request = RaidSimRequest.create({
		requestId: `storybook-sim-run-${++runs}`,
		type: SimType.SimTypeIndividual,
		raid: Raid.create({ parties: [Party.create({ players: [player] })] }),
		encounter: Encounter.create({ duration: DURATION, targets: [target] }),
		simOptions: { iterations: ITERATIONS },
	});
	const result = RaidSimResult.create({
		raidMetrics: RaidMetrics.create({
			dps: playerMetrics.dps,
			hps: playerMetrics.hps,
			parties: [PartyMetrics.create({ dps: playerMetrics.dps, hps: playerMetrics.hps, players: [playerMetrics] })],
		}),
		encounterMetrics: EncounterMetrics.create({ targets: [targetMetrics] }),
		firstIterationDuration: DURATION,
		avgIterationDuration: DURATION,
		iterationsDone: ITERATIONS,
	});
	return { request, result };
};

export const makeSimResult = (options: SimRunOptions = {}): Promise<SimResult> => {
	const { request, result } = makeSimRun(options);
	return SimResult.makeNew(request, result);
};
