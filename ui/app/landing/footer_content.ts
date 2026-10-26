import { VERSIONS_MANIFEST_URL } from '@sim/constants/other';

export type FooterLink = { name: string; url: string };
export type FooterCommunityGroup = { title: string; links: FooterLink[]; columns: 1 | 2 };
export type FooterVersion = { title: string; url: string; status: 'live' | 'wip' };

/** The part of the shared versions manifest (schema version 1) the landing footer reads. */
export type FooterContent = {
	versions: FooterVersion[];
	communities: FooterCommunityGroup[];
	legal: { copyrightStartYear: number; copyrightHolder: string };
};

const isObject = (value: unknown): value is Record<string, unknown> => typeof value === 'object' && value !== null;
const isString = (value: unknown): value is string => typeof value === 'string';
const isArrayOf = <T>(value: unknown, isItem: (item: unknown) => item is T): value is T[] => Array.isArray(value) && value.every(isItem);

const isLink = (value: unknown): value is FooterLink => isObject(value) && isString(value.name) && isString(value.url);
const isGroup = (value: unknown): value is Omit<FooterCommunityGroup, 'columns'> & { columns?: unknown } =>
	isObject(value) && isString(value.title) && isArrayOf(value.links, isLink) && (value.columns === undefined || value.columns === 1 || value.columns === 2);
const isVersion = (value: unknown): value is FooterVersion =>
	isObject(value) && isString(value.title) && isString(value.url) && (value.status === 'live' || value.status === 'wip');

/** The footer content in a manifest, or null when it is not schema version 1 or a field the footer reads is malformed. */
export const parseFooterContent = (manifest: unknown): FooterContent | null => {
	if (!isObject(manifest) || manifest.schemaVersion !== 1) return null;
	const { versions, communities, legal } = manifest;
	if (!isArrayOf(versions, isVersion) || !isArrayOf(communities, isGroup)) return null;
	if (!isObject(legal) || typeof legal.copyrightStartYear !== 'number' || !isString(legal.copyrightHolder)) return null;
	return {
		versions: versions.map(({ title, url, status }) => ({ title, url, status })),
		communities: communities.map(({ title, links, columns }) => ({
			title,
			links: links.map(({ name, url }) => ({ name, url })),
			columns: columns === 2 ? 2 : 1,
		})),
		legal: { copyrightStartYear: legal.copyrightStartYear, copyrightHolder: legal.copyrightHolder },
	};
};

/** Fetches the shared manifest; resolves to null when it cannot be read. */
export const fetchFooterContent = (signal?: AbortSignal): Promise<FooterContent | null> =>
	fetch(VERSIONS_MANIFEST_URL, { signal })
		.then(response => (response.ok ? response.json() : null))
		.then(parseFooterContent)
		.catch(() => null);
