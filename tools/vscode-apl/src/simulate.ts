import * as vscode from 'vscode';

import { isAplSiml, spanFor, spansOf } from './host';
import type { SimResult } from './protocol';
import { request } from './server';
import { differenceText, perFight, standardError } from './sim_text';

// The last sim of each rotation file, with the document version it ran against.
export interface SimRecord {
	result: SimResult;
	version: number;
}

const records = new Map<string, SimRecord>();
const listeners = new Set<(uri: string, record: SimRecord | undefined) => void>();

export function simRecord(uri: string): SimRecord | undefined {
	return records.get(uri);
}

export function onSimRecord(listener: (uri: string, record: SimRecord | undefined) => void): vscode.Disposable {
	listeners.add(listener);
	return { dispose: () => listeners.delete(listener) };
}

function setRecord(uri: string, record: SimRecord): void {
	records.set(uri, record);
	listeners.forEach(listener => listener(uri, record));
}

let decorations: { counts: vscode.TextEditorDecorationType; dead: vscode.TextEditorDecorationType };

// The count after each row's line and the never-ran marker after each row the sim never ran, while the editor's text
// is the version that was simmed. Rows of a used file have no line here.
async function decorate(editor: vscode.TextEditor): Promise<void> {
	const { document } = editor;
	if (!isAplSiml(document.uri.path)) return;
	const record = records.get(document.uri.toString());
	const ran: vscode.DecorationOptions[] = [];
	const never: vscode.DecorationOptions[] = [];
	const spans = record?.version === document.version ? await spansOf(document).catch(() => undefined) : undefined;
	if (record && spans && record.version === document.version) {
		const run = record.result.compare ?? record.result;
		for (const [key, stats] of Object.entries(run.rows)) {
			const line = stats.hidden ? undefined : spanFor(spans, key.split('/'))?.start.line;
			if (line === undefined) continue;
			const range = new vscode.Range(line, Number.MAX_SAFE_INTEGER, line, Number.MAX_SAFE_INTEGER);
			if (stats.executions === 0) never.push({ range, renderOptions: { after: { contentText: 'never ran' } } });
			else ran.push({ range, renderOptions: { after: { contentText: `▸ ${perFight(stats.executions)} per fight` } } });
		}
	}
	editor.setDecorations(decorations.counts, ran);
	editor.setDecorations(decorations.dead, never);
}

const decorateAll = (uri: vscode.Uri) =>
	vscode.window.visibleTextEditors.filter(editor => editor.document.uri.toString() === uri.toString()).forEach(editor => void decorate(editor));

// Sims the editor's text of an .apl.siml in the server. A compare sims the saved file, or HEAD when nothing is
// unsaved, as the base, and the editor's text against it on the same seed.
export async function simulate(document: vscode.TextDocument, compare: boolean): Promise<void> {
	if (!isAplSiml(document.uri.path)) {
		void vscode.window.showWarningMessage('The sim runs an .apl.siml rotation in a WoWSims repository.');
		return;
	}
	const version = document.version;
	const config = vscode.workspace.getConfiguration('wowsims-apl');
	const name = document.uri.path.split('/').at(-1);
	await vscode.window.withProgress(
		{ location: vscode.ProgressLocation.Notification, title: compare ? `Comparing ${name}` : `Simming ${name}`, cancellable: true },
		async (_, token) => {
			try {
				const params = {
					uri: document.uri.toString(),
					compare,
					settingsFile: config.get<string>('simSettingsFile') || undefined,
					iterations: config.get<number>('simIterations', 1000),
					timeline: true,
				};
				const result = await request<SimResult>(document.uri, 'siml/sim', params, token);
				setRecord(document.uri.toString(), { result, version });
				decorateAll(document.uri);
				const run = result.compare ?? result;
				const summary = result.difference
					? `${run.dps.mean.toFixed(1)} DPS, ${differenceText(result.difference)} against ${result.baseline}`
					: `${run.dps.mean.toFixed(1)} DPS ± ${standardError(run).toFixed(1)}`;
				void vscode.window.setStatusBarMessage(`$(pulse) ${name}: ${summary}`, 30000);
				if (result.character.warnings?.length) void vscode.window.showWarningMessage(`Sim character: ${result.character.warnings.join('; ')}`);
			} catch (e) {
				if (token.isCancellationRequested || (e as { code?: number }).code === -32800) return;
				void vscode.window.showWarningMessage((e as Error).message);
			}
		},
	);
}

export function registerSim(context: vscode.ExtensionContext): void {
	decorations = {
		counts: vscode.window.createTextEditorDecorationType({
			after: { margin: '0 0 0 2em', color: new vscode.ThemeColor('editorCodeLens.foreground'), fontStyle: 'italic' },
		}),
		dead: vscode.window.createTextEditorDecorationType({
			after: { margin: '0 0 0 2em', color: new vscode.ThemeColor('editorWarning.foreground'), fontStyle: 'italic' },
		}),
	};
	const target = (uri?: vscode.Uri) => (uri ? vscode.workspace.openTextDocument(uri) : Promise.resolve(vscode.window.activeTextEditor?.document));
	context.subscriptions.push(
		decorations.counts,
		decorations.dead,
		vscode.commands.registerCommand('wowsims-apl.sim', async (uri?: vscode.Uri) => {
			const document = await target(uri);
			if (document) await simulate(document, false);
		}),
		vscode.commands.registerCommand('wowsims-apl.compare', async (uri?: vscode.Uri) => {
			const document = await target(uri);
			if (document) await simulate(document, true);
		}),
		vscode.window.onDidChangeVisibleTextEditors(editors => editors.forEach(editor => void decorate(editor))),
		vscode.workspace.onDidChangeTextDocument(event => {
			if (records.has(event.document.uri.toString()) && event.contentChanges.length > 0) decorateAll(event.document.uri);
		}),
	);
}
