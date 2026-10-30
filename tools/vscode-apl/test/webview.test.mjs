import assert from 'node:assert/strict';
import { mkdtempSync, readFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { after, before, test } from 'node:test';
import { fileURLToPath } from 'node:url';
import vm from 'node:vm';

import { GlobalRegistrator } from '@happy-dom/global-registrator';

import { buildPanel } from '../scripts/bundle.mjs';
import { activateBuilt, fakeVscode, serverRunning, until } from './vscode.mjs';

// The panel, bundled from its sources as the build bundles it, mounted in happy-dom and talking to the extension host
// and its SIML server through the stand-in for VS Code, as it talks to the extension in a webview: what a click or a
// key in the panel does to the text.

const out = mkdtempSync(join(tmpdir(), 'wowsims-apl-panel-'));
const fixture = fileURLToPath(new URL('../../../sim/siml/lsp/testdata/default.apl.siml', import.meta.url));
// The committed name table, which the expected ids are read from.
const names = JSON.parse(readFileSync(new URL('../../../sim/siml/names.json', import.meta.url), 'utf8'));

let stand;
let extension;
let panel;
let document_;
const sent = [];

const rows = list => [...document.querySelectorAll(`[data-row^="${list}/"]`)].filter(row => /^[a-zA-Z]+\/\d+$/.test(row.dataset.row));
const priority = () => rows('priorityList');
const rowText = row => row.querySelector('.summary-main').textContent;
const lastRotation = () => panel.posted.filter(m => m.type === 'rotation').at(-1);
// Waits until the extension has answered every edit the panel sent.
const settled = () => until(() => (lastRotation()?.handled ?? 0) >= sent.filter(m => m.type === 'edit').length);

before(async () => {
	await buildPanel(out);
	GlobalRegistrator.register({ url: 'https://panel.test/' });
	stand = fakeVscode();
	extension = await activateBuilt(stand);
	const uri = stand.vscode.Uri.file(fixture);
	document_ = await stand.vscode.workspace.openTextDocument(uri);
	await serverRunning(stand);
	await stand.state.commands.get('wowsims-apl.openToSide')(uri);
	panel = stand.state.panels.at(-1);
	panel.onPost = message => window.dispatchEvent(new MessageEvent('message', { data: message }));
	window.acquireVsCodeApi = () => ({
		postMessage: message => {
			sent.push(message);
			panel.send(message);
		},
		setState() {},
	});
	document.body.innerHTML = '<div id="root"></div>';
	vm.runInThisContext(readFileSync(join(out, 'webview.js'), 'utf8'));
	await until(() => priority().length === 13);
});

after(async () => {
	await extension.deactivate();
	await GlobalRegistrator.unregister();
	rmSync(out, { recursive: true, force: true });
});

const key = (element, init) => element.dispatchEvent(new KeyboardEvent('keydown', { bubbles: true, cancelable: true, ...init }));
const click = element => element.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true }));

test('Alt+ArrowUp moves the focused row, and the focus goes with it', async () => {
	const summary = priority()[4].querySelector('.summary');
	assert.equal(rowText(priority()[4]), 'Cast Strider Kick');
	summary.focus();
	key(summary, { key: 'ArrowUp', altKey: true });
	// A second key press is a task of its own, after the panel has drawn the first.
	await until(() => document.activeElement.closest('li')?.dataset.row === 'priorityList/3');
	key(document.activeElement, { key: 'ArrowUp', altKey: true });
	await settled();
	assert.equal(rowText(priority()[2]), 'Cast Strider Kick');
	assert.equal(document.activeElement.closest('li').dataset.row, 'priorityList/2');
	const lines = document_.getText().split('\n');
	assert.equal(lines[lines.indexOf('  cast strider_kick') - 1], '  cast hunters_mark if not target.aura(hunters_mark).is_active');
});

test('an open row stays open on its row when a row above it is deleted', async () => {
	click(priority()[5].querySelector('.chevron'));
	await until(() => priority()[5].querySelector('.row-body'));
	const open = rowText(priority()[5]);
	click(priority()[0].querySelector('button[title="Delete"]'));
	await settled();
	assert.equal(rowText(priority()[4]), open);
	assert.ok(priority()[4].querySelector('.row-body'), 'still open');
	assert.equal(priority()[5].querySelector('.row-body'), null);
	assert.doesNotMatch(document_.getText(), /autocast_other_cooldowns/);
});

test('a spell picked for a tagged one is written as its own name, with no tag carried over', async () => {
	stand.change(document_, document_.getText().replace('priority:\n', 'priority:\n  cast innervate\n'));
	// The chip writes the tag, -1, after the name.
	await until(() => rowText(priority()[0]) === 'Cast Innervate-1');
	click(priority()[0].querySelector('.chevron'));
	await until(() => priority()[0].querySelector('.row-body .id-picker .picker-trigger'));
	click(priority()[0].querySelector('.row-body .id-picker .picker-trigger'));
	await until(() => priority()[0].querySelector('.picker-pop input'));
	const input = priority()[0].querySelector('.picker-pop input');
	Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set.call(input, 'raptor');
	input.dispatchEvent(new Event('input', { bubbles: true }));
	const option = () => [...priority()[0].querySelectorAll('.picker-pop li')].find(li => li.textContent.startsWith('Raptor Strike'));
	await until(option);
	// raptor_strike_tag0, the strike the queue casts, is no spell a rotation casts, so a cast's picker leaves it out.
	const strikes = [...priority()[0].querySelectorAll('.picker-pop li')].filter(li => li.textContent.startsWith('Raptor Strike'));
	assert.equal(strikes.length, 1, `the cast offers ${strikes.map(li => li.textContent)}`);
	option().dispatchEvent(new MouseEvent('mousedown', { bubbles: true, cancelable: true }));
	await settled();
	const edit = sent.filter(m => m.type === 'edit').at(-1);
	// The entry's id, with its own tag and rank where it has them.
	const { spellId, tag, rank } = names.classes.hunter.spell.raptor_strike;
	assert.deepEqual(edit.edits[0].value, { spellId, ...(tag && { tag }), ...(rank && { rank }) });
	assert.match(document_.getText(), /\npriority:\n {2}cast raptor_strike\n/);
});

test('a pet picked as the unit of an aura is written with its owner, as the sim reads one', async () => {
	const mark = () => priority().find(row => rowText(row).startsWith("Cast Hunter's Mark"));
	click(mark().querySelector('.chevron'));
	// The condition `not target.aura(hunters_mark).is_active`: open the `not`, then the aura value inside it.
	await until(() => mark().querySelector('.row-body .value-preview'));
	click(mark().querySelector('.row-body .value-preview'));
	await until(() => mark().querySelectorAll('.row-body .value-preview').length > 1);
	click(mark().querySelectorAll('.row-body .value-preview')[1]);
	await until(() => [...mark().querySelectorAll('.unit-picker select')].some(select => select.value === 'CurrentTarget'));
	const select = [...mark().querySelectorAll('.unit-picker select')].find(s => s.value === 'CurrentTarget');
	assert.deepEqual(
		[...select.options].map(option => option.value),
		['', 'Pet', 'CurrentTarget', 'PreviousTarget', 'NextTarget', 'Player', 'Target'],
		"the units the site's aura sources offer",
	);
	select.value = 'Pet';
	select.dispatchEvent(new Event('change', { bubbles: true }));
	await settled();
	assert.deepEqual(sent.filter(m => m.type === 'edit').at(-1).edits[0].value, { type: 'Pet', owner: { type: 'Self' } });
	assert.match(document_.getText(), /cast hunters_mark if not pet1\.aura\(/);
});

test('nothing is sent while the text has errors', async () => {
	stand.change(document_, document_.getText().replace('cast mongoose_bite', 'cast mongoose_bit'));
	await until(() => document.querySelector('.error'));
	assert.match(document.querySelector('.error').textContent, /^The text has errors/);
	assert.equal(document.querySelector('fieldset').disabled, true);
	const edits = sent.filter(m => m.type === 'edit').length;
	const summary = priority()[3].querySelector('.summary');
	summary.focus();
	key(summary, { key: 'ArrowUp', altKey: true });
	click(priority()[1].querySelector('button[title="Delete"]'));
	await new Promise(resolve => setTimeout(resolve, 100));
	assert.equal(sent.filter(m => m.type === 'edit').length, edits);
});

// When blocks and defs: the rows below are those of this text.
const blocksText = [
	'siml 1',
	'spec hunter',
	'',
	'def fits(x) = spell(x).cast_time < 2s',
	'',
	'priority:',
	'  cast hunters_mark',
	'  when current_time > 1s:',
	'    cast aimed_shot if fits(aimed_shot)',
	'    cast multi_shot',
	'  cast arcane_shot',
	'  cast serpent_sting',
	'',
].join('\n');

const head = id => document.querySelector(`[data-when="${id}"]`);
const condText = row => row.querySelector('.summary-cond')?.textContent;
// Types into a box that writes on Enter: after the box's first effects, which set its text, and once the panel has
// drawn what was typed.
const commit = async (input, value) => {
	await new Promise(resolve => setTimeout(resolve, 20));
	input.focus();
	Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set.call(input, value);
	input.dispatchEvent(new Event('input', { bubbles: true }));
	await new Promise(resolve => setTimeout(resolve, 20));
	key(input, { key: 'Enter' });
};
// Opens a row's editor, which a row opened before stays.
const openRow = row => row.querySelector('.row-body') || click(row.querySelector('.chevron'));
// Puts the text above in the document, and waits until the panel has drawn it. The rotation the extension posts can be
// ahead of the page, and rows an earlier test left would take a click as an edit of the version before, which the
// extension drops, so the wait reads the page's rows: one block, its condition, the def and the calls as the text has
// them, and none of what the tests below add.
const useBlocksText = async () => {
	if (document_.getText() !== blocksText) stand.change(document_, blocksText);
	await until(
		() =>
			lastRotation()?.version === document_.version &&
			priority().length === 5 &&
			head(0) &&
			!head(1) &&
			/^when Current Time > 1s/.test(head(0).textContent) &&
			condText(priority()[1]) === 'if fits(aimed_shot)' &&
			condText(priority()[4]) === undefined &&
			document.querySelector('[data-def="fits"]') &&
			!document.querySelector('[data-def="def1"]') &&
			!document.querySelector('fieldset').disabled,
	);
};

test('a when block is a line over its rows, which show their own conditions and their def calls', async () => {
	await useBlocksText();
	assert.match(head(0).textContent, /^when Current Time > 1s/);
	assert.equal(condText(priority()[1]), 'if fits(aimed_shot)');
	assert.equal(condText(priority()[2]), undefined, 'the block gives multi_shot its only condition');
	assert.equal(getComputedStyle(priority()[1]).getPropertyValue('--depth').trim(), '1');
});

test('a def call keeps its call when an argument changes', async () => {
	await useBlocksText();
	openRow(priority()[1]);
	await commit(await until(() => priority()[1].querySelector('.def-arg input')), 'multi_shot');
	await settled();
	await until(() => condText(priority()[1]) === 'if fits(multi_shot)');
	assert.match(document_.getText(), /\n {4}cast aimed_shot if fits\(multi_shot\)\n/);
});

test('picked rows go in a new when block', async () => {
	await useBlocksText();
	click(priority()[3].querySelector('.index'));
	priority()[4]
		.querySelector('.index')
		.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true, shiftKey: true }));
	await until(() => document.querySelector('.pick-bar'));
	click([...document.querySelectorAll('.pick-bar button')].find(b => b.textContent === 'Put in a when block'));
	await settled();
	await until(() => head(1));
	// The panel shows its change at once and the text catches up.
	await until(() => /when true:/.test(document_.getText()));
	assert.match(document_.getText(), /\n {2}when true:\n {4}cast arcane_shot\n {4}cast serpent_sting\n/);
});

test('a block taken apart writes its condition on each of its rows', async () => {
	await useBlocksText();
	click(head(0).querySelector('button[title^="Take the block apart"]'));
	await settled();
	await until(() => !head(0));
	await until(() => !/when current_time > 1s:/.test(document_.getText()));
	assert.match(document_.getText(), /\n {2}cast aimed_shot if current_time > 1s and fits\(aimed_shot\)\n {2}cast multi_shot if current_time > 1s\n/);
});

test("a block's condition edits on its line, and its rows keep theirs", async () => {
	await useBlocksText();
	click(head(0).querySelector('.summary'));
	// The condition `current_time > 1s`: its fields, then the constant.
	click(await until(() => head(0).querySelector('.row-body .value-preview')));
	await commit(await until(() => head(0).querySelector('.row-body input.const')), '3s');
	await settled();
	await until(() => /when current_time > 3s:/.test(document_.getText()));
	assert.match(document_.getText(), /\n {2}when current_time > 3s:\n {4}cast aimed_shot if fits\(aimed_shot\)\n {4}cast multi_shot\n/);
});

test('a def deleted from the defs section leaves its calls as the values they read as', async () => {
	await useBlocksText();
	const line = () => document.querySelector('[data-def="fits"]');
	await until(line);
	click(line().querySelector('button[title^="Delete the def"]'));
	await settled();
	await until(() => !line());
	assert.doesNotMatch(document_.getText(), /def fits/);
	assert.match(document_.getText(), /cast aimed_shot if spell\(aimed_shot\)\.cast_time < 2s\n/);
});

test('a new def from the defs section, then called from a picker', async () => {
	await useBlocksText();
	click([...document.querySelectorAll('.defs .add')].at(-1));
	await settled();
	await until(() => document.querySelector('[data-def="def1"]'));
	assert.match(document_.getText(), /\ndef fits\(x\) = spell\(x\)\.cast_time < 2s\ndef def1 = true\n/);
	openRow(priority()[4]);
	await until(() => priority()[4].querySelector('.row-body .def-pick'));
	click(priority()[4].querySelector('.row-body .def-pick').closest('button'));
	await until(() => priority()[4].querySelector('.picker-pop li'));
	const option = [...priority()[4].querySelectorAll('.picker-pop li')].find(li => li.textContent.startsWith('def1'));
	option.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, cancelable: true }));
	await settled();
	await until(() => /cast serpent_sting if def1\n/.test(document_.getText()));
});
