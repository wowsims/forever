// @vitest-environment node
import { existsSync, readdirSync, readFileSync, statSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

import { describe, expect, it } from 'vitest';

const UI = resolve(dirname(fileURLToPath(import.meta.url)), '../../..');
const ALIASES = ['sim', 'ui-kit', 'features', 'app', 'specs', 'worker', 'generated', 'i18n', 'testing'];

// Runtime imports only: `import type` never reaches the browser, and a bare package is out of scope.
const IMPORT = /(?:import|export)\s+(?!type\s)(?:[^'";]*?\sfrom\s+)?['"]([^'"]+)['"]|import\(\s*['"]([^'"]+)['"]\s*\)/g;

const sources = new Map<string, string>();
const read = (file: string) => {
	if (!sources.has(file)) sources.set(file, readFileSync(file, 'utf8'));
	return sources.get(file)!;
};

const toFile = (base: string) =>
	[base, `${base}.ts`, `${base}.tsx`, `${base}/index.ts`, `${base}/index.tsx`].find(path => existsSync(path) && statSync(path).isFile());

const resolveSpecifier = (from: string, specifier: string): string | undefined => {
	if (specifier.startsWith('.')) return toFile(resolve(dirname(from), specifier));
	const alias = ALIASES.find(name => specifier.startsWith(`@${name}/`));
	return alias ? toFile(resolve(UI, alias, specifier.slice(alias.length + 2))) : undefined;
};

const runtimeGraph = (entries: Array<string>) => {
	const importedBy = new Map<string, string>(entries.map(entry => [entry, entry]));
	const queue = [...entries];
	for (let file = queue.shift(); file; file = queue.shift()) {
		for (const match of read(file).matchAll(IMPORT)) {
			const next = resolveSpecifier(file, match[1] ?? match[2]);
			if (next && /\.tsx?$/.test(next) && !importedBy.has(next)) {
				importedBy.set(next, file);
				queue.push(next);
			}
		}
	}
	return importedBy;
};

const filesUnder = (dir: string, keep: (path: string) => boolean) =>
	readdirSync(dir, { recursive: true, encoding: 'utf8' })
		.filter(keep)
		.map(path => resolve(dir, path));

const STORYBOOK_ENTRIES = [
	resolve(UI, '../.storybook/preview.tsx'),
	...filesUnder(UI, path => path.endsWith('.stories.tsx')),
	...filesUnder(resolve(UI, 'testing/storybook'), path => /\.tsx?$/.test(path) && !/\.test\./.test(path)),
];

describe('Storybook modules', () => {
	it('never load a module that imports vitest, which throws in the Storybook dev server', () => {
		const graph = runtimeGraph(STORYBOOK_ENTRIES);
		const offenders = [...graph.keys()]
			.filter(file => /from ['"]vitest['"]/.test(read(file)))
			.map(file => `${file.slice(UI.length + 1)} (imported by ${graph.get(file)!.slice(UI.length + 1)})`);

		expect(STORYBOOK_ENTRIES.length).toBeGreaterThan(100);
		expect(offenders).toEqual([]);
	});
});
