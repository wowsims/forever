// @vitest-environment node

import path from 'node:path';
import { fileURLToPath } from 'node:url';

// oxlint-disable-next-line import/extensions
import {
	collectTokens,
	findDefaultCss,
	findNonCanonical,
	findUnevenSpacing,
	findVarInClass,
	loadDesignSystemCached,
} from '@testing/tailwind/canonical-classes.mjs';
import { beforeAll, describe, expect, it } from 'vitest';

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');

function formatOffenders(rows: Awaited<ReturnType<typeof findNonCanonical>>): string {
	return rows.map(r => `${r.file}:${r.line} ${r.from} → ${r.to} -- run \`node ui/testing/tailwind/canonical-classes.mjs --write\``).join('\n');
}

function formatVarOffenders(rows: Array<{ file: string; line: number; token: string }>): string {
	return rows.map(r => `${r.file}:${r.line} ${r.token} -- use util-(--x), a named @theme inline token, or a ui-* class`).join('\n');
}

function formatSpacingOffenders(rows: Awaited<ReturnType<typeof findUnevenSpacing>>): string {
	return rows.map(r => `${r.file}:${r.line} ${r.token} -- snap to the nearest even step (ties round down), or \`px\` for a hairline`).join('\n');
}

let offenders: Awaited<ReturnType<typeof findNonCanonical>>;
let varOffenders: ReturnType<typeof findVarInClass>;
let spacingOffenders: Awaited<ReturnType<typeof findUnevenSpacing>>;
let resolvedTokens: number;

beforeAll(async () => {
	offenders = await findNonCanonical(ROOT);
	varOffenders = findVarInClass(ROOT);
	spacingOffenders = await findUnevenSpacing(ROOT);
	const designSystem = await loadDesignSystemCached(findDefaultCss(ROOT));
	const tokens = [...new Set(collectTokens(ROOT).map((occurrence: { token: string }) => occurrence.token))];
	resolvedTokens = designSystem.candidatesToCss(tokens).filter(Boolean).length;
}, 40000);

describe('canonical Tailwind classes', () => {
	it('resolves class tokens from ui/ against the loaded design system', () => {
		expect(resolvedTokens).toBeGreaterThan(0);
	});

	it('has no non-canonical Tailwind class tokens in ui/', () => {
		expect(offenders, formatOffenders(offenders)).toEqual([]);
	});

	it('has no var(...) inside a Tailwind class token in ui/', () => {
		expect(varOffenders, formatVarOffenders(varOffenders)).toEqual([]);
	});

	it('has no uneven spacing-scale step in ui/', () => {
		expect(spacingOffenders, formatSpacingOffenders(spacingOffenders)).toEqual([]);
	});
});
