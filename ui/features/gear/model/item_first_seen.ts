// One entry per item id: the unix second the sim's database first held the item.
const itemFirstSeenUrl = '/forever/assets/database/item_first_seen.json';

const NEWLY_DISCOVERED_WINDOW_MS = 14 * 24 * 60 * 60 * 1000;

export const isNewlyDiscovered = (firstSeenSeconds: number): boolean => Date.now() - firstSeenSeconds * 1000 <= NEWLY_DISCOVERED_WINDOW_MS;

let load: Promise<Map<number, number>> | null = null;

// Fetched once per page. A failed fetch gives an empty map and is never memoized (same reset as
// Database.get), so the next list to open tries again.
export const loadItemFirstSeen = (): Promise<Map<number, number>> => {
	load ??= (async () => {
		try {
			const response = await fetch(itemFirstSeenUrl);
			if (!response.ok) throw new Error(`HTTP ${response.status}`);
			const entries: Array<{ id: number; firstSeen: number }> = await response.json();
			return new Map(entries.map(({ id, firstSeen }) => [id, firstSeen]));
		} catch (error) {
			console.warn(`Item first-seen dates failed to load from ${itemFirstSeenUrl}:`, error);
			load = null;
			return new Map<number, number>();
		}
	})();
	return load;
};
