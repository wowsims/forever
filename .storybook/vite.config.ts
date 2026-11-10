import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';
import i18nextLoader from 'vite-plugin-i18next-loader';

import { simlPlugin } from '../tools/vite/siml.mjs';
import { UI_ALIASES } from '../vite.config.mjs';

export default defineConfig({
	resolve: { alias: { ...UI_ALIASES } },
	plugins: [react(), tailwindcss(), i18nextLoader({ namespaceResolution: 'basename', paths: ['assets/locales'] }), simlPlugin()],
	define: { 'import.meta.env.VITE_GITHUB_REPOSITORY': JSON.stringify(process.env.GITHUB_REPOSITORY ?? '') },
	oxc: { jsx: { runtime: 'automatic', importSource: 'react' } },
});
