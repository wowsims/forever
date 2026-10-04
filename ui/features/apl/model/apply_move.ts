import { carryId, hasId, keyFor, moveItem } from '@features/inputs/list_ids';
import { type APLAction, APLListItem, APLPrepullAction, type APLRotation, type APLValue } from '@generated/proto/apl';
import { randomUUID } from '@sim/utils/misc';

/** The priority pane's tree: two section items, `PREPULL` and `PRIORITY`, each holding its rows. */
export const APL_ROWS = 'apl';
export const PREPULL = 'prepull';
export const PRIORITY = 'priority';
/** The action groups pane's tree: the groups at the top, each group's actions under it. */
export const APL_GROUPS = 'apl-groups';
export const APL_VARIABLES = 'apl-variables';

const VALUES = 'apl-values:';
const ACTIONS = 'apl-actions:';

/** A drag's source and drop target, as `Tree` reports them. Structurally a `TreeMove`. */
export interface AplMove {
	from: { group: string; id: string; index: number; parentId: string | null };
	/** `index` is read against the target list before the dragged item leaves its own. */
	to: { group: string; parentId: string | null; index: number };
}

/** Whether a list offers an action kind: the pre-pull and priority lists offer different sets. */
export type OffersKind = (kind: string, isPrepull: boolean) => boolean;

/** The `Tree.Item` id of a list entry: its list key, which survives reorders and clone replacement. */
export const aplItemId = (item: unknown, index: number): string => String(keyFor(item, index));

/** The group that names the operand list of `owner`, an `and`, `or`, `min` or `max` value. */
export const valueListGroup = (owner: APLValue): string => `${VALUES}${owner.uuid?.value ?? ''}`;

/** The group that names the action list of `owner`, a sequence or strict sequence. */
export const actionListGroup = (owner: APLAction): string => `${ACTIONS}${keyFor(owner, 0)}`;

interface Slot {
	/** Lists of one family exchange items; `prepull` and `priority` convert theirs on the way. */
	family: string;
	list: Array<any>;
	/** The list its owner holds now, which is `list` for as long as the slot is current. */
	read: () => Array<any> | undefined;
	write: (next: Array<any>) => void;
	owner?: object;
}

/** A drag's source list, resolved once per drag; `canApplyMove` takes it so a dragover walks the rotation for the target only. */
export type MoveSource = Slot;

const slot = (family: string, read: () => Array<any> | undefined, write: (next: Array<any>) => void, owner?: object): Slot | undefined => {
	const list = read();
	return list ? { family, list, read, write, owner } : undefined;
};

const find = (node: unknown, match: (candidate: any) => boolean): any => {
	if (!node || typeof node !== 'object') return undefined;
	if (match(node)) return node;
	for (const child of Object.values(node)) {
		const found = find(child, match);
		if (found) return found;
	}
	return undefined;
};

const collect = (node: unknown, match: (candidate: any) => boolean, found: Array<any> = []): Array<any> => {
	if (!node || typeof node !== 'object') return found;
	if (match(node)) found.push(node);
	for (const child of Object.values(node)) collect(child, match, found);
	return found;
};

const contains = (node: unknown, target: object): boolean => node === target || find(node, candidate => candidate === target) !== undefined;

const implOf = (message: { oneofKind?: string } | undefined): any => (message?.oneofKind ? (message as any)[message.oneofKind] : undefined);

const slotAt = (rotation: APLRotation, group: string, parentId: string | null): Slot | undefined => {
	if (group === APL_ROWS && parentId === PREPULL)
		return slot(
			PREPULL,
			() => rotation.prepullActions,
			next => (rotation.prepullActions = next),
		);
	if (group === APL_ROWS && parentId === PRIORITY)
		return slot(
			PRIORITY,
			() => rotation.priorityList,
			next => (rotation.priorityList = next),
		);
	if (group === APL_VARIABLES && parentId === null)
		return slot(
			APL_VARIABLES,
			() => (rotation.valueVariables ??= []),
			next => (rotation.valueVariables = next),
		);
	if (group === APL_GROUPS && parentId === null)
		return slot(
			APL_GROUPS,
			() => (rotation.groups ??= []),
			next => (rotation.groups = next),
		);
	if (group === APL_GROUPS) {
		const owner = (rotation.groups || []).find((candidate, index) => hasId(candidate) && aplItemId(candidate, index) === parentId);
		return (
			owner &&
			slot(
				'group-actions',
				() => owner.actions,
				next => (owner.actions = next),
				owner,
			)
		);
	}
	if (group.startsWith(VALUES)) {
		const uuid = group.slice(VALUES.length);
		const owners: Array<APLValue> = uuid ? collect(rotation, candidate => candidate.uuid?.value === uuid && 'value' in candidate) : [];
		// Two values sharing a uuid name no list at all, rather than whichever comes first.
		const owner = owners.length === 1 ? owners[0] : undefined;
		if (!owner) return undefined;
		const vals = () => {
			const impl = implOf(owner.value);
			return Array.isArray(impl?.vals) ? impl.vals : undefined;
		};
		return slot('values', vals, next => (implOf(owner.value).vals = next), owner);
	}
	if (group.startsWith(ACTIONS)) {
		const key = group.slice(ACTIONS.length);
		const owner: APLAction | undefined = find(rotation, candidate => hasId(candidate) && String(keyFor(candidate, 0)) === key && 'action' in candidate);
		if (!owner) return undefined;
		const actions = () => {
			const impl = implOf(owner.action);
			return Array.isArray(impl?.actions) ? impl.actions : undefined;
		};
		return slot('actions', actions, next => (implOf(owner.action).actions = next), owner);
	}
	return undefined;
};

const ROW_FAMILIES = new Set([PREPULL, PRIORITY]);

const compatible = (from: Slot, to: Slot) => from.family === to.family || (ROW_FAMILIES.has(from.family) && ROW_FAMILIES.has(to.family));

const AT_PULL = '-1s';

const convert = (item: any, from: Slot, to: Slot): any => {
	if (from.family === to.family) return item;
	const next =
		to.family === PRIORITY
			? APLListItem.create({ hide: item.hide })
			: APLPrepullAction.create({
					hide: item.hide,
					doAtValue: { uuid: { value: randomUUID() }, value: { oneofKind: 'const', const: { val: AT_PULL } } },
				});
	next.action = item.action;
	carryId(item, next);
	return next;
};

const atPull = (doAt: APLValue | undefined) => {
	const value = doAt?.value;
	return !value?.oneofKind || (value.oneofKind === 'const' && value.const.val.trim() === AT_PULL);
};

// A crossing converts the row, and each list keeps a field the other has no room for: the pre-pull list
// shows no condition (the sim still reads one), the priority list has no time. A row carrying either stays.
const losesData = (item: any, to: Slot) =>
	to.family === PREPULL ? !!item?.action?.condition?.value?.oneofKind : to.family === PRIORITY && !atPull(item?.doAtValue);

const matches = (item: unknown, index: number, id: string) =>
	(hasId(item) && aplItemId(item, index) === id) || (item as APLValue | undefined)?.uuid?.value === id;

const plan = (rotation: APLRotation, { from, to }: AplMove, offersKind?: OffersKind, resolved?: MoveSource): (() => void) | null => {
	const source = resolved && resolved.read() === resolved.list ? resolved : slotAt(rotation, from.group, from.parentId);
	const target = slotAt(rotation, to.group, to.parentId);
	if (!source || !target || !compatible(source, target)) return null;
	const item = source.list[from.index];
	if (!matches(item, from.index, from.id)) return null;
	if (to.index < 0 || to.index > target.list.length) return null;

	if (source.list === target.list) {
		if (to.index === from.index || to.index === from.index + 1) return null;
		return () => source.write(moveItem(source.list, from.index, to.index));
	}

	if (target.owner && contains(item, target.owner)) return null;
	const kind = item?.action?.action?.oneofKind;
	if (source.family !== target.family && kind && offersKind && !offersKind(kind, target.family === PREPULL)) return null;
	if (source.family !== target.family && losesData(item, target)) return null;
	return () => {
		const remaining = source.list.slice();
		remaining.splice(from.index, 1);
		const landed = target.list.slice();
		landed.splice(to.index, 0, convert(item, source, target));
		source.write(remaining);
		target.write(landed);
	};
};

/**
 * Whether `move` names two lists of this rotation that can exchange the item: the id still names
 * the item at its index, the lists are of one kind, the target is not inside the item, and a row
 * crossing into the pre-pull list or out of it has a kind the other list offers and carries no
 * condition (into pre-pull) or time other than the pull (out of it). A `source` from
 * `resolveMoveSource` is used while its list is still the one its owner holds.
 */
export const canApplyMove = (rotation: APLRotation, move: AplMove, offersKind?: OffersKind, source?: MoveSource): boolean =>
	plan(rotation, move, offersKind, source) !== null;

/** The source list of a drag, for `canApplyMove` to reuse on every dragover of that drag. */
export const resolveMoveSource = (rotation: APLRotation, from: AplMove['from']): MoveSource | undefined => slotAt(rotation, from.group, from.parentId);

/**
 * Applies a drop to the live rotation. Each changed list is replaced by a fresh array on its owner,
 * every item keeps its identity, and a move `canApplyMove` refuses writes nothing. Called inside
 * `player.modifyAplRotation`, so a move notifies once.
 */
export const applyMove = (rotation: APLRotation, move: AplMove, offersKind?: OffersKind): APLRotation => {
	plan(rotation, move, offersKind)?.();
	return rotation;
};
