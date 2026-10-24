// A stand-in for the vscode module, enough to run the extension's host in Node: documents in memory that
// WorkspaceEdits change, webview panels whose messages a test sends and reads, and the registered commands. Every other
// name the API exports is a value that takes any use. The language client is a stand-in too, which starts the server
// binary the extension built and keeps the documents in memory open in it, as VS Code's client does.

import { spawn } from 'node:child_process';
import { mkdtempSync, readFileSync, rmSync } from 'node:fs';
import Module from 'node:module';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';

import { createMessageConnection, StreamMessageReader, StreamMessageWriter } from 'vscode-jsonrpc/node';

import { bundleExtension } from '../scripts/bundle.mjs';

const require = Module.createRequire(import.meta.url);

// Anything at all: callable, constructible, and each of its properties is anything again.
const anything = () =>
	new Proxy(function () {}, {
		get: (_, key) => (key === 'prototype' ? {} : key === Symbol.toPrimitive ? () => 0 : anything()),
		apply: () => anything(),
		construct: () => anything(),
	});

class Emitter {
	listeners = new Set();
	event = listener => {
		this.listeners.add(listener);
		return { dispose: () => this.listeners.delete(listener) };
	};
	fire(value) {
		for (const listener of [...this.listeners]) listener(value);
	}
}

class Uri {
	constructor(path, scheme = 'file') {
		this.scheme = scheme;
		this.path = path;
		this.fsPath = path;
	}
	static file(path) {
		return new Uri(path);
	}
	static parse(text) {
		return text.startsWith('file:') ? new Uri(fileURLToPath(text)) : new Uri(text.slice(text.indexOf(':') + 1), text.slice(0, text.indexOf(':')));
	}
	static joinPath(base, ...parts) {
		return new Uri([base.path, ...parts].join('/'));
	}
	toString() {
		return this.scheme === 'file' ? pathToFileURL(this.path).href : `${this.scheme}:${this.path}`;
	}
}

class Position {
	constructor(line, character) {
		this.line = line;
		this.character = character;
	}
}

class Range {
	constructor(a, b, c, d) {
		[this.start, this.end] = typeof a === 'number' ? [new Position(a, b), new Position(c, d)] : [a, b];
	}
}

class Selection extends Range {}

class WorkspaceEdit {
	edits = [];
	replace(uri, range, newText) {
		this.edits.push({ uri, range, newText });
	}
}

class ThemeColor {
	constructor(id) {
		this.id = id;
	}
}

// A document in memory, as VS Code keeps one: its text, a version each change counts up, and offset conversion.
class TextDocument {
	constructor(uri, text) {
		this.uri = uri;
		this.languageId = uri.path.endsWith('.siml') ? 'siml' : 'plaintext';
		this.fileName = uri.fsPath;
		this.text = text;
		this.version = 1;
		this.isDirty = false;
		this.isClosed = false;
	}
	getText() {
		return this.text;
	}
	offsetAt(position) {
		const lines = this.text.split('\n');
		let offset = 0;
		for (let line = 0; line < position.line && line < lines.length; line++) offset += lines[line].length + 1;
		return Math.min(offset + Math.min(position.character, lines[position.line]?.length ?? 0), this.text.length);
	}
	positionAt(offset) {
		const before = this.text.slice(0, offset);
		const line = before.split('\n').length - 1;
		return new Position(line, offset - (before.lastIndexOf('\n') + 1));
	}
}

// The stand-in and what it records: documents, panels, commands, messages shown, documents opened from disk, the
// WorkspaceEdits applied and the titles of what it asked.
export function fakeVscode() {
	const opens = new Emitter();
	const changes = new Emitter();
	const closes = new Emitter();
	const selections = new Emitter();
	// `answers` are what the quick picks and input boxes answer, in order: a value, or a function of the items or options.
	// `config` holds settings of wowsims-apl; `progress` the cancellable tasks running, each with its `cancel`.
	const state = {
		documents: [],
		panels: [],
		commands: new Map(),
		serializers: [],
		messages: [],
		opened: [],
		applied: [],
		answers: [],
		asked: [],
		clients: [],
		config: {},
		progress: [],
	};
	const answer = async (shown, options) => {
		state.asked.push(options?.title);
		const next = state.answers.shift();
		return typeof next === 'function' ? next(await shown) : next;
	};

	// Replaces a document's text as one change, the way an edit or typing reaches the extension.
	const change = (document, text) => {
		document.text = text;
		document.version++;
		document.isDirty = true;
		changes.fire({ document, contentChanges: [{ text }] });
	};

	const workspace = {
		get textDocuments() {
			return state.documents.filter(d => !d.isClosed);
		},
		async openTextDocument(uri) {
			const open = state.documents.find(d => !d.isClosed && d.uri.toString() === uri.toString());
			if (open) return open;
			const document = new TextDocument(uri, readFileSync(uri.fsPath, 'utf8'));
			state.documents.push(document);
			state.opened.push(uri.toString());
			opens.fire(document);
			return document;
		},
		async applyEdit(edit) {
			state.applied.push(edit.edits);
			for (const [key, edits] of Map.groupBy(edit.edits, e => e.uri.toString())) {
				const document = state.documents.find(d => !d.isClosed && d.uri.toString() === key);
				if (!document) return false;
				let text = document.text;
				const located = edits.map(e => ({ start: document.offsetAt(e.range.start), end: document.offsetAt(e.range.end), newText: e.newText }));
				for (const e of located.sort((a, b) => b.start - a.start)) text = text.slice(0, e.start) + e.newText + text.slice(e.end);
				change(document, text);
			}
			return true;
		},
		onDidOpenTextDocument: opens.event,
		onDidChangeTextDocument: changes.event,
		onDidCloseTextDocument: closes.event,
		workspaceFolders: undefined,
		getConfiguration: () => ({ get: (key, fallback) => state.config[key] ?? fallback }),
	};

	const window = {
		visibleTextEditors: [],
		activeTextEditor: undefined,
		createWebviewPanel(viewType, title) {
			const received = new Emitter();
			const disposed = new Emitter();
			const panel = {
				viewType,
				title,
				posted: [],
				webview: {
					options: {},
					html: '',
					cspSource: 'vscode-webview:',
					asWebviewUri: uri => uri,
					postMessage(message) {
						panel.posted.push(message);
						panel.onPost?.(message);
						return Promise.resolve(true);
					},
					onDidReceiveMessage: received.event,
				},
				// Sends a message from the panel's page to the extension.
				send: message => received.fire(message),
				reveal() {},
				onDidDispose: disposed.event,
				dispose: () => disposed.fire(),
			};
			state.panels.push(panel);
			return panel;
		},
		registerWebviewPanelSerializer(viewType) {
			state.serializers.push(viewType);
			return { dispose() {} };
		},
		showQuickPick: (items, options) => answer(items, options),
		showInputBox: options => answer(options, options),
		showInformationMessage: message => void state.messages.push({ severity: 'info', message }),
		showWarningMessage: message => void state.messages.push({ severity: 'warning', message }),
		showErrorMessage: message => void state.messages.push({ severity: 'error', message }),
		// The extension's output channel, printed to the test's stderr: a server that does not build logs its compiler's
		// errors there.
		createOutputChannel: name =>
			new Proxy(
				{},
				{ get: (_, key) => (['show', 'hide', 'dispose', 'clear'].includes(key) ? () => {} : (...text) => console.error(`[${name}]`, ...text)) },
			),
		setStatusBarMessage: () => ({ dispose() {} }),
		createTextEditorDecorationType: () => ({ dispose() {} }),
		onDidChangeTextEditorSelection: selections.event,
		onDidChangeVisibleTextEditors: () => ({ dispose() {} }),
		async withProgress(options, task) {
			const cancelled = new Emitter();
			const token = { isCancellationRequested: false, onCancellationRequested: cancelled.event };
			const running = {
				title: options.title,
				cancel() {
					token.isCancellationRequested = true;
					cancelled.fire();
				},
			};
			state.progress.push(running);
			try {
				return await task({ report() {} }, token);
			} finally {
				state.progress.splice(state.progress.indexOf(running), 1);
			}
		},
	};

	const commands = {
		registerCommand(id, handler) {
			state.commands.set(id, handler);
			return { dispose() {} };
		},
	};

	const declarations = readFileSync(require.resolve('@types/vscode/index.d.ts'), 'utf8');
	const vscode = Object.fromEntries([...declarations.matchAll(/^\texport (?:class|enum|namespace) (\w+)/gm)].map(([, name]) => [name, anything()]));
	const stand = (target, real) => new Proxy(real, { get: (r, key) => (key in r ? r[key] : anything()) });
	// The language client: it starts the server the extension names and keeps every .siml document of the stand-in
	// open in it, its text sent whole on each change.
	class LanguageClient {
		constructor(id, name, server, options) {
			this.server = server;
			this.options = options;
			this.stderr = '';
			state.clients.push(this);
		}
		async start() {
			const { command, args, options } = this.server;
			this.process = spawn(command, args, { cwd: options.cwd, stdio: ['pipe', 'pipe', 'pipe'] });
			this.process.stderr.on('data', chunk => (this.stderr += chunk));
			this.exited = new Promise(resolve => this.process.on('exit', resolve));
			const connection = createMessageConnection(new StreamMessageReader(this.process.stdout), new StreamMessageWriter(this.process.stdin));
			connection.onRequest(() => null);
			connection.onNotification(() => {});
			connection.listen();
			await connection.sendRequest('initialize', {
				processId: process.pid,
				rootUri: null,
				workspaceFolders: null,
				capabilities: {},
				initializationOptions: this.options.initializationOptions,
			});
			connection.sendNotification('initialized', {});
			this.connection = connection;
			const siml = document => document.languageId === 'siml';
			// The middleware's didOpen, didChange and didClose decide what reaches the server, as in the real client.
			const middleware = this.options.middleware ?? {};
			const through = (hook, data, send) => (middleware[hook] ? middleware[hook](data, send) : send(data));
			this.opened = new Set();
			const open = document =>
				through('didOpen', document, async () => {
					this.opened.add(document.uri.fsPath);
					connection.sendNotification('textDocument/didOpen', {
						textDocument: { uri: document.uri.toString(), languageId: 'siml', version: document.version, text: document.text },
					});
				});
			this.subscriptions = [
				opens.event(document => siml(document) && open(document)),
				changes.event(
					change =>
						siml(change.document) &&
						through('didChange', change, async ({ document }) =>
							connection.sendNotification('textDocument/didChange', {
								textDocument: { uri: document.uri.toString(), version: document.version },
								contentChanges: [{ text: document.text }],
							}),
						),
				),
				closes.event(
					document =>
						siml(document) &&
						through('didClose', document, async () => {
							this.opened.delete(document.uri.fsPath);
							connection.sendNotification('textDocument/didClose', { textDocument: { uri: document.uri.toString() } });
						}),
				),
			];
			workspace.textDocuments.filter(siml).forEach(open);
			this.running = true;
		}
		sendRequest(method, ...params) {
			return this.connection.sendRequest(method, ...params);
		}
		async stop() {
			this.running = false;
			this.subscriptions.forEach(s => s.dispose());
			await this.connection.sendRequest('shutdown');
			await this.connection.sendNotification('exit');
			await this.exited;
			this.connection.dispose();
		}
	}

	Object.assign(vscode, {
		Uri,
		Position,
		Range,
		Selection,
		WorkspaceEdit,
		ThemeColor,
		EventEmitter: class extends Emitter {
			dispose() {}
		},
		ViewColumn: { One: 1, Two: 2, Beside: -2 },
		TextEditorRevealType: { InCenterIfOutsideViewport: 2 },
		TextEditorSelectionChangeKind: { Keyboard: 1, Mouse: 2, Command: 3 },
		ProgressLocation: { Notification: 15 },
		workspace: stand(vscode.workspace, workspace),
		window: stand(vscode.window, window),
		commands: stand(vscode.commands, commands),
	});

	return {
		vscode,
		languageClient: { LanguageClient, RevealOutputChannelOn: { Never: 4 } },
		state,
		change,
		// Closes a document, as VS Code does with one no editor shows.
		close(document) {
			document.isClosed = true;
			closes.fire(document);
		},
		// Moves the cursor of an editor on the document, as a click in the text does.
		select(document, offset) {
			const position = document.positionAt(offset);
			selections.fire({ textEditor: { document }, selections: [{ active: position }], kind: 2 });
		},
	};
}

// The extension bundled from its sources as the build bundles it, with the language client left out for the
// stand-in's.
const bundled = (async () => {
	const dir = mkdtempSync(join(tmpdir(), 'wowsims-apl-extension-'));
	const file = join(dir, 'extension.js');
	await bundleExtension(file, ['vscode-languageclient/node']);
	process.on('exit', () => rmSync(dir, { recursive: true, force: true }));
	return file;
})();

// Where the extension builds the server in the tests: kept between runs, so a build after the first relinks nothing.
const storage = join(tmpdir(), 'wowsims-apl-test-storage');

// Loads the extension with the stand-in and activates it as installed at `installed`; the module, whose deactivate
// stops the servers it started.
export async function activateBuilt(stand, installed = fileURLToPath(new URL('../../../.vscode/extensions/wowsims-apl', import.meta.url))) {
	const built = await bundled;
	const load = Module._load;
	Module._load = function (request, ...rest) {
		if (request === 'vscode') return stand.vscode;
		if (request === 'vscode-languageclient/node') return stand.languageClient;
		return load.call(this, request, ...rest);
	};
	try {
		delete require.cache[built];
		const extension = require(built);
		extension.activate({ subscriptions: [], extensionUri: Uri.file(installed), globalStorageUri: Uri.file(storage) });
		return extension;
	} finally {
		Module._load = load;
	}
}

// Waits for the stand-in's first server to run, which the first .siml document opened builds and starts; fails as soon
// as the extension reports that it did not build.
export async function serverRunning(stand) {
	const end = Date.now() + 600000;
	while (!stand.state.clients[0]?.running) {
		const failed = stand.state.messages.find(m => m.severity === 'error' && /did not build|No go binary/.test(m.message));
		if (failed) throw new Error(failed.message);
		if (Date.now() > end) throw new Error('the SIML server did not start');
		await new Promise(resolve => setTimeout(resolve, 50));
	}
}
