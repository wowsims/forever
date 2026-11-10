import { readFileSync, readdirSync } from 'node:fs';
import { basename, dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { build } from 'esbuild';

const source = dirname(dirname(fileURLToPath(import.meta.url)));
const repository = join(source, '..', '..');

// The site's i18n config imports its strings from a vite plugin's virtual module; the panel takes the
// English ones, which is the language the site's APL labels are written in.
const i18nLoader = {
	name: 'i18next-loader',
	setup(bundle) {
		bundle.onResolve({ filter: /^virtual:i18next-loader$/ }, args => ({ path: args.path, namespace: 'i18next-loader' }));
		bundle.onLoad({ filter: /.*/, namespace: 'i18next-loader' }, () => {
			const dir = join(repository, 'assets', 'locales', 'en');
			const en = Object.fromEntries(
				readdirSync(dir)
					.filter(file => file.endsWith('.json'))
					.sort()
					.map(file => [basename(file, '.json'), JSON.parse(readFileSync(join(dir, file), 'utf8'))]),
			);
			return { contents: `export default ${JSON.stringify({ en })};`, loader: 'js' };
		});
	},
};

// The panel's page, webview.js and webview.css in `outdir`. It imports the site's APL model through the repository's
// path aliases and packages.
export function buildPanel(outdir) {
	return build({
		absWorkingDir: source,
		entryPoints: { webview: 'src/webview/main.tsx' },
		outdir,
		bundle: true,
		platform: 'browser',
		format: 'iife',
		target: 'es2022',
		tsconfig: join(repository, 'tsconfig.json'),
		jsx: 'automatic',
		plugins: [i18nLoader],
		define: { 'process.env.NODE_ENV': '"production"', 'import.meta.env': '{}' },
		loader: { '.svg': 'empty', '.png': 'empty', '.jpg': 'empty', '.webp': 'empty' },
		minify: true,
		legalComments: 'none',
		logLevel: 'warning',
	});
}

// The extension as one Node bundle, which VS Code loads. `external` leaves more modules to the loader, as the tests do
// with the language client.
export function bundleExtension(outfile, external = []) {
	return build({
		absWorkingDir: source,
		entryPoints: ['src/extension.ts'],
		outfile,
		bundle: true,
		platform: 'node',
		format: 'cjs',
		target: 'node20',
		external: ['vscode', ...external],
		minify: true,
		logLevel: 'warning',
	});
}
