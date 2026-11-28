import { getLang } from '@i18n/locale_service';

// Wowhead serves each expansion under its own url segment and its own `dataEnv`
// id. Porting this file to a sibling sim repo means changing the one line below:
// the domain, and every url built from it, follows. The literal-union key means
// an unmapped id is a compile error rather than a `.../undefined/...` url.
const WOWHEAD_EXPANSIONS = {
	5: 'tbc',
	15: 'mop-classic',
	16: 'forever',
} as const;

export const WOWHEAD_EXPANSION_ENV: keyof typeof WOWHEAD_EXPANSIONS = 16;
export const WOWHEAD_DOMAIN = WOWHEAD_EXPANSIONS[WOWHEAD_EXPANSION_ENV];

export function getWowheadLanguagePrefix(): string {
	const lang = getLang();
	return lang === 'en' ? '' : `${lang}/`;
}

// Every wowhead link this app builds hangs off one of these. The entity links
// keep the bare host they have always used; the gear planner page is the `www.`
// form we show to users verbatim in the importer.
export const WOWHEAD_BASE_URL = `https://wowhead.com/${WOWHEAD_DOMAIN}`;
export const WOWHEAD_GEAR_PLANNER_URL = `https://www.wowhead.com/${WOWHEAD_DOMAIN}/gear-planner`;
export const WOWHEAD_ICON_BASE_URL = 'https://wow.zamimg.com/images/wow/icons';

type WowheadEntity = 'item' | 'spell' | 'quest' | 'npc' | 'zone' | 'faction';

// `https://wowhead.com/<domain>/<lang>/<entity>=<id>` — the language segment
// is empty for English.
export function wowheadEntityUrl(entity: WowheadEntity, id: number, rank = 0, definitionId = 0): string {
	const url = `${WOWHEAD_BASE_URL}/${getWowheadLanguagePrefix()}${entity}=${id}`;
	const params = new URLSearchParams();
	if (definitionId > 0) params.set('def', String(definitionId));
	if (rank > 0) params.set('rank', String(rank));
	const query = params.toString();
	return query ? `${url}?${query}` : url;
}

// The icons deploy decodes from the game files (tools/icons), 56 px like zamimg's large ones, served beside the
// site. A name the manifest does not list comes from zamimg, and so does every name while the manifest loads,
// when it fails to load, and on a build that has no icons folder.
const LOCAL_ICON_BASE_URL = '/forever/assets/icons';
let localIcons: ReadonlySet<string> = new Set();

// A missing manifest fails to parse on Cloudflare Pages, which answers a missing file with the site's HTML.
export async function loadLocalIcons(): Promise<void> {
	try {
		const response = await fetch(`${LOCAL_ICON_BASE_URL}/manifest.json`);
		if (!response.ok) throw new Error(`HTTP ${response.status}`);
		localIcons = new Set(await response.json());
	} catch (error) {
		console.warn(`No local icons, so every icon comes from zamimg: ${error}`);
	}
}

export function wowheadIconUrl(iconLabel: string, size: 'large' | 'medium' | 'small' = 'large'): string {
	if (localIcons.has(iconLabel)) return `${LOCAL_ICON_BASE_URL}/${iconLabel}.jpg`;
	return `${WOWHEAD_ICON_BASE_URL}/${size}/${iconLabel}.jpg`;
}
