import { getGearIdentityKey } from '@sim/proto/items';
import { subscribeGated } from '@sim/state/batch';
import { Emitter } from '@sim/state/events';
import { PLAYER_FIELDS, type PlayerField, SimRunKind, type SimState, type SimStore } from '@sim/state/sim_store';
import { shallowArrayEquals } from '@sim/state/subscriptions';

export type ResultFreshness = 'fresh' | 'stale' | null;

// Player fields that never change a sim result: the name is a label, the EP fields feed stat
// weights, and `currentStats` is computed from the other fields and lands after a compute-stats
// round trip, so counting it would turn a result stale after the fact.
const NON_INPUT_FIELDS: ReadonlyArray<PlayerField> = ['name', 'epWeights', 'epRatios', 'epRefStat', 'currentStats'];
const PLAYER_INPUT_FIELDS = PLAYER_FIELDS.filter(field => !NON_INPUT_FIELDS.includes(field) && field !== 'gear');

// What an individual sim request is built from. The encounter and raid slices are replace-on-write,
// so their references stand in for their contents. Gear compares by content, with the key the batch
// sim identifies gear sets by, so a swap that is undone reads fresh again; the other player fields
// compare by version counter. The sim slice reaches the request only as the iteration count and the
// seed, which change the precision and the sample, not the answer.
export function resultInputsKey(s: SimState): Array<unknown> {
	const key: Array<unknown> = [s.encounter, s.raid];
	for (const [storeKey, player] of Object.entries(s.players)) {
		key.push(storeKey, getGearIdentityKey(player.gear.asSpec()), ...PLAYER_INPUT_FIELDS.map(field => player.v[field]));
	}
	return key;
}

export class ResultFreshnessTracker {
	private readonly resultChanged = new Emitter<void>();
	// Store states are immutable, so holding the state a run started from is its snapshot.
	private runStart: SimState | null = null;
	private resultInputs: Array<unknown> | null = null;
	private keyedState: SimState | null = null;
	private liveKey: Array<unknown> = [];

	constructor(private readonly store: SimStore) {
		// Raw and per write rather than gated or compared by value: a start writes the running flag
		// twice, the second time after the previous request is cleared and right before the request
		// is built, and that later write is the one whose state the run simulates.
		store.subscribe(
			s => s.runs[SimRunKind.IndividualSim],
			run => {
				if (run.isRunning && !run.isAborting) this.runStart = store.getState();
			},
		);
	}

	// The current result came from the latest individual run, so it is judged against that run's
	// start. A result set with no run behind it is judged against the inputs as they are.
	resultLanded() {
		this.resultInputs = resultInputsKey(this.runStart ?? this.store.getState());
		this.resultChanged.emit();
	}

	resultMatchesInputs() {
		this.resultInputs = this.key(this.store.getState());
		this.resultChanged.emit();
	}

	readonly subscribe = (listener: () => void): (() => void) => {
		const offInputs = subscribeGated(this.store.subscribe, s => this.key(s), listener, shallowArrayEquals);
		const offResult = this.resultChanged.on(listener);
		return () => {
			offInputs();
			offResult();
		};
	};

	readonly getSnapshot = (): ResultFreshness => {
		if (!this.resultInputs) return null;
		return shallowArrayEquals(this.resultInputs, this.key(this.store.getState())) ? 'fresh' : 'stale';
	};

	// Selectors run on every store write, so the live key is built once per state.
	private key(s: SimState): Array<unknown> {
		if (s !== this.keyedState) {
			this.keyedState = s;
			this.liveKey = resultInputsKey(s);
		}
		return this.liveKey;
	}
}

// One tracker per store, shared by the results manager (which feeds it) and the hook (which reads it).
const trackers = new WeakMap<SimStore, ResultFreshnessTracker>();

export function resultFreshnessTracker(store: SimStore): ResultFreshnessTracker {
	let tracker = trackers.get(store);
	if (!tracker) trackers.set(store, (tracker = new ResultFreshnessTracker(store)));
	return tracker;
}
