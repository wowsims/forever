import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';

import i18n from '@i18n/config';
import { render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { LocaleHtml } from './LocaleHtml';

const html = (markup: string) => render(<LocaleHtml html={markup} />).container.innerHTML;

describe('LocaleHtml', () => {
	it.each([
		['a locale tag as a real element carrying its own attributes', '<span className="bold">Batch</span> plain', '<span class="bold">Batch</span> plain'],
		[
			'void and repeated tags apart',
			'<span className="bold">a</span><br /><span className="text-brand">b</span>',
			'<span class="bold">a</span><br><span class="text-brand">b</span>',
		],
		[
			'anchors with href and target',
			'<a href="https://raidbots.com" target="_blank" rel="noopener noreferrer">R</a>',
			'<a href="https://raidbots.com" target="_blank" rel="noopener noreferrer">R</a>',
		],
		['nested block markup', '<p>one</p><ul><li>two</li></ul>', '<p>one</p><ul><li>two</li></ul>'],
		['a separator-bearing sentence intact', 'Stats: haste. Then <span className="bold">more</span>', 'Stats: haste. Then <span class="bold">more</span>'],
	])('renders %s', (_, markup, expected) => {
		expect(html(markup)).toBe(expected);
	});

	it('logs no missing key for markup that is already translated', () => {
		const { logger } = (i18n as unknown as { translator: { logger: { debug: boolean } } }).translator;
		const log = vi.spyOn(console, 'log').mockImplementation(() => undefined);
		const debug = logger.debug;
		logger.debug = true;
		try {
			html('<span className="bold">Batch</span> plain');
			expect(log.mock.calls.flat().join(' ')).not.toContain('missingKey');
		} finally {
			logger.debug = debug;
			log.mockRestore();
		}
	});
});

const LOCALES = join(process.cwd(), 'assets/locales');
const HTML_TAG = /<[a-zA-Z/][^>]*>/;

// Keys that carry markup *and* interpolate. LocaleHtml and Trans resolve the string before parsing it, so a
// value that looks like a tag is re-parsed — it renders as text rather than markup, but it still renders
// wrong. Every value spliced into these is a number, a duration, a stat or resource name or the engine's
// reason text, none of which carries a tag.
const INTERPOLATING_HTML_KEYS = [
	'bulk_tab.progress.iterations_complete',
	'results_tab.details.damage_result.damage_amount',
	'results_tab.details.damage_result.health_amount',
	'results_tab.details.logs.lines.aura_stacks',
	'results_tab.details.logs.lines.auto_delay',
	'results_tab.details.logs.lines.cast_began',
	'results_tab.details.logs.lines.cast_cancelled',
	'results_tab.details.logs.lines.cast_failed',
	'results_tab.details.logs.lines.health_lost',
	'results_tab.details.logs.lines.health_recovered',
	'results_tab.details.logs.lines.pushback_casting',
	'results_tab.details.logs.lines.pushback_channeling',
	'results_tab.details.logs.lines.resource_gained',
	'results_tab.details.logs.lines.resource_spent',
	'results_tab.details.logs.lines.spell_queued',
	'results_tab.details.logs.lines.stats_gained',
	'results_tab.details.logs.lines.stats_lost',
];

const flatten = (value: unknown, path = ''): Array<[string, string]> =>
	typeof value === 'string'
		? [[path, value]]
		: value && typeof value === 'object'
			? Object.entries(value).flatMap(([key, child]) => flatten(child, path ? `${path}.${key}` : key))
			: [];

describe('locale markup', () => {
	const locales = readdirSync(LOCALES, { withFileTypes: true })
		.filter(entry => entry.isDirectory())
		.map(entry => entry.name);

	it.each(locales)('%s interpolates nothing into markup that is not already accounted for', locale => {
		const dir = join(LOCALES, locale);
		const offenders = readdirSync(dir)
			.filter(file => file.endsWith('.json'))
			.flatMap(file => flatten(JSON.parse(readFileSync(join(dir, file), 'utf-8'))))
			.filter(([, text]) => HTML_TAG.test(text) && text.includes('{{'))
			.map(([key]) => key)
			.filter(key => !INTERPOLATING_HTML_KEYS.includes(key));

		expect(offenders).toEqual([]);
	});
});
