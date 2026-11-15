import path from 'node:path';

import type { StorybookConfig } from '@storybook/react-vite';

const here = import.meta.dirname;

const config: StorybookConfig = {
	framework: {
		name: '@storybook/react-vite',
		// builder-vite loads ./vite.config.mts by default; that config carries the site build's
		// plugins (specPages, watchAndRun, checker), so Storybook gets a config of its own.
		options: { builder: { viteConfigPath: path.join(here, 'vite.config.ts') } },
	},
	stories: ['../ui/**/*.mdx', '../ui/**/*.stories.tsx'],
	addons: ['@storybook/addon-docs'],
	// Fonts and spec backgrounds are root-absolute under /forever/assets.
	staticDirs: [{ from: '../assets', to: '/forever/assets' }],
	docs: { defaultName: 'Docs' },
	typescript: { reactDocgen: 'react-docgen' },
	core: { disableTelemetry: true },
};

export default config;
