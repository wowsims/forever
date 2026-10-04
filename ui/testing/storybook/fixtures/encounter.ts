import { Target as TargetProto } from '@generated/proto/common';
import type { Encounter } from '@sim/raid/encounter';
import { createSimStore, patchSlice } from '@sim/state/sim_store';

/**
 * The `Encounter` members the target pickers read, on a real store: a write patches the
 * `encounter` slice, so the pickers subscribed to `encounter.targets` re-read and a story stays
 * editable. No preset targets, so the NPC and AI menus offer only their custom entries.
 */
export const makeStoryEncounter = (targets: Array<TargetProto>, { threat = true }: { threat?: boolean } = {}): Encounter => {
	const store = createSimStore();
	patchSlice(store, 'encounter', { targets });
	patchSlice(store, 'ui', { showThreatMetrics: threat });
	const read = () => store.getState().encounter.targets;
	const encounter = {
		sim: {
			store,
			db: { getAllPresetTargets: () => [] },
			getShowDamageMetrics: () => true,
			getShowHealingMetrics: () => false,
			getShowThreatMetrics: () => store.getState().ui.showThreatMetrics,
		},
		getTargets: read,
		getTarget: (index: number) => read()[index],
		setTargets: (next: Array<TargetProto>) => patchSlice(store, 'encounter', { targets: next }),
		modifyTarget: (index: number, apply: (target: TargetProto) => void) => {
			const draft = TargetProto.clone(read()[index]);
			apply(draft);
			patchSlice(store, 'encounter', { targets: read().map((target, at) => (at === index ? draft : target)) });
		},
		applyPresetTarget: () => {},
	};
	return encounter as unknown as Encounter;
};
