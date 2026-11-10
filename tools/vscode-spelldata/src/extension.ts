import { existsSync } from 'node:fs';
import { join } from 'node:path';

import * as vscode from 'vscode';
import { ExecuteCommandRequest, LanguageClient, LanguageClientOptions, ServerOptions, TransportKind } from 'vscode-languageclient/node';

import { findGo } from '../../vscode-common/go';

// The languages package.json activates the extension on.
const LANGUAGES = ['go', 'json', 'typescript', 'typescriptreact'];

let client: LanguageClient | undefined;

export async function activate(context: vscode.ExtensionContext): Promise<void> {
	const changeRank = vscode.commands.registerCommand('wowsims-spelldata.changeRank', async (args?: { parameter: string; function: string; rank: number }) => {
		// From the command palette there is no rank to pick; only a hover's link carries one.
		if (client === undefined || args === undefined) {
			return;
		}

		const editor = vscode.window.activeTextEditor;
		if (editor === undefined) {
			return;
		}

		await client.sendRequest(ExecuteCommandRequest.type, {
			command: 'wowsims-spelldata.setRank',
			arguments: [{ uri: editor.document.uri.toString(), parameter: args.parameter, function: args.function, rank: args.rank }],
		});

		await vscode.commands.executeCommand('editor.action.hideHover');

		setTimeout(() => {
			void vscode.commands.executeCommand('editor.action.showHover');
		}, 50);
	});

	context.subscriptions.push(changeRank);

	const root = vscode.workspace.workspaceFolders?.map(folder => folder.uri.fsPath).find(path => existsSync(join(path, 'go.mod')));
	if (root === undefined) {
		return;
	}

	const settings = vscode.workspace.getConfiguration('wowsims-spelldata');
	const go = findGo(settings.get<string>('goBinary', 'go'));
	if (go === undefined) {
		void vscode.window.showErrorMessage('WoWSims Spelldata: no Go binary found. Set "wowsims-spelldata.goBinary" to its full path.');
		return;
	}
	const serverOptions: ServerOptions = {
		command: go,
		args: ['run', './tools/spelldata/launch', '-lsp'],
		options: { cwd: root },
		transport: TransportKind.stdio,
	};
	const outputChannel = vscode.window.createOutputChannel('WoWSims Spelldata', { log: true });
	const clientOptions: LanguageClientOptions = {
		documentSelector: LANGUAGES.map(language => ({ scheme: 'file', language })),
		// The server writes the hover's rank selector as links to wowsims-spelldata.changeRank only for an
		// editor that registers it.
		initializationOptions: { trace: settings.get<string>('trace', 'on'), rankLinks: true },
		outputChannel,
		middleware: {
			// VS Code runs a command link in a hover only where the markdown trusts that command.
			provideHover: async (document, position, token, next) => {
				const hover = await next(document, position, token);
				for (const content of hover?.contents ?? []) {
					if (content instanceof vscode.MarkdownString) {
						content.isTrusted = {
							enabledCommands: ['wowsims-spelldata.changeRank'],
						};
					}
				}
				return hover;
			},
		},
	};

	client = new LanguageClient('wowsims-spelldata', 'WoWSims Spelldata', serverOptions, clientOptions);
	context.subscriptions.push(outputChannel);
	await client.start();
}

export function deactivate(): Promise<void> | undefined {
	return client?.needsStop() ? client.stop() : undefined;
}
