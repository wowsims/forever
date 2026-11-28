import { BUILD_INFO } from '@sim/constants/build_info';

import type { BuildInfo } from '../../tools/vite/build_info.mjs';

/**
 * A release as a build reads it: Feral Cat with every part, Mage, a one-spec class, with its presets,
 * and DPS Warrior with only its game build, as a shallow clone reads it. Vitest and Storybook define
 * no VITE_BUILD_INFO, so BUILD_INFO is empty there.
 */
export const RELEASE: BuildInfo = {
	date: '2026-10-18',
	commit: '2fe052cd96a4b1f0e2d3c4b5a697887766554433',
	gameVersion: '1.60.1.70450',
	specs: {
		'druid/feralcat': {
			gameVersion: '1.60.1.70425',
			presets: { commit: '8109a4ce0d1e2f3a4b5c6d7e8f90a1b2c3d4e5f6', date: '2026-10-17' },
			engine: { commit: '474c4981a2b3c4d5e6f708192a3b4c5d6e7f8091', date: '2026-10-16' },
		},
		'mage/dps': {
			gameVersion: '1.60.1.70450',
			presets: { commit: 'f3f5118a0b1c2d3e4f5061728394a5b6c7d8e9f0', date: '2026-10-15' },
		},
		'warrior/dps': { gameVersion: '1.60.1.70450' },
	},
};

/** Puts `info` in BUILD_INFO, for a story's `beforeEach` or a test; returns the restore. */
export const withRelease = (info: BuildInfo): (() => void) => {
	const saved = { ...BUILD_INFO };
	Object.assign(BUILD_INFO, info);
	return () => {
		Object.assign(BUILD_INFO, saved);
	};
};
