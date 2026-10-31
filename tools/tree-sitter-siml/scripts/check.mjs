// The grammar's checks: the committed parser is what `tree-sitter generate` writes from grammar.js, the corpus under
// test/corpus parses as it records, every .siml file of the repository parses with no error, except the parser's error
// cases, the Zed extension's queries compile against the grammar, each of its snippets parses in some place a snippet
// goes, and the commit its extension.toml fetches the grammar from holds this grammar. The pinned CLI generates the
// parser, so every checkout writes the same src/parser.c.
import { spawnSync } from 'node:child_process';
import { cpSync, mkdtempSync, readdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const grammar = dirname(dirname(fileURLToPath(import.meta.url)));
const repository = join(grammar, '..', '..');
const cli = join(grammar, 'node_modules', 'tree-sitter-cli', 'cli.js');
const zed = join(repository, 'tools', 'zed-siml');
const generated = [
	'parser.c',
	'grammar.json',
	'node-types.json',
	join('tree_sitter', 'parser.h'),
	join('tree_sitter', 'alloc.h'),
	join('tree_sitter', 'array.h'),
];

function run(cwd, ...args) {
	const result = spawnSync(process.execPath, [cli, ...args], { cwd, stdio: 'inherit' });
	if (result.error) console.error(result.error.message);
	return result.status === 0;
}

// The text a snippet inserts when every tab stop keeps its default.
function expand(body) {
	const text = body.join('\n');
	const defaults = {};
	for (const [, stop, value] of text.matchAll(/\$\{(\d+):([^}]*)\}/g)) defaults[stop] ??= value;
	return text
		.replace(/\$\{(\d+):([^}]*)\}/g, (_, stop) => defaults[stop])
		.replace(/\$\{?(\d+)\}?/g, (_, stop) => defaults[stop] ?? '')
		.replace(/\\\$/g, '$');
}

// The places a snippet goes: a whole file, a top-level line, a row of a section, a prepull row or a row's condition.
function placements(text) {
	const indented = text.replace(/^/gm, '  ');
	return [
		text,
		`siml 1\nspec hunter\n${text}\n`,
		`siml 1\nspec hunter\npriority:\n${indented}\n`,
		`siml 1\nspec hunter\nprepull:\n${indented}\n`,
		`siml 1\nspec hunter\npriority:\n  cast x if ${text}\n`,
	];
}

let failed = false;
const scratch = mkdtempSync(join(tmpdir(), 'tree-sitter-siml-'));
try {
	for (const file of ['grammar.js', 'tree-sitter.json', 'package.json', join('src', 'scanner.c')]) {
		cpSync(join(grammar, file), join(scratch, file));
	}
	if (!run(scratch, 'generate')) {
		process.exit(1);
	}
	for (const file of generated) {
		if (!readFileSync(join(scratch, 'src', file)).equals(readFileSync(join(grammar, 'src', file)))) {
			console.error(`src/${file} is not what \`npm run generate\` writes from grammar.js: run it and commit the result`);
			failed = true;
		}
	}

	failed = !run(grammar, 'test') || failed;

	const queries = join(zed, 'languages', 'siml');
	const sample = join(repository, 'ui', 'specs', 'hunter', 'dps', 'apls', 'default.apl.siml');
	for (const query of readdirSync(queries).filter(f => f.endsWith('.scm'))) {
		const result = spawnSync(process.execPath, [cli, 'query', '--grammar-path', grammar, join(queries, query), sample], {
			cwd: grammar,
			stdio: ['ignore', 'ignore', 'inherit'],
		});
		if (result.status !== 0) {
			console.error(`tools/zed-siml/languages/siml/${query} does not compile against the grammar`);
			failed = true;
		}
	}

	// The CLI's own error count misses an error node under a node it reads as clean, so each printed tree is searched.
	const library = join(scratch, process.platform === 'win32' ? 'siml.dll' : process.platform === 'darwin' ? 'siml.dylib' : 'siml.so');
	if (!run(grammar, 'build', '--output', library, grammar)) {
		process.exit(1);
	}
	function unreadable(files) {
		const paths = join(scratch, 'paths.txt');
		writeFileSync(paths, files.join('\n'));
		const parsed = spawnSync(process.execPath, [cli, 'parse', '--lib-path', library, '--lang-name', 'siml', '--no-ranges', '--paths', paths], {
			encoding: 'utf8',
			maxBuffer: 1 << 30,
			stdio: ['ignore', 'pipe', 'inherit'],
		});
		const trees = parsed.stdout.split(/^(?=\(source_file)/m).filter(tree => tree.startsWith('(source_file'));
		if (parsed.status !== 0 && trees.length === 0) {
			console.error('tree-sitter parse failed');
			process.exit(1);
		}
		return new Set(files.filter((_, i) => /\((ERROR|MISSING)\b/.test(trees[i] ?? '(ERROR')));
	}

	const snippets = Object.entries(JSON.parse(readFileSync(join(zed, 'snippets', 'siml.json'), 'utf8')));
	const candidates = snippets.flatMap(([name, snippet], i) =>
		placements(expand(snippet.body)).map((text, j) => {
			const file = join(scratch, `snippet-${i}-${j}.siml`);
			writeFileSync(file, text);
			return { name, file };
		}),
	);
	const broken = unreadable(candidates.map(c => c.file));
	for (const [name] of snippets) {
		if (candidates.every(c => c.name !== name || broken.has(c.file))) {
			console.error(`The snippet "${name}" of tools/zed-siml/snippets/siml.json parses in no place a snippet goes`);
			failed = true;
		}
	}

	const listed = spawnSync('git', ['ls-files', '-z', '--', '*.siml'], { cwd: repository, encoding: 'utf8' });
	if (listed.status !== 0) {
		console.error(listed.stderr);
		process.exit(1);
	}
	const files = listed.stdout
		.split('\0')
		.filter(f => f && !f.startsWith('sim/siml/testdata/errors/'))
		.map(f => join(repository, f));
	console.log(`Parsing ${files.length} .siml files`);
	for (const file of unreadable(files)) {
		console.error(`${file} does not parse: \`npx tree-sitter parse ${file}\` shows the error nodes`);
		failed = true;
	}
	// Zed fetches the grammar from the commit extension.toml names, so that commit's src/ must be this one: a change to
	// the grammar lands, and a later commit points rev at it.
	const manifest = readFileSync(join(zed, 'extension.toml'), 'utf8');
	const entry = manifest.match(/\[grammars\.siml\]\s*\nrepository = "([^"]*)"\s*\nrev = "([0-9a-f]{40})"\s*\npath = "([^"]*)"/);
	if (!entry || entry[1] !== 'https://github.com/wowsims/forever' || entry[3] !== 'tools/tree-sitter-siml') {
		console.error(
			'tools/zed-siml/extension.toml names the grammar as [grammars.siml] with this repository, a full rev and path = "tools/tree-sitter-siml"',
		);
		failed = true;
	} else {
		const rev = entry[2];
		const git = (...args) => spawnSync('git', args, { cwd: repository, encoding: 'utf8' });
		if (git('cat-file', '-e', `${rev}^{commit}`).status !== 0) {
			git('fetch', '--quiet', '--no-tags', '--depth', '1', 'origin', rev);
		}
		const diff = git('diff', '--quiet', rev, '--', join('tools', 'tree-sitter-siml', 'src'));
		if (diff.status !== 0) {
			console.error(
				diff.status === 1
					? `tools/tree-sitter-siml/src differs from the commit tools/zed-siml/extension.toml fetches (${rev}): once the grammar's commit is pushed, set rev to it`
					: `The commit tools/zed-siml/extension.toml fetches the grammar from (${rev}) is not in the repository: ${diff.stderr.trim()}`,
			);
			failed = true;
		}
	}
} finally {
	rmSync(scratch, { recursive: true, force: true });
}
process.exit(failed ? 1 : 0);
