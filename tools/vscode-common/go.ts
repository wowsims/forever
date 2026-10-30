import { statSync } from 'node:fs';
import { homedir } from 'node:os';
import { delimiter, join } from 'node:path';

// The go binary an extension's goBinary setting names: `go` searches for an install, anything else is the binary.
// VS Code started from a desktop launcher or a WSL/remote server does not read the shell profile, so a Go install the
// terminal finds is often missing from its PATH. A configured path that is not on this machine, as one set in the shared
// workspace settings for another machine can be, falls back to the same search.
export function findGo(configured: string): string | undefined {
	if (configured !== 'go' && isFile(configured)) {
		return configured;
	}
	const exe = process.platform === 'win32' ? 'go.exe' : 'go';
	const candidates = [
		...(process.env.PATH ?? '').split(delimiter).map(dir => join(dir, exe)),
		...(process.env.GOROOT ? [join(process.env.GOROOT, 'bin', exe)] : []),
		'/usr/local/go/bin/go',
		'/usr/lib/go/bin/go',
		'/opt/homebrew/bin/go',
		join(homedir(), 'go', 'bin', exe),
		'C:\\Program Files\\Go\\bin\\go.exe',
	];
	return candidates.find(path => path !== exe && isFile(path));
}

function isFile(path: string): boolean {
	try {
		return statSync(path).isFile();
	} catch {
		return false;
	}
}
