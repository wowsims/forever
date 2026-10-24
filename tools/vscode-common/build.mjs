import { existsSync, mkdtempSync, readFileSync, readdirSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, relative } from 'node:path';

// Builds an extension into its committed folder under .vscode/extensions, or with `--check` builds it into a
// scratch folder and fails when the committed outputs differ from it or the folder holds a file the build does not
// write. `outputs` are the committed outputs, `kept` the other files the folder may hold.
export async function buildExtension({ installed, outputs, kept = [], makeTarget, buildInto }) {
	if (!process.argv.includes('--check')) {
		await buildInto(installed);
		console.log(`built ${relative(process.cwd(), installed)}`);
		return;
	}
	const fresh = mkdtempSync(join(tmpdir(), `wowsims-${makeTarget}-`));
	try {
		await buildInto(fresh);
		const stale = outputs.filter(file => {
			const committed = join(installed, file);
			return !existsSync(committed) || !readFileSync(committed).equals(readFileSync(join(fresh, file)));
		});
		const extra = existsSync(installed) ? filesUnder(installed).filter(file => !outputs.includes(file) && !kept.includes(file)) : [];
		if (stale.length > 0 || extra.length > 0) {
			const where = relative(process.cwd(), installed);
			for (const file of stale) {
				console.error(`${join(where, file)} is not what the build writes`);
			}
			for (const file of extra) {
				console.error(`${join(where, file)} is not a build output`);
			}
			console.error(`run \`make ${makeTarget}\` and commit the result`);
			process.exit(1);
		}
		console.log(`${relative(process.cwd(), installed)} matches a fresh build`);
	} finally {
		rmSync(fresh, { recursive: true, force: true });
	}
}

function filesUnder(dir) {
	return readdirSync(dir, { recursive: true, withFileTypes: true })
		.filter(entry => entry.isFile())
		.map(entry => relative(dir, join(entry.parentPath, entry.name)));
}
