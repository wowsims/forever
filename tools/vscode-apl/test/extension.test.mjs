import assert from 'node:assert/strict';
import { mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { after, before, test } from 'node:test';
import { fileURLToPath } from 'node:url';

import { activateBuilt, fakeVscode, serverRunning, until } from './vscode.mjs';

// The extension's host, run against a stand-in vscode module and the SIML server it builds: the panel's messages go in
// as its page would send them, and the documents are in memory.

const fixtures = fileURLToPath(new URL('../../../sim/siml/lsp/testdata/', import.meta.url));
const manifest = JSON.parse(readFileSync(new URL('../package.json', import.meta.url), 'utf8'));
// The committed name table, which the expected ids are read from.
const names = JSON.parse(readFileSync(new URL('../../../sim/siml/names.json', import.meta.url), 'utf8'));
const tableSpell = (className, name) => names.classes[className].spell[name];

let stand;
let extension;

before(async () => {
	stand = fakeVscode();
	extension = await activateBuilt(stand);
	await reopen('default.apl.siml');
	await serverRunning(stand);
});

after(() => extension.deactivate());

// A fixture's document read from disk again, the one an earlier test left open closed.
async function reopen(file) {
	const uri = stand.vscode.Uri.file(`${fixtures}${file}`);
	for (const open of stand.vscode.workspace.textDocuments) if (open.uri.toString() === uri.toString()) stand.close(open);
	return stand.vscode.workspace.openTextDocument(uri);
}

// Opens the panel on a fixture, read from disk again, and answers its first `ready`: the panel and its document.
async function openPanel(file) {
	const document = await reopen(file);
	const { uri } = document;
	await stand.state.commands.get('wowsims-apl.openToSide')(uri);
	const panel = stand.state.panels.at(-1);
	panel.send({ type: 'ready' });
	await until(() => panel.posted.some(m => m.type === 'rotation'));
	return { panel, document };
}

const rotations = panel => panel.posted.filter(m => m.type === 'rotation');
const lastRotation = panel => rotations(panel).at(-1);

// Sends an edit and waits for the rotation that answers it.
async function edit(panel, message) {
	panel.send({ type: 'edit', ...message });
	await until(() => (lastRotation(panel)?.handled ?? 0) >= message.seq);
	return lastRotation(panel);
}

// The lines of `after` that are not lines of `before`, and the other way round.
function changedLines(before, after) {
	const a = before.split('\n');
	const b = after.split('\n');
	const count = lines => lines.reduce((m, line) => m.set(line, (m.get(line) ?? 0) + 1), new Map());
	const [ca, cb] = [count(a), count(b)];
	return {
		removed: [...ca].filter(([line, n]) => n > (cb.get(line) ?? 0)).map(([line]) => line),
		added: [...cb].filter(([line, n]) => n > (ca.get(line) ?? 0)).map(([line]) => line),
	};
}

test('the extension registers every contributed command and the panel, and starts the server it built', () => {
	assert.deepEqual([...stand.state.commands.keys()].sort(), manifest.contributes.commands.map(c => c.command).sort());
	assert.deepEqual(stand.state.serializers, ['wowsims-apl.editor']);
	const [client] = stand.state.clients;
	assert.deepEqual(client.server.args, ['siml', 'lsp']);
	assert.equal(client.server.options.cwd, fileURLToPath(new URL('../../../', import.meta.url)).replace(/[\\/]$/, ''));
});

test('the panel opens on an .apl.siml and reads its rotation, its findings and the names its spec registers', async () => {
	const { panel, document } = await openPanel('default.apl.siml');
	const rotation = lastRotation(panel);
	assert.equal(rotation.version, document.version);
	assert.equal(rotation.readOnly, undefined);
	assert.equal(rotation.className, 'hunter');
	assert.equal(rotation.rotation.priorityList.length, 13);
	assert.ok(Array.isArray(rotation.findings));
	const init = panel.posted.find(m => m.type === 'init');
	assert.equal(init.fileName, 'default.apl.siml');
	assert.equal(init.db.spells[tableSpell('warrior', 'defensive_stance').spellId]?.[0], 'Defensive Stance');
	// The names come through their own request, after the rotation.
	await until(() => panel.posted.some(m => m.type === 'registered'));
	const registered = panel.posted.find(m => m.type === 'registered').registered;
	assert.equal(registered.status, 'ready');
	assert.equal(registered.spec, 'hunter');
	const mark = tableSpell('hunter', 'hunters_mark');
	assert.ok(registered.registered.spells.some(n => n.name === 'hunters_mark' && n.id.spellId === mark.spellId && n.id.rank === mark.rank));
	assert.ok(registered.registered.items.every(n => n.name.startsWith('item.') && n.id.itemId > 0));
	panel.dispose();
});

test('a moved row changes only the lines of that row', async () => {
	const { panel, document } = await openPanel('default.apl.siml');
	const before = document.getText();
	const { version } = lastRotation(panel);
	const applied = stand.state.applied.length;
	const answer = await edit(panel, { version, seq: 1, edits: [{ op: 'move', path: ['priorityList'], from: 4, to: 1 }] });
	assert.equal(document.version, version + 1, 'one change');
	assert.equal(answer.rotation.priorityList[1].action.castSpell.spellId.spellId, tableSpell('hunter', 'strider_kick').spellId, 'Strider Kick is second');
	// One WorkspaceEdit: the row's line goes in at its new place and comes out of its old one.
	assert.equal(stand.state.applied.length, applied + 1);
	const edits = stand.state.applied.at(-1).map(e => [e.range.start.line, e.range.end.line, e.newText]);
	const row = before.split('\n').indexOf('  cast strider_kick');
	assert.deepEqual(edits, [
		[row - 3, row - 3, '  cast strider_kick\n'],
		[row, row + 1, ''],
	]);
	assert.deepEqual(changedLines(before, document.getText()), { removed: [], added: [] }, 'the same lines in another order');
	panel.dispose();
});

test('an edit through an index that does not exist is refused, and the text stays as it is', async () => {
	const { panel, document } = await openPanel('default.apl.siml');
	const before = document.getText();
	const { version } = lastRotation(panel);
	const warnings = stand.state.messages.length;
	const answer = await edit(panel, {
		version,
		seq: 1,
		edits: [{ op: 'set', path: ['groups', 0, 'actions', -1, 'action', 'condition'], value: { gcdIsReady: {} } }],
	});
	assert.equal(document.getText(), before);
	assert.equal(answer.handled, 1);
	assert.match(stand.state.messages.at(-1).message, /groups\[0\]\.actions\[-1\] does not exist/);
	assert.equal(stand.state.messages.length, warnings + 1);
	panel.dispose();
});

test('while the text has errors the panel is read-only, keeps no rotation, and an edit sent anyway is not made', async () => {
	const { panel, document } = await openPanel('default.apl.siml');
	stand.change(document, document.getText().replace('  cast mongoose_bite\n', '  cast mongoose_bit\n'));
	await until(() => lastRotation(panel).version === document.version);
	const broken = lastRotation(panel);
	assert.equal(broken.rotation, undefined);
	assert.match(broken.readOnly, /^The text has errors \(line 27: no spell is named mongoose_bit; did you mean mongoose_bite\?\)\. Edits wait/);
	const before = document.getText();
	await edit(panel, { version: document.version, seq: 1, edits: [{ op: 'set', path: ['priorityList', 0, 'hide'], value: true }] });
	assert.equal(document.getText(), before);
	assert.match(stand.state.messages.at(-1).message, /the text has errors/);
	panel.dispose();
});

test("the panel's own edits made in a row are all made; one against a version someone else moved past is dropped", async () => {
	const { panel, document } = await openPanel('default.apl.siml');
	const { version } = lastRotation(panel);
	// Two edits sent before the first answer, both against the version the panel read.
	panel.send({ type: 'edit', version, seq: 1, edits: [{ op: 'set', path: ['priorityList', 0, 'hide'], value: true }] });
	panel.send({ type: 'edit', version, seq: 2, edits: [{ op: 'set', path: ['priorityList', 2, 'hide'], value: true }] });
	await until(() => lastRotation(panel).handled === 2);
	assert.match(document.getText(), /\n {2}# autocast_other_cooldowns if current_time >= 5s\n/);
	assert.match(document.getText(), /\n {2}# group "Weave" if \$melee_weave\n/);
	// Someone types in the text; an edit the panel made before it saw that is dropped.
	stand.change(document, document.getText().replace('var ${Use Aimed Shot} = false', 'var ${Use Aimed Shot} = true'));
	const typed = document.getText();
	const answer = await edit(panel, { version: version + 2, seq: 3, edits: [{ op: 'set', path: ['priorityList', 3, 'hide'], value: true }] });
	assert.equal(document.getText(), typed);
	assert.equal(answer.version, document.version);
	panel.dispose();
});

test("a panel's edit is dropped when the text changes while the server splices it", async () => {
	const { panel, document } = await openPanel('default.apl.siml');
	const { version } = lastRotation(panel);
	const [client] = stand.state.clients;
	const sendRequest = client.sendRequest;
	// Someone types while the server answers siml/splice: its edits are for the text before that.
	client.sendRequest = async function (method, ...params) {
		const answer = await sendRequest.call(this, method, ...params);
		if (method === 'siml/splice') stand.change(document, document.getText().replace('${Time to weave} = 400ms', '${Time to weave} = 500ms'));
		return answer;
	};
	try {
		const applied = stand.state.applied.length;
		const answer = await edit(panel, { version, seq: 1, edits: [{ op: 'move', path: ['priorityList'], from: 4, to: 1 }] });
		assert.equal(stand.state.applied.length, applied, 'no edit applied');
		assert.equal(answer.version, document.version);
		assert.equal(
			answer.rotation.priorityList[4].action.castSpell.spellId.spellId,
			tableSpell('hunter', 'strider_kick').spellId,
			'Strider Kick stays fifth',
		);
	} finally {
		client.sendRequest = sendRequest;
		panel.dispose();
	}
});

test('the panel finds its document by URI on every message, after VS Code closes the one it opened with', async () => {
	const { panel, document } = await openPanel('default.apl.siml');
	const { version } = lastRotation(panel);
	stand.close(document);
	const opened = stand.state.opened.length;
	await edit(panel, { version, seq: 1, edits: [{ op: 'set', path: ['priorityList', 0, 'hide'], value: true }] });
	assert.equal(stand.state.opened.length, opened + 1, 'opened again from disk');
	const reopened = stand.vscode.workspace.textDocuments.find(d => d.uri.toString() === document.uri.toString());
	assert.notEqual(reopened, document);
	assert.match(reopened.getText(), /\n {2}# autocast_other_cooldowns/);
	assert.doesNotMatch(document.getText(), /# autocast_other_cooldowns/, 'not the closed one');
	panel.dispose();
});

test('a rotation put together with another file is read-only, and shows the rows it uses', async () => {
	const { panel } = await openPanel('hawks.apl.siml');
	const rotation = lastRotation(panel);
	assert.equal(rotation.readOnly, 'The rotation is put together with default.apl.siml; edit it as text.');
	assert.equal(rotation.rotation.priorityList.length, 13);
	assert.ok(rotation.rotation.valueVariables.some(v => v.name === 'Use Summon Hawk' && v.value.const.val === 'true'));
	panel.dispose();
});

test("a rotation put together with another file follows that file's editor", async () => {
	const used = await reopen('default.apl.siml');
	const { panel } = await openPanel('hawks.apl.siml');
	assert.equal(lastRotation(panel).rotation.priorityList.length, 13);
	stand.change(used, used.getText().replace('  cast strider_kick\n', ''));
	await until(() => lastRotation(panel).rotation.priorityList.length === 12);
	// Closed unsaved, it is read from disk again.
	stand.close(used);
	await until(() => lastRotation(panel).rotation.priorityList.length === 13);
	panel.dispose();
});

test('a sim of a rotation that uses a file with unsaved changes is refused, as the sim reads that file from disk', async () => {
	const used = await reopen('default.apl.siml');
	const using = await reopen('hawks.apl.siml');
	stand.change(used, used.getText().replace('  cast strider_kick\n', ''));
	await stand.state.commands.get('wowsims-apl.sim')(using.uri);
	assert.deepEqual(stand.state.messages.at(-1), {
		severity: 'warning',
		message: 'default.apl.siml has unsaved changes; the sim reads it from disk. Save it first.',
	});
});

test('the cursor in the text selects the row it is in', async () => {
	const { panel, document } = await openPanel('default.apl.siml');
	stand.select(document, document.getText().indexOf('cast serpent_sting') + 3);
	await until(() => panel.posted.some(m => m.type === 'cursor'));
	assert.deepEqual(panel.posted.find(m => m.type === 'cursor').path.slice(0, 2), ['priorityList', 5]);
	panel.dispose();
});

test('Insert Pattern asks for every value, its default offered, and writes the spell by its name', async () => {
	const { panel, document } = await openPanel('default.apl.siml');
	stand.state.asked.length = 0;
	stand.state.answers.push(
		items => items.find(item => item.pattern.id === 'keep-debuff'),
		items => items.find(item => item.description === 'hunters_mark'),
		undefined, // The debuff is optional: the spell's own.
		options => (options.value === '3s' ? '2s' : undefined),
	);
	const applied = stand.state.applied.length;
	panel.send({ type: 'pattern', path: ['priorityList', 0] });
	await until(() => stand.state.applied.length > applied);
	assert.deepEqual(stand.state.asked, ['Insert APL pattern', 'Spell', 'Debuff (default: the spell) (Escape leaves it out)', 'Refresh at']);
	assert.deepEqual(
		stand.state.applied.at(-1).map(e => e.newText),
		['  cast hunters_mark if target.aura(hunters_mark).remaining_time <= 2s\n'],
	);
	assert.equal(document.getText().split('\n').indexOf('  cast hunters_mark if target.aura(hunters_mark).remaining_time <= 2s'), 24);
	panel.dispose();
});

test('Insert Pattern puts no condition on a prepull row, whose condition the sim does not read', async () => {
	const { panel, document } = await openPanel('default.apl.siml');
	const before = document.getText();
	stand.state.answers.push(items => items.find(item => item.pattern.id === 'outside-execute'));
	panel.send({ type: 'pattern', path: ['prepullActions', 0] });
	await until(() => stand.state.messages.at(-1)?.message.startsWith('Outside the execute phase is a condition'));
	assert.equal(document.getText(), before);
	panel.dispose();
});

test('the panel does not open on a gear set, and says the server still serves the file', async () => {
	const gear = await stand.vscode.workspace.openTextDocument(
		stand.vscode.Uri.file(fileURLToPath(new URL('../../../ui/specs/hunter/dps/gear_sets/p1.gear.siml', import.meta.url))),
	);
	const panels = stand.state.panels.length;
	await stand.state.commands.get('wowsims-apl.openToSide')(gear.uri);
	assert.equal(stand.state.panels.length, panels);
	assert.match(stand.state.messages.at(-1).message, /opens on \.apl\.siml rotations only, and p1\.gear\.siml is not one\. The SIML server still checks/);
	stand.close(gear);
});

test('WoWSims: Restart SIML Server stops the server and starts a new one, which answers the panel', async () => {
	const before = stand.state.clients.length;
	const old = stand.state.clients.at(-1);
	await stand.state.commands.get('wowsims-apl.restartServer')();
	assert.equal(old.running, false);
	assert.equal(stand.state.clients.length, before + 1);
	assert.equal(stand.state.clients.at(-1).running, true);
	const { panel } = await openPanel('default.apl.siml');
	assert.ok(lastRotation(panel).rotation, 'the new server reads the rotation');
	panel.dispose();
});

test("the panel's page loads its bundle, or builds it with make vscode-apl, and after a failed build tries again when the panel opens again", async () => {
	const installed = mkdtempSync(join(tmpdir(), 'wowsims-apl-'));
	const other = fakeVscode();
	const otherExtension = await activateBuilt(other, installed);
	try {
		const page = async file => {
			const document = await other.vscode.workspace.openTextDocument(other.vscode.Uri.file(`${fixtures}${file}`));
			await other.state.commands.get('wowsims-apl.openToSide')(document.uri);
			return other.state.panels.at(-1).webview.html;
		};
		// The extension sits outside a repository here, so the build fails.
		assert.match(await page('default.apl.siml'), /Building the panel with <code>make vscode-apl<\/code>/);
		await until(() => /did not build/.test(other.state.panels.at(-1).webview.html));
		mkdirSync(join(installed, 'out'));
		writeFileSync(join(installed, 'out', 'webview.js'), '');
		assert.match(await page('default.apl.siml'), /<script nonce="[^"]+" src="[^"]+webview\.js"><\/script>/, 'the same panel, its page built again');
		assert.equal(other.state.panels.length, 1);
	} finally {
		await otherExtension.deactivate();
		rmSync(installed, { recursive: true, force: true });
	}
});
