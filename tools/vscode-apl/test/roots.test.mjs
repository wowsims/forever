import assert from 'node:assert/strict';
import { chmodSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import Module from 'node:module';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { test } from 'node:test';

import { activateBuilt, fakeVscode, until } from './vscode.mjs';

// A file in a repository nested inside another, such as a worktree under .claude/worktrees, goes to the server of the
// nearest repository only: the outer server is never told it is open. The servers here are a stand-in that answers
// `initialize` and `shutdown`.

const require = Module.createRequire(import.meta.url);

// A repository at `root` as the extension finds one, with a .siml file in it.
function repository(root) {
	mkdirSync(join(root, 'sim', 'siml'), { recursive: true });
	writeFileSync(join(root, 'sim', 'siml', 'names.json'), '{}\n');
	mkdirSync(join(root, 'ui'), { recursive: true });
	writeFileSync(join(root, 'ui', 'a.siml'), 'siml 1\n');
	return join(root, 'ui', 'a.siml');
}

// The stand-in server: an executable that answers the requests the language client sends to start and stop it.
function stubServer(dir) {
	const file = join(dir, 'server.cjs');
	writeFileSync(
		file,
		`#!/usr/bin/env node
const rpc = require(${JSON.stringify(require.resolve('vscode-jsonrpc/node'))});
const connection = rpc.createMessageConnection(new rpc.StreamMessageReader(process.stdin), new rpc.StreamMessageWriter(process.stdout));
connection.onRequest('initialize', () => ({ capabilities: {} }));
connection.onRequest('shutdown', () => null);
connection.onNotification('exit', () => process.exit(0));
connection.listen();
`,
	);
	chmodSync(file, 0o755);
	return file;
}

for (const order of [
	['inner', 'outer'],
	['outer', 'inner'],
]) {
	test(`a file in a nested repository goes to its own server only, opened ${order.join(' then ')}`, { skip: process.platform === 'win32' }, async () => {
		const dir = mkdtempSync(join(tmpdir(), 'wowsims-apl-roots-'));
		const outer = join(dir, 'outer');
		const inner = join(outer, '.claude', 'worktrees', 'wt');
		const files = { outer: repository(outer), inner: repository(inner) };
		const stand = fakeVscode();
		stand.state.config.server = stubServer(dir);
		const extension = await activateBuilt(stand);
		try {
			for (const which of order) await stand.vscode.workspace.openTextDocument(stand.vscode.Uri.file(files[which]));
			const client = root => stand.state.clients.find(c => c.running && c.server.options.cwd === root);
			await until(() => client(outer)?.opened && client(inner)?.opened);
			await until(() => client(outer).opened.size + client(inner).opened.size === 2);
			assert.equal(stand.state.clients.length, 2, 'one server per repository');
			assert.deepEqual([...client(outer).opened], [files.outer]);
			assert.deepEqual([...client(inner).opened], [files.inner]);
		} finally {
			await extension.deactivate();
			rmSync(dir, { recursive: true, force: true });
		}
	});
}
