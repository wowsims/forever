// The Rotation Master export string: standard Base64 of a LibCompress stream (tag 1 store, 2 LZW, 3 Huffman) holding
// an AceSerializer rev. 1 value, optionally led by "RM1:". It is the string the addon's Import/Export box shows and
// that rotationmaster.neuromancy.net hands out. This file reads and writes only that wire format, from its published
// behaviour; it carries no code from either library and evaluates nothing (the data is inert).
//
// A Lua table becomes an array when its keys are exactly 1..n and a null-prototype object otherwise; Lua strings are
// UTF-8 bytes and become JS strings. Everything read is bounded (sizes, nesting, tokens), so a hostile string costs
// bounded work and memory.

export type RmValue = null | boolean | number | string | RmValue[] | { [key: string]: RmValue };

export type DecodeResult =
	| { ok: true; value: RmValue; /** The root has format "RotationMaster" (an RM1: envelope) rather than a bare rotation body. */ envelope: boolean }
	| { ok: false; reason: string };
export type EncodeResult = { ok: true; text: string } | { ok: false; reason: string };

export const CODEC_LIMITS = Object.freeze({
	input: 2097152,
	compressed: 1048577,
	serialized: 1048576,
	tokens: 20000,
	tableDepth: 64,
});

const ALPHABET = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/';
export const EXPORT_COLUMNS = 76;

// Every whitespace and zero-width character the addon removes from a pasted export (chat, mail and forums re-space it).
const REMOVED = new RegExp('[\\t-\\r \\u0085\\u00a0\\u00ad\\u1680\\u180e\\u2000-\\u200f\\u2028-\\u202f\\u205f-\\u2064\\u2066-\\u2069\\u3000\\ufeff]+', 'g');
export const stripExportWhitespace = (text: string): string => text.replace(REMOVED, '');

/** The export broken into lines of EXPORT_COLUMNS characters, the way the addon and the site show it. */
export const wrapExport = (text: string, columns = EXPORT_COLUMNS): string => {
	const flat = text.replace(/[\r\n]/g, '');
	if (flat.length <= columns) return flat;
	const lines: string[] = [];
	for (let at = 0; at < flat.length; at += columns) lines.push(flat.slice(at, at + columns));
	return lines.join('\n');
};

const fail = (reason: string): { ok: false; reason: string } => ({ ok: false, reason });
type Stage<T> = { ok: true; value: T } | { ok: false; reason: string };

// ---- Base64 ------------------------------------------------------------------------------------------------------
function fromBase64(text: string): Stage<Uint8Array> {
	if (!text.length || text.length % 4 || /[^A-Za-z0-9+/=]/.test(text)) return fail('invalid-base64');
	const pad = text.endsWith('==') ? 2 : text.endsWith('=') ? 1 : 0;
	if (/=/.test(text.slice(0, text.length - pad))) return fail('invalid-base64-padding');
	const length = (text.length / 4) * 3 - pad;
	if (length > CODEC_LIMITS.compressed) return fail('compressed-size-limit');
	const out = new Uint8Array(length);
	let at = 0;
	for (let i = 0; i < text.length; i += 4) {
		const a = ALPHABET.indexOf(text[i]);
		const b = ALPHABET.indexOf(text[i + 1]);
		const c = text[i + 2] === '=' ? 0 : ALPHABET.indexOf(text[i + 2]);
		const d = text[i + 3] === '=' ? 0 : ALPHABET.indexOf(text[i + 3]);
		const code = a * 262144 + b * 4096 + c * 64 + d;
		out[at++] = Math.floor(code / 65536);
		if (at < length) out[at++] = Math.floor(code / 256) % 256;
		if (at < length) out[at++] = code % 256;
	}
	return { ok: true, value: out };
}

function toBase64(bytes: Uint8Array): string {
	const parts: string[] = [];
	for (let i = 0; i < bytes.length; i += 3) {
		const a = bytes[i];
		const b = bytes[i + 1];
		const c = bytes[i + 2];
		parts.push(
			ALPHABET[a >>> 2] +
				ALPHABET[((a & 3) << 4) | ((b ?? 0) >>> 4)] +
				(b === undefined ? '=' : ALPHABET[((b & 15) << 2) | ((c ?? 0) >>> 6)]) +
				(c === undefined ? '=' : ALPHABET[c & 63]),
		);
	}
	return parts.join('');
}

// ---- LibCompress streams -------------------------------------------------------------------------------------------
function* lzwCodes(input: Uint8Array): Generator<number> {
	let p = 1;
	while (p < input.length) {
		const marker = input[p++];
		if (marker < 250) {
			yield marker;
			continue;
		}
		const digits = 256 - marker;
		if (p + digits > input.length) throw new Error('invalid-lzw-code');
		let code = 0;
		let factor = 1;
		for (let n = 0; n < digits; n++) {
			const digit = input[p++];
			if (digit === 0) throw new Error('invalid-lzw-code');
			code += (digit - 1) * factor;
			factor *= 255;
		}
		yield code;
	}
}

function inflateLzw(input: Uint8Array): Stage<Uint8Array> {
	// The first pass counts the output without building any phrase.
	const lengths = new Uint32Array(input.length + 256);
	lengths.fill(1, 0, 256);
	let next = 256;
	let previous = -1;
	let total = 0;
	for (const code of lzwCodes(input)) {
		let length: number;
		if (previous < 0) {
			if (code > 255) throw new Error('invalid-lzw-code');
			length = 1;
		} else if (code === next) length = lengths[previous] + 1;
		else {
			if (code >= next) throw new Error('invalid-lzw-code');
			length = lengths[code];
		}
		total += length;
		if (total > CODEC_LIMITS.serialized) return fail('decompressed-size-limit');
		if (previous >= 0) lengths[next++] = lengths[previous] + 1;
		previous = code;
	}
	if (previous < 0) return fail('empty-lzw-stream');
	const output = new Uint8Array(total);
	const prefix = new Int32Array(next);
	const suffix = new Uint8Array(next);
	for (let i = 0; i < 256; i++) {
		prefix[i] = -1;
		suffix[i] = i;
	}
	let position = 0;
	next = 256;
	previous = -1;
	const emit = (code: number): number => {
		const end = position + lengths[code];
		let cursor = end;
		let walk = code;
		while (walk >= 256) {
			output[--cursor] = suffix[walk];
			walk = prefix[walk];
		}
		output[--cursor] = walk;
		position = end;
		return output[cursor];
	};
	for (const code of lzwCodes(input)) {
		let first: number;
		if (code === next) {
			first = emit(previous);
			output[position++] = first;
		} else first = emit(code);
		if (previous >= 0) {
			prefix[next] = previous;
			suffix[next] = first;
			next++;
		}
		previous = code;
	}
	return { ok: true, value: output };
}

function inflateHuffman(input: Uint8Array): Stage<Uint8Array> {
	if (input.length < 5) return fail('invalid-huffman-header');
	const size = input[2] + input[3] * 256 + input[4] * 65536;
	if (size > CODEC_LIMITS.serialized) return fail('decompressed-size-limit');
	if (size === 0) return { ok: true, value: new Uint8Array() };
	let offset = 40;
	const bitAt = (at: number): number => Math.floor(input[Math.floor(at / 8)] / 2 ** (at % 8)) % 2;
	const bit = (): number | null => (offset >= input.length * 8 ? null : bitAt(offset++));
	const byLength: Array<Map<number, number>> = Array.from({ length: 33 }, () => new Map());
	let maximum = 0;
	let minimum = 33;
	for (let entry = 0; entry < input[1] + 1; entry++) {
		let symbol = 0;
		for (let j = 0; j < 8; j++) {
			const b = bit();
			if (b === null) return fail('invalid-huffman-codebook');
			symbol += b * 2 ** j;
		}
		let escaped = 0;
		let code = 0;
		let length = 0;
		for (;;) {
			const first = bit();
			if (first === null) return fail('invalid-huffman-codebook');
			if (first === 1) {
				const second = bit();
				if (second === null) return fail('invalid-huffman-codebook');
				if (second === 1) {
					if (escaped === 0) return fail('invalid-huffman-codebook');
					break;
				}
				escaped += 2;
				code += 2 ** length;
			} else escaped++;
			if (escaped > 32) return fail('invalid-huffman-codebook');
			length++;
		}
		byLength[length].set(code, symbol);
		minimum = Math.min(minimum, length);
		maximum = Math.max(maximum, length);
	}
	const output = new Uint8Array(size);
	// The game's reader can consume one zero-padding byte after the end of the stream.
	const end = input.length * 8 + 8;
	for (let p = 0; p < size; p++) {
		let code = 0;
		let found = false;
		for (let length = 1; length <= maximum; length++) {
			if (offset >= end) return fail('decompression-size-mismatch');
			const b = offset < input.length * 8 ? bitAt(offset) : 0;
			offset++;
			code += b * 2 ** (length - 1);
			const symbol = length >= minimum ? byLength[length].get(code) : undefined;
			if (symbol !== undefined) {
				output[p] = symbol;
				found = true;
				break;
			}
		}
		if (!found) return fail('decompress-error');
	}
	return { ok: true, value: output };
}

function inflate(input: Uint8Array): Stage<Uint8Array> {
	if (input.length > CODEC_LIMITS.compressed) return fail('compressed-size-limit');
	try {
		if (input[0] === 1) {
			if (input.length - 1 > CODEC_LIMITS.serialized) return fail('decompressed-size-limit');
			return { ok: true, value: input.slice(1) };
		}
		if (input[0] === 2) return inflateLzw(input);
		if (input[0] === 3) return inflateHuffman(input);
		return fail('unsupported-compression');
	} catch (error) {
		if (error instanceof Error && error.message === 'invalid-lzw-code') return fail('invalid-lzw-code');
		throw error;
	}
}

/**
 * LZW (wire tag 2): 256 single-byte codes, then every new phrase gets the next code. A code is one byte when it is
 * 1..249, otherwise a marker (256 minus the digit count) followed by little-endian base-255 digits stored plus one.
 * Returns null unless the result is smaller than store framing, so the caller falls back to tag 1.
 */
export function deflateLzw(raw: Uint8Array): Uint8Array | null {
	const n = raw.length;
	if (!n) return null;
	const out = new Uint8Array(n + 1);
	out[0] = 2;
	let at = 1;
	const dict = new Map<number, number>();
	let next = 256;
	let w = raw[0];
	const emit = (code: number): boolean => {
		if (code > 0 && code < 250) {
			if (at >= n) return false;
			out[at++] = code;
			return true;
		}
		const digits: number[] = [];
		let x = code;
		do {
			digits.push(x % 255);
			x = Math.floor(x / 255);
		} while (x > 0);
		if (at + 1 + digits.length > n) return false;
		out[at++] = 256 - digits.length;
		for (const d of digits) out[at++] = d + 1;
		return true;
	};
	for (let i = 1; i < n; i++) {
		const c = raw[i];
		const key = w * 256 + c;
		const found = dict.get(key);
		if (found !== undefined) {
			w = found;
			continue;
		}
		dict.set(key, next++);
		if (!emit(w)) return null;
		w = c;
	}
	if (!emit(w)) return null;
	return at < n + 1 ? out.slice(0, at) : null;
}

// ---- AceSerializer rev. 1 --------------------------------------------------------------------------------------------
type Tok = { control: string; text: string };

function aceNumber(text: string, legacyInfinity = false): number | undefined {
	if (/^[+-]?(?:nan|inf(?:inity)?)$/i.test(text)) return /nan/i.test(text) ? NaN : text[0] === '-' ? -Infinity : Infinity;
	if (legacyInfinity && (text === '1.#INF' || text === '-1.#INF')) return text[0] === '-' ? -Infinity : Infinity;
	// Hexadecimal integers only: the addon never writes hex floats.
	if (/^[+-]?0x[0-9a-f]+$/i.test(text)) return (text[0] === '-' ? -1 : 1) * Number('0x' + text.replace(/^[+-]?0x/i, ''));
	if (!/^[+-]?(?:\d+(?:\.\d*)?|\.\d+)(?:e[+-]?\d+)?$/i.test(text)) return undefined;
	return Number(text);
}

function aceUnescape(text: string): string {
	let out = '';
	for (let i = 0; i < text.length; i++) {
		const c = text.charCodeAt(i);
		if (c !== 126 || i === text.length - 1) {
			out += text[i];
			continue;
		}
		const next = text.charCodeAt(++i);
		let value: number | undefined;
		if (next < 122) value = next - 64;
		else value = { 122: 30, 123: 127, 124: 126, 125: 94 }[next as 122 | 123 | 124 | 125];
		if (value === undefined || value < 0) throw new Error('deserialize-error');
		out += String.fromCharCode(value);
	}
	return out;
}

const bytesToBinary = (bytes: Uint8Array): string => {
	const chunks: string[] = [];
	for (let i = 0; i < bytes.length; i += 8192) chunks.push(String.fromCharCode(...bytes.subarray(i, i + 8192)));
	return chunks.join('');
};
/** A Lua byte string (one char per byte) as the UTF-8 text it holds. */
const utf8 = (binary: string): string => {
	for (let i = 0; i < binary.length; i++) if (binary.charCodeAt(i) > 127) return new TextDecoder().decode(Uint8Array.from(binary, c => c.charCodeAt(0)));
	return binary;
};

interface Frame {
	map: Map<string | number, unknown>;
	key: string | number | null;
	wantsKey: boolean;
}

function parseAce1(input: Uint8Array): Stage<Map<string | number, unknown>> {
	if (input.length > CODEC_LIMITS.serialized) return fail('serialized-size-limit');
	const clean = bytesToBinary(input).replace(/[\x00-\x20\x7f]/g, '');
	if (!clean.startsWith('^1')) return fail('unsupported-serializer');
	const tokens: Tok[] = [];
	let position = 0;
	let depth = 0;
	let root = false;
	let closed = false;
	let terminated = false;
	while (position < clean.length) {
		const begin = clean.indexOf('^', position);
		if (begin < 0 || begin + 1 >= clean.length) break;
		const next = clean.indexOf('^', begin + 2);
		const control = clean.slice(begin, begin + 2);
		const text = clean.slice(begin + 2, next < 0 ? clean.length : next);
		position = next < 0 ? clean.length : next;
		tokens.push({ control, text });
		if (tokens.length > CODEC_LIMITS.tokens) return fail('serializer-token-limit');
		if (tokens.length === 1) {
			if (control !== '^1') return fail('unsupported-serializer');
		} else if (terminated) return fail('serializer-trailing-data');
		else if (!root) {
			if (control !== '^T') return fail('expected-root-table');
			root = true;
			depth = 1;
		} else if (closed) {
			if (control !== '^^') return fail('serializer-multiple-values');
			terminated = true;
		} else if (control === '^T') {
			if (++depth > CODEC_LIMITS.tableDepth) return fail('serializer-depth-limit');
		} else if (control === '^t') {
			if (--depth === 0) closed = true;
		} else if (control === '^^') return fail('serializer-unclosed-table');
		else if (!['^S', '^N', '^F', '^f', '^B', '^b', '^Z'].includes(control)) return fail('invalid-serializer-token');
	}
	if (!terminated) return fail('serializer-missing-terminator');
	const stack: Frame[] = [];
	let result: unknown;
	const attach = (value: unknown): void => {
		if (!stack.length) {
			result = value;
			return;
		}
		const frame = stack[stack.length - 1];
		if (value === null || (typeof value === 'number' && Number.isNaN(value) && frame.wantsKey)) throw new Error('deserialize-error');
		if (frame.wantsKey) {
			if (typeof value !== 'string' && typeof value !== 'number') throw new Error('deserialize-error');
			frame.key = value;
			frame.wantsKey = false;
		} else {
			frame.map.set(frame.key as string | number, value);
			frame.wantsKey = true;
		}
	};
	try {
		for (let i = 1; i < tokens.length - 1; i++) {
			const { control, text } = tokens[i];
			if (control === '^T') {
				const map = new Map<string | number, unknown>();
				attach(map);
				stack.push({ map, key: null, wantsKey: true });
			} else if (control === '^t') {
				const frame = stack.pop();
				if (!frame || !frame.wantsKey) throw new Error('deserialize-error');
			} else {
				let value: unknown;
				if (control === '^S') value = aceUnescape(text);
				else if (control === '^N') value = aceNumber(text, true);
				else if (control === '^F') {
					const exponent = tokens[++i];
					if (exponent?.control !== '^f') throw new Error('deserialize-error');
					const m = aceNumber(text);
					const e = aceNumber(exponent.text);
					if (m === undefined || e === undefined) throw new Error('deserialize-error');
					value = m * 2 ** e;
				} else if (control === '^B') value = true;
				else if (control === '^b') value = false;
				else if (control === '^Z') value = null;
				else throw new Error('deserialize-error');
				if (value === undefined) throw new Error('deserialize-error');
				attach(value);
			}
		}
		if (stack.length || !(result instanceof Map)) return fail('deserialize-error');
		return { ok: true, value: result };
	} catch (error) {
		if (error instanceof Error && error.message === 'deserialize-error') return fail('deserialize-error');
		throw error;
	}
}

// A Lua table (Map) as plain data: keys exactly 1..n make an array; anything else an object with string keys.
function toPlain(table: Map<string | number, unknown>, depth: number, budget: { nodes: number }): RmValue {
	if (depth > CODEC_LIMITS.tableDepth || ++budget.nodes > 100000) throw new Error('data-limit');
	const convert = (v: unknown): RmValue => {
		if (v instanceof Map) return toPlain(v, depth + 1, budget);
		if (typeof v === 'string') return utf8(v);
		if (typeof v === 'number' && !Number.isFinite(v)) throw new Error('non-finite-number');
		return v as RmValue;
	};
	const keys = [...table.keys()];
	// Only positive whole numbers are table positions; any other number as a key is not rotation data.
	if (keys.some(k => typeof k === 'number' && !(Number.isInteger(k) && k >= 1))) throw new Error('non-portable-key');
	const dense = keys.every(k => typeof k === 'number' && Number.isInteger(k) && k >= 1 && k <= keys.length);
	if (dense && new Set(keys).size === keys.length) {
		const list: RmValue[] = new Array(keys.length);
		for (const [k, v] of table) list[(k as number) - 1] = convert(v);
		return list;
	}
	const out: { [key: string]: RmValue } = Object.create(null);
	for (const [k, v] of table) out[typeof k === 'number' ? String(k) : utf8(k)] = convert(v);
	return out;
}

/** Reads an export string: pasted text with any wrapping, optionally led by RM1:. */
export function decodeRmExport(text: string): DecodeResult {
	if (typeof text !== 'string') return fail('expected-import-string');
	if (text.length > CODEC_LIMITS.input) return fail('import-size-limit');
	let payload = stripExportWhitespace(text);
	const prefix = /^RM\d+:/.exec(payload)?.[0] ?? null;
	if (prefix && prefix !== 'RM1:') return fail('unsupported-transport-version');
	if (prefix) payload = payload.slice(prefix.length);
	const compressed = fromBase64(payload);
	if (!compressed.ok) return compressed;
	const raw = inflate(compressed.value);
	if (!raw.ok) return raw;
	const parsed = parseAce1(raw.value);
	if (!parsed.ok) return parsed;
	let value: RmValue;
	try {
		value = toPlain(parsed.value, 0, { nodes: 0 });
	} catch (error) {
		return fail(error instanceof Error ? error.message : 'invalid-structure');
	}
	const envelope = !Array.isArray(value) && value !== null && typeof value === 'object' && (value as { format?: unknown }).format === 'RotationMaster';
	if (prefix && !envelope) return fail('envelope-marker-mismatch');
	return { ok: true, value, envelope };
}

// ---- writing -----------------------------------------------------------------------------------------------------------
function serialize(root: RmValue): Stage<Uint8Array> {
	const bytes: number[] = [];
	let tokens = 0;
	let reason: string | null = null;
	const byte = (b: number): boolean => {
		if (bytes.length >= CODEC_LIMITS.serialized) {
			reason = 'serialized-size-limit';
			return false;
		}
		bytes.push(b);
		return true;
	};
	const token = (kind: number): boolean => {
		if (++tokens > CODEC_LIMITS.tokens) {
			reason = 'serializer-token-limit';
			return false;
		}
		return byte(94) && byte(kind);
	};
	const text = (value: string): boolean => {
		for (let i = 0; i < value.length; i++) if (!byte(value.charCodeAt(i))) return false;
		return true;
	};
	const writeString = (value: string): boolean => {
		if (!token(83)) return false;
		for (const code of new TextEncoder().encode(value)) {
			let escape: number | undefined;
			if (code <= 32) escape = code === 30 ? 122 : code + 64;
			else if (code === 94) escape = 125;
			else if (code === 126) escape = 124;
			else if (code === 127) escape = 123;
			if (escape !== undefined) {
				if (!byte(126) || !byte(escape)) return false;
			} else if (!byte(code)) return false;
		}
		return true;
	};
	type Task = { value: RmValue | undefined; depth: number; key?: string | number; close?: boolean };
	if (!token(49)) return fail(reason ?? 'serialize-error');
	const stack: Task[] = [{ value: root, depth: 1 }];
	while (stack.length) {
		const task = stack.pop() as Task;
		if (task.close) {
			if (!token(116)) return fail(reason ?? 'serialize-error');
			continue;
		}
		if (task.key !== undefined) {
			// a table key: numbers (list positions) or strings
			const ok = typeof task.key === 'number' ? token(78) && text(String(task.key)) : writeString(task.key);
			if (!ok) return fail(reason ?? 'serialize-error');
			continue;
		}
		const value = task.value;
		if (value === null || value === undefined) return fail('non-roundtrippable-nil');
		if (typeof value === 'string') {
			if (!writeString(value)) return fail(reason ?? 'serialize-error');
		} else if (typeof value === 'number') {
			if (!Number.isFinite(value)) return fail('non-finite-number');
			const spelling = value === 0 && 1 / value === -Infinity ? '-0' : String(value);
			if (!token(78) || !text(spelling)) return fail(reason ?? 'serialize-error');
		} else if (typeof value === 'boolean') {
			if (!token(value ? 66 : 98)) return fail(reason ?? 'serialize-error');
		} else {
			if (task.depth > CODEC_LIMITS.tableDepth) return fail('serializer-depth-limit');
			if (!token(84)) return fail(reason ?? 'serialize-error');
			stack.push({ value: undefined, depth: 0, close: true });
			const entries: Array<[string | number, RmValue]> = Array.isArray(value)
				? value.map((v, i): [number, RmValue] => [i + 1, v])
				: Object.entries(value).filter(([, v]) => v !== undefined);
			// Numeric keys first, then string keys by byte order, so the same data always writes the same text.
			entries.sort((a, b) => (typeof a[0] === typeof b[0] ? (a[0] < b[0] ? -1 : a[0] > b[0] ? 1 : 0) : typeof a[0] === 'number' ? -1 : 1));
			for (let i = entries.length - 1; i >= 0; i--) {
				stack.push({ value: entries[i][1], depth: task.depth + 1 });
				stack.push({ value: undefined, depth: 0, key: entries[i][0] });
			}
		}
	}
	if (!token(94)) return fail(reason ?? 'serialize-error');
	return { ok: true, value: Uint8Array.from(bytes) };
}

/**
 * Writes an export string. The root must be a table. An object whose `format` is "RotationMaster" gets the RM1:
 * prefix; a bare rotation body (what the game's older versions export) gets none. compress (default true) picks LZW
 * when it is smaller than storing.
 */
export function encodeRmExport(value: RmValue, { compress = true }: { compress?: boolean } = {}): EncodeResult {
	if (value === null || typeof value !== 'object') return fail('expected-root-table');
	const serialized = serialize(value);
	if (!serialized.ok) return serialized;
	const stored = new Uint8Array(serialized.value.length + 1);
	stored[0] = 1;
	stored.set(serialized.value, 1);
	let stream: Uint8Array = stored;
	if (compress) {
		const lzw = deflateLzw(serialized.value);
		if (lzw && lzw.length < stream.length) stream = lzw;
	}
	const prefix = !Array.isArray(value) && (value as { format?: unknown }).format === 'RotationMaster' ? 'RM1:' : '';
	return { ok: true, text: prefix + toBase64(stream) };
}
