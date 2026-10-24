import type { ACTION_ID_SET } from '@features/apl/model/action_id_sets';
import { createContext, useContext } from 'react';

import type { ActionIdJson, Registered, RegisteredName } from '../protocol';
import { isObject, type Json } from './model';

export const RegisteredContext = createContext<Registered>({ status: 'loading' });

export const useRegistered = () => useContext(RegisteredContext);

// An action id's identity: what it names and its tag. The rank is the spell id's own and not compared.
export function idKey(json: Json | ActionIdJson | undefined): string | undefined {
	if (!isObject(json as Json | undefined)) return undefined;
	const id = json as ActionIdJson;
	const tag = typeof id.tag === 'number' ? id.tag : 0;
	if (typeof id.spellId === 'number') return `spell:${id.spellId}:${tag}`;
	if (typeof id.itemId === 'number') return `item:${id.itemId}:${tag}`;
	if (typeof id.otherId === 'string') return `other:${id.otherId}:${tag}`;
	return undefined;
}

// The sets whose fields hold an item, and those that take an item as well as a spell.
const ITEM_ONLY = new Set<ACTION_ID_SET>(['potions', 'conjured_items']);
const ITEM_SETS = new Set<ACTION_ID_SET>(['spells', 'castable_spells']);

// The auras a unit of `type` (a UnitReference type's JSON name) carries, as the parser reads an aura's name on it: the
// pet's, the player's (with no unit too), or the target's.
function auraNamespaceOf(type: string | undefined): 'aura' | 'pet_aura' | 'target_aura' {
	if (type === 'Pet') return 'pet_aura';
	return type === undefined || ['Self', 'Player', 'AllPlayers', 'Unknown'].includes(type) ? 'aura' : 'target_aura';
}

// What an ActionID field offers, in the namespace the parser reads its name in (SPEC.md §7.2): an aura id among the auras of the unit the field's JSON names (the player's with none), a
// boss spell's id among the target's spells, and any other id among the spec's spells, with the items with a use
// where the site's set takes one. Each prints as its name.
export function offeredFor(
	registered: Extract<Registered, { status: 'ready' }>,
	field: { kind: string; name: string; set: ACTION_ID_SET },
	unit: Json | undefined,
): RegisteredName[] {
	const { spells, auras, targetAuras, petAuras, targetSpells, items } = registered.registered;
	if (ITEM_ONLY.has(field.set)) return items;
	if (field.name === 'auraId') {
		const ns = auraNamespaceOf(isObject(unit) && typeof unit.type === 'string' ? unit.type : undefined);
		return ns === 'pet_aura' ? petAuras : ns === 'target_aura' ? targetAuras : auras;
	}
	if (field.kind.startsWith('bossSpell')) return targetSpells;
	return ITEM_SETS.has(field.set) ? [...spells, ...items] : spells;
}

const keyCache = new WeakMap<Registered, Set<string>>();

// Every id the spec registers, as a spell or as an aura on any unit; built once for each answer.
export function registeredKeys(registered: Registered): Set<string> | undefined {
	if (registered.status !== 'ready') return undefined;
	let keys = keyCache.get(registered);
	if (!keys) {
		const { spells, auras, petAuras, targetAuras, targetSpells } = registered.registered;
		keys = new Set([...spells, ...auras, ...petAuras, ...targetAuras, ...targetSpells].map(entry => idKey(entry.id)).filter((key): key is string => !!key));
		keyCache.set(registered, keys);
	}
	return keys;
}
