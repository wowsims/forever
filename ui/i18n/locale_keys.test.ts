import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join } from 'node:path';

import { describe, expect, it } from 'vitest';

const ROOT = join(import.meta.dirname, '..', '..');
const UI = join(ROOT, 'ui');
const LOCALES = ['en'];
const DEFAULT_NAMESPACE = 'translation';

const sourceFiles = (dir: string): Array<string> =>
	readdirSync(dir).flatMap(entry => {
		const path = join(dir, entry);
		if (statSync(path).isDirectory()) return entry === 'node_modules' ? [] : sourceFiles(path);
		return /\.tsx?$/.test(entry) && !/\.test\.tsx?$/.test(entry) ? [path] : [];
	});

const bundle = (locale: string, namespace: string): unknown => JSON.parse(readFileSync(join(ROOT, 'assets', 'locales', locale, `${namespace}.json`), 'utf8'));

const resolve = (node: unknown, key: string): unknown =>
	key.split('.').reduce<unknown>((current, part) => (current && typeof current === 'object' ? (current as Record<string, unknown>)[part] : undefined), node);

/** The key itself, or its plural forms: a key asked for with a `count` lives as `key_one` and `key_other`. */
const resolveKey = (node: unknown, key: string): unknown => resolve(node, key) ?? resolve(node, `${key}_other`);

/** A literal `i18n.t('a.b')`, with the namespace from an `{ ns: 'x' }` option when one is given. */
const requests = (): Array<{ key: string; namespace: string; file: string }> => {
	const pattern = /i18n\.t\(\s*'([^']+)'([^)]*)\)/g;
	const found = new Map<string, { key: string; namespace: string; file: string }>();
	for (const file of sourceFiles(UI)) {
		const source = readFileSync(file, 'utf8');
		for (const [, key, options] of source.matchAll(pattern)) {
			if (key.includes('${')) continue;
			const namespace = options.match(/\bns:\s*'([^']+)'/)?.[1] ?? DEFAULT_NAMESPACE;
			found.set(`${namespace}:${key}`, { key, namespace, file: file.slice(ROOT.length + 1) });
		}
	}
	return [...found.values()].sort((a, b) => `${a.namespace}:${a.key}`.localeCompare(`${b.namespace}:${b.key}`));
};

describe('locale keys', () => {
	it('finds every key ui/ asks for in en', () => {
		const bundles = new Map(
			LOCALES.flatMap(locale => [...new Set(requests().map(r => r.namespace))].map(ns => [`${locale}:${ns}`, bundle(locale, ns)] as const)),
		);
		const unresolved = requests()
			.filter(({ key, namespace }) => LOCALES.some(locale => resolveKey(bundles.get(`${locale}:${namespace}`), key) === undefined))
			.map(({ key, namespace, file }) => `${namespace}:${key} (${file})`);

		expect(unresolved).toEqual([]);
	});

	it('asks for a non-trivial number of keys, so a broken scan cannot pass', () => {
		// The tree asks for ~1060 distinct keys; the floor only guards against a scan
		// that silently stops finding files.
		expect(requests().length).toBeGreaterThan(500);
	});
});
