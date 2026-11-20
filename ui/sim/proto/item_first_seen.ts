const NEWLY_DISCOVERED_WINDOW_MS = 7 * 24 * 60 * 60 * 1000;

/** Whether the sim's database first held an item, at that unix second, within the last week. */
export const isNewlyDiscovered = (firstSeenSeconds: number): boolean => Date.now() - firstSeenSeconds * 1000 <= NEWLY_DISCOVERED_WINDOW_MS;
