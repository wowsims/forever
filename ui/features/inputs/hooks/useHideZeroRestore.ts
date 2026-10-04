import { useEffect, useRef } from 'react';

import type { InputState } from './useInput';

export const useHideZeroRestore = <V>(
	state: Pick<InputState<any, V>, 'value' | 'setValue' | 'hidden' | 'revision'>,
	zero: V,
	restoreWhen: (current: V) => boolean = () => true,
	/** Runs after every notification the hook acts on, with the value it left the source holding. */
	onSettled?: (current: V) => void,
) => {
	const stored = useRef<V | undefined>(undefined);
	const lastRevision = useRef<number | null>(null);
	useEffect(() => {
		if (lastRevision.current === state.revision) return; // StrictMode replay: same revision, no-op
		const construction = lastRevision.current === null;
		lastRevision.current = state.revision;
		if (construction) return;
		let current = state.value;
		if (state.hidden) {
			if (stored.current === undefined) {
				stored.current = state.value;
				current = zero;
				state.setValue(current);
			}
		} else if (stored.current !== undefined) {
			const restored = stored.current;
			stored.current = undefined;
			if (restoreWhen(state.value)) {
				current = restored;
				state.setValue(current);
			}
		}
		onSettled?.(current);
	});
	return {
		clearStored: () => {
			stored.current = undefined;
		},
	};
};
