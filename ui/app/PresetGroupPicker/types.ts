import type { Phase } from '@sim/constants/other';
import type { ReactNode } from 'react';

export interface PresetGroupItem {
	key: string;
	phase?: Phase;
	group?: string;
	node: ReactNode;
}

export interface PresetGroupSection {
	title: string;
	tooltip?: string;
	items: ReadonlyArray<PresetGroupItem>;
	/** Rendered under the section's chips, outside the phase filter (the user's own saved sets). */
	footer?: ReactNode;
	/** Rendered at the right of the section's heading. */
	headerChildren?: ReactNode;
}
