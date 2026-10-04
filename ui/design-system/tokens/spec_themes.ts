import specsCss from '../../styles/theme/specs.css?raw';

export interface SpecTheme {
	key: string;
	className: string;
	name: string;
	/** Root-absolute, as `specs.css` writes it: `/forever/assets/img/…`. */
	backgroundImage?: string;
}

const SPEC_CLASS = /\.([a-z]+(?:-[a-z]+)*)-sim-ui\b/g;
const RULE = /([^{}]+)\{([^{}]*)\}/g;
const BACKGROUND_IMAGE = /--theme-background-image:\s*url\(([^)]+)\)/;

const titleCase = (key: string) =>
	key
		.split('-')
		.map(word => word[0].toUpperCase() + word.slice(1))
		.join(' ');

const parseSpecThemes = (css: string): Array<SpecTheme> => {
	const themes = new Map<string, SpecTheme>();
	for (const [, selectors, body] of css.matchAll(RULE)) {
		const backgroundImage = BACKGROUND_IMAGE.exec(body)?.[1];
		for (const [className, key] of selectors.matchAll(SPEC_CLASS)) {
			const theme = themes.get(key) ?? { key, className: className.slice(1), name: titleCase(key) };
			if (backgroundImage) theme.backgroundImage = backgroundImage;
			themes.set(key, theme);
		}
	}
	return [...themes.values()];
};

/** Every `<spec>-sim-ui` class `specs.css` selects, in file order. */
export const SPEC_THEMES: ReadonlyArray<SpecTheme> = parseSpecThemes(specsCss);
