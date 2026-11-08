// One entry per item id: the unix second the sim first found the item in the game files.
const itemFirstSeenUrl = '/forever/assets/database/item_first_seen.json';

const NEWLY_DISCOVERED_WINDOW_MS = 14 * 24 * 60 * 60 * 1000;

export const isNewlyDiscovered = (firstSeenSeconds: number, nowMs: number): boolean => nowMs - firstSeenSeconds * 1000 <= NEWLY_DISCOVERED_WINDOW_MS;

let load: Promise<Map<number, number>> | null = null;

// Fetched once per page; a failed fetch gives an empty map and is not retried.
export const loadItemFirstSeen = (): Promise<Map<number, number>> => {
	load ??= (async () => {
		try {
			const response = await fetch(itemFirstSeenUrl);
			if (!response.ok) throw new Error(`HTTP ${response.status}`);
			const entries: Array<{ id: number; firstSeen: number }> = await response.json();
			return new Map(entries.map(({ id, firstSeen }) => [id, firstSeen]));
		} catch (error) {
			console.warn(`Item first-seen dates failed to load from ${itemFirstSeenUrl}:`, error);
			return new Map<number, number>();
		}
	})();
	return load;
};
