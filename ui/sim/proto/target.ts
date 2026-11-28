import { MobType, type Target } from '@generated/proto/common';

import { CREATURE_FAMILY_MOB_TYPES } from './sim_tables_auto_gen';

/** The target's own mob type, or its family's where it has none, as the sim counts it in every versus check. */
export const targetMobType = (target: Target): MobType =>
	target.mobType === MobType.MobTypeUnknown ? (CREATURE_FAMILY_MOB_TYPES[target.creatureFamily] ?? MobType.MobTypeUnknown) : target.mobType;
