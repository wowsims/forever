import type { ResolvedIcon } from '@ui-kit/Field/types';
import type { WowheadIconProps } from '@ui-kit/WowheadIcon';

export enum IconEnumPickerDirection {
	Vertical = 'vertical',
	Horizontal = 'Horizontal',
}

export interface IconEnumOption<T> {
	value: T;
	// One of these should be set.
	// If icon is set, shows that icon.
	// If color is set, shows that color.
	// If iconUrl is set, shows that icon as grayscale
	icon?: ResolvedIcon;
	color?: string;
	iconUrl?: string;
	// Text to be displayed on the icon.
	text?: string;
	// Hover tooltip.
	tooltip?: string;
	// The `data-wowhead` tooltip attributes, spread onto the option's swatch, and onto the trigger's while it is selected.
	wowhead?: WowheadIconProps['wowhead'];
	// A hidden option is left out of the menu, and selecting its value shows as no selection.
	hidden?: boolean;
}
