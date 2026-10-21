import { ProtoVersion } from '@generated/proto/common';
import { readMessageOption } from '@protobuf-ts/runtime';

export enum Phase {
	Phase1 = 1,
	Phase2,
	Phase3,
	Phase4,
	Phase5,
}

export const CURRENT_PHASE = Phase.Phase1;

export enum LaunchStatus {
	Unlaunched,
	// Gear, enchants, talents and the gear optimizer work. No simulation, and none planned.
	GearPlanner,
	Alpha,
	Beta,
	Launched,
}

export const CURRENT_API_VERSION: number = readMessageOption(ProtoVersion, 'proto.current_version_number')! as number;

// Github pages serves our site under the /forever directory
export const REPO_NAME = 'forever';
export const REPO_URL = `https://github.com/wowsims/${REPO_NAME}`;
export const REPO_RELEASES_URL = `${REPO_URL}/releases`;
// The repository whose CI built this page (see vite.config.mts). Empty for a build made outside
// GitHub Actions, which then offers no issue links at all.
export const ISSUES_REPO: string = import.meta.env.VITE_GITHUB_REPOSITORY;
export const REPO_NEW_ISSUE_URL = `https://github.com/${ISSUES_REPO}/issues/new`;
export const REPO_CHOOSE_NEW_ISSUE_URL = `${REPO_NEW_ISSUE_URL}/choose`;

export const WOWSIMS_URL = 'https://www.wowsims.com/';
export const DISCORD_URL = 'https://discord.gg/p3DgvmnDCS';
export const GITHUB_ORG_URL = 'https://github.com/wowsims/';
export const ADDON_URL = 'https://www.curseforge.com/wow/addons/wowsimsexporter';
export const PATREON_URL = 'https://www.patreon.com/wowsims';

export const SOCIALS = [
	{ key: 'discord', href: DISCORD_URL, className: 'ui-social-link', icon: 'discord', tooltip: 'info.discord' },
	{ key: 'github', href: REPO_URL, className: 'ui-social-link', icon: 'github', tooltip: 'info.github' },
	{
		key: 'patreon',
		href: PATREON_URL,
		className: 'ui-social-link',
		icon: 'patreon',
		tooltip: 'info.patreon',
		label: ' Patreon',
	},
] as const;

export type Social = (typeof SOCIALS)[number];

// Root-relative path of the individual sim page for the given spec. Resolve it
// against the page origin at the point of use (see SimTitleDropdown) — this
// layer has no `window`. Lives here rather than in proto/utils so that
// player/specs/<class>.ts (which calls it at module scope) does not import the
// proto <-> player/specs cycle; the evaluation order of that cycle decides
// whether PlayerSpecs' lookup table is populated.
export function getSpecSitePath(classString: string, specString: string): string {
	return `/${REPO_NAME}/${classString}/${specString}/`;
}

export const LOCAL_STORAGE_PREFIX = '__forever';

export enum SortDirection {
	ASC,
	DESC,
}
