import { parseThemeVars, type ThemeVar, VAR_RE } from './parse_theme';

export interface ThemeToken extends ThemeVar {
	file: string;
	/** A plain `:root` variable in `vars.css`, not a `@theme` token: no utility is generated for it. */
	runtime: boolean;
}

const sources = import.meta.glob<string>('../../styles/theme/*.css', { query: '?raw', import: 'default', eager: true });

export const parseRootVars = (css: string): ThemeVar[] =>
	[...css.matchAll(VAR_RE)].map(m => ({ name: m[1], value: m[2].replace(/\s+/g, ' '), inline: false, prefix: m[1].slice(2).split('-')[0] }));

/** Every token under `ui/styles/theme/`, the `@theme` blocks plus the runtime variables of `vars.css`. */
export const THEME_TOKENS: ReadonlyArray<ThemeToken> = Object.entries(sources).flatMap(([key, css]) => {
	const file = key.slice(key.lastIndexOf('/') + 1);
	const runtime = file === 'vars.css';
	return (runtime ? parseRootVars(css) : parseThemeVars(css)).map(token => ({ ...token, file, runtime }));
});

/** Tokens whose name starts with `--<prefix>-`, minus those starting with `--<exclude>-`. */
export const tokensNamed = (prefix: string, exclude?: string): ThemeToken[] =>
	THEME_TOKENS.filter(token => token.name.startsWith(`--${prefix}-`) && !(exclude && token.name.startsWith(`--${exclude}-`)));
