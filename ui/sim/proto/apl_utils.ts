import { APLAction, APLPrepullAction, APLRotation, APLRotation_Type } from '@generated/proto/apl';
import { ActionID as ActionIdProto, Cooldowns, Spec } from '@generated/proto/common';

import { Player } from '../player/player';
import { simpleSpecRotationFromJson } from './spec_functions';

type APLRenameTarget =
	| { type: 'variable'; oldName: string; newName: string }
	| { type: 'placeholder'; oldName: string; newName: string }
	| { type: 'group'; oldName: string; newName: string };

/**
 * Recursively walks an APL rotation object and renames all references to a variable, placeholder, or group.
 * Matches on oneofKind discriminators (variableRef, variablePlaceholder, groupReference)
 * so it never needs updating when new action/value types are added to the proto.
 */
export function renameAPLReference(obj: unknown, target: APLRenameTarget): void {
	if (!obj || typeof obj !== 'object') return;

	if (Array.isArray(obj)) {
		for (const item of obj) renameAPLReference(item, target);
		return;
	}

	const record = obj as Record<string, unknown>;

	if (target.type === 'variable') {
		if (record.oneofKind === 'variableRef') {
			const ref = record.variableRef as Record<string, unknown>;
			if (ref?.name === target.oldName) ref.name = target.newName;
			return; // Leaf node
		}
		// variableRef can appear inside groupReference variable values — keep traversing
		if (record.oneofKind === 'variablePlaceholder') {
			return; // Not affected by variable renames
		}
	}

	if (target.type === 'placeholder') {
		if (record.oneofKind === 'variablePlaceholder') {
			const ph = record.variablePlaceholder as Record<string, unknown>;
			if (ph?.name === target.oldName) ph.name = target.newName;
			return; // Leaf node
		}
		// groupReference.variables[] entries have names that map to placeholders
		if (record.oneofKind === 'groupReference') {
			const gr = record.groupReference as Record<string, unknown>;
			if (Array.isArray(gr?.variables)) {
				for (const v of gr.variables as Array<Record<string, unknown>>) {
					if (v.name === target.oldName) v.name = target.newName;
				}
			}
			// Keep traversing — nested values could contain more placeholders
		}
		if (record.oneofKind === 'variableRef') {
			return; // Not affected by placeholder renames
		}
	}

	if (target.type === 'group') {
		if (record.oneofKind === 'groupReference') {
			const gr = record.groupReference as Record<string, unknown>;
			if (gr?.groupName === target.oldName) gr.groupName = target.newName;
			return; // No nested group references inside a group reference
		}
		// A group is also named by the `Action Group is used` condition, which stores it as a bare
		// `{name}` the generic recursion below cannot recognise. Without this arm a rename leaves the
		// condition pointing at a group that no longer exists.
		if (record.oneofKind === 'actionGroupUsed') {
			const agu = record.actionGroupUsed as Record<string, unknown>;
			if (agu?.name === target.oldName) agu.name = target.newName;
			return; // Leaf node
		}
	}

	for (const value of Object.values(record)) {
		renameAPLReference(value, target);
	}
}

// Ensure that the auto rotation type can be matched. Only `type` moves, and the comparisons below
// only read, so this is a shallow copy — cloning the message would walk the whole rotation tree.
const matchableRotation = (rotation?: APLRotation): APLRotation | undefined =>
	rotation?.type === APLRotation_Type.TypeAuto ? { ...rotation, type: APLRotation_Type.TypeAPL } : rotation;

export const isEqualAPLRotation = (player: Player<Spec>, rotation?: APLRotation, otherRotation?: APLRotation): boolean => {
	if (!!rotation != !!otherRotation) return false;

	const matchable = matchableRotation(rotation);
	const otherMatchable = matchableRotation(otherRotation);
	if (otherMatchable?.type === APLRotation_Type.TypeSimple && otherMatchable.simple) {
		return (
			!!matchable?.simple &&
			player.specTypeFunctions.rotationEquals(
				simpleSpecRotationFromJson(player.specTypeFunctions, otherMatchable.simple.specRotationJson),
				simpleSpecRotationFromJson(player.specTypeFunctions, matchable.simple.specRotationJson),
			) &&
			// Cooldown timings are part of a Simple rotation, so two that differ only there are not
			// the same rotation. Without this the mage mana-gem and Serpent-Coil builds report as
			// already applied while the user's schedule is still the old one.
			Cooldowns.equals(otherMatchable.simple.cooldowns, matchable.simple.cooldowns)
		);
	} else {
		return APLRotation.equals(otherMatchable, matchable);
	}
};

// The offensive potion before the pull, which a stat potion is drunk for so its buff covers the opening.
export function prepullPotionAction(doAt?: string): APLPrepullAction {
	return APLPrepullAction.fromJsonString(
		`{"action":{"castSpell":{"spellId":{"otherId":"OtherActionOffensivePotion"}}},"doAtValue":{"const":{"val":"${doAt || '-1s'}"}}}`,
	);
}

export function autocastCooldownsAction(startAt?: string): APLAction {
	if (startAt) {
		return APLAction.fromJsonString(
			`{"condition":{"cmp":{"op":"OpGt","lhs":{"currentTime":{}},"rhs":{"const":{"val":"${startAt}"}}}},"autocastOtherCooldowns":{}}`,
		);
	} else {
		return APLAction.fromJsonString(`{"autocastOtherCooldowns":{}}`);
	}
}

export function scheduledCooldownAction(schedule: string, actionId: ActionIdProto): APLAction {
	return APLAction.fromJsonString(`{"schedule":{"schedule":"${schedule}","innerAction":{"castSpell":{"spellId":${ActionIdProto.toJsonString(actionId)}}}}}`);
}

export function simpleCooldownActions(cooldowns: Cooldowns): Array<APLAction> {
	return cooldowns.cooldowns
		.filter(cd => cd.id)
		.map(cd => {
			const schedule = cd.timings.map(timing => timing.toFixed(1) + 's').join(', ');
			return scheduledCooldownAction(schedule, cd.id!);
		});
}

/**
 * Hides castSpell actions, in the priority list and in action groups, for every cooldown the user
 * has scheduled with at least one timing, so a preset's own logic for that spell cannot fire before
 * or between the user's timings. Items are hidden, not removed, so list indices stay stable.
 * Mutates the given rotation, so pass a clone.
 * Matches on the spell, item or other id and ignores tag and rank, because the sim takes the whole
 * shared-timer family out of autocast anyway.
 * A cooldown with no timings is left alone: its schedule action is invalid and gets dropped, so
 * hiding the preset's line would mean the spell is never cast.
 */
export function hideScheduledCooldownCasts(rotation: APLRotation, cooldowns: Cooldowns): void {
	const scheduled = new Set(
		cooldowns.cooldowns
			.filter(cd => cd.timings.length > 0)
			.map(cd => rawIdKey(cd.id))
			.filter(key => key !== null),
	);
	if (!scheduled.size) return;

	for (const item of [...rotation.priorityList, ...rotation.groups.flatMap(group => group.actions)]) {
		const action = item.action?.action;
		if (action?.oneofKind !== 'castSpell') continue;
		const key = rawIdKey(action.castSpell.spellId);
		if (key !== null && scheduled.has(key)) item.hide = true;
	}
}

function rawIdKey(id: ActionIdProto | undefined): string | null {
	const rawId = id?.rawId;
	switch (rawId?.oneofKind) {
		case 'spellId':
			return `spell:${rawId.spellId}`;
		case 'itemId':
			return `item:${rawId.itemId}`;
		case 'otherId':
			return `other:${rawId.otherId}`;
		default:
			return null;
	}
}

/**
 * A copy of the rotation with every spell id `ids` maps swapped for the id it maps to, wherever an action or
 * value names it: a cast, a queued cast (tag 1), a dot or an aura. An id with tag -1 is an external buff copy,
 * such as the raid's Battle Shout, which stands for the rank the raid applies, so it is never swapped.
 * A spec's simple rotation uses it to pick between a trainer rank and the rank a class book teaches.
 */
export function withSpellIds(rotation: APLRotation, ids: Map<number, number>): APLRotation {
	const copy = APLRotation.clone(rotation);
	const visit = (node: unknown) => {
		if (!node || typeof node !== 'object') return;
		const id = node as Partial<ActionIdProto>;
		if (id.rawId?.oneofKind === 'spellId') {
			if (id.tag !== -1) id.rawId.spellId = ids.get(id.rawId.spellId) ?? id.rawId.spellId;
			return;
		}
		for (const child of Object.values(node)) visit(child);
	};
	visit(copy);
	return copy;
}

export function standardCooldownDefaults(cooldowns: Cooldowns, prepotAt?: string, startAutocastCDsAt?: string): [Array<APLPrepullAction>, Array<APLAction>] {
	return [[prepullPotionAction(prepotAt)], [autocastCooldownsAction(startAutocastCDsAt), simpleCooldownActions(cooldowns)].flat()];
}
