import { randomBytes } from 'node:crypto';
import { existsSync } from 'node:fs';
import { basename } from 'node:path';

import * as vscode from 'vscode';

import { applyEdits, documentFor, isAplSiml, pathAt, readRotation, samePath, spanFor, spansOf, vscodeRange } from './host';
import { insertPattern } from './insert_pattern';
import type { AplDatabase, FromPanel, JsonPath, PanelState, Range, Registered, TextEdit, ToPanel } from './protocol';
import { buildStep, databaseFor, log, onServerStarted, request } from './server';
import { onSimRecord, type SimRecord, simRecord, simulate } from './simulate';

const VIEW_TYPE = 'wowsims-apl.editor';
const ICON_HOST = 'https://wow.zamimg.com';
// How long the panel waits for typing to pause before it reads the text again.
const SETTLE_MS = 60;

// One panel per rotation file; opening the editor for a file that has one reveals it.
const panels = new Map<string, vscode.WebviewPanel>();

// The pages shown while the panel's bundle is missing: while it builds, and when the build failed. Opening the panel
// again after a failed build builds it again.
const message = (text: string) => `<!doctype html>
<html lang="en">
<body>
<p>${text}</p>
</body>
</html>`;
const BUILDING = message('Building the panel with <code>make vscode-apl</code>; it opens when the build ends.');
const FAILED = message(
	'The panel did not build: the WoWSims SIML output has the errors of <code>make vscode-apl</code>. Open the panel again to build it again.',
);

// The build of the panel's bundle running, which every panel opened meanwhile waits on.
let building: Promise<void> | undefined;

// Builds the panel's bundle with `make vscode-apl` in the repository the extension sits in
// (.vscode/extensions/wowsims-apl), which also builds what it needs: the packages and the generated code. The build
// stops with the server builds, when the extension deactivates, and after their timeout.
function buildPage(extensionUri: vscode.Uri): Promise<void> {
	const root = vscode.Uri.joinPath(extensionUri, '..', '..', '..').fsPath;
	building ??= Promise.resolve(
		vscode.window.withProgress({ location: vscode.ProgressLocation.Notification, title: 'Building the APL editor panel (make vscode-apl)' }, () =>
			buildStep('make', ['vscode-apl'], root, { what: 'The APL editor panel', again: 'Open the panel again' }),
		),
	).finally(() => (building = undefined));
	return building;
}

// Shows the panel's page, building its bundle first when the checkout has none: `make vscode-apl` writes it and it is
// not committed. Only a panel still open when the build ends takes the page: a closed one is disposed, and throws.
function showPage(panel: vscode.WebviewPanel, key: string, extensionUri: vscode.Uri): void {
	if (existsSync(vscode.Uri.joinPath(extensionUri, 'out', 'webview.js').fsPath)) {
		panel.webview.html = html(panel.webview, extensionUri);
		return;
	}
	panel.webview.html = BUILDING;
	const live = () => panels.get(key) === panel;
	buildPage(extensionUri).then(
		() => {
			if (live()) panel.webview.html = html(panel.webview, extensionUri);
		},
		(error: Error) => {
			log(error.message);
			if (live()) panel.webview.html = FAILED;
		},
	);
}

// The panel's page, from the bundle `make vscode-apl` builds.
function html(webview: vscode.Webview, extensionUri: vscode.Uri): string {
	const nonce = randomBytes(16).toString('base64');
	const asset = (file: string) => webview.asWebviewUri(vscode.Uri.joinPath(extensionUri, 'out', file));
	const csp = [
		`default-src 'none'`,
		`img-src ${webview.cspSource} ${ICON_HOST} data:`,
		`style-src ${webview.cspSource} 'unsafe-inline'`,
		`script-src 'nonce-${nonce}'`,
	].join('; ');
	return `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta http-equiv="Content-Security-Policy" content="${csp}">
<meta name="viewport" content="width=device-width, initial-scale=1">
<link rel="stylesheet" href="${asset('webview.css')}">
</head>
<body>
<div id="root"></div>
<script nonce="${nonce}" src="${asset('webview.js')}"></script>
</body>
</html>`;
}

// Ties a panel to its file. The panel holds the file's URI only, and finds the open document by it on every message.
function bind(panel: vscode.WebviewPanel, uri: vscode.Uri, extensionUri: vscode.Uri): void {
	const key = uri.toString();
	panels.set(key, panel);
	panel.webview.options = { enableScripts: true, localResourceRoots: [vscode.Uri.joinPath(extensionUri, 'out')] };
	panel.title = `APL: ${basename(uri.fsPath)}`;
	showPage(panel, key, extensionUri);

	const post = (message: ToPanel) => void panel.webview.postMessage(message);
	// The last of the panel's edits made or refused, and the document versions its own edits made.
	let handled = 0;
	const own = new Set<number>();
	// The spec and name table whose names the pickers hold, and the database read the page holds; read again when
	// one changes, undefined before the first read.
	let spec: string | null | undefined;
	let names: string | undefined;
	let database: string | undefined;
	// The files the rotation's `use` lines read, whose text it is read with.
	let used: string[] = [];
	const uses = (file: vscode.Uri) => file.scheme === 'file' && used.some(path => samePath(path, file.fsPath));

	const postRegistered = async () => {
		const registered = await request<Registered>(uri, 'siml/registered', { uri: key }).catch((e: Error): Registered => ({
			status: 'unavailable',
			reason: e.message,
		}));
		post({ type: 'registered', registered });
	};
	const postRotation = async (document: vscode.TextDocument) => {
		const version = document.version;
		const read = await readRotation(document).catch((e: Error) => e);
		if (read instanceof Error) {
			post({ type: 'rotation', version, handled, readOnly: read.message, findings: [] });
			return;
		}
		used = read.used;
		const { rotation, readOnly, className, findings, structure } = read;
		post({ type: 'rotation', version: read.version, handled, rotation, readOnly, className, findings, structure });
		if ((read.spec ?? null) !== spec || read.names !== names) {
			spec = read.spec ?? null;
			names = read.names;
			await postRegistered();
		}
		if (database !== undefined && read.database !== database) {
			database = read.database;
			await postInit();
		}
		database = read.database;
	};
	const postSim = (version: number, record: SimRecord | undefined = simRecord(key), running = false) =>
		post({ type: 'sim', running, sim: record && { result: record.result, stale: record.version !== version } });

	// While an edit of the panel is being made, the text's changes are not posted: the edit posts the text it leaves,
	// with the count of edits made that it holds.
	let editing = false;
	let settle: ReturnType<typeof setTimeout> | undefined;
	const changed = () => {
		clearTimeout(settle);
		settle = setTimeout(async () => {
			if (editing) return;
			const document = await documentFor(uri);
			await postRotation(document);
			postSim(document.version);
		}, SETTLE_MS);
	};

	// The panel's edits, one at a time in the order it sent them. An edit made against a version that only the panel's
	// own edits have moved past is made on the text as it is; the panel computed it with those edits applied. One that
	// someone else's change moved past is dropped.
	const make = async (message: Extract<FromPanel, { type: 'edit' }>, document: vscode.TextDocument) => {
		for (let version = message.version + 1; version <= document.version; version++) if (!own.has(version)) return;
		if (message.version > document.version) return;
		const spliced = document.version;
		let edits: TextEdit[];
		try {
			({ edits } = await request<{ edits: TextEdit[] }>(uri, 'siml/splice', { uri: key, version: spliced, edits: message.edits }));
		} catch (e) {
			void vscode.window.showWarningMessage(`The panel's edit was not made: ${(e as Error).message}.`);
			return;
		}
		if (edits.length === 0) return;
		// Text that changed while the server worked is not the text the edits are for; the edit is dropped as above.
		const version = await applyEdits(document, spliced, edits);
		if (version !== undefined) own.add(version);
	};
	const edit = async (message: Extract<FromPanel, { type: 'edit' }>) => {
		editing = true;
		try {
			await make(message, await documentFor(uri));
		} finally {
			editing = false;
			handled = message.seq;
			await postRotation(await documentFor(uri));
		}
	};
	let edits = Promise.resolve();

	const reveal = async (path: (string | number)[], at?: Range) => {
		const document = await documentFor(uri);
		const range = at ?? spanFor(await spansOf(document).catch(() => []), path);
		if (!range) return;
		const visible = vscode.window.visibleTextEditors.find(editor => editor.document.uri.toString() === key);
		const editor = await vscode.window.showTextDocument(document, { viewColumn: visible?.viewColumn ?? vscode.ViewColumn.One, preserveFocus: true });
		const selection = vscodeRange(range);
		editor.selection = new vscode.Selection(selection.start, selection.end);
		editor.revealRange(new vscode.Range(selection.start, selection.start), vscode.TextEditorRevealType.InCenterIfOutsideViewport);
	};

	// The database, posted when the page loads and again when a server starts; a failed read posts it empty.
	let loaded = false;
	const postInit = async () => {
		const db = await databaseFor(uri, database).catch((): AplDatabase => ({ spells: {}, items: {} }));
		post({ type: 'init', db, fileName: basename(uri.fsPath), uri: key });
	};

	const disposables: vscode.Disposable[] = [
		panel.webview.onDidReceiveMessage(async (message: FromPanel) => {
			switch (message.type) {
				case 'ready': {
					// A page loaded again counts its edits from the start.
					const document = await documentFor(uri);
					handled = 0;
					spec = undefined;
					names = undefined;
					loaded = true;
					await postInit();
					await postRotation(document);
					postSim(document.version);
					break;
				}
				case 'edit':
					edits = edits.then(() => edit(message)).catch(e => void vscode.window.showErrorMessage(`The panel's edit failed: ${(e as Error).message}`));
					break;
				case 'sim': {
					const document = await documentFor(uri);
					postSim(document.version, undefined, true);
					await simulate(document, message.compare);
					postSim(document.version);
					break;
				}
				case 'pattern':
					await insertPattern(uri, message.path);
					break;
				case 'reveal':
					await reveal(message.path);
					break;
				case 'revealRange':
					await reveal([], message.range);
					break;
			}
		}),
		vscode.workspace.onDidChangeTextDocument(event => {
			if (event.contentChanges.length > 0 && (event.document.uri.toString() === key || uses(event.document.uri))) changed();
		}),
		// A used file closed unsaved is read from disk again.
		vscode.workspace.onDidCloseTextDocument(document => {
			if (uses(document.uri)) changed();
		}),
		onSimRecord(changedUri => {
			if (changedUri === key) changed();
		}),
		// The server checks the file again when the name table or the database changes, and a restarted server reads
		// it and the database afresh.
		vscode.languages.onDidChangeDiagnostics(event => {
			if (event.uris.some(changedUri => changedUri.toString() === key)) changed();
		}),
		onServerStarted(() => {
			database = undefined;
			if (loaded) void postInit();
			spec = undefined;
			changed();
		}),
		vscode.window.onDidChangeTextEditorSelection(async event => {
			const document = event.textEditor.document;
			if (document.uri.toString() !== key || event.kind === vscode.TextEditorSelectionChangeKind.Command) return;
			const path = await pathAt(document.uri, event.selections[0].active).catch(() => undefined);
			if (path) post({ type: 'cursor', path });
		}),
	];

	panel.onDidDispose(() => {
		clearTimeout(settle);
		panels.delete(key);
		disposables.forEach(d => d.dispose());
	});
}

// Opens the panel beside an .apl.siml file, or reveals the one it has.
async function openToSide(extensionUri: vscode.Uri, uri: vscode.Uri | undefined): Promise<void> {
	const chosen = uri ?? vscode.window.activeTextEditor?.document.uri;
	if (!chosen || !isAplSiml(chosen.path)) {
		const message = chosen
			? `The APL editor panel opens on .apl.siml rotations only, and ${basename(chosen.fsPath)} is not one. The SIML server still checks, completes, highlights and formats it.`
			: 'The APL editor panel opens on an .apl.siml rotation: open one first.';
		log(message);
		void vscode.window.showInformationMessage(message);
		return;
	}
	const open = panels.get(chosen.toString());
	if (open) {
		if (open.webview.html === FAILED) showPage(open, chosen.toString(), extensionUri);
		open.reveal(vscode.ViewColumn.Beside, true);
		return;
	}
	const panel = vscode.window.createWebviewPanel(VIEW_TYPE, basename(chosen.fsPath), { viewColumn: vscode.ViewColumn.Beside, preserveFocus: true });
	bind(panel, chosen, extensionUri);
}

export function registerPanel(context: vscode.ExtensionContext): void {
	context.subscriptions.push(
		vscode.commands.registerCommand('wowsims-apl.openToSide', (uri?: vscode.Uri) => openToSide(context.extensionUri, uri)),
		vscode.window.registerWebviewPanelSerializer(VIEW_TYPE, {
			async deserializeWebviewPanel(panel: vscode.WebviewPanel, state?: PanelState) {
				if (state?.uri && isAplSiml(state.uri)) bind(panel, vscode.Uri.parse(state.uri), context.extensionUri);
				else panel.dispose();
			},
		}),
	);
}
