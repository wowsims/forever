import colorsCss from '../../styles/theme/colors.css?raw';
import { parseThemeVars, type ThemeVar } from './parse_theme';

export interface ColorFamily {
	name: string;
	/** The second segment of every `--color-<segment>-*` token that belongs to the family. */
	segments: ReadonlyArray<string>;
}

export const COLOR_FAMILIES: ReadonlyArray<ColorFamily> = [
	{ name: 'Base', segments: ['black', 'white', 'grey', 'gray', 'blue', 'orange', 'maroon', 'frozen', 'light', 'off'] },
	{ name: 'Surface', segments: ['background', 'foreground', 'surface', 'overlay', 'panel', 'border', 'muted', 'scrim', 'input', 'table', 'spinner', 'sim'] },
	{ name: 'Primary', segments: ['primary'] },
	{ name: 'Secondary', segments: ['secondary'] },
	{ name: 'Danger', segments: ['danger'] },
	{ name: 'Cancel', segments: ['cancel'] },
	{ name: 'Link', segments: ['link'] },
	{ name: 'Status', segments: ['success', 'warning', 'info', 'brand', 'expansion'] },
	{ name: 'Social', segments: ['discord'] },
	{ name: 'Landing', segments: ['landing'] },
	{ name: 'Class', segments: ['class'] },
	{ name: 'School', segments: ['school'] },
	{ name: 'Quality', segments: ['quality'] },
	{ name: 'Resource', segments: ['resource'] },
	{ name: 'Damage', segments: ['damage'] },
	{ name: 'Timeline', segments: ['timeline'] },
	{ name: 'Chart series', segments: ['chart'] },
	{ name: 'Faction', segments: ['faction'] },
	{ name: 'Talent', segments: ['talent'] },
	{ name: 'Combat replay', segments: ['cr'] },
];

export const COLOR_TOKENS: ReadonlyArray<ThemeVar> = parseThemeVars(colorsCss).filter(token => token.prefix === 'color');

export const familyOf = (token: ThemeVar): ColorFamily | undefined => {
	const segment = token.name.split('-')[3];
	return COLOR_FAMILIES.find(family => family.segments.includes(segment));
};
