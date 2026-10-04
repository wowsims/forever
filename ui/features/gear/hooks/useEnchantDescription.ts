import type { UIEnchant as Enchant } from '@generated/proto/ui';
import { getEnchantDescription } from '@sim/proto/enchants';
import { useEffect, useState } from 'react';

// The enchant's effect text, or its name where it has none.
export const useEnchantDescription = (enchant: Enchant | null | undefined): string => {
	const [description, setDescription] = useState('');

	useEffect(() => {
		if (!enchant) {
			setDescription('');
			return;
		}
		let live = true;
		getEnchantDescription(enchant)
			.then(text => {
				if (live) setDescription(text);
			})
			.catch(() => {
				if (live) setDescription(enchant.name);
			});
		return () => {
			live = false;
		};
	}, [enchant]);

	return description;
};
