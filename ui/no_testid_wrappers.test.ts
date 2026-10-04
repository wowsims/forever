// @vitest-environment node

import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { describe, expect, it } from 'vitest';

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const SCAN_DIR = 'ui';

const TAG_OPEN_RE = /<(div|span|section)(?=[\s>/])/g;

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

// A whole-line comment (`//`, `*`, `/*`) is blanked before scanning, so a mention of a tag name in
// prose never counts as a hit; every real hit here is an element in JSX.
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

// Finds the end of the opening tag starting at `start` (the `<` of `<div`, `<span` or `<section`),
// tracking brace and string depth so a `data-testid={toggle.testId}` expression or a quoted `>`
// inside an attribute value never ends the tag early. Returns the index just past the tag's `>`,
// and the raw attribute text between the tag name and that `>`.
function readTag(text: string, start: number): { end: number; attrs: string } | null {
	const nameMatch = /^<(div|span|section)/.exec(text.slice(start));
	if (!nameMatch) return null;
	let i = start + nameMatch[0].length;
	const attrsStart = i;
	let braceDepth = 0;
	let quote: string | null = null;
	for (; i < text.length; i++) {
		const ch = text[i];
		if (quote) {
			if (ch === quote) quote = null;
			continue;
		}
		if (ch === '"' || ch === "'") {
			quote = ch;
			continue;
		}
		if (ch === '{') {
			braceDepth++;
			continue;
		}
		if (ch === '}') {
			braceDepth--;
			continue;
		}
		if (braceDepth === 0 && ch === '>') {
			const attrs = text.slice(attrsStart, i).replace(/\/$/, '');
			return { end: i + 1, attrs };
		}
	}
	return null;
}

export interface TestidWrapperHit {
	file: string;
	line: number;
	tag: string;
}

// An opening tag whose only attributes are `data-testid` and, optionally, `key` exists only to
// carry the test id: tests must find it by role, label or text, or the test id must move onto the
// element that actually carries the identity.
function isTestidOnly(attrs: string): boolean {
	const tokens = attrs.trim().length === 0 ? [] : attrs.trim().split(/\s+(?=[\w-]+=)/);
	if (tokens.length === 0) return false;
	let hasTestid = false;
	for (const token of tokens) {
		if (/^data-testid=("[^"]*"|\{[^}]*\})$/.test(token)) {
			hasTestid = true;
			continue;
		}
		if (/^key=\{[^}]*\}$/.test(token)) continue;
		return false;
	}
	return hasTestid;
}

export function findTestidWrappers(root: string): TestidWrapperHit[] {
	const files: string[] = [];
	walk(path.join(root, SCAN_DIR), files);

	const hits: TestidWrapperHit[] = [];
	for (const file of files) {
		const relPath = path.relative(root, file);
		const text = blankCommentLines(fs.readFileSync(file, 'utf8'));

		TAG_OPEN_RE.lastIndex = 0;
		let match: RegExpExecArray | null;
		while ((match = TAG_OPEN_RE.exec(text))) {
			const tagStart = match.index;
			const tag = readTag(text, tagStart);
			if (!tag) continue;
			if (isTestidOnly(tag.attrs)) hits.push({ file: relPath, line: lineOf(text, tagStart), tag: text.slice(tagStart, tag.end) });
			TAG_OPEN_RE.lastIndex = tag.end;
		}
	}

	hits.sort((a, b) => (a.file === b.file ? a.line - b.line : a.file < b.file ? -1 : 1));
	return hits;
}

function formatHits(rows: TestidWrapperHit[]): string {
	return rows
		.map(
			r =>
				`${r.file}:${r.line} ${r.tag} -- no element exists only to carry a test id; find it by role, label or text, or move the test id onto the element that carries the identity`,
		)
		.join('\n');
}

describe('testid-only wrappers', () => {
	it('has no div, span or section whose only attributes are data-testid and optionally key', () => {
		const hits = findTestidWrappers(ROOT);
		expect(hits, formatHits(hits)).toEqual([]);
	});

	it('flags a multi-line tag and a key+testid tag, and ignores a className or role tag', () => {
		const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'testid-wrappers-'));
		try {
			fs.mkdirSync(path.join(tmp, 'ui/features'), { recursive: true });
			fs.writeFileSync(
				path.join(tmp, 'ui/features/x.tsx'),
				[
					'export const X = () => (',
					'\t<div>',
					'\t\t<div',
					'\t\t\tdata-testid="wrapper-only">child</div>',
					'\t\t{items.map(item => (',
					'\t\t\t<span key={item.id} data-testid={item.testId}>{item.value}</span>',
					'\t\t))}',
					'\t\t<div className="ui-row" data-testid="has-class">kept</div>',
					'\t\t<section role="status" data-testid="has-role">kept</section>',
					'\t\t// <div data-testid="in-a-comment">not real</div>',
					'\t</div>',
					');',
				].join('\n'),
			);

			const hits = findTestidWrappers(tmp);
			expect(hits.map(h => h.tag)).toEqual(['<div\n\t\t\tdata-testid="wrapper-only">', '<span key={item.id} data-testid={item.testId}>']);
		} finally {
			fs.rmSync(tmp, { recursive: true, force: true });
		}
	});
});
