import type { JsonValue } from '@protobuf-ts/runtime';
import { freeName } from '@sim/utils/misc';
import { readSavedDataEntry, type SavedDataCodec } from '@ui-kit/hooks/useSavedData';

export const MAX_NAME_LENGTH = 100;
// Names other code could read back through the prototype chain.
const RESERVED_NAMES = new Set(['__proto__', 'constructor', 'prototype']);

export interface ImportRow<T> {
	name: string;
	saveAs: string;
	// The saved entry holding the same contents, on a row already saved.
	savedAs?: string;
	data: T;
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

// The panel's save input takes a name of any length, so an export can hold one over the cap. The cut leaves no half of a
// surrogate pair behind.
const cappedName = (name: string): string =>
	name.length <= MAX_NAME_LENGTH
		? name
		: name
				.slice(0, MAX_NAME_LENGTH)
				.replace(/[\uD800-\uDBFF]$/, '')
				.trimEnd();

const readEntry = <T>(rawName: string, json: JsonValue, { codec, check }: ImportPlanSources<T>) => {
	const name = rawName.trim();
	if (!name || RESERVED_NAMES.has(name)) return null;
	try {
		const { data, json: normalised } = readSavedDataEntry(name, json, codec);
		check?.(data);
		return { name, wanted: cappedName(name), data, json: normalised };
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
	const savedAs = ({ wanted, json }: { wanted: string; json: string }) => (savedJson.get(wanted) === json ? wanted : savedNameOf.get(json));

	const holders = new Map(readable.map(entry => [entry, savedAs(entry)]));
	const fresh = readable.filter(entry => holders.get(entry) === undefined);
	const known = readable.filter(entry => holders.get(entry) !== undefined);
	const taken = new Set([...sources.storedNames, ...sources.presetNames]);
	const saveAs = new Map<(typeof readable)[number], string>();
	// Every free name is claimed before any is numbered, so the numbers step around the names later in the file. A row
	// already saved comes last, since it saves only if the user ticks it.
	for (const entry of fresh) {
		if (taken.has(entry.wanted)) continue;
		taken.add(entry.wanted);
		saveAs.set(entry, entry.wanted);
	}
	for (const entry of [...fresh, ...known]) {
		if (saveAs.has(entry)) continue;
		const free = freeName(entry.wanted, taken);
		taken.add(free);
		saveAs.set(entry, free);
	}

	const rows = readable.map((entry): ImportRow<T> => {
		const holder = holders.get(entry);
		return { name: entry.name, saveAs: saveAs.get(entry)!, data: entry.data, savedAs: holder, ticked: holder === undefined };
	});

	return { rows, unreadable: read.length - readable.length };
};

export const tickedRows = <T>({ rows }: ImportPlan<T>): Array<ImportRow<T>> => rows.filter(row => row.ticked);

// Only the rows the player ticked or unticked, by name; the rest keep what the plan says.
export const withTicksKept = <T>(plan: ImportPlan<T>, ticks: ReadonlyMap<string, boolean>): ImportPlan<T> => ({
	...plan,
	rows: plan.rows.map(row => ({ ...row, ticked: ticks.get(row.name) ?? row.ticked })),
});
