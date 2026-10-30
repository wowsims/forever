import { execFile } from 'node:child_process';
import { createHash } from 'node:crypto';
import { existsSync } from 'node:fs';
import { dirname, join } from 'node:path';

import * as vscode from 'vscode';
import { LanguageClient, RevealOutputChannelOn } from 'vscode-languageclient/node';

import { findGo } from '../../vscode-common/go';
import type { AplDatabase } from './protocol';

// The SIML language server, `wowsimcli siml lsp`, built from the repository each file sits in and started there: one
// per repository, which serves its *.siml files and answers the panel's requests about them.

const clients = new Map<string, Promise<LanguageClient>>();
const started = new vscode.EventEmitter<void>();
// Fires when a server starts, after a restart too: what was read from the one before is read again.
export const onServerStarted = started.event;

// The builds running, which a restart or deactivation stops.
const builds = new Set<AbortController>();
// How long a build may take before it is stopped: a cold build of the repository takes a few minutes.
const BUILD_TIMEOUT_MS = 10 * 60 * 1000;

// What a build is of, for its errors, and what starts it again.
export interface Build {
	what: string;
	again: string;
}
const SERVER_BUILD: Build = { what: 'The SIML server', again: 'Run WoWSims: Restart SIML Server' };

// One step of a build in root, which a restart or deactivation stops, as does BUILD_TIMEOUT_MS. A failed step writes
// its output to the WoWSims SIML output, and its error says why `what` did not build and how to build it `again`.
export function buildStep(file: string, args: string[], root: string, { what, again }: Build): Promise<void> {
	const build = new AbortController();
	builds.add(build);
	return new Promise<void>((resolve, reject) =>
		execFile(file, args, { cwd: root, maxBuffer: 64 << 20, signal: build.signal, timeout: BUILD_TIMEOUT_MS }, (error, stdout, stderr) => {
			if (!error) return resolve();
			if (build.signal.aborted) return reject(new Error(`${what} build was stopped.`));
			if (error.killed)
				return reject(new Error(`${what} did not build in ${root} within ${BUILD_TIMEOUT_MS / 60000} minutes. ${again} to build it again.`));
			logFailedBuild(`${file} ${args.join(' ')} in ${root}\n${stdout}${stderr || error.message}`);
			reject(new Error(`${what} did not build in ${root}: the WoWSims SIML output has the errors. ${again} once they are fixed.`));
		}),
	).finally(() => builds.delete(build));
}

let context: vscode.ExtensionContext;
let output: vscode.LogOutputChannel;
// Cleared when the extension deactivates, after which nothing starts a server again.
let active = true;

// The repository each folder is in, found once: every change of a document asks. WoWSims: Restart SIML Server
// looks again.
const roots = new Map<string, string | undefined>();

// The repository a file is in: the nearest folder above it with sim/siml/names.json.
function repositoryRoot(file: string): string | undefined {
	const start = dirname(file);
	if (!roots.has(start)) {
		let found: string | undefined;
		for (let dir = start; ; dir = dirname(dir)) {
			if (existsSync(join(dir, 'sim', 'siml', 'names.json'))) {
				found = dir;
				break;
			}
			if (dirname(dir) === dir) break;
		}
		roots.set(start, found);
	}
	return roots.get(start);
}

// The repository a document is read in; for an untitled one, the first workspace folder's.
function rootOf(uri: vscode.Uri): string | undefined {
	if (uri.scheme === 'file') return repositoryRoot(uri.fsPath);
	for (const folder of vscode.workspace.workspaceFolders ?? []) {
		const root = repositoryRoot(join(folder.uri.fsPath, 'x'));
		if (root) return root;
	}
	return undefined;
}

const settings = () => vscode.workspace.getConfiguration('wowsims-apl');

// The server binary for a repository: the `wowsims-apl.server` setting's, or one built from the repository into the
// extension's storage. The server imports the generated Go protos, which git does not hold, so make writes them first
// when proto/ is newer, as after pulling a branch that changes a proto; it does nothing when they are current, and
// `go build` relinks nothing when the binary is up to date.
async function serverBinary(root: string): Promise<string> {
	const prebuilt = settings().get<string>('server', '');
	if (prebuilt) return prebuilt;
	const go = findGo(settings().get<string>('goBinary', 'go'));
	if (!go) throw new Error('No go binary was found to build the SIML server; set wowsims-apl.goBinary.');
	const hash = createHash('sha256').update(root).digest('hex').slice(0, 16);
	const binary = join(vscode.Uri.joinPath(context.globalStorageUri, hash).fsPath, process.platform === 'win32' ? 'wowsimcli.exe' : 'wowsimcli');
	await vscode.window.withProgress({ location: vscode.ProgressLocation.Notification, title: 'Building the SIML server' }, async () => {
		await buildStep('make', ['sim/core/proto/api.pb.go'], root, SERVER_BUILD);
		await buildStep(go, ['build', '--tags=with_db', '-o', binary, './cmd/wowsimcli'], root, SERVER_BUILD);
	});
	return binary;
}

async function start(root: string): Promise<LanguageClient> {
	const command = await serverBinary(root);
	// A file in a repository nested in this one, such as a worktree under .claude/worktrees, is that repository's: this
	// server is never told it is open, so it answers nothing about it and the nested one's server serves it alone.
	const ours = (document: vscode.TextDocument) => document.uri.scheme !== 'file' || repositoryRoot(document.uri.fsPath) === root;
	const client = new LanguageClient(
		'wowsims-siml',
		'WoWSims SIML',
		{ command, args: ['siml', 'lsp'], options: { cwd: root } },
		{
			documentSelector: [
				{ scheme: 'file', language: 'siml', pattern: { baseUri: vscode.Uri.file(root).toString(), pattern: '**/*' } },
				// Untitled files are read in the first workspace folder's repository, as the server reads them.
				...(rootOf(vscode.Uri.parse('untitled:x')) === root ? [{ scheme: 'untitled', language: 'siml' }] : []),
			],
			synchronize: { configurationSection: 'wowsims-apl' },
			outputChannel: output,
			revealOutputChannelOn: RevealOutputChannelOn.Never,
			middleware: {
				didOpen: (document, next) => (ours(document) ? next(document) : Promise.resolve()),
				didChange: (change, next) => (ours(change.document) ? next(change) : Promise.resolve()),
				didClose: (document, next) => (ours(document) ? next(document) : Promise.resolve()),
			},
		},
	);
	await client.start();
	output.info(`The SIML server for ${root} is running: ${command} siml lsp`);
	started.fire();
	return client;
}

// The server of the repository a document is in, built and started on first use. Rejects with a message for the
// reader when the document is in no repository or the server does not build or start; a failed start stays failed
// until WoWSims: Restart SIML Server.
export function serverFor(uri: vscode.Uri): Promise<LanguageClient> {
	const root = rootOf(uri);
	if (!active) return Promise.reject(new Error('The extension is stopping.'));
	if (!root) return Promise.reject(new Error('The file is not in a WoWSims repository: no sim/siml/names.json above it.'));
	let client = clients.get(root);
	if (!client) {
		client = start(root);
		clients.set(root, client);
		// A start a restart replaced fails unreported.
		const starting = client;
		client.catch(e => clients.get(root) === starting && void vscode.window.showErrorMessage((e as Error).message));
	}
	return client;
}

// Sends one of the server's requests about a document; `token` cancels it.
export async function request<R>(uri: vscode.Uri, method: string, params: object, token?: vscode.CancellationToken): Promise<R> {
	const client = await serverFor(uri);
	return token ? client.sendRequest<R>(method, params, token) : client.sendRequest<R>(method, params);
}

// The database's names and icons, read once from each server for each database read it holds (`database` of
// siml/rotation); without one, the read held last. A failed read is asked again next time.
const databases = new WeakMap<LanguageClient, { read?: string; db: Promise<AplDatabase> }>();
export async function databaseFor(uri: vscode.Uri, read?: string): Promise<AplDatabase> {
	const client = await serverFor(uri);
	const held = databases.get(client);
	if (held && (read === undefined || held.read === undefined || held.read === read)) {
		held.read ??= read;
		return held.db;
	}
	const db = client.sendRequest<AplDatabase>('siml/database', { uri: uri.toString() });
	databases.set(client, { read, db });
	db.catch(() => databases.delete(client));
	return db;
}

// Stops every server, which ends the sims they run; with `last`, for good.
export async function stopServers(last = false): Promise<void> {
	active &&= !last;
	for (const build of builds) build.abort();
	const running = [...clients.values()];
	clients.clear();
	await Promise.all(running.map(client => client.then(c => c.stop()).catch(() => undefined)));
}

// Starts the server of a .siml document as it opens, and says in the WoWSims SIML output whether it serves the file:
// every .siml file in a repository gets the checks, completion, hover, highlighting and formatting, while the APL
// editor panel opens on .apl.siml rotations only.
function served(document: vscode.TextDocument): void {
	if (document.languageId !== 'siml' || (document.uri.scheme !== 'file' && document.uri.scheme !== 'untitled')) return;
	const file = document.uri.scheme === 'file' ? document.uri.fsPath : document.uri.toString();
	const starting = serverFor(document.uri);
	starting.then(
		() =>
			output.info(
				`${file}: served by the SIML server${/\.apl\.siml$/.test(document.uri.path) ? '' : '; the APL editor panel opens on .apl.siml rotations only'}`,
			),
		error => {
			// A start a restart or the extension's stop ended: the next start says whether it serves the file.
			const root = rootOf(document.uri);
			if (root && clients.get(root) !== starting) return;
			output.warn(`${file}: not served: ${(error as Error).message}`);
		},
	);
}

// Writes a line to the WoWSims SIML output.
export function log(text: string): void {
	output.info(text);
}

// Writes a failed build's output to the WoWSims SIML output and shows it.
export function logFailedBuild(text: string): void {
	output.error(text);
	output.show(true);
}

export function registerServer(extensionContext: vscode.ExtensionContext): void {
	context = extensionContext;
	output = vscode.window.createOutputChannel('WoWSims SIML', { log: true });
	context.subscriptions.push(
		output,
		started,
		vscode.workspace.onDidOpenTextDocument(served),
		vscode.commands.registerCommand('wowsims-apl.restartServer', async () => {
			const running = [...clients.keys()];
			output.info('Restarting the SIML server: it stops, builds again when its code changed, and starts');
			await stopServers();
			roots.clear();
			for (const document of vscode.workspace.textDocuments) served(document);
			await Promise.all(running.map(root => serverFor(vscode.Uri.file(join(root, 'x'))).catch(() => undefined)));
		}),
	);
	vscode.workspace.textDocuments.forEach(served);
}
