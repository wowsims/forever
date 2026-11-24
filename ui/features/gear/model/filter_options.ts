import type { AreaType, CreatureFamily, ItemRandomSuffix, MobType, Stat } from '@generated/proto/common';
import type { UIItem } from '@generated/proto/ui';
import { itemOwnStats } from '@sim/proto/items';

export interface FilterOptions {
	areaTypes: AreaType[];
	mobTypes: MobType[];
	creatureFamilies: CreatureFamily[];
	itemStats: Stat[];
}

/** What the Filters menu offers for `items`, each list sorted: a stat counts when the item or a random suffix it can roll carries it. */
export const filterOptionsOf = (items: readonly UIItem[], suffixById: (id: number) => ItemRandomSuffix | undefined): FilterOptions => {
	const listed = <T extends number>(typesOf: (item: UIItem) => Array<T>) => [...new Set(items.flatMap(typesOf))].sort((a, b) => a - b);
	// Each random suffix is read once, however many items can roll it.
	const carried = new Set<Stat>(items.flatMap(item => [...itemOwnStats(item)]));
	for (const id of new Set(items.flatMap(item => item.randomSuffixOptions))) {
		suffixById(id)?.stats.forEach((value, stat) => {
			if (value > 0) carried.add(stat);
		});
	}
	return {
		areaTypes: listed(item => item.areaTypes),
		mobTypes: listed(item => item.mobTypes),
		creatureFamilies: listed(item => item.creatureFamilies),
		itemStats: [...carried].sort((a, b) => a - b),
	};
};
