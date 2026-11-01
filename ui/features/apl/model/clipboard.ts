import { APLGroup, APLListItem, APLValueVariable } from '@generated/proto/apl';
import type { IMessageType } from '@protobuf-ts/runtime';

import { freshUuids } from './fresh_uuids';

export interface AplClipItems {
	action: APLListItem;
	group: APLGroup;
	variable: APLValueVariable;
}

export type AplClipKind = keyof AplClipItems;
export type AplClip = { [K in AplClipKind]: { kind: K; item: AplClipItems[K] } }[AplClipKind];

const MESSAGES: Record<AplClipKind, IMessageType<any>> = { action: APLListItem, group: APLGroup, variable: APLValueVariable };

/** Clipboard text: `{ "kind": ..., "value": <the item's proto JSON> }`, so a paste can tell an action from a group or a variable. */
export const aplClipToText = ({ kind, item }: AplClip): string => JSON.stringify({ kind, value: MESSAGES[kind].toJson(item) }, null, 2);

/**
 * The item the text holds, with new uuids, or `undefined` when the text is not a copied action, group or variable.
 * Fields this build does not know are dropped, so text copied on another build of the sim still pastes.
 */
export const aplClipFromText = (text: string): AplClip | undefined => {
	try {
		const { kind, value } = JSON.parse(text);
		if (!Object.prototype.hasOwnProperty.call(MESSAGES, kind)) return undefined;
		const item = MESSAGES[kind as AplClipKind].fromJson(value, { ignoreUnknownFields: true });
		// A row with no action chosen yet, as New Action makes, is still a row.
		if (!(kind === 'action' ? item.action : item.name)) return undefined;
		return { kind, item: freshUuids(item) } as AplClip;
	} catch {
		return undefined;
	}
};
