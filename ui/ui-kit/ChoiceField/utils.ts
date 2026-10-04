import type { ChoiceEntry, ChoiceOption } from './types';

export const isChoiceOption = <V>(entry: ChoiceEntry<V>): entry is ChoiceOption<V> => 'value' in entry;

export const hasSubmenus = <V>(entries: ReadonlyArray<ChoiceEntry<V>>) => entries.some(entry => entry.submenu !== undefined);

/** Every selectable option in menu order, a submenu's own option before its contents. */
export const flattenChoices = <V>(entries: ReadonlyArray<ChoiceEntry<V>>): Array<ChoiceOption<V>> =>
	entries.flatMap(entry => [...(isChoiceOption(entry) ? [entry] : []), ...(entry.submenu ? flattenChoices(entry.submenu) : [])]);
