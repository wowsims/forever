import { mod } from '@sim/utils/math';

/** The entry ArrowUp or ArrowDown moves to from `current`, wrapping at either end; null for any other key or an entry the rail does not hold. */
export const railStep = <T>(rail: readonly T[], current: T, key: string): T | null => {
	if (key !== 'ArrowUp' && key !== 'ArrowDown') return null;
	const index = rail.indexOf(current);
	return index < 0 ? null : rail[mod(index + (key === 'ArrowUp' ? -1 : 1), rail.length)];
};
