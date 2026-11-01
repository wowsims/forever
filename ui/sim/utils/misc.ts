// `crypto.randomUUID` exists only in a secure context. A page served over plain http from a LAN host
// (`make dev` with `--host`) builds the v4 id from `getRandomValues`, which every context has.
export const randomUUID = (): string => {
	if (typeof crypto.randomUUID === 'function') return crypto.randomUUID();
	const bytes = crypto.getRandomValues(new Uint8Array(16));
	bytes[6] = (bytes[6] & 0x0f) | 0x40;
	bytes[8] = (bytes[8] & 0x3f) | 0x80;
	const hex = Array.from(bytes, b => b.toString(16).padStart(2, '0')).join('');
	return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
};

// eslint-disable-next-line @typescript-eslint/no-empty-function
export const noop = () => {};

export const sleep = (ms: number) => new Promise(r => setTimeout(r, ms));

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
