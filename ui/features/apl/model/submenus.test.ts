import type { ChoiceEntry, ChoiceOption } from '@ui-kit/ChoiceField';
import { describe, expect, it, vi } from 'vitest';

import { pathChoices, type PathOption } from './submenus';

// Only `resources` is translated here, so the untranslated branch is exercised too.
vi.mock('@i18n/config', () => ({ default: { t: (key: string) => (key === 'rotation_tab.apl.submenus.resources' ? 'Resources' : key) } }));

const sameId = (a: string, b: string) => a === b;

/** A readable shape for the tree: options by value, a submenu as `label(...)`, one opened by an option as `value*(...)`. */
const shape = <V>(entries: ReadonlyArray<ChoiceEntry<V>>): string =>
	entries
		.map(entry => {
			const head = 'value' in entry ? String(entry.value) : String(entry.label);
			if (!entry.submenu) return head;
			return `${'value' in entry ? `${head}*` : head}(${shape(entry.submenu)})`;
		})
		.join(',');

const choiceTree = <V>(options: ReadonlyArray<PathOption<V>>, equals: (a: V, b: V) => boolean) => pathChoices(options, equals).optionTree!();

describe('submenu labels', () => {
	const filed = (segment: string) => choiceTree([{ value: 'a', label: 'a', submenu: [segment] }], sameId)[0].label;

	it('translates a lowercase_underscore segment through the APL submenu namespace', () => {
		expect(filed('resources')).toBe('Resources »');
	});

	it('leaves an already-worded segment alone but still marks it as a submenu', () => {
		expect(filed('Groups')).toBe('Groups »');
	});

	it('falls back to the segment when the namespace has no entry', () => {
		expect(filed('logic')).toBe('logic »');
	});

	it('has nothing to say about a value segment, whose own option labels the submenu instead', () => {
		const tree = choiceTree([{ value: { type: 2 }, label: 'Pet', submenu: [{ type: 1 }] }], (a, b) => a.type === b.type);
		expect(tree[0].label).toBe('');
	});
});

describe('option tree', () => {
	const option = (value: string, submenu?: Array<string>, extra: Partial<PathOption<string>> = {}): PathOption<string> => ({
		value,
		label: value,
		submenu,
		...extra,
	});

	it('keeps options without a path in order at the root', () => {
		expect(shape(choiceTree([option('a'), option('b'), option('c', ['logic'])], sameId))).toBe('a,b,logic »(c)');
	});

	it('files an option under the category its path names', () => {
		expect(shape(choiceTree([option('a'), option('b', ['logic'])], sameId))).toBe('a,logic »(b)');
	});

	it('creates a category once and keeps it in first-mention position', () => {
		expect(shape(choiceTree([option('a', ['logic']), option('b'), option('c', ['logic'])], sameId))).toBe('logic »(a,c),b');
	});

	it('nests a two-segment path, which is the deepest any caller uses', () => {
		const tree = choiceTree([option('chi', ['resources', 'chi']), option('rage', ['resources', 'rage'])], sameId);
		expect(shape(tree)).toBe('Resources »(chi »(chi),rage »(rage))');
	});

	it('keeps the rest of the option: label, tooltip, icon and class', () => {
		const [logic] = choiceTree([option('a', ['logic'], { label: 'All of', tooltip: '<p>and</p>', icon: 'user', className: 'text-danger' })], sameId);
		expect(logic.submenu![0]).toEqual({ value: 'a', label: 'All of', tooltip: '<p>and</p>', icon: 'user', className: 'text-danger' });
	});

	// The pet case. A *string* segment is always a category name, so a value segment can only be
	// exercised with non-string values — which is what every real caller has.
	type Ref = { type: number };
	const byType = (a: Ref, b: Ref) => a.type === b.type;
	const self: PathOption<Ref> = { value: { type: 1 }, label: 'Self' };
	const pet = (owner: Ref): PathOption<Ref> => ({ value: { type: 2 }, label: 'Pet', submenu: [owner] });

	it('makes the option a value path names open that submenu, and keeps it selectable', () => {
		const tree = choiceTree([self, pet(self.value)], byType);

		expect(tree).toHaveLength(1);
		const owner = tree[0] as ChoiceOption<Ref>;
		expect(owner.value).toEqual({ type: 1 });
		expect(owner.label).toBe('Self');
		expect(owner.submenu!.map(entry => entry.label)).toEqual(['Pet']);
	});

	it('files a pet listed before its owner under the owner all the same', () => {
		const tree = choiceTree([pet(self.value), self], byType);

		expect(tree).toHaveLength(1);
		expect((tree[0] as ChoiceOption<Ref>).label).toBe('Self');
		expect(tree[0].submenu!.map(entry => entry.label)).toEqual(['Pet']);
	});

	it('reads a value segment through `equals`, not by identity', () => {
		// A *different object* with the same type: it must still find the owner.
		const tree = choiceTree([self, pet({ type: 1 })], byType);

		expect(tree).toHaveLength(1);
		expect((tree[0] as ChoiceOption<Ref>).label).toBe('Self');
	});
});

describe('pathChoices', () => {
	const options: Array<PathOption<string>> = [
		{ value: 'a', label: 'A', tooltip: 'first' },
		{ value: 'b', label: 'B', submenu: ['logic'] },
	];

	it('hands the trigger the options in order, without their paths', () => {
		expect(pathChoices(options, sameId).options).toEqual([
			{ value: 'a', label: 'A', tooltip: 'first' },
			{ value: 'b', label: 'B' },
		]);
	});

	it('builds the tree only when the builder is called', () => {
		const { optionTree } = pathChoices(options, sameId);

		expect(shape(optionTree!())).toBe('a,logic »(b)');
	});

	it('gives a list with no path no builder, so it stays a select', () => {
		expect(pathChoices([options[0], { value: 'c', label: 'C', submenu: [] }], sameId).optionTree).toBeUndefined();
	});
});
