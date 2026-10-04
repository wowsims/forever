import { type AriaRole, createContext, type RefObject } from 'react';

import type { FieldDensity, FieldLayout } from '../types';

export interface FieldContextValue {
	id: string | undefined;
	role: AriaRole | undefined;
	layout: FieldLayout | undefined;
	density: FieldDensity | undefined;
	disabled: boolean;
	grow: boolean;
	/** The ids of the root's descriptions, space-separated; `undefined` without one. */
	describedBy: string | undefined;
	registerDescription: (id: string) => () => void;
	/** Generated once in the root, whether or not a label renders: an icon field's label carries it, and its control is named by it. */
	labelId: string;
	/** The root's control, an icon field's label clicks through this. */
	controlRef: RefObject<HTMLElement | null>;
}

export const FieldContext = createContext<FieldContextValue | null>(null);
