import { useState } from 'react';

import { type LocalStorageParser, useTypedLocalStorage } from './useTypedLocalStorage';

export const parseStringArray: LocalStorageParser<string[]> = value =>
	Array.isArray(value) && value.every(entry => typeof entry === 'string') ? (value as string[]) : undefined;

/** React state seeded from storage and written through: a blocked `localStorage` writes nothing and reads back null, and the state still has to move. */
export const useLocalStorageState = <T>(key: string, parse: LocalStorageParser<T>, initial: () => T): [T, (next: T) => void] => {
	const [stored, setStored] = useTypedLocalStorage<T>(key, parse);
	const [value, setValue] = useState(() => stored ?? initial());

	return [
		value,
		next => {
			setValue(next);
			setStored(next);
		},
	];
};
