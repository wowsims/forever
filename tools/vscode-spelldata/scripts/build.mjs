import { spawnSync } from 'node:child_process';
import { copyFileSync, mkdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { build } from 'esbuild';

import { buildExtension } from '../../vscode-common/build.mjs';
import { manifestText } from './manifest.mjs';

const source = dirname(dirname(fileURLToPath(import.meta.url)));
const installed = join(source, '..', '..', '.vscode', 'extensions', 'wowsims-spelldata');
const outputs = ['package.json', 'README.md', join('out', 'extension.js')];

async function buildInto(target) {
	const tsc = spawnSync(join(source, 'node_modules', '.bin', 'tsc'), ['--noEmit', '-p', source], { stdio: 'inherit' });
	if (tsc.status !== 0) {
		process.exit(tsc.status ?? 1);
	}

	rmSync(target, { recursive: true, force: true });
	mkdirSync(join(target, 'out'), { recursive: true });
	await build({
		absWorkingDir: source,
		entryPoints: ['src/extension.ts'],
		outfile: join(target, 'out', 'extension.js'),
		bundle: true,
		platform: 'node',
		format: 'cjs',
		target: 'node20',
		external: ['vscode'],
		minify: true,
		logLevel: 'warning',
	});
	writeFileSync(join(target, 'package.json'), manifestText(JSON.parse(readFileSync(join(source, 'package.json'), 'utf8'))));
	copyFileSync(join(source, 'README.md'), join(target, 'README.md'));
}

await buildExtension({ installed, outputs, makeTarget: 'vscode-spelldata', buildInto });
