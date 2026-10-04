export interface CssRule {
	selector: string;
	/** The whole rule as written, `selector { … }`, dedented. */
	text: string;
	start: number;
}

export interface UiClass {
	name: string;
	rules: string[];
}

export interface UiClassFile {
	file: string;
	classes: UiClass[];
}

const dedent = (text: string) => {
	const lines = text.split('\n');
	const indent = /^\s*/.exec(lines[lines.length - 1])![0];
	return lines.map((line, i) => (i > 0 && line.startsWith(indent) ? line.slice(indent.length) : line)).join('\n');
};

/** Every block in `css`, nested ones included, by a brace-balanced walk that skips quoted strings. */
export const parseCssRules = (css: string): CssRule[] => {
	const rules: CssRule[] = [];
	const open: Array<{ selector: string; start: number }> = [];
	let segment = 0;
	for (let i = 0; i < css.length; i++) {
		const ch = css[i];
		if (ch === '"' || ch === "'") {
			i = css.indexOf(ch, i + 1);
			if (i < 0) break;
		} else if (ch === ';') {
			segment = i + 1;
		} else if (ch === '{') {
			const raw = css.slice(segment, i);
			const start = segment + raw.length - raw.trimStart().length;
			open.push({ selector: raw.trim(), start });
			segment = i + 1;
		} else if (ch === '}') {
			const rule = open.pop();
			if (rule) rules.push({ ...rule, text: dedent(css.slice(rule.start, i + 1)) });
			segment = i + 1;
		}
	}
	return rules.sort((a, b) => a.start - b.start);
};

const UI_CLASS_RE = /\.(ui-[a-z0-9-]+)/g;

const parseUiClasses = (css: string): UiClass[] => {
	const classes = new Map<string, UiClass>();
	for (const rule of parseCssRules(css)) {
		for (const name of new Set([...rule.selector.matchAll(UI_CLASS_RE)].map(m => m[1]))) {
			const entry = classes.get(name) ?? { name, rules: [] };
			entry.rules.push(rule.text);
			classes.set(name, entry);
		}
	}
	return [...classes.values()];
};

const sources = import.meta.glob<string>(['../../**/*.css', '!../../styles/**', '!../../generated/**'], { query: '?raw', import: 'default', eager: true });

/** Every `.ui-*` class the co-located component stylesheets select, grouped by file. */
export const UI_CLASS_FILES: ReadonlyArray<UiClassFile> = Object.entries(sources)
	.map(([key, css]) => ({ file: key.replace(/^(\.\.\/)+/, 'ui/'), classes: parseUiClasses(css) }))
	.filter(entry => entry.classes.length > 0)
	.sort((a, b) => a.file.localeCompare(b.file));
