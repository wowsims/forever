// @vitest-environment node
import { existsSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { describe, expect, it } from 'vitest';

import { SPEC_THEMES } from './spec_themes';

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../..');

describe('SPEC_THEMES', () => {
	it('lists every spec theme once', () => {
		expect(SPEC_THEMES).toHaveLength(18);
		expect(new Set(SPEC_THEMES.map(theme => theme.className)).size).toBe(18);
	});

	it('points every background image at a file under assets/img', () => {
		for (const theme of SPEC_THEMES) {
			expect(theme.backgroundImage, theme.key).toMatch(/^\/forever\/assets\/img\//);
			const file = path.join(repoRoot, theme.backgroundImage!.replace(/^\/forever\//, ''));
			expect(existsSync(file), file).toBe(true);
		}
	});
});
