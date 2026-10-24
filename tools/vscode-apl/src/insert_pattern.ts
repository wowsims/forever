import * as vscode from 'vscode';

import { applyEdits, documentFor, isAplSiml, readRotation } from './host';
import type { AplDatabase, AplPattern, AplPlaceholder, JsonObject, JsonPath, Registered, RegisteredName, TextEdit } from './protocol';
import { databaseFor, request } from './server';

interface NamePick extends vscode.QuickPickItem {
	id: RegisteredName['id'];
}

// The names a placeholder offers: the spec's spells, or its auras on the unit the pattern reads them on.
type Names = Record<'spell' | 'aura' | 'target_aura', NamePick[]>;

// The value of a placeholder, asked with its default filled in; undefined when the reader cancels, and null when an
// optional one is left out.
async function ask(placeholder: AplPlaceholder, names: Names, taken: { groups: string[]; variables: string[] }): Promise<unknown> {
	const fallback = placeholder.default === undefined ? '' : String(placeholder.default);
	const skip = placeholder.optional ? ' (Escape leaves it out)' : '';
	let answer: unknown;
	switch (placeholder.kind) {
		case 'spell':
		case 'aura':
		case 'target_aura':
			answer = (
				await vscode.window.showQuickPick(names[placeholder.kind], {
					title: `${placeholder.label}${skip}`,
					matchOnDescription: true,
					placeHolder: 'Search by name',
				})
			)?.id;
			break;
		case 'choice': {
			const options = (placeholder.options ?? []).map(option => ({ label: option, picked: option === placeholder.default }));
			const first = options.findIndex(option => option.picked);
			answer = (
				await vscode.window.showQuickPick(first > 0 ? [options[first], ...options.filter((_, i) => i !== first)] : options, {
					title: placeholder.label,
				})
			)?.label;
			break;
		}
		// The server reads a time as a const does and refuses the pattern when it reads none.
		case 'duration':
			answer = await vscode.window.showInputBox({ title: `${placeholder.label}${skip}`, value: fallback });
			break;
		case 'number': {
			const value = await vscode.window.showInputBox({
				title: placeholder.label,
				value: fallback,
				validateInput: v => (v.trim() !== '' && Number.isFinite(Number(v)) ? undefined : 'A number'),
			});
			answer = value === undefined ? undefined : Number(value);
			break;
		}
		case 'group':
		case 'variable': {
			const used = placeholder.kind === 'group' ? taken.groups : taken.variables;
			if (placeholder.existing) {
				answer = await vscode.window.showQuickPick(used, { title: `${placeholder.label}${skip}` });
				break;
			}
			// A new group or variable takes a name the rotation does not have: references read the first of two.
			answer = await vscode.window.showInputBox({
				title: placeholder.label,
				value: fallback,
				validateInput: v => (!v.trim() ? 'A name' : used.includes(v) ? `The rotation has a ${placeholder.kind} named ${v} already` : undefined),
			});
			break;
		}
	}
	return answer === undefined && placeholder.optional ? null : answer;
}

// The spec's names for the pickers, each with its ActionID, tag and rank included, and its game name from the database.
function namePicks(registered: Registered, db: AplDatabase): Names {
	if (registered.status !== 'ready') return { spell: [], aura: [], target_aura: [] };
	const picks = (entries: RegisteredName[]) =>
		entries.map(entry => ({
			label: (entry.id.spellId !== undefined && db.spells[entry.id.spellId]?.[0]) || entry.name,
			description: entry.name,
			id: entry.id,
		}));
	const { spells, auras, targetAuras } = registered.registered;
	return { spell: picks(spells), aura: picks(auras), target_aura: picks(targetAuras) };
}

// Whether a path is in a priority or group row, which a condition pattern goes onto.
const onRow = (path: JsonPath) =>
	(path[0] === 'priorityList' && typeof path[1] === 'number' && path[1] >= 0) ||
	(path[0] === 'groups' && path[2] === 'actions' && typeof path[3] === 'number' && path[3] >= 0);

// Inserts a pattern at a rotation path: after the row it is in, or onto that row's condition. The values are asked
// first; the server places the pattern in the text as it read it then, and refuses one whose text moved meanwhile.
export async function insertPattern(uri: vscode.Uri, path: JsonPath): Promise<void> {
	try {
		const patterns = await request<AplPattern[]>(uri, 'siml/patterns', {});
		const pattern = (
			await vscode.window.showQuickPick(
				patterns.map(p => ({ label: p.label, detail: p.description, description: p.target, pattern: p })),
				{ title: 'Insert APL pattern', matchOnDetail: true },
			)
		)?.pattern;
		if (!pattern) return;
		if (pattern.target === 'condition' && !onRow(path)) {
			void vscode.window.showInformationMessage(`${pattern.label} is a condition: put the cursor on a priority or group row to add it there.`);
			return;
		}
		const document = await documentFor(uri);
		const version = document.version;
		const read = await readRotation(document);
		if (read.readOnly) {
			void vscode.window.showWarningMessage(`Cannot insert a pattern: ${read.readOnly}`);
			return;
		}
		const rotation: JsonObject = read.rotation ?? {};
		const names = (list: unknown) => (Array.isArray(list) ? list.map(item => String((item as JsonObject)?.name ?? '')) : []);
		const taken = { groups: names(rotation.groups), variables: names(rotation.valueVariables) };
		const db = await databaseFor(uri);
		const registered = await request<Registered>(uri, 'siml/registered', { uri: uri.toString() });
		const values = await askAll(pattern, namePicks(registered, db), taken);
		if (!values) return;
		const { edits } = await request<{ edits: TextEdit[] }>(uri, 'siml/insertPattern', { uri: uri.toString(), version, id: pattern.id, values, path });
		if ((await applyEdits(await documentFor(uri), version, edits)) === undefined)
			void vscode.window.showWarningMessage(`${pattern.label} was not inserted: the file changed while it was being made. Insert it again.`);
	} catch (e) {
		void vscode.window.showWarningMessage((e as Error).message);
	}
}

// Every placeholder's value, or undefined when the reader cancels one that is not optional.
async function askAll(pattern: AplPattern, names: Names, taken: { groups: string[]; variables: string[] }): Promise<Record<string, unknown> | undefined> {
	const values: Record<string, unknown> = {};
	for (const placeholder of pattern.placeholders) {
		const value = await ask(placeholder, names, taken);
		if (value === undefined) return undefined;
		if (value !== null && value !== '') values[placeholder.name] = value;
	}
	return values;
}

export function registerInsertPattern(context: vscode.ExtensionContext): void {
	context.subscriptions.push(
		vscode.commands.registerCommand('wowsims-apl.insertPattern', async () => {
			const editor = vscode.window.activeTextEditor;
			if (!editor || !isAplSiml(editor.document.uri.path)) return;
			const { uri } = editor.document;
			const { line, character } = editor.selection.active;
			const path = await request<JsonPath>(uri, 'siml/pathAt', { uri: uri.toString(), position: { line, character } }).catch(() => []);
			await insertPattern(uri, path);
		}),
	);
}
