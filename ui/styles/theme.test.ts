import { readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { describe, expect, it } from 'vitest';

const here = path.dirname(fileURLToPath(import.meta.url));
const css = readFileSync(path.join(here, 'theme', 'colors.css'), 'utf-8');
const colorTokensSource = readFileSync(path.join(here, '..', 'ui-kit', 'utils', 'color_tokens.ts'), 'utf-8');

const classColors = new Map<string, [number, number, number]>();
for (const m of css.matchAll(/--color-class-([a-z-]+):\s*rgb\((\d+),\s*(\d+),\s*(\d+)\)/g)) {
	classColors.set(m[1], [Number(m[2]), Number(m[3]), Number(m[4])]);
}

const foregrounds = new Map<string, string>();
for (const m of css.matchAll(/--color-class-([a-z-]+)-foreground:\s*(#fff|#000|var\(--color-white\)|var\(--color-black\))/g)) {
	foregrounds.set(m[1], m[2] === '#fff' || m[2] === 'var(--color-white)' ? '#fff' : '#000');
}

const srgbToLinear = (c: number) => {
	const v = c / 255;
	return v <= 0.03928 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4;
};

const relativeLuminance = ([r, g, b]: [number, number, number]) => 0.2126 * srgbToLinear(r) + 0.7152 * srgbToLinear(g) + 0.0722 * srgbToLinear(b);

const contrastRatio = (a: [number, number, number], b: [number, number, number]) => {
	const la = relativeLuminance(a);
	const lb = relativeLuminance(b);
	const lighter = Math.max(la, lb);
	const darker = Math.min(la, lb);
	return (lighter + 0.05) / (darker + 0.05);
};

const expectedForeground = (color: [number, number, number]) => {
	const white: [number, number, number] = [255, 255, 255];
	const black: [number, number, number] = [0, 0, 0];
	const whiteRatio = contrastRatio(color, white);
	const blackRatio = contrastRatio(color, black);
	if (whiteRatio >= 4.5) return '#fff';
	if (blackRatio >= 4.5) return '#000';
	return whiteRatio > blackRatio ? '#fff' : '#000';
};

describe('theme.css class foregrounds', () => {
	it('finds exactly nine class colour/foreground pairs, one per class', () => {
		expect(classColors.size).toBe(9);
		expect(foregrounds.size).toBe(9);
	});

	it.each([...classColors.entries()])('%s foreground matches WCAG contrast-based choice', (name, color) => {
		expect(foregrounds.get(name)).toBe(expectedForeground(color));
	});
});

describe('ColorToken', () => {
	it('lists every --color- token in colors.css, and nothing else', () => {
		const declared = [...css.matchAll(/^\s*--color-([a-z0-9-]+):/gm)].map(match => match[1]);
		const listed = [...colorTokensSource.matchAll(/^\s*'([a-z0-9-]+)',?$/gm)].map(match => match[1]);

		expect(listed.sort()).toEqual(declared.sort());
	});
});
