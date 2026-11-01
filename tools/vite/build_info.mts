// The release a build is: the commit it is built from, the UTC date it is built on, and for each spec the
// game build of the data its sim reads and the last commits that changed its presets and its engine.
// vite.config.mts defines it into the page, so every build, the deploy's included, states its own release
// with nothing written into the source tree. ui/sim/constants/build_info.ts reads it with these types.
import { execFileSync } from 'node:child_process';
import { existsSync, readFileSync } from 'node:fs';
import path from 'node:path';
import { gunzipSync } from 'node:zlib';

import { discoverSpecPages } from './spec_pages.mts';

export interface Change {
	/** The full hash of the commit. */
	commit: string;
	/** The commit's UTC date, as YYYY-MM-DD. */
	date: string;
}

/** One spec's release, each part left out when the build cannot read it. */
export interface SpecRelease {
	/** The newest game build among the spell data parts its sim reads: its class, racials, items and shared. */
	gameVersion?: string;
	/** The last change to its presets: spec.ts, presets.ts, the apls, gear_sets and builds folders, and the class's shared presets. */
	presets?: Change;
	/** The last change to its engine: its class's sim code and spell data, its own sim folder and expected results, and sim/core and sim/common. */
	engine?: Change;
}

export interface BuildInfo {
	/** The UTC date of the build, as YYYY-MM-DD. */
	date: string;
	/** The full hash of the commit built, or empty when nothing names it. */
	commit: string;
	/** The newest game build the spell store was generated from, such as 1.60.1.70450, or empty. */
	gameVersion: string;
	/** Each spec's release, keyed `<class>/<spec>` as its page path is. */
	specs: Record<string, SpecRelease>;
}

export const readBuildInfo = (root: string): BuildInfo => {
	const builds = spellDataBuilds(root);
	return {
		date: new Date().toISOString().slice(0, 10),
		commit: buildCommit(root),
		gameVersion: newest(Object.values(builds)),
		specs: specReleases(root, builds),
	};
};

// Dates are UTC, as the build date is, whatever the committer's or the builder's time zone.
const git = (root: string, args: string[]): string =>
	execFileSync('git', args, { cwd: root, stdio: ['ignore', 'pipe', 'ignore'], env: { ...process.env, TZ: 'UTC' } })
		.toString()
		.trim();

// BUILD_COMMIT names the commit when the checkout is not it: a pull request's checkout is the merge
// commit, and its preview names the head. Otherwise git names the checkout; GITHUB_SHA covers a tree
// without .git.
const buildCommit = (root: string): string => {
	if (process.env.BUILD_COMMIT) {
		return process.env.BUILD_COMMIT;
	}
	try {
		return git(root, ['rev-parse', 'HEAD']);
	} catch {
		return process.env.GITHUB_SHA ?? '';
	}
};

// assets/db_inputs/spell_store_inputs.json is gzipped despite its name. Its Builds map stamps each part
// of the store, a class, items, racials or shared, with the game build it was generated from. The parts
// are regenerated one at a time, so they need not agree.
const spellDataBuilds = (root: string): Record<string, string> => {
	try {
		const inputs = JSON.parse(gunzipSync(readFileSync(path.join(root, 'assets/db_inputs/spell_store_inputs.json'))).toString());
		return inputs.Builds ?? {};
	} catch {
		return {};
	}
};

// Build numbers compare part by part as numbers, so 1.60.1.9999 is older than 1.60.1.70450.
const newest = (builds: string[]): string => [...builds].sort((a, b) => a.localeCompare(b, 'en', { numeric: true })).at(-1) ?? '';

// A spec's presets and engine are each the last commit over a set of paths:
//   - presets: spec.ts (the defaults and the preset lists the pickers offer), presets.ts and the apls,
//     gear_sets and builds folders in ui/specs/<class>/<spec>, and ui/specs/<class>/shared/presets.ts,
//     which the warrior specs take their default buffs from;
//   - engine: the files directly in sim/<class>, its generated spell data among them and the expected
//     results of a class with no spec folders, sim/<class>/<spec> with its expected results when the
//     spec has a sim folder, and sim/core and sim/common, which every spec's sim runs on: the outcome
//     rolls, the APL, the racials, the item effects, enchants and buffs, and the shared spell data.
// The log starts at the checkout, which is what the bundle is built from: a pull request's preview is
// built from its merge commit, so a master change its branch lacks is still dated; where both sides
// touched the paths, the row names that merge commit, which lives only on refs/pull/<n>/merge. A
// shallow clone has no history to read, so it reports no change rather than its one commit for every
// spec.
const specReleases = (root: string, builds: Record<string, string>): Record<string, SpecRelease> => {
	let shallow = true;
	try {
		shallow = git(root, ['rev-parse', '--is-shallow-repository']) === 'true';
	} catch {
		// Not a git tree: no changes to read.
	}
	const lastChange = (paths: string[]): Change | undefined => {
		if (shallow) {
			return undefined;
		}
		try {
			const [commit, date] = git(root, ['log', '-1', '--format=%H %cd', '--date=format-local:%Y-%m-%d', 'HEAD', '--', ...paths]).split(' ');
			return commit ? { commit, date } : undefined;
		} catch {
			return undefined;
		}
	};
	const specs: Record<string, SpecRelease> = {};
	for (const key of discoverSpecPages(path.join(root, 'ui'))) {
		const [className, specName] = key.split('/');
		const specSim = `sim/${className}/${specName}`;
		const release: SpecRelease = {
			gameVersion: newest([className, 'racials', 'items', 'shared'].flatMap(part => (builds[part] ? [builds[part]] : []))) || undefined,
			presets: lastChange([
				...['spec.ts', 'spec.tsx', 'presets.ts', 'apls', 'gear_sets', 'builds'].map(name => `ui/specs/${key}/${name}`),
				`ui/specs/${className}/shared/presets.ts`,
			]),
			engine: lastChange([`:(glob)sim/${className}/*`, ...(existsSync(path.join(root, specSim)) ? [specSim] : []), 'sim/core', 'sim/common']),
		};
		specs[key] = release;
	}
	return specs;
};
