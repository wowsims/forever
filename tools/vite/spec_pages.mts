// The 18 spec pages are one page served at 18 URLs.
//
// `ui/index_template.html` carries nothing per-spec. Its title, description and link tags hold the
// Forever-wide text, and every asset reference is root-absolute (`/forever/...` once vite rewrites
// it), so the build processes it ONCE as an ordinary html input and this plugin copies the processed
// result to `<class>/<spec>/index.html` for every spec, with that spec's title, description and URL
// written into the head. Link previews read only the html, so these cannot wait for the page's
// script. Nothing per-spec is written into the source tree; in dev the same file is served through
// `transformIndexHtml` for the same URLs, with the same rewrite.
import fs from 'node:fs/promises';
import { readFileSync } from 'node:fs';
import path from 'node:path';

import { glob } from 'glob';
import type { Plugin } from 'vite';

/** The one spec page, relative to vite's root (`ui/`). Also its build output path. */
export const SPEC_PAGE_TEMPLATE = 'index_template.html';

/** The production origin; canonical and `og:url` links must be absolute. */
const SITE_ORIGIN = 'https://www.wowsims.com';

interface SpecPageMetadata {
	title: string;
	description: string;
	url: string;
}

/**
 * A spec page's title, description and URL, from the same English locale strings the page sets at
 * runtime (`updateSimPageMetadata`). A class with one spec has no spec name and uses the
 * `_class` strings.
 */
function specPageMetadata(localesRoot: string, base: string, spec: string): SpecPageMetadata {
	const read = (namespace: string) => JSON.parse(readFileSync(path.join(localesRoot, 'en', `${namespace}.json`), 'utf-8'));
	const character = read('character');
	const sim = read('translation').sim;
	const [className, specName] = spec.split('/');
	const values: Record<string, string> = { class: character.classes[className], spec: character.specs[className]?.[specName] ?? '' };
	if (!values.class) throw new Error(`spec-pages: no class name for ${spec} in character.json.`);
	const suffix = values.spec ? '' : '_class';
	const interpolate = (text: string) => text.replace(/\{\{(\w+)\}\}/g, (_, key) => values[key]);
	return {
		title: interpolate(sim[`title${suffix}`]),
		description: interpolate(sim[`description${suffix}`]),
		url: `${SITE_ORIGIN}${base}${spec}/`,
	};
}

const escapeHtml = (text: string) => text.replace(/&/g, '&amp;').replace(/"/g, '&quot;').replace(/</g, '&lt;').replace(/>/g, '&gt;');

/** Writes `meta` into the template's head; a tag the template lacks fails the build. */
function withSpecMetadata(html: string, meta: SpecPageMetadata): string {
	const title = escapeHtml(meta.title);
	const description = escapeHtml(meta.description);
	// The formatter wraps a long tag one attribute per line, so attributes are split by any whitespace.
	const tag = (name: string, key: string, valueAttr: string) => new RegExp(`(<${name}\\s+${key}\\s+${valueAttr}=")[^"]*(")`);
	const replacements: [RegExp, string][] = [
		[/(<title>)[^<]*(<\/title>)/, title],
		[tag('meta', 'name="description"', 'content'), description],
		[tag('link', 'rel="canonical"', 'href'), meta.url],
		[tag('meta', 'property="og:url"', 'content'), meta.url],
		[tag('meta', 'property="og:title"', 'content'), title],
		[tag('meta', 'property="og:description"', 'content'), description],
		[tag('meta', 'name="twitter:title"', 'content'), title],
		[tag('meta', 'name="twitter:description"', 'content'), description],
	];
	return replacements.reduce((page, [pattern, value]) => {
		if (!pattern.test(page)) throw new Error(`spec-pages: ${SPEC_PAGE_TEMPLATE} has no tag matching ${pattern}.`);
		return page.replace(pattern, (_, open, close) => `${open}${value}${close}`);
	}, html);
}

/**
 * Every spec that owns a page, as `<class>/<spec>`, from the same source of truth the
 * makefile's `PAGE_INDECES` used: `ui/specs/<class>/<spec>/spec.ts(x)`.
 */
export function discoverSpecPages(uiRoot: string): string[] {
	const specsRoot = path.join(uiRoot, 'specs');
	return glob
		.sync(path.join(specsRoot, '*/*/spec.{ts,tsx}').replace(/\\/g, '/'))
		.map(specFile => path.relative(specsRoot, path.dirname(specFile)).split(path.sep).join('/'))
		.sort();
}

export function specPages(uiRoot: string): Plugin {
	const specs = discoverSpecPages(uiRoot);
	if (!specs.length) throw new Error(`spec-pages: no spec pages found under ${path.join(uiRoot, 'specs')}.`);
	const templatePath = path.resolve(uiRoot, SPEC_PAGE_TEMPLATE);
	const localesRoot = path.resolve(uiRoot, '../assets/locales');
	let base = '/';

	return {
		name: 'spec-pages',
		// `vite:build-html` emits the processed page from its own `generateBundle`, which runs
		// after unenforced plugins' — only a `post` plugin sees it in the bundle.
		enforce: 'post',

		configResolved(config) {
			base = config.base;
		},

		generateBundle(_options, bundle) {
			const processed = bundle[SPEC_PAGE_TEMPLATE];
			if (!processed || processed.type !== 'asset') {
				throw new Error(`spec-pages: ${SPEC_PAGE_TEMPLATE} is not a processed html input of this build.`);
			}
			// The template's own path is not a page; only its copies are.
			delete bundle[SPEC_PAGE_TEMPLATE];
			for (const spec of specs) {
				this.emitFile({
					type: 'asset',
					fileName: `${spec}/index.html`,
					source: withSpecMetadata(String(processed.source), specPageMetadata(localesRoot, base, spec)),
				});
			}
		},

		configureServer(server) {
			const pages = new Map(specs.flatMap(spec => [`${base}${spec}/`, `${base}${spec}/index.html`].map(url => [url, spec])));
			// The production static host answers the bare `${base}<class>/<spec>` with a 301 to the
			// trailing-slash form. Dev has to do the same: an unredirected bare url misses `pages`
			// and falls through to vite's SPA fallback, which serves the *landing* page.
			const bare = new Set(specs.map(spec => `${base}${spec}`));

			server.middlewares.use((req, res, next) => {
				const { pathname, search } = new URL(req.url!, 'http://localhost');
				if (bare.has(pathname)) {
					res.writeHead(301, { Location: `${pathname}/${search}` });
					res.end();
					return;
				}
				const spec = pages.get(pathname);
				if (!spec) {
					next();
					return;
				}
				// This middleware runs before vite's `base` middleware, so the path still carries
				// `/forever/`; `transformIndexHtml` wants it stripped (that is what vite's own
				// `indexHtmlMiddleware` passes, and the html-proxy ids for the page's inline
				// module scripts are keyed on it).
				const url = pathname.slice(base.length - 1).replace(/\/$/, '/index.html');
				fs.readFile(templatePath, 'utf-8')
					.then(html => server.transformIndexHtml(url, html, req.originalUrl))
					.then(html => withSpecMetadata(html, specPageMetadata(localesRoot, base, spec)))
					.then(html => {
						res.writeHead(200, { 'Content-Type': 'text/html' });
						res.end(html);
					})
					.catch(next);
			});
		},
	};
}
