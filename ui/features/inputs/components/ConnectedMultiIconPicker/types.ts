import type { Player } from '@sim/player/player';
import type { ActionId } from '@sim/proto/action_id';
import type { MultiIconPickerProps } from '@ui-kit/MultiIconPicker';

import type { KeptInput } from '../../input';
import type { IconPickerConfig } from '../ConnectedIconPicker/types';

export type MultiIconPickerItemConfig<ModObject> = IconPickerConfig<ModObject, any>;

export interface MultiIconPickerConfig<ModObject> extends KeptInput<ModObject, boolean>, Pick<MultiIconPickerProps, 'direction' | 'tooltip'> {
	inputs: Array<MultiIconPickerItemConfig<ModObject>>;
	label?: string;
	categoryId?: ActionId;
	showWhen?: (obj: Player<any>) => boolean;
	// The menu's columns when it opens below, from the inputs the player is offered.
	numColumns?: (obj: ModObject) => number;
}
