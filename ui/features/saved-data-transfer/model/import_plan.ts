import type { JsonValue } from '@protobuf-ts/runtime';
import type { SavedDataCodec } from '@ui-kit/hooks/useSavedData';

export const MAX_IMPORT_ENTRIES = 500;
export const MAX_NAME_LENGTH = 100;
// Names other code could read back through the prototype chain.
const RESERVED_NAMES = new Set(['__proto__', 'constructor', 'prototype']);

export type ImportRowStatus = 'new' | 'renamed' | 'already_saved';

export interface ImportRow<T> {
	name: string;
	saveAs: string;
	// The saved entry holding the same contents, on an `already_saved` row.
	savedAs?: string;
	data: T;
	status: ImportRowStatus;
	ticked: boolean;
}

export interface ImportPlan<T> {
	rows: Array<ImportRow<T>>;
	unreadable: number;
}

export interface ImportPlanSources<T> {
	codec: SavedDataCodec<T>;
	// Throws on an entry the codec reads that the rest of the UI could not use.
	check?: (data: T) => void;
	saved: ReadonlyArray<{ name: string; json: string }>;
	// Every name in the slot, the entries the codec rejects included, so an import never overwrites one.
	storedNames: readonly string[];
	presetNames: readonly string[];
}

const freeName = (name: string, taken: ReadonlySet<string>): string => {
	let number = 2;
	while (taken.has(`${name} (${number})`)) number++;
	return `${name} (${number})`;
};

const readEntry = <T>(rawName: string, json: JsonValue, { codec, check }: ImportPlanSources<T>) => {
	const name = rawName.trim();
	if (!name || name.length > MAX_NAME_LENGTH || RESERVED_NAMES.has(name)) return null;
	try {
		const data = codec.fromJson(json);
		check?.(data);
		return { name, data, json: JSON.stringify(codec.toJson(data)) };
	} catch {
		return null;
	}
};

export const planImport = <T>(entries: Record<string, JsonValue>, sources: ImportPlanSources<T>): ImportPlan<T> => {
	const read = Object.entries(entries).map(([name, json]) => readEntry(name, json, sources));
	const readable = read.filter(entry => entry !== null);

	const savedJson = new Map(sources.saved.map(({ name, json }) => [name, json]));
	const savedNameOf = new Map<string, string>();
	for (const { name, json } of sources.saved) if (!savedNameOf.has(json)) savedNameOf.set(json, name);
	const savedAs = ({ name, json }: { name: string; json: string }) => (savedJson.get(name) === json ? name : savedNameOf.get(json));

	const fresh = readable.filter(entry => savedAs(entry) === undefined);
	const known = readable.filter(entry => savedAs(entry) !== undefined);
	const taken = new Set([...sources.storedNames, ...sources.presetNames]);
	const saveAs = new Map<(typeof readable)[number], string>();
	// Every free name is claimed before any is numbered, so the numbers step around the names later in the file. A row
	// already saved comes last, since it saves only if the user ticks it.
	for (const entry of fresh) {
		if (taken.has(entry.name)) continue;
		taken.add(entry.name);
		saveAs.set(entry, entry.name);
	}
	for (const entry of [...fresh, ...known]) {
		if (saveAs.has(entry)) continue;
		const free = taken.has(entry.name) ? freeName(entry.name, taken) : entry.name;
		taken.add(free);
		saveAs.set(entry, free);
	}

	const rows = readable.map((entry): ImportRow<T> => {
		const row = { name: entry.name, saveAs: saveAs.get(entry)!, data: entry.data };
		const holder = savedAs(entry);
		if (holder !== undefined) return { ...row, savedAs: holder, status: 'already_saved', ticked: false };
		return { ...row, status: row.saveAs === row.name ? 'new' : 'renamed', ticked: true };
	});

	return { rows, unreadable: read.length - readable.length };
};

export const tickedRows = <T>({ rows }: ImportPlan<T>): Array<ImportRow<T>> => rows.filter(row => row.ticked);

export const withRowTicked = <T>(plan: ImportPlan<T>, index: number, ticked: boolean): ImportPlan<T> => ({
	...plan,
	rows: plan.rows.map((row, i) => (i === index ? { ...row, ticked } : row)),
});
