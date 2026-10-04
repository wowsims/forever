export interface ThemeVar {
	name: string; // --color-primary
	value: string; // raw source text, display only
	inline: boolean; // declared in a `@theme static inline` block
	prefix: string; // color | spacing | text | shadow | …
}

const BLOCK_RE = /@theme\s+static(\s+inline)?\s*\{([\s\S]*?)\n\}/g;
export const VAR_RE = /^\s*(--[a-z0-9-]+):\s*([\s\S]+?);\s*$/gm;

/** Resolves a token's raw value to its rendered px length at the given root font size: a `--spacing(n)`
 * step is `n` times a quarter of the root (Tailwind's `--spacing` is `0.25rem`), an `Nrem` length is `N`
 * times the root, and an `Npx` length is its own number; anything else has no resolved length. */
export const resolvePxValue = (value: string, rootFontSizePx: number): number | null => {
	const spacing = value.match(/^--spacing\(([\d.-]+)\)$/);
	if (spacing) return Math.round(Number(spacing[1]) * rootFontSizePx * 0.25 * 1000) / 1000;
	const rem = value.match(/^([\d.-]+)rem$/);
	if (rem) return Math.round(Number(rem[1]) * rootFontSizePx * 1000) / 1000;
	const px = value.match(/^([\d.-]+)px$/);
	if (px) return Number(px[1]);
	return null;
};

export const parseThemeVars = (css: string): ThemeVar[] => {
	const out: ThemeVar[] = [];
	for (const block of css.matchAll(BLOCK_RE)) {
		const inline = !!block[1];
		for (const m of block[2].matchAll(VAR_RE)) {
			if (m[1].endsWith('-*')) continue;
			out.push({ name: m[1], value: m[2].replace(/\s+/g, ' '), inline, prefix: m[1].slice(2).split('-')[0] });
		}
	}
	return out;
};
