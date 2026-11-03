import { type AriaRole, createContext, type RefObject } from 'react';

import type { ControlSize, FieldLayout } from '../types';

export interface FieldContextValue {
	id: string | undefined;
	role: AriaRole | undefined;
	layout: FieldLayout | undefined;
	/** The root's own size, else the size of the root it is nested in. */
	size: ControlSize | undefined;
	disabled: boolean;
	grow: boolean;
	/** The ids of the root's descriptions, space-separated; `undefined` without one. */
	describedBy: string | undefined;
	registerDescription: (id: string) => () => void;
	/** Generated once in the root, whether or not a label renders: an icon field's label carries it, and its control is named by it. */
	labelId: string;
	/** The root's control, an icon field's label clicks through this. */
	controlRef: RefObject<HTMLElement | null>;
	/** Opens and closes the label's tooltip at the label, for a control whose focus shows it; null without a label tooltip. */
	labelTooltip: RefObject<{ open: () => void; close: () => void } | null>;
}

export const FieldContext = createContext<FieldContextValue | null>(null);
