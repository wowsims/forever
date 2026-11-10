// Unit-test config. The site build stays in vite.config.mts; this only adds a DOM and the
// test-file glob, and reuses the same aliases so a test resolves imports exactly as the app does.
import path from 'path';
import { defineConfig } from 'vitest/config';

import { simlPlugin } from './tools/vite/siml.mts';
import { UI_ALIASES } from './vite.config.mts';

const here = import.meta.dirname;

export default defineConfig({
	resolve: {
		alias: {
			'virtual:i18next-loader': path.resolve(here, 'tools/vite/stub-i18n.js'),
			...UI_ALIASES,
		},
	},
	plugins: [simlPlugin()],
	// Same JSX contract as the app.
	oxc: {
		jsx: { runtime: 'automatic', importSource: 'react' },
	},
	test: {
		environment: 'happy-dom',
		setupFiles: [path.resolve(here, 'vitest.setup.ts')],
		include: ['ui/**/*.test.ts', 'ui/**/*.test.tsx'],
		restoreMocks: true,
		// Vitest empties every CSS import it does not include, `?raw` too; SPEC_THEMES and COLOR_TOKENS parse these.
		css: { include: [/\/ui\/styles\/theme\/(specs|colors)\.css\?raw$/] },
		// An externalised dependency imports chart.js natively, past `vi.mock('chart.js')`, and the
		// real one cannot build its scales under happy-dom.
		server: { deps: { inline: ['react-chartjs-2'] } },
	},
});
