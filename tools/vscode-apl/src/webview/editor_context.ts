import { createContext, useContext } from 'react';

import type { AplEdit, Range } from '../protocol';
import type { JsonPath } from './model';

export interface EditorApi {
	// Makes the edits on the panel's rotation and sends them to the extension. While the panel is read-only it does
	// nothing, and says so with false.
	edit: (...edits: AplEdit[]) => boolean;
	reveal: (path: JsonPath) => void;
	// Selects text of the file: a when line or a def.
	revealRange: (range: Range) => void;
	// Names the rotation defines, for the name pickers.
	groupNames: string[];
	variableNames: string[];
	// Every action id the rotation names, which the spell picker lists first.
	usedIds: string[];
	// The variables each group declares, which a reference to the group can set.
	groupVariables: Record<string, string[]>;
}

export const EditorContext = createContext<EditorApi>({
	edit: () => false,
	reveal: () => {},
	revealRange: () => {},
	groupNames: [],
	variableNames: [],
	usedIds: [],
	groupVariables: {},
});

export const useEditor = () => useContext(EditorContext);
