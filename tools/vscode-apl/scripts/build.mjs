import { spawnSync } from 'node:child_process';
import { copyFileSync, mkdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

import { buildExtension } from '../../vscode-common/build.mjs';
import { buildPanel, bundleExtension } from './bundle.mjs';
import { manifestText } from './manifest.mjs';

const source = dirname(dirname(fileURLToPath(import.meta.url)));
const repository = join(source, '..', '..');
const installed = join(repository, '.vscode', 'extensions', 'wowsims-apl');
// The committed outputs, which the extension's sources decide; `--check` holds the committed build to them.
const outputs = ['package.json', 'README.md', 'language-configuration.json', join('out', 'extension.js')];
// The grammar is committed too, but written by `go run ./cmd/wowsimcli siml grammar` (the makefile's vscode-apl rule)
// and checked by that command's test.
const grammar = join('syntaxes', 'siml.tmLanguage.json');
// The panel's page also takes in the generated protos, the site's APL model and its English strings, so it is built
// here but not committed (.gitignore): a change to those needs no extension rebuild in the same commit.
const panel = [join('out', 'webview.js'), join('out', 'webview.css')];

// tsc runs through node: the `.bin` entry is a shell script on Windows, which cannot be spawned without a shell.
function typecheck(project) {
	const tsc = spawnSync(process.execPath, [join(source, 'node_modules', 'typescript', 'bin', 'tsc'), '--noEmit', '-p', join(source, project)], {
		stdio: 'inherit',
	});
	if (tsc.error) console.error(tsc.error.message);
	if (tsc.status !== 0) {
		process.exit(tsc.status ?? 1);
	}
}

async function buildInto(target) {
	typecheck('tsconfig.json');
	typecheck('tsconfig.webview.json');

	rmSync(join(target, 'out'), { recursive: true, force: true });
	mkdirSync(join(target, 'out'), { recursive: true });
	await bundleExtension(join(target, 'out', 'extension.js'));
	await buildPanel(join(target, 'out'));
	writeFileSync(join(target, 'package.json'), manifestText(JSON.parse(readFileSync(join(source, 'package.json'), 'utf8'))));
	copyFileSync(join(source, 'README.md'), join(target, 'README.md'));
	copyFileSync(join(source, 'language-configuration.json'), join(target, 'language-configuration.json'));
}

await buildExtension({ installed, outputs, kept: [...panel, grammar], makeTarget: 'vscode-apl', buildInto });
