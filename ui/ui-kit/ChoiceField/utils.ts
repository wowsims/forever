import type { ChoiceEntry, ChoiceOption, ChoiceSection } from './types';

export const isChoiceOption = <V>(entry: ChoiceEntry<V>): entry is ChoiceOption<V> => 'value' in entry;

export const isChoiceSection = <V>(entry: ChoiceEntry<V>): entry is ChoiceSection<V> => 'options' in entry;

export const hasSubmenus = <V>(entries: ReadonlyArray<ChoiceEntry<V>>) => entries.some(entry => entry.submenu !== undefined);

/** Every selectable option in menu order, a submenu's own option before its contents. */
export const flattenChoices = <V>(entries: ReadonlyArray<ChoiceEntry<V>>): Array<ChoiceOption<V>> =>
	entries.flatMap(entry =>
		isChoiceSection(entry) ? [...entry.options] : [...(isChoiceOption(entry) ? [entry] : []), ...(entry.submenu ? flattenChoices(entry.submenu) : [])],
	);
