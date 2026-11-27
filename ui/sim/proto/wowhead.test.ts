import { afterEach, describe, expect, it, vi } from 'vitest';

// The local icon set lives in module scope, so every case loads a fresh copy of the module.
const freshWowhead = async () => {
	vi.resetModules();
	return import('./wowhead');
};

const ZAMIMG = 'https://wow.zamimg.com/images/wow/icons';

describe('wowheadIconUrl', () => {
	afterEach(() => {
		vi.unstubAllGlobals();
		vi.restoreAllMocks();
	});

	it('serves an icon the manifest lists from the site and every other icon from zamimg', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn(() => Promise.resolve(Response.json(['inv_sword_39']))),
		);
		const { loadLocalIcons, wowheadIconUrl } = await freshWowhead();
		await loadLocalIcons();

		expect(wowheadIconUrl('inv_sword_39')).toBe('/forever/assets/icons/inv_sword_39.jpg');
		expect(wowheadIconUrl('inv_axe_01')).toBe(`${ZAMIMG}/large/inv_axe_01.jpg`);
	});

	// Cloudflare Pages answers a missing file with the site's HTML and a 200.
	it.each([
		['the site HTML in place of the manifest', () => new Response('<!doctype html>', { headers: { 'content-type': 'text/html' } })],
		['a 404 for the manifest', () => new Response('Not found', { status: 404 })],
	])('keeps every icon on zamimg given %s', async (_, response) => {
		vi.stubGlobal(
			'fetch',
			vi.fn(() => Promise.resolve(response())),
		);
		const warn = vi.spyOn(console, 'warn').mockImplementation(() => {});
		const { loadLocalIcons, wowheadIconUrl } = await freshWowhead();
		await loadLocalIcons();

		expect(wowheadIconUrl('inv_sword_39')).toBe(`${ZAMIMG}/large/inv_sword_39.jpg`);
		expect(warn).toHaveBeenCalledOnce();
	});
});
