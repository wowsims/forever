import { useApl } from '@features/apl/context/AplContext';
import { type InputState, useInput } from '@features/inputs/hooks/useInput';
import type { InputConfig } from '@features/inputs/input';
import type { Player } from '@sim/player/player';
import type { FieldFrameProps } from '@ui-kit/Field';
import { useId } from 'react';

export interface AplInput<T> extends InputState<T> {
	/** The caller's label, layout and classes for its `Field`, with the id filled in. */
	frameProps: FieldFrameProps & { id: string };
}

/**
 * `useInput` for a picker inside the APL tree, filling in the two things no such picker is given:
 *
 * - **the rotation as the change source**, or inside a list row that row's `rowSource`. Without it
 *   `useStoreSubscribe` never re-reads and the picker freezes at its first render.
 * - **an id.** `Field` puts it on the label's `htmlFor`.
 */
export const useAplInput = <T>(player: Player<any>, config: InputConfig<Player<any>, T>): AplInput<T> => {
	const generatedId = useId();
	const { changeSource } = useApl();
	const id = config.id || generatedId;
	const frameProps = {
		id,
		label: config.label,
		labelTooltip: config.labelTooltip,
		description: config.description,
		layout: config.layout,
		className: config.extraClassNames,
	};
	return { ...useInput(player, { ...config, storeSubscribe: changeSource, id }), frameProps };
};
