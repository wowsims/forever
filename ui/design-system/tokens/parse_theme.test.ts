// @vitest-environment node
import { readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { describe, expect, it } from 'vitest';

import { parseThemeVars, resolvePxValue } from './parse_theme';

const themeDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../styles/theme');
const read = (file: string) => readFileSync(path.join(themeDir, file), 'utf8');

describe('parseThemeVars', () => {
	it('finds every --color- declaration in colors.css', () => {
		const css = read('colors.css');
		const declared = css.split('\n').filter(line => /^\s*--color-[a-z0-9-]+:/.test(line)).length;
		const colors = parseThemeVars(css).filter(token => token.name.startsWith('--color-'));
		expect(colors).toHaveLength(declared);
		expect(colors.every(token => token.prefix === 'color')).toBe(true);
	});

	it('flags tokens from a static inline block', () => {
		const tokens = parseThemeVars(read('colors.css'));
		expect(tokens.find(token => token.name === '--color-primary')?.inline).toBe(false);
		expect(tokens.find(token => token.name === '--color-primary-hover')?.inline).toBe(true);
	});

	it('reads a multi-line value as one entry', () => {
		const progress = parseThemeVars(read('effects.css')).filter(token => token.name === '--background-image-progress');
		expect(progress).toHaveLength(1);
		expect(progress[0].value).toMatch(/^linear-gradient\( 90deg, var\(--color-success\) 0%, .* var\(--color-success\) 100% \)$/);
	});

	it('skips namespace resets', () => {
		const names = parseThemeVars(read('effects.css')).map(token => token.name);
		expect(names).not.toContain('--radius-*');
		expect(names).toContain('--radius-xs');
	});
});

describe('resolvePxValue', () => {
	it('resolves a --spacing(n) step at the 14px root', () => {
		expect(resolvePxValue('--spacing(18)', 14)).toBe(63);
	});
});
