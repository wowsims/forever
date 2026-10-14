import { readFileSync } from 'node:fs';
import { join } from 'node:path';

import { describe, expect, it } from 'vitest';

import { protoStatNameI18nKeys } from './entity_mapping';

const ROOT = join(import.meta.dirname, '..', '..');

// The names the sim writes for its stats (`Stat.StatName`), which the combat log and the random suffixes carry.
const goStatNames = (): Array<string> =>
	[...readFileSync(join(ROOT, 'sim', 'core', 'stats', 'stats.go'), 'utf8').matchAll(/case \w+:\n\t\treturn "(\w+)"/g)]
		.map(([, name]) => name)
		.filter(name => name !== 'none');

describe('protoStatNameI18nKeys', () => {
	it('gives every stat name the sim writes a key in common.stats', () => {
		const stats = (
			JSON.parse(readFileSync(join(ROOT, 'assets', 'locales', 'en', 'translation.json'), 'utf8')) as { common: { stats: Record<string, string> } }
		).common.stats;
		const missing = goStatNames().filter(name => stats[protoStatNameI18nKeys[name] ?? name.toLowerCase()] === undefined);

		expect(missing).toEqual([]);
	});

	it('reads the stat names from the sim, so a broken scan cannot pass', () => {
		expect(goStatNames().length).toBeGreaterThan(30);
	});
});
