import assert from 'node:assert/strict';
import { after, before, test } from 'node:test';
import { fileURLToPath } from 'node:url';

import { activateBuilt, fakeVscode, serverRunning, until } from './vscode.mjs';

// Sims from the editor, which the SIML server runs: the hunter's default rotation, whose path picks the hunter's
// default build, as ui/specs/hunter/dps/apls/default.apl.siml.

const fixture = fileURLToPath(new URL('../../../sim/siml/lsp/testdata/default.apl.siml', import.meta.url));

let stand;
let extension;
let document;

before(async () => {
	stand = fakeVscode();
	extension = await activateBuilt(stand);
	document = await stand.vscode.workspace.openTextDocument(stand.vscode.Uri.file(fixture));
	await serverRunning(stand);
});

after(() => extension.deactivate());

// A sim takes minutes on a loaded machine.
const untilSimmed = condition => until(condition, 300000);

test('a compare sims HEAD and the text on one seed, and the panel shows the row counts', async () => {
	stand.state.config.simIterations = 20;
	await stand.state.commands.get('wowsims-apl.openToSide')(document.uri);
	const panel = stand.state.panels.at(-1);
	panel.send({ type: 'ready' });
	await untilSimmed(() => panel.posted.some(m => m.type === 'rotation'));
	panel.send({ type: 'sim', compare: true });
	await untilSimmed(() => panel.posted.some(m => m.type === 'sim' && m.sim && !m.running));
	const { sim } = panel.posted.findLast(m => m.type === 'sim');
	assert.equal(sim.result.baseline, 'HEAD');
	assert.equal(sim.stale, false);
	assert.equal(sim.result.iterations, 20);
	assert.equal(sim.result.rows['prepullActions/0'].executions, 1);
	assert.equal(sim.result.difference.delta, 0, 'the same text on the same seed');
	assert.equal(sim.result.timeline[0].path, 'prepullActions/0');
	assert.ok(sim.result.spells.some(spell => spell.name === "Hunter's Mark"));
	panel.dispose();
});

test('a text that does not read is not simmed', async () => {
	const messages = stand.state.messages.length;
	stand.change(document, document.getText().replace('  cast mongoose_bite\n', '  cast mongoose_bit\n'));
	await stand.state.commands.get('wowsims-apl.sim')(document.uri);
	assert.equal(stand.state.messages.length, messages + 1);
	assert.match(stand.state.messages.at(-1).message, /^The rotation does not read \(line 27: no spell is named mongoose_bit/);
	stand.change(document, document.getText().replace('  cast mongoose_bit\n', '  cast mongoose_bite\n'));
});

test('a cancelled sim stops in the server and says nothing', async () => {
	stand.state.config.simIterations = 100000;
	const messages = stand.state.messages.length;
	const sim = stand.state.commands.get('wowsims-apl.sim')(document.uri);
	// The sim's progress is the only one running: the server is built, and the tests install a panel bundle.
	const simming = () => stand.state.progress[0];
	await untilSimmed(() => simming());
	await new Promise(resolve => setTimeout(resolve, 300));
	const started = Date.now();
	simming().cancel();
	await sim;
	assert.ok(Date.now() - started < 30000, 'cancelled, not run to the end');
	assert.equal(stand.state.messages.length, messages);
});
