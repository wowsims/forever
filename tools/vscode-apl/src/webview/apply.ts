// The panel's own copy of the rotation: each edit is made on it at once, as the extension's splice makes it in the
// text, so the next click is made against what the panel shows. Rows keyed by their path follow the edits too.

import type { AplEdit, JsonPath } from '../protocol';
import { isObject, type Json, type JsonObject } from './model';

// `value` with `change` made at `path`; a change that gives undefined removes the property or element. A path
// through an index that does not exist changes nothing.
function update(value: Json | undefined, path: JsonPath, change: (old: Json | undefined) => Json | undefined): Json | undefined {
	if (path.length === 0) return change(value);
	const [head, ...rest] = path;
	if (typeof head === 'number') {
		if (!Array.isArray(value) || head < 0 || head >= value.length) return value;
		const list = [...value];
		const next = update(list[head], rest, change);
		if (next === undefined) list.splice(head, 1);
		else list[head] = next;
		return list;
	}
	const object: JsonObject = isObject(value) ? { ...value } : {};
	const next = update(object[head], rest, change);
	if (next === undefined) delete object[head];
	else object[head] = next;
	return object;
}

const valueAt = (json: Json | undefined, path: JsonPath): Json | undefined =>
	path.reduce<Json | undefined>((node, key) => (Array.isArray(node) || isObject(node) ? (node as Record<string | number, Json>)[key] : undefined), json);

// Edits of when blocks and defs, and def calls, change JSON the panel cannot work out; it waits for the text's.
export function applyEdit(rotation: JsonObject, edit: AplEdit): JsonObject {
	if ('path' in edit && edit.path[0] === 'whens') return rotation;
	switch (edit.op) {
		case 'set':
			if (edit.call !== undefined) return rotation;
			if (edit.value === undefined && valueAt(rotation, edit.path) === undefined) return rotation;
			return update(rotation, edit.path, () => edit.value as Json | undefined) as JsonObject;
		case 'insert':
			return update(rotation, edit.path, list => {
				const items = Array.isArray(list) ? [...list] : [];
				items.splice(Math.min(Math.max(edit.index, 0), items.length), 0, edit.value as Json);
				return items;
			}) as JsonObject;
		case 'move':
			return update(rotation, edit.path, list => {
				if (!Array.isArray(list) || edit.from < 0 || edit.from >= list.length) return list;
				const items = [...list];
				const [moved] = items.splice(edit.from, 1);
				items.splice(edit.to, 0, moved);
				return items;
			}) as JsonObject;
		default:
			return rotation;
	}
}

// Where an element of the list an edit changes ends up: its new index, or undefined when the edit removes it.
function indexMap(edit: AplEdit): { list: string; to: (i: number) => number | undefined } | undefined {
	if ('path' in edit && edit.path[0] === 'whens') return undefined;
	switch (edit.op) {
		case 'insert':
			return { list: edit.path.join('/'), to: i => (i >= edit.index ? i + 1 : i) };
		case 'move': {
			const { from, to } = edit;
			return {
				list: edit.path.join('/'),
				to: i => (i === from ? to : from < to && i > from && i <= to ? i - 1 : from > to && i >= to && i < from ? i + 1 : i),
			};
		}
		case 'set': {
			const last = edit.path[edit.path.length - 1];
			if (edit.value !== undefined || typeof last !== 'number') return undefined;
			return { list: edit.path.slice(0, -1).join('/'), to: i => (i === last ? undefined : i > last ? i - 1 : i) };
		}
		default:
			return undefined;
	}
}

// Keys of the form 'priorityList/3' or 'groups/1/actions/0', moved with the rows and groups an edit inserts, moves
// or removes. A key under a removed element goes.
export function shiftKeys(keys: ReadonlySet<string>, edit: AplEdit): Set<string> {
	const map = indexMap(edit);
	if (!map) return new Set(keys);
	const prefix = `${map.list}/`;
	const out = new Set<string>();
	for (const key of keys) {
		const [head, ...rest] = key.startsWith(prefix) ? key.slice(prefix.length).split('/') : [];
		if (head === undefined || !/^\d+$/.test(head)) {
			out.add(key);
			continue;
		}
		const index = map.to(Number(head));
		if (index !== undefined) out.add([`${prefix}${index}`, ...rest].join('/'));
	}
	return out;
}
