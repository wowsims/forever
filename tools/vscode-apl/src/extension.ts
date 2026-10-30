import type * as vscode from 'vscode';

import { registerHost } from './host';
import { registerInsertPattern } from './insert_pattern';
import { registerPanel } from './panel';
import { registerServer, stopServers } from './server';
import { registerSim } from './simulate';

export function activate(context: vscode.ExtensionContext): void {
	registerHost(context);
	registerSim(context);
	registerInsertPattern(context);
	registerPanel(context);
	registerServer(context);
}

export function deactivate(): Promise<void> {
	return stopServers(true);
}
