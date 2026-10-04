import { usePlayer } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';

import { ConnectedBooleanPicker } from '../ConnectedBooleanPicker';
import { ConnectedChoiceField } from '../ConnectedChoiceField';
import { ConnectedNumberField } from '../ConnectedTextField';
import type { ConnectedInputConfig } from './types';

export interface ConnectedInputProps {
	config: ConnectedInputConfig<Player<any>>;
	/**
	 * Lay the label and the control on one line. The rotation tab's simple-rotation section sets it
	 * for every input; the settings blocks do not.
	 */
	inline?: boolean;
}

export const ConnectedInput = ({ config, inline }: ConnectedInputProps) => {
	const player = usePlayer();
	const shared = inline ? { ...config, layout: 'inline' as const } : config;

	switch (shared.type) {
		case 'number':
			return <ConnectedNumberField modObject={player} config={shared} />;
		case 'boolean':
			return <ConnectedBooleanPicker modObject={player} config={{ ...shared, reverse: true }} />;
		case 'enum':
			return <ConnectedChoiceField modObject={player} config={shared} />;
	}
};
