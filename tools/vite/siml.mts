// SIML files as modules (sim/siml/SPEC.md). Importing `x.apl.siml` gives the APLRotation JSON, `x.gear.siml` the
// EquipmentSpec JSON and any other `.siml` the IndividualSimSettings JSON: the JSON of the message each file kind
// loads into (SPEC.md §1), which reads as the same message as a `.json` file of it.
//
// The Go reader reads each file: one `wowsimcli siml serve` per process, the binary $WOWSIMCLI names, else
// bin/wowsimcli, which `make` builds and keeps fresh, else `go run ./cmd/wowsimcli`, so the build needs Go.
import { type ChildProcess, spawn } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';
import readline from 'node:readline';
import { isDeepStrictEqual } from 'node:util';

import { normalizePath, type Plugin } from 'vite';

const ROOT = path.resolve(import.meta.dirname, '../..');
const NAMES = path.join(ROOT, 'sim/siml/names.json');
const BUILT = path.join(ROOT, 'bin', process.platform === 'win32' ? 'wowsimcli.exe' : 'wowsimcli');

interface Diagnostic {
	line: number;
	col: number;
	severity: string;
	code: string;
	message: string;
}

// What `wowsimcli siml serve` answers for a file: `errors` fail the import, which under ui/specs holds its warnings
// and raw refs too (SPEC.md §7.3, §13), and `uses` are the files its use lines read.
interface Response {
	id: number;
	json?: unknown;
	errors: Diagnostic[];
	warnings: Diagnostic[];
	uses: string[];
	error?: string;
}

const relative = (file: string) => normalizePath(path.relative(ROOT, file));

type Pending = Map<number, { resolve: (r: Response) => void; reject: (e: Error) => void }>;

let server: { child: ChildProcess; pending: Pending } | undefined;
let nextId = 0;

const stop = () => {
	server?.child.kill();
	server = undefined;
};
process.once('exit', stop);

const read = (file: string, text: string): Promise<Response> => {
	if (!server) {
		const args = ['siml', 'serve', '-names', NAMES];
		const binary = process.env.WOWSIMCLI || (fs.existsSync(BUILT) ? BUILT : undefined);
		const [command, argv] = binary ? [binary, args] : ['go', ['run', './cmd/wowsimcli', ...args]];
		const child = spawn(command, argv, { cwd: ROOT, stdio: ['pipe', 'pipe', 'inherit'] });
		const pending: Pending = new Map();
		const fail = (error: Error) => {
			if (server?.child === child) server = undefined;
			for (const { reject } of pending.values()) reject(error);
			pending.clear();
		};
		child.on('error', error => fail(new Error(`${command} does not start, and SIML files need Go or $WOWSIMCLI: ${error.message}`)));
		child.on('exit', code => fail(new Error(`wowsimcli siml serve exited with code ${code}`)));
		// A write to a child that has exited fails with EPIPE: the next read starts another, and the exit fails the
		// requests this one owes.
		child.stdin!.on('error', () => {
			if (server?.child === child) server = undefined;
			child.kill();
		});
		readline.createInterface({ input: child.stdout! }).on('line', line => {
			let response: Response;
			try {
				response = JSON.parse(line) as Response;
			} catch {
				// A line that is not an answer leaves every request unmatched: fail them all and start another server.
				fail(new Error(`wowsimcli siml serve wrote a line that is not JSON: ${line.slice(0, 200)}`));
				child.kill();
				return;
			}
			pending.get(response.id)?.resolve(response);
			pending.delete(response.id);
			// The child keeps the process alive only while it owes an answer.
			if (!pending.size) for (const handle of [child, child.stdin, child.stdout] as { unref(): void }[]) handle.unref();
		});
		server = { child, pending };
	}
	const { child, pending } = server;
	for (const handle of [child, child.stdin, child.stdout] as { ref(): void }[]) handle.ref();
	const id = ++nextId;
	return new Promise((resolve, reject) => {
		pending.set(id, { resolve, reject });
		child.stdin!.write(`${JSON.stringify({ id, path: file, text })}\n`);
	});
};

// The parts of a full file that a file it uses may give whole: the field of its player, and the kind of file.
const SHARED_PARTS = [
	['rotation', '.apl.siml'],
	['equipment', '.gear.siml'],
] as const;

/**
 * The module of `json`, read from `file`. A part of its player that one of the `uses` gives as it stands is that
 * file's module, so a rotation or gear set several builds use ships once.
 */
const moduleCode = async (file: string, json: unknown, uses: string[]): Promise<string> => {
	const player = (json as { player?: Record<string, unknown> } | undefined)?.player;
	const shared: Array<[field: string, specifier: string]> = [];
	for (const [field, kind] of SHARED_PARTS) {
		if (player?.[field] === undefined) continue;
		for (const used of uses.filter(used => used.endsWith(kind))) {
			if (!isDeepStrictEqual((await read(used, fs.readFileSync(used, 'utf8'))).json, player[field])) continue;
			const specifier = normalizePath(path.relative(path.dirname(file), used));
			shared.push([field, specifier.startsWith('.') ? specifier : `./${specifier}`]);
			break;
		}
	}
	if (!shared.length) return `export default ${JSON.stringify(json)};\n`;

	const own = { ...(json as object), player: Object.fromEntries(Object.entries(player!).filter(([key]) => !shared.some(([field]) => field === key))) };
	return [
		...shared.map(([, specifier], i) => `import shared${i} from ${JSON.stringify(specifier)};`),
		`const settings = ${JSON.stringify(own)};`,
		...shared.map(([field], i) => `settings.player.${field} = shared${i};`),
		'export default settings;',
		'',
	].join('\n');
};

export function simlPlugin(): Plugin {
	return {
		name: 'siml',

		// Only a bare `.siml` id: one with a query, such as `?raw`, is another plugin's.
		async transform(text, id) {
			if (!id.endsWith('.siml')) return null;
			this.addWatchFile(NAMES);
			const response = await read(id, text);
			for (const used of response.uses) this.addWatchFile(used);
			if (response.error) this.error(`${relative(id)}: ${response.error}`);
			const where = (d: Diagnostic) => `${relative(id)}:${d.line}:${d.col}`;
			if (response.errors.length) {
				const [first] = response.errors;
				this.error({
					message: response.errors.map(d => `${where(d)}: ${d.severity} ${d.code}: ${d.message}`).join('\n'),
					loc: { file: id, line: first.line, column: first.col },
				});
			}
			for (const d of response.warnings) this.warn(`${where(d)}: ${d.code}: ${d.message}`);
			return { code: await moduleCode(id, response.json, response.uses), map: { mappings: '' }, moduleType: 'js' };
		},

		closeBundle: stop,
	};
}
