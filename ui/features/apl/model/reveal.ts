import type { APLAction, APLListItem, APLPrepullAction, APLRotation } from '@generated/proto/apl';
import type { AplEntryPath } from '@sim/proto/combat_log';
import { Emitter } from '@sim/state/events';

import { resolveAplEntry, type ResolvedAplEntry } from './apl_entry';

/** A cast's entry to open in the Rotation tab: its row and the steps down to the one that fired, as objects of the live rotation. */
export interface AplRevealRequest {
	readonly kind: ResolvedAplEntry['kind'];
	readonly item: APLListItem | APLPrepullAction;
	readonly steps: ReadonlyArray<APLAction>;
}

// One page holds one sim, so the Timeline and the Rotation tab share the module's one emitter.
export const aplReveals = new Emitter<AplRevealRequest>();

/**
 * Sends the entry `path` names in `rotation`, the player's own, to the Rotation tab, whose rows find themselves
 * in it by identity. False, with nothing sent, when the path no longer fits the rotation.
 */
export function requestAplReveal(rotation: APLRotation, path: AplEntryPath): boolean {
	const entry = resolveAplEntry(rotation, path);
	if (!entry) return false;
	aplReveals.emit({ kind: entry.kind, item: entry.item, steps: entry.stepActions });
	return true;
}
