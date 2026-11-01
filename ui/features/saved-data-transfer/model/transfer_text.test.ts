import { describe, expect, it } from 'vitest';

import { decodeSavedData, encodeSavedData, MAX_TEXT_LENGTH, type SavedDataExport, SavedDataTextError } from './transfer_text';

const EXPORT: SavedDataExport = {
	type: 'talents',
	spec: 'SpecDpsWarrior',
	entries: { Raid: { talentsString: '2305' }, 'Früh 火': { talentsString: '1' } },
};

const deflate = async (text: string) => {
	const stream = new Blob([text]).stream().pipeThrough(new CompressionStream('deflate'));
	const bytes = new Uint8Array(await new Response(stream).arrayBuffer());
	return btoa(String.fromCharCode(...bytes));
};

const problemOf = (text: string) =>
	decodeSavedData(text).then(
		() => null,
		(error: unknown) => (error instanceof SavedDataTextError ? error.problem : error),
	);

describe('saved data text', () => {
	it('round-trips an export, non-ASCII names included', async () => {
		const text = await encodeSavedData(EXPORT);

		expect(text.startsWith('wowsims-saved:v1:')).toBe(true);
		expect(await decodeSavedData(text)).toEqual(EXPORT);
	});

	it('round-trips more data than one String.fromCharCode call could take', async () => {
		const entries = Object.fromEntries(Array.from({ length: 4000 }, (_, i) => [`Set ${i}`, { talentsString: `${Math.sin(i)}${Math.cos(i)}` }]));

		expect((await decodeSavedData(await encodeSavedData({ ...EXPORT, entries }))).entries).toEqual(entries);
	});

	it('reads text with line breaks around it', async () => {
		expect(await decodeSavedData(`\r\n  ${await encodeSavedData(EXPORT)}\n`)).toEqual(EXPORT);
	});

	it.each([
		['no prefix', 'hello'],
		['JSON', '{"type":"talents"}'],
		['base64 that is not deflate', 'wowsims-saved:v1:aGVsbG8='],
		['characters base64 does not have', 'wowsims-saved:v1:***'],
		['nothing after the prefix', 'wowsims-saved:v1:'],
	])('refuses %s as unreadable', async (_, text) => {
		expect(await problemOf(text)).toBe('unreadable');
	});

	it('refuses cut-off text as unreadable', async () => {
		const text = await encodeSavedData(EXPORT);

		expect(await problemOf(text.slice(0, text.length - 12))).toBe('unreadable');
	});

	it('names an unknown format version', async () => {
		const text = await encodeSavedData(EXPORT);

		expect(await problemOf(text.replace('v1', 'v2'))).toBe('version');
	});

	it.each([
		['a list', []],
		['no entries', { type: 'talents', spec: 'SpecDpsWarrior' }],
		['entries as a list', { type: 'talents', spec: 'SpecDpsWarrior', entries: [] }],
		['an unknown type', { type: 'weapons', spec: 'SpecDpsWarrior', entries: {} }],
		['a type on the prototype chain', { type: 'constructor', spec: 'SpecDpsWarrior', entries: {} }],
		['a numeric spec', { type: 'talents', spec: 1, entries: {} }],
	])('refuses a payload of %s as unreadable', async (_, payload) => {
		expect(await problemOf(`wowsims-saved:v1:${await deflate(JSON.stringify(payload))}`)).toBe('unreadable');
	});

	it('refuses text past the length cap before reading it', async () => {
		expect(await problemOf(`wowsims-saved:v1:${'A'.repeat(MAX_TEXT_LENGTH)}`)).toBe('too_large');
	});

	it('stops inflating at the size cap', async () => {
		const bomb = await deflate(`{"pad":"${'0'.repeat(25 * 1024 * 1024)}"}`);

		expect(bomb.length).toBeLessThan(MAX_TEXT_LENGTH);
		expect(await problemOf(`wowsims-saved:v1:${bomb}`)).toBe('too_large');
	});

	it('keeps a __proto__ name as data, off the prototype chain', async () => {
		const payload = `{"type":"talents","spec":"SpecDpsWarrior","entries":{"__proto__":{"polluted":true}}}`;

		const { entries } = await decodeSavedData(`wowsims-saved:v1:${await deflate(payload)}`);

		expect(Object.keys(entries)).toEqual(['__proto__']);
		expect(({} as Record<string, unknown>).polluted).toBeUndefined();
		expect(Object.getPrototypeOf(entries)).toBe(Object.prototype);
	});
});
