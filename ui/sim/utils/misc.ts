import { v4 as uuidv4 } from 'uuid';

export const randomUUID = () => uuidv4();

// eslint-disable-next-line @typescript-eslint/no-empty-function
export const noop = () => {};

export const sleep = (ms: number) => new Promise(r => setTimeout(r, ms));

/** `name` when it is free, otherwise the first free `base (2)`, `base (3)`, ..., where `base` is `name` without a ` (n)` it already ends in. */
export const freeName = (name: string, taken: Iterable<string>): string => {
	const names = taken instanceof Set ? taken : new Set(taken);
	if (!names.has(name)) return name;
	const base = name.replace(/ \(\d+\)$/, '');
	for (let n = 2; ; n++) if (!names.has(`${base} (${n})`)) return `${base} (${n})`;
};

/** A one-entry memo: each call returns the last value, or runs `compute` again when one of `deps` changed identity since. */
export const memoLast = <T>() => {
	let last: { deps: ReadonlyArray<unknown>; value: T } | undefined;
	return (deps: ReadonlyArray<unknown>, compute: () => T): T => {
		if (!last || deps.length !== last.deps.length || deps.some((dep, index) => !Object.is(dep, last!.deps[index]))) last = { deps, value: compute() };
		return last.value;
	};
};

/**
 * Synchronous 64-bit string hash (FNV-1a paired with a djb2 variant). Collision-resistant enough
 * for local cache keys and content-derived seeds; unlike crypto.subtle it costs no async round-trip.
 */
export function hashString(value: string): string {
	let h1 = 0x811c9dc5;
	let h2 = 0xcbf29ce4;
	for (let i = 0; i < value.length; i++) {
		const charCode = value.charCodeAt(i);
		h1 = Math.imul(h1 ^ charCode, 0x01000193) >>> 0;
		h2 = (Math.imul(h2, 33) ^ charCode) >>> 0;
	}
	return h1.toString(16).padStart(8, '0') + h2.toString(16).padStart(8, '0');
}
