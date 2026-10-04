import { Unstyled } from '@storybook/addon-docs/blocks';
import { CLASS_TEXT } from '@ui-kit/utils/colors';
import clsx from 'clsx';

import { SPEC_THEMES, type SpecTheme } from '../tokens/spec_themes';

/** The class a spec key belongs to: the last dash-separated word (`feral-bear-druid` → `druid`). */
const classOf = (theme: SpecTheme) => theme.key.slice(theme.key.lastIndexOf('-') + 1);

const titleCase = (word: string) => word[0].toUpperCase() + word.slice(1);

const SpecThemeCard = ({ theme }: { theme: SpecTheme }) => (
	<div className={clsx('sim-ui', theme.className, 'flex min-h-56 flex-col gap-2 bg-sim bg-cover bg-center p-2 text-foreground')}>
		<strong>{theme.name}</strong>
		<code className="text-ui">.{theme.className}</code>
		<span className="mt-auto self-start bg-primary px-2 py-1 text-primary-foreground">bg-primary text-primary-foreground</span>
		<code className="truncate text-ui text-link" title={theme.backgroundImage}>
			{theme.backgroundImage}
		</code>
	</div>
);

const byClass = (): ReadonlyArray<[string, ReadonlyArray<SpecTheme>]> => {
	const groups = new Map<string, Array<SpecTheme>>();
	for (const theme of SPEC_THEMES) {
		const name = classOf(theme);
		groups.set(name, [...(groups.get(name) ?? []), theme]);
	}
	return [...groups.entries()].sort(([a], [b]) => a.localeCompare(b));
};

export const SpecThemeCards = () => (
	<Unstyled>
		<div className="flex flex-col gap-6">
			{byClass().map(([name, themes]) => (
				<section key={name} className="flex flex-col gap-2">
					<h3 className={clsx('font-bold', CLASS_TEXT[name])}>{titleCase(name)}</h3>
					<div className="grid grid-cols-1 gap-2 md:grid-cols-2">
						{themes.map(theme => (
							<SpecThemeCard key={theme.key} theme={theme} />
						))}
					</div>
				</section>
			))}
		</div>
	</Unstyled>
);
