import { readFileSync } from 'node:fs';
import { join } from 'node:path';

import { describe, expect, it } from 'vitest';

import { getTargetInputI18nKey, protoStatNameI18nKeys } from './entity_mapping';

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

// The labels of the inputs the Dynamic Adds boss declares.
const dynamicAddsInputLabels = (): Array<string> =>
	[...readFileSync(join(ROOT, 'sim', 'encounters', 'dynamic_adds_ai.go'), 'utf8').matchAll(/Label:\s+"([^"]+)"/g)].map(([, label]) => label);

describe('getTargetInputI18nKey', () => {
	it('gives every Dynamic Adds input a label and tooltip in settings_tab.encounter.target_inputs', () => {
		const inputs = (
			JSON.parse(readFileSync(join(ROOT, 'assets', 'locales', 'en', 'translation.json'), 'utf8')) as {
				settings_tab: { encounter: { target_inputs: Record<string, { label?: string; tooltip?: string }> } };
			}
		).settings_tab.encounter.target_inputs;
		const labels = dynamicAddsInputLabels();
		const missing = labels.filter(label => !inputs[getTargetInputI18nKey(label)]?.label || !inputs[getTargetInputI18nKey(label)]?.tooltip);

		expect(labels).toContain('Kill share %');
		expect(missing).toEqual([]);
	});
});
