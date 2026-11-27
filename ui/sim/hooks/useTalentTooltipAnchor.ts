import { hasLoadedSpellTooltip, hasSpellTooltip, spellTooltipAnchor } from '@sim/proto/tooltips';
import { useEffect, useMemo, useState } from 'react';

/** The props that open a talent's own tooltip on its anchor, once the tooltips have loaded and hold it; none before, which leaves wowhead's. */
export const useTalentTooltipAnchor = (spellId: number): Record<string, string> => {
	const [ownTooltip, setOwnTooltip] = useState(() => hasLoadedSpellTooltip(spellId));

	useEffect(() => {
		let current = true;
		hasSpellTooltip(spellId).then(has => current && setOwnTooltip(has));
		return () => {
			current = false;
		};
	}, [spellId]);

	return useMemo(() => (ownTooltip ? spellTooltipAnchor(spellId, { talent: true }) : {}), [ownTooltip, spellId]);
};
