import { createContext, useContext } from 'react';

import type { AplDatabase } from '../protocol';
import type { Json } from './model';
import { isObject } from './model';
import { idKey, registeredKeys, useRegistered } from './registered';

export const DbContext = createContext<AplDatabase>({ spells: {}, items: {} });

export const useDb = () => useContext(DbContext);

export function iconUrl(icon: string): string {
	return `https://wow.zamimg.com/images/wow/icons/medium/${icon}.jpg`;
}

export interface ResolvedId {
	kind: 'spell' | 'item' | 'other';
	id: number | string;
	name: string;
	icon?: string;
	tag?: number;
}

// A proto ActionID as JSON (`{ "spellId": 1 }`, `{ "itemId": 2 }`, `{ "otherId": "OtherActionAttack" }`)
// with the name and icon the database gives it.
export function resolveId(db: AplDatabase, json: Json | undefined): ResolvedId | undefined {
	if (!isObject(json)) {
		return undefined;
	}
	const tag = typeof json.tag === 'number' ? json.tag : undefined;
	if (typeof json.spellId === 'number') {
		const [name, icon] = db.spells[json.spellId] ?? [`Spell ${json.spellId}`, undefined];
		return { kind: 'spell', id: json.spellId, name, icon, tag };
	}
	if (typeof json.itemId === 'number') {
		const [name, icon] = db.items[json.itemId] ?? [`Item ${json.itemId}`, undefined];
		return { kind: 'item', id: json.itemId, name, icon, tag };
	}
	if (typeof json.otherId === 'string') {
		return { kind: 'other', id: json.otherId, name: json.otherId.replace(/^OtherAction/, ''), tag };
	}
	return undefined;
}

export function IdChip({ json }: { json: Json | undefined }) {
	const resolved = resolveId(useDb(), json);
	const registered = useRegistered();
	const keys = registeredKeys(registered);
	if (!resolved) {
		return <span className="chip chip-empty">none</span>;
	}
	const known = resolved.kind === 'other' || !resolved.name.startsWith(resolved.kind === 'spell' ? 'Spell ' : 'Item ');
	// Items register with the gear, so only a spell is held to the spec's list.
	const unregistered = resolved.kind === 'spell' && keys !== undefined && !keys.has(idKey(json) ?? '');
	const title = `${resolved.kind} ${resolved.id}${resolved.tag ? ` · tag ${resolved.tag}` : ''}`;
	return (
		<span
			className={`chip chip-${resolved.kind}${known ? '' : ' chip-unknown'}${unregistered ? ' chip-unregistered' : ''}`}
			title={
				unregistered && registered.status === 'ready'
					? `${title}: the sim does not register this for ${registered.spec}; an item or set bonus may`
					: title
			}>
			{unregistered ? <span className="warn">!</span> : null}
			{resolved.icon ? <img src={iconUrl(resolved.icon)} alt="" /> : null}
			{resolved.name}
			{resolved.tag ? <sup>{resolved.tag}</sup> : null}
		</span>
	);
}
