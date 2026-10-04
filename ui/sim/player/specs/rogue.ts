import { Class, Spec } from '@generated/proto/common';

import { CURRENT_PHASE, getSpecSitePath, LaunchStatus } from '../../constants/other';
import { isSpecDualWieldCapable } from '../classes/capabilities';
import { IconSize } from '../player_class';
import { PlayerSpec, SimStatus } from '../player_spec';

export class Rogue extends PlayerSpec<Spec.SpecRogue> {
	static specIndex = 0;
	static specID = Spec.SpecRogue as Spec.SpecRogue;
	static classID = Class.ClassRogue as Class.ClassRogue;
	static friendlyName = 'Rogue';
	static simLink = getSpecSitePath('rogue', 'dps');

	static isTankSpec = false;
	static isHealingSpec = false;
	static isRangedDpsSpec = false;
	static isMeleeDpsSpec = true;
	static canDualWield = isSpecDualWieldCapable(Rogue.specID);

	static launch: SimStatus = {
		phase: CURRENT_PHASE,
		status: LaunchStatus.Alpha,
	};

	readonly specIndex = Rogue.specIndex;
	readonly specID = Rogue.specID;
	readonly classID = Rogue.classID;
	readonly friendlyName = Rogue.friendlyName;
	readonly simLink = Rogue.simLink;

	readonly isTankSpec = Rogue.isTankSpec;
	readonly isHealingSpec = Rogue.isHealingSpec;
	readonly isRangedDpsSpec = Rogue.isRangedDpsSpec;
	readonly isMeleeDpsSpec = Rogue.isMeleeDpsSpec;

	readonly canDualWield = Rogue.canDualWield;

	readonly launch = Rogue.launch;

	static getIcon = (size: IconSize): string => {
		return `https://wow.zamimg.com/images/wow/icons/${size}/class_rogue.jpg`;
	};

	getIcon = (size: IconSize): string => {
		return Rogue.getIcon(size);
	};
}
