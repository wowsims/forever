// A point can always be taken back, even when that breaks a talent: the broken talents keep their points
// and the sim still applies them, so a build can be tested with a point taken out.

import { CHARACTER_LEVEL } from '@sim/constants/mechanics';
import { findBrokenTalents } from '@sim/talents/broken_talents';
import type { TalentsConfig } from '@sim/talents/config';
import type { TalentPoints } from '@sim/talents/talents_string';
import { totalPointsSpent, withTalentPoints } from '@sim/talents/talents_string';

// A character earns one talent point per level from 10 onwards: 51 at level 60.
export const MAX_POINTS_PLAYER = CHARACTER_LEVEL - 9;

export const canSetPoints = <TalentsProto>(
	config: TalentsConfig<TalentsProto>,
	points: TalentPoints,
	treeIdx: number,
	talentIdx: number,
	newPoints: number,
): boolean => {
	const oldPoints = points[treeIdx][talentIdx];
	if (newPoints <= oldPoints) return true;
	if (totalPointsSpent(points) + (newPoints - oldPoints) > MAX_POINTS_PLAYER) return false;

	return !findBrokenTalents(config, withTalentPoints(points, treeIdx, talentIdx, newPoints))[treeIdx][talentIdx];
};
