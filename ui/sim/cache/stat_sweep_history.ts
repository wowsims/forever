import type { Spec } from '@generated/proto/common';
import { type DBSchema, type IDBPDatabase, openDB } from 'idb';

import { LOCAL_STORAGE_PREFIX } from '../constants/other';
import type { StatSweepMetric } from '../state/sim_store';

const STAT_SWEEP_DB_NAME = `${LOCAL_STORAGE_PREFIX}_stat-sweep`;
// Every upgrade drops the stored runs, so a change to the record shape never meets runs saved in the old one.
const STAT_SWEEP_DB_VERSION = 3;
const RUNS_STORE = 'runs';

export interface MetricSummary {
	avg: number;
	stdev: number;
	/** Iterations behind `avg`. */
	n: number;
}

export interface StatSweepPoint {
	/** The step's change in rating for a % stat, whose % that rating gives (10 hit rating = 1% hit); bonus points for any other stat. Every stat shares this axis. */
	offset: number;
	/** The bonus the step adds to the stat in its own units: `offset`, or for a % stat the % that rating gives. */
	bonus: number;
	/** The swept stat's final total with the step's bonus, as the sim computes it. */
	statValue: number;
	metrics: Record<StatSweepMetric, MetricSummary>;
}

export interface StatSweepRecord {
	id: string;
	/** Epoch milliseconds. */
	createdAt: number;
	specId: Spec;
	label: string;
	/** The chart colour slot the run keeps across reloads. */
	colorSlot: number;
	/** `UnitStat.toJson()`: IndexedDB's structured clone drops class prototypes, so a stored UnitStat would come back without its methods. */
	unitStat: object;
	/** The metric plotted when the run started, which Target ± sizes the steps by. */
	metric: StatSweepMetric;
	baseStatValue: number;
	points: StatSweepPoint[];
	/** The step past which more of the stat left every metric unchanged, so the run skipped the steps above the next one. */
	capped?: Pick<StatSweepPoint, 'offset' | 'bonus'>;
	/** Shared by every run one sweep saved; a run without one is a batch of its own. */
	batchId?: string;
	/** The sweep's stats in run order with its range, the same on every run of the batch. */
	batchLabel?: string;
}

interface StatSweepDb extends DBSchema {
	[RUNS_STORE]: {
		key: string;
		value: StatSweepRecord;
		indexes: { specId: number };
	};
}

// A fresh connection per call, closed when it is done, so a version upgrade in another tab is never blocked.
const openStatSweepDb = (): Promise<IDBPDatabase<StatSweepDb>> =>
	openDB<StatSweepDb>(STAT_SWEEP_DB_NAME, STAT_SWEEP_DB_VERSION, {
		upgrade(db) {
			if (db.objectStoreNames.contains(RUNS_STORE)) db.deleteObjectStore(RUNS_STORE);
			const store = db.createObjectStore(RUNS_STORE, { keyPath: 'id' });
			store.createIndex('specId', 'specId');
		},
	});

export async function saveRun(run: StatSweepRecord): Promise<void> {
	const db = await openStatSweepDb();
	try {
		await db.put(RUNS_STORE, run);
	} finally {
		db.close();
	}
}

/** The spec's saved runs, newest first. */
export async function listRuns(specId: Spec): Promise<StatSweepRecord[]> {
	const db = await openStatSweepDb();
	try {
		const runs = await db.getAllFromIndex(RUNS_STORE, 'specId', specId);
		return runs.sort((a, b) => b.createdAt - a.createdAt);
	} finally {
		db.close();
	}
}

/** Either every run goes or none does. */
export async function deleteRuns(ids: readonly string[]): Promise<void> {
	const db = await openStatSweepDb();
	try {
		const tx = db.transaction(RUNS_STORE, 'readwrite');
		await Promise.all([...ids.map(id => tx.store.delete(id)), tx.done]);
	} finally {
		db.close();
	}
}
