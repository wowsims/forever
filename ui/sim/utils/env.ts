// Which deployment this page is running as.

export type Environments = 'local' | 'external';

// Pure classification of a hostname. The browser probe that reads the actual
// page hostname lives in ui/ui-kit/utils/dom.ts; anything with an `Env` in hand
// (e.g. Sim) passes `env.location.hostname` here instead.
export const environmentOf = (hostname: string): Environments => (hostname.includes('localhost') ? 'local' : 'external');
export const isDevMode = () => {
	return import.meta.env.DEV;
};
// The dev server, or a page served from the project's Cloudflare Pages host, where the
// pull request previews live. Keyed on the hostname so wowsims.com, a page embedding
// it and a copy hosted anywhere else never show what is not ready yet.
export const isPreviewOrDevMode = (hostname: string) =>
	isDevMode() || hostname === 'wowsims-forever.pages.dev' || hostname.endsWith('.wowsims-forever.pages.dev');
