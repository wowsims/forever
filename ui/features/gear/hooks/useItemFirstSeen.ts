import { useEffect, useState } from 'react';

import { loadItemFirstSeen } from '../model/item_first_seen';

// Empty until the file arrives; the loader never rejects.
export const useItemFirstSeen = (): ReadonlyMap<number, number> => {
	const [firstSeen, setFirstSeen] = useState<ReadonlyMap<number, number>>(() => new Map());

	useEffect(() => {
		loadItemFirstSeen().then(setFirstSeen);
	}, []);

	return firstSeen;
};
