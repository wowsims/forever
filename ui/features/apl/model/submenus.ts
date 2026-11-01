import i18n from '@i18n/config';
import type { ChoiceEntry, ChoiceOption } from '@ui-kit/ChoiceField';

/**
 * An option filed by path. A string segment names a category submenu; a value segment files it
 * under the option carrying that value, which is how a pet is filed under its owner and a spell's
 * lower ranks under its highest.
 */
export interface PathOption<V> extends Omit<ChoiceOption<V>, 'submenu'> {
	submenu?: ReadonlyArray<string | V>;
}

/**
 * Translates a lowercase_underscore segment through the APL submenu namespace and appends a `»`. A
 * segment that is not such a string — a value segment, or a category already in the user's
 * language — is used as it stands.
 */
const submenuLabel = (segment: unknown): string => {
	if (typeof segment !== 'string') return '';
	const group = /^group_(\d+)$/.exec(segment);
	if (group) return `${i18n.t('rotation_tab.apl.submenus.group', { group: group[1] })} »`;
	if (!/^[a-z_]+$/.test(segment)) return `${segment} »`;
	const key = `rotation_tab.apl.submenus.${segment}`;
	const translated = i18n.t(key);
	return `${translated === key ? segment : translated} »`;
};

interface Submenu<V> {
	segment: string | V;
	entries: Array<ChoiceEntry<V>>;
}

/**
 * Folds a flat option list into the nested choices its `submenu` paths describe, each submenu where
 * its first option appears.
 *
 * Two flavours of path segment, both of which the APL pickers use:
 * - a **string** names a category submenu (`['resources', 'mana']`, two deep at most today);
 * - a **value** makes the option carrying that value the submenu's own option, still selectable.
 */
const choiceTree = <V>(options: ReadonlyArray<PathOption<V>>, equals: (a: V, b: V) => boolean): Array<ChoiceEntry<V>> => {
	const root: Array<ChoiceEntry<V>> = [];
	const submenus = new Map<Array<ChoiceEntry<V>>, Array<Submenu<V>>>();

	const matches = (submenu: Submenu<V>, segment: string | V) =>
		typeof segment === 'string' ? submenu.segment === segment : typeof submenu.segment !== 'string' && equals(submenu.segment as V, segment);

	const open = (parent: Array<ChoiceEntry<V>>, segment: string | V): Submenu<V> => {
		const siblings = submenus.get(parent) ?? [];
		const found = siblings.find(submenu => matches(submenu, segment));
		if (found) return found;
		const submenu: Submenu<V> = { segment, entries: [] };
		parent.push({ label: submenuLabel(segment), submenu: submenu.entries });
		submenus.set(parent, [...siblings, submenu]);
		return submenu;
	};

	const owners = options.flatMap(option => (option.submenu ?? []).filter((segment): segment is V => typeof segment !== 'string'));
	const opensSubmenu = (option: PathOption<V>) => owners.some(owner => equals(owner, option.value));

	options.forEach(({ submenu: path = [], ...option }) => {
		const parent = path.reduce((entries, segment) => open(entries, segment).entries, root);
		if (!opensSubmenu(option)) {
			parent.push(option);
			return;
		}
		const { entries } = open(parent, option.value);
		const position = parent.findIndex(entry => entry.submenu === entries);
		parent[position] = { ...option, submenu: entries };
	});

	return root;
};

/**
 * A choice field's `options` and `optionTree` for a path-filed list: the options without their paths,
 * which the trigger reads, and a `choiceTree` builder the field runs only while its menu is open. A
 * list with no path gets no builder and stays a select.
 */
export const pathChoices = <V>(options: ReadonlyArray<PathOption<V>>, equals: (a: V, b: V) => boolean) => ({
	options: options.map(({ submenu: _path, ...option }): ChoiceOption<V> => option),
	optionTree: options.some(option => option.submenu?.length) ? () => choiceTree(options, equals) : undefined,
});
