import type { JsonValue } from '@protobuf-ts/runtime';

export const SAVED_DATA_TYPES = ['gear', 'talents', 'rotation', 'settings', 'encounter'] as const;
export type SavedDataType = (typeof SAVED_DATA_TYPES)[number];

export interface SavedDataExport {
	type: SavedDataType;
	// The `Spec` enum name, as proto JSON writes an enum.
	spec: string;
	// `Record<name, toJson(data)>`, the form `useSavedData` stores.
	entries: Record<string, JsonValue>;
}

export type SavedDataTextProblem = 'unreadable' | 'version' | 'too_large';

export class SavedDataTextError extends Error {
	readonly problem: SavedDataTextProblem;

	constructor(problem: SavedDataTextProblem) {
		super(problem);
		this.name = 'SavedDataTextError';
		this.problem = problem;
	}
}

const VERSION = 1;
const PREFIX = `wowsims-saved:v${VERSION}:`;
const PREFIX_PATTERN = /^wowsims-saved:v(\d+):/;
// About one origin's localStorage before compression, so anything that was stored fits.
export const MAX_TEXT_LENGTH = 5 * 1024 * 1024;
const MAX_JSON_BYTES = 20 * 1024 * 1024;
// Bytes per `String.fromCharCode` call, far below the engines' argument limits.
const CHAR_CODE_CHUNK = 0x8000;

const pipeThrough = async (bytes: Uint8Array<ArrayBuffer>, stream: CompressionStream | DecompressionStream, limit = Infinity): Promise<Uint8Array> => {
	const writer = stream.writable.getWriter();
	// Corrupt input rejects these as well as the read below, which is the one that reports it.
	writer.write(bytes).catch(() => {});
	writer.close().catch(() => {});

	const reader = stream.readable.getReader();
	const chunks: Uint8Array[] = [];
	let length = 0;
	for (let read = await reader.read(); !read.done; read = await reader.read()) {
		length += read.value.length;
		if (length > limit) {
			await reader.cancel();
			throw new SavedDataTextError('too_large');
		}
		chunks.push(read.value);
	}

	const joined = new Uint8Array(length);
	let offset = 0;
	for (const chunk of chunks) {
		joined.set(chunk, offset);
		offset += chunk.length;
	}
	return joined;
};

const isRecord = (value: unknown): value is Record<string, unknown> => value !== null && typeof value === 'object' && !Array.isArray(value);

const isSavedDataExport = (value: unknown): value is SavedDataExport =>
	isRecord(value) && SAVED_DATA_TYPES.includes(value.type as SavedDataType) && typeof value.spec === 'string' && isRecord(value.entries);

export const encodeSavedData = async (data: SavedDataExport): Promise<string> => {
	const deflated = await pipeThrough(new TextEncoder().encode(JSON.stringify(data)), new CompressionStream('deflate'));
	let binary = '';
	for (let i = 0; i < deflated.length; i += CHAR_CODE_CHUNK) {
		binary += String.fromCharCode(...deflated.subarray(i, i + CHAR_CODE_CHUNK));
	}
	return PREFIX + btoa(binary);
};

// The inflated size is capped while it is read, since a few kilobytes of deflate can expand a thousandfold.
export const decodeSavedData = async (text: string): Promise<SavedDataExport> => {
	const trimmed = text.trim();
	if (trimmed.length > MAX_TEXT_LENGTH) throw new SavedDataTextError('too_large');
	const prefix = PREFIX_PATTERN.exec(trimmed);
	if (!prefix) throw new SavedDataTextError('unreadable');
	if (Number(prefix[1]) !== VERSION) throw new SavedDataTextError('version');

	let json: unknown;
	try {
		const binary = atob(trimmed.slice(prefix[0].length));
		const deflated = new Uint8Array(binary.length);
		for (let i = 0; i < binary.length; i++) deflated[i] = binary.charCodeAt(i);
		json = JSON.parse(new TextDecoder().decode(await pipeThrough(deflated, new DecompressionStream('deflate'), MAX_JSON_BYTES)));
	} catch (error) {
		throw error instanceof SavedDataTextError ? error : new SavedDataTextError('unreadable');
	}
	if (!isSavedDataExport(json)) throw new SavedDataTextError('unreadable');
	return json;
};
