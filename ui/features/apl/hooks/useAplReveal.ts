import { useEffect, useEffectEvent } from 'react';

import { type AplRevealRequest, aplReveals } from '../model/reveal';

/** Calls `onReveal` for each entry a cast's click opens while the caller is mounted; a click from before it mounted is gone. */
export const useAplReveal = (onReveal: (request: AplRevealRequest) => void) => {
	const reveal = useEffectEvent(onReveal);
	useEffect(() => aplReveals.on(request => reveal(request)), []);
};
