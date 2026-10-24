import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';

import { MANIFEST_KEYS, generatedManifest, manifestText } from '../scripts/manifest.mjs';

const pkg = JSON.parse(readFileSync(new URL('../package.json', import.meta.url), 'utf8'));

test('the scripts and the dependencies stay out of the manifest', () => {
	assert.equal(manifestText({ ...pkg, scripts: { other: 'x' }, devDependencies: {} }), manifestText(pkg));
});

test('the manifest carries only what VS Code reads, and says it is generated', () => {
	const manifest = generatedManifest(pkg);
	assert.deepEqual(Object.keys(manifest), [...MANIFEST_KEYS, 'x-generated']);
	assert.equal(manifest['x-generated'], 'make vscode-apl');
	assert.equal(manifest.main, './out/extension.js');
	assert.deepEqual(manifest.activationEvents, ['onWebviewPanel:wowsims-apl.editor', 'onLanguage:siml']);
});

test('the panel, the sims and Insert Pattern are offered on .apl.siml files', () => {
	const offered = new Map();
	for (const item of [...pkg.contributes.menus['editor/title'], ...pkg.contributes.menus.commandPalette, ...pkg.contributes.keybindings]) {
		offered.set(item.command, [...(offered.get(item.command) ?? []), item.when]);
	}
	for (const [command, whens] of offered) {
		assert.ok(
			whens.every(when => when.includes('\\.apl\\.siml$')),
			`${command}: ${whens.join(' | ')}`,
		);
	}
});
