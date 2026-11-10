import { resolve } from 'node:path';

import * as vscode from 'vscode';

import type { JsonPath, Range, RotationResult } from './protocol';
import { onServerStarted, request } from './server';

export const isAplSiml = (path: string) => path.endsWith('.apl.siml');

// Whether two paths name the same file, without case on Windows.
export function samePath(a: string, b: string): boolean {
	const [x, y] = [resolve(a), resolve(b)];
	return process.platform === 'win32' ? x.toLowerCase() === y.toLowerCase() : x === y;
}

// The open document of a file, looked up when it is needed: VS Code closes a document that no editor shows, and a
// later open gives a new one.
export async function documentFor(uri: vscode.Uri): Promise<vscode.TextDocument> {
	const key = uri.toString();
	return vscode.workspace.textDocuments.find(d => d.uri.toString() === key) ?? (await vscode.workspace.openTextDocument(uri));
}

// The last rotation the server read for each document, whose spans the cursor, Reveal and the sim's counts look up
// while the document is at its version. The server's own reads are never kept here: siml/rotation reads them afresh.
const reads = new Map<string, RotationResult>();

// The rotation a document's text reads as, from the server.
export async function readRotation(document: vscode.TextDocument): Promise<RotationResult> {
	const read = await request<RotationResult>(document.uri, 'siml/rotation', { uri: document.uri.toString() });
	reads.set(document.uri.toString(), read);
	return read;
}

// The text of each JSON path of the document at its version.
export async function spansOf(document: vscode.TextDocument): Promise<RotationResult['spans']> {
	const read = reads.get(document.uri.toString());
	return read?.version === document.version ? read.spans : (await readRotation(document)).spans;
}

// The JSON path of the rotation at a position, from the server.
export const pathAt = (uri: vscode.Uri, { line, character }: vscode.Position) =>
	request<JsonPath>(uri, 'siml/pathAt', { uri: uri.toString(), position: { line, character } });

// The text a JSON path is written at.
export function spanFor(spans: RotationResult['spans'], path: JsonPath): Range | undefined {
	const key = path.join('/');
	return spans.find(span => span.path.join('/') === key)?.range;
}

export function forgetDocument(document: vscode.TextDocument): void {
	reads.delete(document.uri.toString());
}

// A protocol range, as VS Code's.
export function vscodeRange(range: Range): vscode.Range {
	return new vscode.Range(range.start.line, range.start.character, range.end.line, range.end.character);
}

// Applies the server's text edits, made against the document at `version`, as one change; the version it leaves, or
// undefined when the document has moved past `version` or VS Code refused it.
export async function applyEdits(document: vscode.TextDocument, version: number, edits: { range: Range; newText: string }[]): Promise<number | undefined> {
	if (document.isClosed || document.version !== version) return undefined;
	const edit = new vscode.WorkspaceEdit();
	for (const e of edits) edit.replace(document.uri, vscodeRange(e.range), e.newText);
	return (await vscode.workspace.applyEdit(edit)) ? document.version : undefined;
}

export function registerHost(context: vscode.ExtensionContext): void {
	context.subscriptions.push(
		vscode.workspace.onDidCloseTextDocument(forgetDocument),
		onServerStarted(() => reads.clear()),
	);
}
