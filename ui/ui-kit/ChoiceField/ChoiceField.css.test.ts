// @vitest-environment node
import { readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { compile, optimize } from '@tailwindcss/node';
import { beforeAll, describe, expect, it } from 'vitest';

const styles = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../styles');

let css = '';

beforeAll(async () => {
	const compiler = await compile(readFileSync(path.join(styles, 'main.css'), 'utf8'), { base: styles, onDependency: () => {} });
	css = optimize(compiler.build([]), { minify: false }).code;
});

const declarations = (selector: string): Record<string, string> => {
	const found: Record<string, string> = {};
	for (const [, selectors, body] of css.matchAll(/([^{}]+)\{([^{}]*)\}/g)) {
		if (selectors.trim() !== selector) continue;
		for (const declaration of body.split(';')) {
			const colon = declaration.indexOf(':');
			if (colon > 0) found[declaration.slice(0, colon).trim()] = declaration.slice(colon + 1).trim();
		}
	}
	return found;
};

// happy-dom neither lays out nor applies `@layer` rules, so the width a start-aligned column leaves the
// select is read off the compiled rules rather than a rendered box.
describe('ChoiceField.css', () => {
	it('holds the select look to a minimum width, so a start-aligned column cannot shrink it to its label', () => {
		const select = declarations('.ui-choice-field-trigger[data-variant="select"]');
		expect(select['min-width']).toBe('calc(var(--spacing) * 28)');
		expect(select.width).toBe('100%');
		expect(select.contain).toBe('inline-size');
	});

	it('sizes an inline select to its label, with no minimum', () => {
		const inline = declarations('.ui-field[data-layout="inline"] > .ui-choice-field-trigger');
		expect(inline.width).toBe('auto');
		expect(inline['min-width']).toBe('auto');
		expect(inline.contain).toBe('none');
	});
});
