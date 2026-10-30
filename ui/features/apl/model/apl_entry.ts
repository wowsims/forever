import type { APLAction, APLListItem, APLPrepullAction, APLRotation } from '@generated/proto/apl';
import type { AplEntryPath, AplEntrySegment } from '@sim/proto/combat_log';
import type { SimResult, UnitMetrics } from '@sim/proto/sim_result';
import { MAX_PARTY_SIZE } from '@sim/raid/party';

/** The rotation a run ran, and whether the player's rotation differs from it now. */
export interface AplRun {
	readonly rotation: APLRotation;
	/**
	 * Asked when a tooltip shows or a cast is clicked, so an edit in the Rotation tab compares nothing. It reads the
	 * store unsubscribed, which holds because both happen in the Results tab, where the rotation cannot change.
	 */
	readonly changedSinceRun: () => boolean;
}

/** The rotation the run's request carried for `player`. An Auto or Simple run carries one too, but its casts name no entry path. */
export function runRotation(result: SimResult, player: UnitMetrics): APLRotation | undefined {
	return result.request.raid?.parties[Math.floor(player.index / MAX_PARTY_SIZE)]?.players[player.index % MAX_PARTY_SIZE]?.rotation;
}

export interface ResolvedAplEntry {
	readonly kind: Exclude<AplEntrySegment['kind'], 'step'>;
	readonly item: APLListItem | APLPrepullAction;
	/** The indexes of the steps after the row, outermost first. */
	readonly stepIndexes: ReadonlyArray<number>;
	/** The actions those steps name, in the same order. */
	readonly stepActions: ReadonlyArray<APLAction>;
	readonly crumbs: ReadonlyArray<AplEntrySegment & { readonly groupName?: string }>;
}

const stepsOf = (action: APLAction | undefined): APLAction[] | undefined => {
	const impl = action?.action;
	if (impl?.oneofKind === 'sequence') return impl.sequence.actions;
	if (impl?.oneofKind === 'strictSequence') return impl.strictSequence.actions;
	// A Schedule's inner action runs for the Schedule's own entry path, so the steps of a sequence inside it follow that entry path.
	if (impl?.oneofKind === 'schedule') return stepsOf(impl.schedule.innerAction);
	return undefined;
};

/**
 * The editor row a cast's path names, taken from the last segment that is not a step, with the steps after it.
 * Null when any segment no longer fits the rotation, a step included.
 */
export function resolveAplEntry(rotation: APLRotation, path: AplEntryPath): ResolvedAplEntry | null {
	let row: { item: APLListItem | APLPrepullAction; kind: ResolvedAplEntry['kind'] } | undefined;
	let action: APLAction | undefined;
	let stepIndexes: number[] = [];
	let stepActions: APLAction[] = [];
	for (const segment of path) {
		if (segment.kind === 'step') {
			const next = stepsOf(action)?.[segment.index];
			if (!next) return null;
			action = next;
			stepIndexes.push(segment.index);
			stepActions.push(next);
			continue;
		}
		const item =
			segment.kind === 'prepull'
				? rotation.prepullActions[segment.index]
				: segment.kind === 'group'
					? rotation.groups[segment.group]?.actions[segment.index]
					: rotation.priorityList[segment.index];
		if (!item) return null;
		row = { item, kind: segment.kind };
		action = item.action;
		stepIndexes = [];
		stepActions = [];
	}
	if (!row) return null;

	const crumbs = path.map(segment => (segment.kind === 'group' ? { ...segment, groupName: rotation.groups[segment.group].name } : segment));
	return { ...row, stepIndexes, stepActions, crumbs };
}
