// @vitest-environment node

import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { describe, expect, it } from 'vitest';

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const SCAN_DIRS = ['ui/features', 'ui/app', 'ui/specs'];

const ANCHOR_RE = /<a(?=[\s>/])/g;
const BUTTON_RE = /<button\b/g;

function isProductionTsx(relPath: string): boolean {
	return relPath.endsWith('.tsx') && !relPath.endsWith('.test.tsx') && !relPath.endsWith('.stories.tsx');
}

function walk(dir: string, out: string[]): void {
	if (!fs.existsSync(dir)) return;
	for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
		const full = path.join(dir, entry.name);
		if (entry.isDirectory()) walk(full, out);
		else if (isProductionTsx(entry.name)) out.push(full);
	}
}

// A whole-line comment (`//`, `*`, `/*`) is blanked before scanning, so a mention of `<a>`/`<button`
// in prose never counts as a hit; every real hit here is an element in JSX.
function blankCommentLines(text: string): string {
	return text
		.split('\n')
		.map(line => (/^\s*(\/\/|\*|\/\*)/.test(line) ? '' : line))
		.join('\n');
}

function lineOf(text: string, index: number): number {
	let line = 1;
	for (let i = 0; i < index; i++) if (text[i] === '\n') line++;
	return line;
}

export interface RawControlHit {
	file: string;
	line: number;
	token: string;
}

export function findRawControls(root: string): RawControlHit[] {
	const files: string[] = [];
	for (const dir of SCAN_DIRS) walk(path.join(root, dir), files);

	const hits: RawControlHit[] = [];
	for (const file of files) {
		const relPath = path.relative(root, file);
		const text = blankCommentLines(fs.readFileSync(file, 'utf8'));

		for (const [re, token] of [
			[ANCHOR_RE, '<a'],
			[BUTTON_RE, '<button'],
		] as const) {
			re.lastIndex = 0;
			let match: RegExpExecArray | null;
			while ((match = re.exec(text))) hits.push({ file: relPath, line: lineOf(text, match.index), token });
		}
	}

	hits.sort((a, b) => (a.file === b.file ? a.line - b.line : a.file < b.file ? -1 : 1));
	return hits;
}

function formatHits(rows: RawControlHit[]): string {
	return rows.map(r => `${r.file}:${r.line} ${r.token} -- use a kit control (Button, WowheadAnchor) for every control in features, app and specs`).join('\n');
}

describe('raw controls', () => {
	it('has no raw <a> or <button> in ui/features, ui/app or ui/specs production .tsx', () => {
		const hits = findRawControls(ROOT);
		expect(hits, formatHits(hits)).toEqual([]);
	});

	it('finds a split-across-lines <a> and a <button>, and does not flag <abbr>/<aside> or a commented-out <a>', () => {
		const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'raw-controls-'));
		try {
			fs.mkdirSync(path.join(tmp, 'ui/features'), { recursive: true });
			fs.writeFileSync(
				path.join(tmp, 'ui/features/x.tsx'),
				[
					'export const X = () => (',
					'\t<div>',
					'\t\t<a',
					'\t\t\thref="https://example.com">link</a>',
					'\t\t<button type="button">click</button>',
					'\t\t<abbr title="World of Warcraft">WoW</abbr>',
					'\t\t<aside>note</aside>',
					'\t\t// <a> mentioned in prose, not JSX',
					'\t</div>',
					');',
				].join('\n'),
			);

			const hits = findRawControls(tmp);
			expect(hits.map(h => h.token)).toEqual(['<a', '<button']);
		} finally {
			fs.rmSync(tmp, { recursive: true, force: true });
		}
	});
});
