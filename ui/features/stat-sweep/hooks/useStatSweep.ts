import { deleteRuns, listRuns, type StatSweepRecord } from '@sim/cache/stat_sweep_history';
import { usePlayer, useSim } from '@sim/context/SimHostContext';
import { useSimRun } from '@sim/hooks/useSimRun';
import { UnitStat } from '@sim/proto/stats';
import { SimRunKind } from '@sim/state/sim_store';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

import { runLabel, type StatSweepSeries } from '../components/StatSweepChart';
import type { StatSweepHistoryEntry } from '../components/StatSweepHistory';
import { SERIES_SLOT_COUNT } from '../model/colors';
import { runStatSweep, type StatSweepArgs } from '../model/run';
import {
	displayColorSlots,
	mergeSweepHistory,
	nextColorSlot,
	type StatSweepCsvRun,
	type StatSweepProgress,
	type SweepBatch,
	sweepBatches,
} from '../model/stat_sweep';

interface SweepView {
	history: StatSweepRecord[];
	/** The run in progress, drawn as its points arrive. */
	live: StatSweepRecord | null;
	overlaid: ReadonlyArray<string>;
}

/** The stat sweep's runs and history while its dialog is open; a run still going when it unmounts is cancelled. */
export const useStatSweep = () => {
	const sim = useSim();
	const player = usePlayer();
	const { isRunning, abort } = useSimRun(SimRunKind.StatSweep);
	const [view, setView] = useState<SweepView>({ history: [], live: null, overlaid: [] });
	const history = useRef(view.history);
	history.current = view.history;

	useEffect(() => {
		let active = true;
		listRuns(player.getSpec())
			.then(stored => {
				if (!active) return;
				setView(current => {
					const merged = mergeSweepHistory(current.history, stored);
					// With no run in progress and none ticked yet, the newest stored sweep is drawn.
					const overlaid =
						current.live || current.overlaid.length || !merged.length ? current.overlaid : sweepBatches(merged)[0].runs.map(run => run.id);
					return { ...current, history: merged, overlaid };
				});
			})
			.catch(console.error);
		return () => {
			active = false;
		};
	}, [player]);

	// A sweep runs only while its dialog is mounted, so one does not keep simulating after power tools are turned off.
	useEffect(
		() => () => {
			sim.runs.abort(SimRunKind.StatSweep).catch(console.error);
		},
		[sim],
	);

	const run = useCallback(
		(args: Omit<StatSweepArgs, 'firstColorSlot'>, onProgress: (progress: StatSweepProgress) => void) => {
			// Only this batch's runs are drawn; the user may tick older ones again afterwards.
			setView(current => ({ ...current, overlaid: [] }));
			return runStatSweep(
				{ sim, specId: player.getSpec() },
				// The history read here holds none of this run's stats, so each stat takes the slot after the one before.
				{ ...args, firstColorSlot: nextColorSlot(history.current, SERIES_SLOT_COUNT) },
				{
					onProgress,
					onRunUpdate: live =>
						setView(current => ({
							...current,
							live,
							overlaid: current.overlaid.includes(live.id) ? current.overlaid : [...current.overlaid, live.id],
						})),
					onRunEnd: finished =>
						setView(current =>
							finished.points.length
								? { ...current, live: null, history: [finished, ...current.history] }
								: { ...current, live: null, overlaid: current.overlaid.filter(id => id !== finished.id) },
						),
				},
			);
		},
		[sim, player],
	);

	const cancel = useCallback(() => {
		abort().catch(console.error);
	}, [abort]);

	const remove = useCallback((ids: readonly string[]) => {
		setView(current => ({
			...current,
			history: current.history.filter(run => !ids.includes(run.id)),
			overlaid: current.overlaid.filter(other => !ids.includes(other)),
		}));
		deleteRuns(ids).catch(console.error);
	}, []);

	const setOverlaid = useCallback(
		(ids: readonly string[], overlaid: boolean) =>
			setView(current => {
				const others = current.overlaid.filter(other => !ids.includes(other));
				return { ...current, overlaid: overlaid ? [...others, ...ids] : others };
			}),
		[],
	);

	// The slot of each shown run, oldest first so a clash moves the newer run. A live run keeps its slot from its start, so
	// its points arriving leave this alone.
	const liveId = view.live?.id;
	const liveColorSlot = view.live?.colorSlot;
	const colorSlots = useMemo(() => {
		const shown = [
			...(liveId !== undefined && liveColorSlot !== undefined ? [{ id: liveId, colorSlot: liveColorSlot }] : []),
			...view.history.filter(run => view.overlaid.includes(run.id)),
		];
		return displayColorSlots(shown.reverse(), SERIES_SLOT_COUNT);
	}, [view.history, view.overlaid, liveId, liveColorSlot]);

	const statLabel = useCallback((run: StatSweepRecord) => UnitStat.fromJson(run.unitStat).getFullName(player.getClass()), [player]);
	const series = useMemo<Array<StatSweepSeries & StatSweepCsvRun>>(
		() =>
			[...(view.live ? [view.live] : []), ...view.history.filter(run => view.overlaid.includes(run.id))].map(run => ({
				id: run.id,
				label: run.label,
				batchLabel: run.batchLabel ?? '',
				colorSlot: colorSlots[run.id],
				statLabel: statLabel(run),
				baseStatValue: run.baseStatValue,
				points: run.points,
				capped: run.capped,
			})),
		[view, colorSlots, statLabel],
	);
	const savedBatches = useMemo<SweepBatch<Omit<StatSweepHistoryEntry, 'overlaid' | 'colorSlot'>>[]>(
		() =>
			sweepBatches(view.history).map(({ id, runs }) => ({
				id,
				runs: runs.map(run => ({
					id: run.id,
					label: runLabel(run),
					statLabel: statLabel(run),
					createdAt: run.createdAt,
					metric: run.metric,
				})),
			})),
		[view.history, statLabel],
	);
	const historyBatches = useMemo<SweepBatch<StatSweepHistoryEntry>[]>(() => {
		const overlaid = new Set(view.overlaid);
		return savedBatches.map(({ id, runs }) => ({
			id,
			runs: runs.map(run => ({ ...run, overlaid: overlaid.has(run.id), colorSlot: colorSlots[run.id] })),
		}));
	}, [savedBatches, view.overlaid, colorSlots]);

	return { isRunning, series, historyBatches, run, cancel, remove, setOverlaid };
};

export type StatSweep = ReturnType<typeof useStatSweep>;
