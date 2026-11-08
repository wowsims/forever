import { useEffect, useState } from 'react';

import { loadItemFirstSeen } from '../model/item_first_seen';

export const useItemFirstSeen = (): ReadonlyMap<number, number> | null => {
	const [firstSeen, setFirstSeen] = useState<ReadonlyMap<number, number> | null>(null);

	useEffect(() => {
		let live = true;
		loadItemFirstSeen().then(map => {
			if (live) setFirstSeen(map);
		});
		return () => {
			live = false;
		};
	}, []);

	return firstSeen;
};
