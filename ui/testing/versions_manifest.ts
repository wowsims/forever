import { VERSIONS_MANIFEST_URL } from '@sim/constants/other';

/** A schema version 1 manifest as wowsims.com publishes it, with one 'wip' version, a short community group and a two-column one. */
export const VERSIONS_MANIFEST = {
	schemaVersion: 1,
	homepage: 'https://www.wowsims.com',
	versions: [
		{ slug: 'forever', title: 'Forever', acronym: 'Forever', url: 'https://www.wowsims.com/forever/', themeColor: '#4fc3f7', status: 'live' },
		{ slug: 'mop', title: 'Mists of Pandaria', acronym: 'MoP', url: 'https://www.wowsims.com/mop/', themeColor: '#00FF98', status: 'live' },
		{ slug: 'next', title: 'Next Expansion', acronym: 'Next', url: 'https://www.wowsims.com/next/', themeColor: '#ffffff', status: 'wip' },
	],
	legal: {
		copyrightStartYear: 2020,
		copyrightHolder: 'WoWSims contributors',
		disclaimer: 'A fan-made project, not affiliated with Blizzard Entertainment.',
	},
	communities: [
		{ title: 'Partner Servers', links: [{ name: 'Example Realm', url: 'https://discord.gg/example' }] },
		{
			title: 'Class Discords',
			columns: 2,
			links: ['Death Knight', 'Druid', 'Hunter', 'Mage', 'Monk', 'Paladin', 'Priest'].map(name => ({
				name,
				url: `https://discord.gg/${name.toLowerCase().replace(' ', '')}`,
			})),
		},
	],
};

/**
 * Answers fetches of the versions manifest with `body`, or fails them when `body` is an Error, for a story's
 * `beforeEach` or a test; other URLs go to the real fetch. Returns the restore.
 */
export const withVersionsManifest = (body: unknown): (() => void) => {
	const realFetch = globalThis.fetch;
	globalThis.fetch = (input, init) => {
		const url = input instanceof Request ? input.url : String(input);
		if (url !== VERSIONS_MANIFEST_URL) return realFetch(input, init);
		if (body instanceof Error) return Promise.reject(body);
		return Promise.resolve(new Response(JSON.stringify(body), { headers: { 'Content-Type': 'application/json' } }));
	};
	return () => {
		globalThis.fetch = realFetch;
	};
};
