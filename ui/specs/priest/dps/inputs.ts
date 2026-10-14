import * as InputHelpers from '@features/inputs/input_helpers';
import { ActionId } from '@sim/proto/action_id';
import { PriestSpecs } from '@sim/proto/spec_types';

// Configuration for the DPS spec's own UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.

export const ShadowformInput = <SpecType extends PriestSpecs>() =>
	InputHelpers.makeClassOptionsBooleanIconInput<SpecType>({
		fieldName: 'preShadowform',
		label: 'Shadowform',
		id: ActionId.fromSpellId(15473),
		// `showWhen` reads talents, so the picker has to watch them as well as the default specOptions.
		storeField: ['specOptions', 'talentsString'] as const,
		showWhen: player => player.getTalents().shadowform,
	});
