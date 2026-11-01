import type { JsonValue } from '@protobuf-ts/runtime';
import { useCallback, useMemo, useRef } from 'react';

import { useTypedLocalStorage } from './useTypedLocalStorage';

// Every generated `MessageType` satisfies this structurally, so a named hook passes the proto itself.
export interface SavedDataCodec<T> {
	toJson: (data: T) => JsonValue;
	fromJson: (json: any) => T;
}

export interface SavedDataEntry<T> {
	name: string;
	data: T;
	// Serialised form, for comparing an entry against the live subject.
	json: string;
}

export interface UseSavedDataResult<T> {
	entries: Array<SavedDataEntry<T>>;
	// Every name in the slot, the ones the codec rejected included, which `entries` leaves out.
	storedNames: string[];
	save: (name: string, data: T) => void;
	// All in one write, with one answer: false when storage refused it, and nothing of the batch is kept.
	saveMany: (items: ReadonlyArray<{ name: string; data: T }>) => boolean;
	remove: (name: string) => void;
	codec: SavedDataCodec<T>;
}

type StoredRecord = Record<string, JsonValue>;

const parseStoredRecord = (value: unknown): StoredRecord | undefined =>
	value !== null && typeof value === 'object' && !Array.isArray(value) ? (value as StoredRecord) : undefined;

// Throws when the codec cannot read the entry.
export const readSavedDataEntry = <T>(name: string, json: JsonValue, codec: SavedDataCodec<T>): SavedDataEntry<T> => {
	const data = codec.fromJson(json);
	return { name, data, json: JSON.stringify(codec.toJson(data)) };
};

// The entries of a parsed slot, for a reader outside React; anything but a record holds none.
export const savedDataEntries = <T>(value: unknown, codec: SavedDataCodec<T>): Array<SavedDataEntry<T>> => {
	const stored = parseStoredRecord(value);
	const parsed: Array<SavedDataEntry<T>> = [];
	for (const name in stored) {
		try {
			parsed.push(readSavedDataEntry(name, stored[name], codec));
		} catch (error) {
			// One unreadable entry must not cost the user the rest of the slot.
			console.warn('Failed parsing saved data: ', stored[name], error);
		}
	}
	return parsed;
};

// One saved-data slot: `Record<name, toJson(data)>` under one key, which is the shape every
// `SavedDataManager` has always written. Only the key and the codec vary between slots.
export const useSavedData = <T>(key: string, codec: SavedDataCodec<T>): UseSavedDataResult<T> => {
	const [stored, setStored] = useTypedLocalStorage<StoredRecord>(key, parseStoredRecord);
	// Held in a ref so a caller may pass the proto inline without churning every callback below.
	const codecRef = useRef(codec);
	codecRef.current = codec;
	const storedRef = useRef(stored);
	storedRef.current = stored;

	const entries = useMemo(() => savedDataEntries(stored, codecRef.current), [stored]);
	const storedNames = useMemo(() => Object.keys(stored ?? {}), [stored]);
	// Every write patches the stored record rather than rebuilding it from `entries`, which would
	// silently drop every entry the codec rejected the moment the user saves anything else. The ref moves with each
	// write, so several in one handler keep one another's entries.
	const write = useCallback(
		(patch: (record: StoredRecord) => StoredRecord) => {
			const next = patch(storedRef.current ?? {});
			const ok = setStored(next);
			if (ok) storedRef.current = next;
			return ok;
		},
		[setStored],
	);

	const save = useCallback(
		(name: string, data: T) => {
			write(record => ({ ...record, [name]: codecRef.current.toJson(data) }));
		},
		[write],
	);

	const saveMany = useCallback(
		(items: ReadonlyArray<{ name: string; data: T }>) =>
			write(record => ({ ...record, ...Object.fromEntries(items.map(({ name, data }) => [name, codecRef.current.toJson(data)])) })),
		[write],
	);

	const remove = useCallback(
		(name: string) => {
			write(record => {
				const next = { ...record };
				delete next[name];
				return next;
			});
		},
		[write],
	);

	return { entries, storedNames, save, saveMany, remove, codec };
};
