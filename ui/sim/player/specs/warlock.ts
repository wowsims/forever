import { Class, Spec } from '@generated/proto/common';

import { CURRENT_PHASE, getSpecSitePath, LaunchStatus } from '../../constants/other';
import { isSpecDualWieldCapable } from '../classes/capabilities';
import { IconSize } from '../player_class';
import { PlayerSpec, SimStatus } from '../player_spec';

export class Warlock extends PlayerSpec<Spec.SpecWarlock> {
	static specIndex = 0;
	static specID = Spec.SpecWarlock as Spec.SpecWarlock;
	static classID = Class.ClassWarlock as Class.ClassWarlock;
	static friendlyName = 'Warlock';
	static simLink = getSpecSitePath('warlock', 'dps');

	static isTankSpec = false;
	static isHealingSpec = false;
	static isRangedDpsSpec = true;
	static isMeleeDpsSpec = false;

	static canDualWield = isSpecDualWieldCapable(Warlock.specID);

	static launch: SimStatus = {
		phase: CURRENT_PHASE,
		status: LaunchStatus.Alpha,
	};

	readonly specIndex = Warlock.specIndex;
	readonly specID = Warlock.specID;
	readonly classID = Warlock.classID;
	readonly friendlyName = Warlock.friendlyName;
	readonly simLink = Warlock.simLink;

	readonly isTankSpec = Warlock.isTankSpec;
	readonly isHealingSpec = Warlock.isHealingSpec;
	readonly isRangedDpsSpec = Warlock.isRangedDpsSpec;
	readonly isMeleeDpsSpec = Warlock.isMeleeDpsSpec;

	readonly canDualWield = Warlock.canDualWield;

	readonly launch = Warlock.launch;

	static getIcon = (size: IconSize): string => {
		return `https://wow.zamimg.com/images/wow/icons/${size}/class_warlock.jpg`;
	};

	getIcon = (size: IconSize): string => {
		return Warlock.getIcon(size);
	};
}
